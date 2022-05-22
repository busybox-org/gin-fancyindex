package internal

import (
	"encoding/json"
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

type accessLog struct {
	TimeStamp  string `json:"timestamp"`
	ClientIP   string `json:"client_ip"`
	Method     string `json:"method"`
	Path       string `json:"path"`
	Protocol   string `json:"protocol"`
	StatusCode int    `json:"status"`
	Latency    int64  `json:"duration"`
	BodySize   int    `json:"body_size"`
}

type Engine struct {
	*gin.Engine
}

func Router(conf *Config) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	e := &Engine{
		gin.New(),
	}
	e.MaxMultipartMemory = 32 << 20 // 32 MiB
	e.Use(
		gin.Recovery(),
		gzip.Gzip(gzip.DefaultCompression),
		cors.Default(),
		func(c *gin.Context) {
			c.Header("Server", "Gin")
			c.Header("X-Server", "Gin")
			c.Header("X-Powered-By", "XMapst")
			c.Header("Cache-Control", "no-cache, no-store, max-age=0, must-revalidate, value")
			c.Header("Expires", "Thu, 01 Jan 1970 00:00:00 GMT")
			c.Header("Last-Modified", time.Now().UTC().Format(http.TimeFormat))
			c.Header("Pragma", "no-cache")
			c.Next()
		},
		gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
			log := &accessLog{
				TimeStamp:  param.TimeStamp.Format(time.RFC3339),
				ClientIP:   param.ClientIP,
				Method:     param.Method,
				Path:       param.Path,
				Protocol:   param.Request.Proto,
				StatusCode: param.StatusCode,
				Latency:    int64(param.Latency),
				BodySize:   param.BodySize,
			}
			bs, err := json.Marshal(log)
			if err != nil {
				logrus.Error(err)
				return ""
			}
			// your custom format
			return string(bs) + "\n"
		}),
	)
	// no auth
	e.StaticFS(conf.RelativePath, conf.Root, conf.Auth)
	if conf.Auth {
		g := e.Group("/", gin.BasicAuth(gin.Accounts{
			conf.User: conf.Pass,
		}))
		{
			// upload filepath
			g.POST("/*filepath", func(c *gin.Context) {
				c.Request.URL.Path = strings.TrimPrefix(c.Request.URL.Path, conf.RelativePath)
				c.Request.URL.Path = path.Join("/", c.Request.URL.Path)
				filePath := filepath.Join(conf.Root, path.Clean(c.Request.URL.Path))
				if err := saveFiles(c, filePath); err != nil {
					c.String(http.StatusInternalServerError, err.Error())
					return
				}
				e.HandleContext(c)
			})
			// delete filepath
			g.DELETE("/*filepath", func(c *gin.Context) {
				c.Request.URL.Path = strings.TrimPrefix(c.Request.URL.Path, conf.RelativePath)
				c.Request.URL.Path = path.Join("/", c.Request.URL.Path)
				filePath := filepath.Join(conf.Root, path.Clean(c.Request.URL.Path))
				if !FileOrPathExist(filePath) {
					c.String(http.StatusNotFound, fmt.Sprintf("%s not found", filePath))
					return
				}
				err := os.RemoveAll(filePath)
				if err != nil {
					c.String(http.StatusInternalServerError, err.Error())
					return
				}
				c.Request.URL.Path = "/"
				e.HandleContext(c)
			})
		}
	}
	return e.Engine
}
