package api

import (
	"context"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/wetor/AnimeGo/configs"
	"github.com/wetor/AnimeGo/internal/api"
	"github.com/wetor/AnimeGo/internal/utils"
	webModels "github.com/wetor/AnimeGo/internal/web/models"
	"github.com/wetor/AnimeGo/pkg/errors"
	"github.com/wetor/AnimeGo/pkg/log"
)

var (
	Ctx                           context.Context
	AccessKey                     string
	Cache                         api.Cacher
	Config                        *configs.Config
	BangumiCache                  api.CacheGetter
	BangumiCacheLock              *sync.Mutex
	FilterManager                 api.FilterManager
	DownloaderManagerCacheDeleter api.DownloaderManagerCacheDeleter
)

type Options struct {
	Ctx                           context.Context
	AccessKey                     string
	Cache                         api.Cacher
	Config                        *configs.Config
	BangumiCache                  api.CacheGetter
	BangumiCacheLock              *sync.Mutex
	FilterManager                 api.FilterManager
	DownloaderManagerCacheDeleter api.DownloaderManagerCacheDeleter
}

func Init(opts *Options) {
	Ctx = opts.Ctx
	AccessKey = opts.AccessKey
	Cache = opts.Cache
	Config = opts.Config
	BangumiCache = opts.BangumiCache
	BangumiCacheLock = opts.BangumiCacheLock
	FilterManager = opts.FilterManager
	DownloaderManagerCacheDeleter = opts.DownloaderManagerCacheDeleter
}

// checkRequest 绑定request结构体
//
//	Description 包含登录信息
//	Param c *gin.Context
//	Param data any 返回结构体指针
//	Return bool
func checkRequest(c *gin.Context, data any) bool {
	if err := c.ShouldBind(data); err != nil {
		log.Warnf("参数错误，err: %s", errors.NewAniErrorD(err))
		c.JSON(webModels.Fail("参数错误"))
		return false
	}
	if err := c.ShouldBindQuery(data); err != nil {
		log.Warnf("Query参数错误，err: %s", errors.NewAniErrorD(err))
		c.JSON(webModels.Fail("Query参数错误"))
		return false
	}
	if err := c.ShouldBindUri(data); err != nil {
		log.Warnf("Uri参数错误，err: %s", errors.NewAniErrorD(err))
		c.JSON(webModels.Fail("Uri参数错误"))
		return false
	}

	key, has := c.Get("access_key")
	localKey := utils.Sha256(AccessKey)
	if has && key != localKey {
		log.Warnf("", errors.NewAniError("Access key错误！"))
		c.JSON(webModels.Fail("Access key错误"))
		return false
	}
	return true
}

// Ping godoc
//
//	@Summary Ping
//	@Description Pong
//	@Tags web
//	@Accept  json
//	@Produce  json
//	@Success 200 {object} webModels.Response
//	@Router /ping [get]
func Ping(c *gin.Context) {
	c.JSON(webModels.Succ("pong", gin.H{
		"version": os.Getenv("ANIMEGO_VERSION"),
		"time":    time.Now().Unix(),
	}))
}

// SHA256 godoc
//
//	@Summary SHA256计算
//	@Description SHA256计算
//	@Tags web
//	@Accept  json
//	@Produce  json
//	@Param access_key query string true "原文本"
//	@Success 200 {object} webModels.Response{data=string}
//	@Router /sha256 [get]
func SHA256(c *gin.Context) {
	c.JSON(webModels.Succ("Access-Key", utils.Sha256(c.Query("access_key"))))
}
