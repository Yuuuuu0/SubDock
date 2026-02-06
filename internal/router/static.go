package router

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/gin-gonic/gin"
)

//go:embed all:dist
var distFS embed.FS

func serveStatic(r *gin.Engine) {
	distSubFS, _ := fs.Sub(distFS, "dist")
	staticHandler := http.FileServer(http.FS(distSubFS))

	r.GET("/", func(c *gin.Context) {
		c.FileFromFS("/", http.FS(distSubFS))
	})

	r.GET("/assets/*filepath", func(c *gin.Context) {
		c.FileFromFS(c.Request.URL.Path, http.FS(distSubFS))
	})

	// 显式处理 favicon，避免文件缺失时被 NoRoute 回退到 index.html
	r.GET("/favicon.ico", func(c *gin.Context) {
		if f, err := distSubFS.Open("favicon.ico"); err == nil {
			f.Close()
			c.FileFromFS("/favicon.ico", http.FS(distSubFS))
			return
		}
		c.Status(http.StatusNotFound)
	})

	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		if f, err := distSubFS.Open(path[1:]); err == nil {
			f.Close()
			staticHandler.ServeHTTP(c.Writer, c.Request)
			return
		}
		c.FileFromFS("/", http.FS(distSubFS))
	})
}
