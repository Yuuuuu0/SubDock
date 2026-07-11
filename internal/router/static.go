package router

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

//go:embed all:dist
var distFS embed.FS

// serveStatic 注册嵌入式前端资源，并仅对非 API 的 GET/HEAD 路径执行 SPA 回退。
func serveStatic(r *gin.Engine) {
	distSubFS, err := fs.Sub(distFS, "dist")
	if err != nil {
		panic("加载嵌入式前端资源失败: " + err.Error())
	}
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
		if path == "/api" || strings.HasPrefix(path, "/api/") {
			c.JSON(http.StatusNotFound, gin.H{"error": "接口不存在"})
			return
		}
		if c.Request.Method != http.MethodGet && c.Request.Method != http.MethodHead {
			c.Status(http.StatusNotFound)
			return
		}
		if f, err := distSubFS.Open(path[1:]); err == nil {
			f.Close()
			staticHandler.ServeHTTP(c.Writer, c.Request)
			return
		}
		c.FileFromFS("/", http.FS(distSubFS))
	})
}
