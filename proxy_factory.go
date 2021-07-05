package krakend

import (
  jsonschema "github.com/devopsfaith/krakend-jsonschema"
  lua "github.com/devopsfaith/krakend-lua/proxy"
  metrics "github.com/devopsfaith/krakend-metrics/gin"
  opencensus "github.com/devopsfaith/krakend-opencensus"
  "github.com/luraproject/lura/logging"
  "github.com/luraproject/lura/proxy"
  acl "github.com/avamonitoring/avanet-gateway-access-logging/logging/proxy"
)

// NewProxyFactory returns a new ProxyFactory wrapping the injected BackendFactory with the default proxy stack and a metrics collector
func NewProxyFactory(logger logging.Logger, backendFactory proxy.BackendFactory, metricCollector *metrics.Metrics) proxy.Factory {
  proxyFactory := proxy.NewDefaultFactory(backendFactory, logger)
  proxyFactory = proxy.NewShadowFactory(proxyFactory)
  proxyFactory = jsonschema.ProxyFactory(proxyFactory)
  proxyFactory = lua.ProxyFactory(logger, proxyFactory)
  proxyFactory = opencensus.ProxyFactory(proxyFactory)
  proxyFactory = acl.ProxyFactory(logger, proxyFactory)
  return proxyFactory
}

type proxyFactory struct{}

func (p proxyFactory) NewProxyFactory(logger logging.Logger, backendFactory proxy.BackendFactory, metricCollector *metrics.Metrics) proxy.Factory {
  return NewProxyFactory(logger, backendFactory, metricCollector)
}
