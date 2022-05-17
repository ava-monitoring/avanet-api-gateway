package krakend

import (
	"time"
	"github.com/gin-gonic/gin"

	botdetector "github.com/devopsfaith/krakend-botdetector/v2/gin"
	httpsecure "github.com/devopsfaith/krakend-httpsecure/v2/gin"
	lua "github.com/devopsfaith/krakend-lua/v2/router/gin"
	opencensus "github.com/devopsfaith/krakend-opencensus/v2/router/gin"
	"github.com/go-logfmt/logfmt"
	"github.com/luraproject/lura/v2/config"
	"github.com/luraproject/lura/v2/core"
	luragin "github.com/luraproject/lura/v2/router/gin"
	"github.com/luraproject/lura/v2/transport/http/server"
)

// See https://github.com/gin-gonic/gin/blob/d062a6a6155236883f4c3292379ab94b1eac8b05/logger.go#L143 for original log format
var customLogFormatter = func(param gin.LogFormatterParams) string {
	if param.Latency > time.Minute {
		param.Latency = param.Latency.Truncate(time.Second)
	}
	clientIp := param.ClientIP
	// Extract origin IP set by Nginx Ingress using the PROXY protocol (or by AWS LB at Layer 7)
	// https://docs.nginx.com/nginx/admin-guide/load-balancer/using-proxy-protocol/
	// https://docs.aws.amazon.com/elasticloadbalancing/latest/classic/enable-proxy-protocol.html
	realIp := param.Request.Header.Get("X-Real-IP")
	forwardedFor := param.Request.Header.Get("X-Forwarded-For")
	if realIp != "" {
		clientIp = realIp
	} else if forwardedFor != "" {
		clientIp = forwardedFor
	}
	output, err := logfmt.MarshalKeyvals("time", param.TimeStamp.Format("2006/01/02 - 15:04:05"),
		"status", param.StatusCode,
		"latency", param.Latency,
		"clientIP", clientIp,
		"method", param.Method,
		"path", param.Path,
		"errorMessage", param.ErrorMessage)
	if err != nil {
		return "Error formatting log: " + err.Error()
	}
	return string(output) + "\n"
}

// NewEngine creates a new gin engine with some default values and a secure middleware
func NewEngine(cfg config.ServiceConfig, opt luragin.EngineOptions) *gin.Engine {
	if !cfg.Debug {
		gin.SetMode(gin.ReleaseMode)
	}
	
	engine := gin.New()
	engine.Use(gin.LoggerWithConfig(gin.LoggerConfig{
				Formatter: customLogFormatter, // Use logfmt format
				Output: opt.Writer,
				SkipPaths: []string{"/__health"}, // Do not log health checks
		}), gin.Recovery())

	engine.NoRoute(opencensus.HandlerFunc(&config.EndpointConfig{Endpoint: "NoRoute"}, defaultHandler, nil))
	engine.NoMethod(opencensus.HandlerFunc(&config.EndpointConfig{Endpoint: "NoMethod"}, defaultHandler, nil))

	logPrefix := "[SERVICE: Gin]"
	if err := httpsecure.Register(cfg.ExtraConfig, engine); err != nil && err != httpsecure.ErrNoConfig {
		opt.Logger.Warning(logPrefix+"[HTTPsecure]", err)
	} else if err == nil {
		opt.Logger.Debug(logPrefix + "[HTTPsecure] Successfuly loaded module")
	}

	lua.Register(opt.Logger, cfg.ExtraConfig, engine)

	botdetector.Register(cfg, opt.Logger, engine)

	return engine
}

func defaultHandler(c *gin.Context) {
	c.Header(core.KrakendHeaderName, core.KrakendHeaderValue)
	c.Header(server.CompleteResponseHeaderName, server.HeaderIncompleteResponseValue)
}

type engineFactory struct{}

func (engineFactory) NewEngine(cfg config.ServiceConfig, opt luragin.EngineOptions) *gin.Engine {
	return NewEngine(cfg, opt)
}
