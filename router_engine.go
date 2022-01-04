package krakend

import (
	"io"
	"time"

	botdetector "github.com/devopsfaith/krakend-botdetector/gin"
	httpsecure "github.com/devopsfaith/krakend-httpsecure/gin"
	lua "github.com/devopsfaith/krakend-lua/router/gin"
	"github.com/gin-gonic/gin"
	"github.com/go-logfmt/logfmt"
	"github.com/luraproject/lura/config"
	"github.com/luraproject/lura/logging"
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
func NewEngine(cfg config.ServiceConfig, logger logging.Logger, w io.Writer) *gin.Engine {
	if !cfg.Debug {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()
	engine.Use(gin.LoggerWithConfig(gin.LoggerConfig{
				Formatter: customLogFormatter, // Use logfmt format
				Output: w,
				SkipPaths: []string{"/__health"}, // Do not log health checks
		}), gin.Recovery())

	engine.RedirectTrailingSlash = true
	engine.RedirectFixedPath = true
	engine.HandleMethodNotAllowed = true

	if err := httpsecure.Register(cfg.ExtraConfig, engine); err != nil {
		logger.Warning(err)
	}

	lua.Register(logger, cfg.ExtraConfig, engine)

	botdetector.Register(cfg, logger, engine)

	return engine
}

type engineFactory struct{}

func (e engineFactory) NewEngine(cfg config.ServiceConfig, l logging.Logger, w io.Writer) *gin.Engine {
	return NewEngine(cfg, l, w)
}
