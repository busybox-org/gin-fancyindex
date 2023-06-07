package internal

import (
	"path"
	"strings"

	"github.com/gin-gonic/gin"
	fancyindex "github.com/xmapst/gin-fancyindex"
)

func (e *Engine) StaticFS(relativePath, root string) {
	if strings.Contains(relativePath, ":") || strings.Contains(relativePath, "*") {
		panic("URL parameters can not be used when serving a static folder")
	}
	handler := e.createStaticHandler(relativePath, root)
	urlPattern := path.Join(relativePath, "/*filepath")

	// Register GET and HEAD handlers
	e.GET(urlPattern, handler)
	e.HEAD(urlPattern, handler)
}

func (e *Engine) createStaticHandler(relativePath, root string) gin.HandlerFunc {
	fileServer := fancyindex.New(relativePath, root)
	return func(c *gin.Context) {
		fileServer.ServeHTTP(c.Writer, c.Request)
	}
}
