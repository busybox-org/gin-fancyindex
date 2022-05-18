package fancyindex

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

func StaticFS(g *gin.Engine, root string) {
	g.Use(func(c *gin.Context) {
		c.Header("Server", "Gin")
		c.Header("X-Server", "Gin")
		c.Header("X-Powered-By", "XMapst")
		c.Header("Cache-Control", "no-cache, no-store, max-age=0, must-revalidate, value")
		c.Header("Expires", "Thu, 01 Jan 1970 00:00:00 GMT")
		c.Header("Last-Modified", time.Now().UTC().Format(http.TimeFormat))
		c.Header("Pragma", "no-cache")
		c.Next()
	})
	g.GET("/*filepath", createStaticHandler(root))
	g.HEAD("/*filepath", createStaticHandler(root))
}

func createStaticHandler(root string) gin.HandlerFunc {
	fileServer := Browser(root)
	return func(c *gin.Context) {
		fileServer.ServeHTTP(c.Writer, c.Request)
	}
}
