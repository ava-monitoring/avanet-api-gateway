package krakend

import (
	"context"

	cb "github.com/krakendio/krakend-circuitbreaker/v2/gobreaker/proxy"
	lua "github.com/krakendio/krakend-lua/v2/proxy"
	martian "github.com/krakendio/krakend-martian/v2"
	metrics "github.com/krakendio/krakend-metrics/v2/gin"
	oauth2client "github.com/krakendio/krakend-oauth2-clientcredentials/v2"
	opencensus "github.com/krakendio/krakend-opencensus/v2"
	pubsub "github.com/krakendio/krakend-pubsub/v2"
	juju "github.com/krakendio/krakend-ratelimit/v2/juju/proxy"
	"github.com/luraproject/lura/v2/config"
	"github.com/luraproject/lura/v2/logging"
	"github.com/luraproject/lura/v2/proxy"
	"github.com/luraproject/lura/v2/transport/http/client"
	httprequestexecutor "github.com/luraproject/lura/v2/transport/http/client/plugin"
	
	acp "github.com/avamonitoring/avanet-gateway-access-control/acp/proxy"
  	acl "github.com/avamonitoring/avanet-gateway-access-logging/logging/proxy"
)

// NewBackendFactory creates a BackendFactory by stacking middlewares
func NewBackendFactory(logger logging.Logger, metricCollector *metrics.Metrics) proxy.BackendFactory {
  	return NewBackendFactoryWithContext(context.Background(), logger, metricCollector)
}

// NewBackendFactory creates a BackendFactory by stacking all the available middlewares and injecting the received context
func NewBackendFactoryWithContext(ctx context.Context, logger logging.Logger, metricCollector *metrics.Metrics) proxy.BackendFactory {
  	requestExecutorFactory := func(cfg *config.Backend) client.HTTPRequestExecutor {
    		var clientFactory client.HTTPClientFactory
    		if _, ok := cfg.ExtraConfig[oauth2client.Namespace]; ok {
      			clientFactory = oauth2client.NewHTTPClient(cfg)
    		} else {
      			clientFactory = client.NewHTTPClient
    		}
    		return acl.HTTPRequestExecutor(logger, true, clientFactory)
  	}

  	requestExecutorFactory = httprequestexecutor.HTTPRequestExecutor(logger, requestExecutorFactory)
  	backendFactory := martian.NewConfiguredBackendFactory(logger, requestExecutorFactory)
  	bf := pubsub.NewBackendFactory(ctx, logger, backendFactory)
  	backendFactory = bf.New
  	backendFactory = lua.BackendFactory(logger, backendFactory)
  	backendFactory = juju.BackendFactory(logger, backendFactory)
  	backendFactory = cb.BackendFactory(backendFactory, logger)
  	backendFactory = metricCollector.BackendFactory("backend", backendFactory)
  	backendFactory = opencensus.BackendFactory(backendFactory)
  	backendFactory = acp.BackendFactory(logger, backendFactory)
  	backendFactory = acl.BackendFactory(logger, backendFactory)
  	return backendFactory
}

type backendFactory struct{}

func (b backendFactory) NewBackendFactory(ctx context.Context, l logging.Logger, m *metrics.Metrics) proxy.BackendFactory {
  	return NewBackendFactoryWithContext(ctx, l, m)
}
