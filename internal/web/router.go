package web

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/wetor/AnimeGo/internal/web/api"
	"github.com/wetor/AnimeGo/internal/web/models"
	"github.com/wetor/AnimeGo/pkg/errors"
	"github.com/wetor/AnimeGo/pkg/log"
)

func Run(ctx context.Context) {
	WG.Add(1)
	go func() {
		defer WG.Done()
		r := gin.New()
		r.Use(Cors())                     // 跨域中间件
		r.Use(GinLogger(log.GetLogger())) // 日志中间件
		r.Use(GinRecovery(log.GetLogger(), true, func(c *gin.Context, recovered any) {
			if err, ok := recovered.(error); ok {
				log.Debugf("服务器错误，err: %v", errors.NewAniErrorD(err))
				c.JSON(models.ErrSvr("服务器错误"))
			} else {
				log.Debugf(recovered.(string))
				c.JSON(models.ErrSvr(recovered.(string)))
			}
		})) // 错误处理中间件
		r.GET("/ping", api.Ping)
		r.GET("/sha256", api.SHA256)
		InitSwagger(r)
		apiRoot := r.Group("/api")
		apiRoot.Use(KeyAuth())
		apiRoot.POST("/rss", api.Rss)
		apiRoot.POST("/plugin/config", api.PluginConfigPost)
		apiRoot.GET("/plugin/config", api.PluginConfigGet)

		apiRoot.GET("/config", api.ConfigGet)
		apiRoot.PUT("/config", api.ConfigPut)

		apiRoot.GET("/bolt", api.BoltList)
		apiRoot.GET("/bolt/value", api.Bolt)
		apiRoot.DELETE("/bolt/value", api.BoltDelete)

		s := &http.Server{
			Addr:    fmt.Sprintf("%s:%d", Host, Port),
			Handler: r,
		}
		go func() {
			if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Debugf("", err)
				log.Warnf("启动web服务失败")
			}
		}()
		log.Infof("AnimeGo Web服务已启动: http://%s", s.Addr)
		select {
		case <-ctx.Done():
			if err := s.Close(); err != nil {
				log.Debugf("", err)
				log.Warnf("关闭web服务失败")
			}
			log.Debugf("正常退出 web")
		}
	}()
}
