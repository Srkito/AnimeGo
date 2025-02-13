package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/wetor/AnimeGo/assets"
	"github.com/wetor/AnimeGo/configs"
	_ "github.com/wetor/AnimeGo/docs"
	"github.com/wetor/AnimeGo/internal/animego/anidata"
	anidataBangumi "github.com/wetor/AnimeGo/internal/animego/anidata/bangumi"
	anidataMikan "github.com/wetor/AnimeGo/internal/animego/anidata/mikan"
	anidataThemoviedb "github.com/wetor/AnimeGo/internal/animego/anidata/themoviedb"
	"github.com/wetor/AnimeGo/internal/animego/anisource"
	"github.com/wetor/AnimeGo/internal/animego/anisource/mikan"
	"github.com/wetor/AnimeGo/internal/animego/downloader"
	"github.com/wetor/AnimeGo/internal/animego/downloader/qbittorrent"
	feedRss "github.com/wetor/AnimeGo/internal/animego/feed/rss"
	"github.com/wetor/AnimeGo/internal/animego/filter/plugin"
	"github.com/wetor/AnimeGo/internal/animego/manager"
	downloaderMgr "github.com/wetor/AnimeGo/internal/animego/manager/downloader"
	filterMgr "github.com/wetor/AnimeGo/internal/animego/manager/filter"
	"github.com/wetor/AnimeGo/internal/constant"
	"github.com/wetor/AnimeGo/internal/logger"
	"github.com/wetor/AnimeGo/internal/models"
	"github.com/wetor/AnimeGo/internal/plugin/python/lib"
	"github.com/wetor/AnimeGo/internal/schedule"
	"github.com/wetor/AnimeGo/internal/schedule/task"
	scheduleUtils "github.com/wetor/AnimeGo/internal/schedule/utils"
	"github.com/wetor/AnimeGo/internal/utils"
	"github.com/wetor/AnimeGo/internal/web"
	"github.com/wetor/AnimeGo/internal/web/api"
	"github.com/wetor/AnimeGo/pkg/cache"
	log2 "github.com/wetor/AnimeGo/pkg/log"
	"github.com/wetor/AnimeGo/pkg/request"
	"github.com/wetor/AnimeGo/pkg/xpath"
	"github.com/wetor/AnimeGo/third_party/gpython"
)

const (
	DefaultConfigFile = "data/animego.yaml"
)

var (
	version   = "dev"
	buildTime = "dev"
)

var (
	ctx, cancel = context.WithCancel(context.Background())
	configFile  string
	debug       bool
	webapi      bool

	WG                sync.WaitGroup
	BangumiCacheMutex sync.Mutex
)

func init() {
	var err error
	err = os.Setenv("ANIMEGO_VERSION", fmt.Sprintf("%s-%s", version, buildTime))
	if err != nil {
		panic(err)
	}
}

func main() {
	printInfo()

	flag.StringVar(&configFile, "config", DefaultConfigFile, "配置文件路径；配置文件中的相对路径均是相对与程序的位置")
	flag.BoolVar(&debug, "debug", false, "Debug模式，将会显示更多的日志")
	flag.BoolVar(&webapi, "web", true, "启用Web API，默认启用")
	flag.Parse()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT)
	go func() {
		for s := range sigs {
			switch s {
			case syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT:
				log2.Infof("收到退出信号: %v", s)
				doExit()
			default:
				log2.Infof("收到其他信号: %v", s)
			}
		}
	}()
	Main(ctx)
}

func InitDefaultConfig() {
	if utils.IsExist(configFile) {
		// 尝试升级配置文件
		if configs.UpdateConfig(configFile, true) {
			os.Exit(0)
		}
		return
	}
	log.Printf("未找到配置文件（%s），开始初始化默认配置\n", configFile)
	conf := configs.DefaultConfig()
	if utils.IsExist(conf.Setting.DataPath) {
		log.Printf("默认data_path文件夹（%s）已存在，无法完成初始化\n", conf.Setting.DataPath)
		os.Exit(0)
	}
	err := utils.CreateMutiDir(conf.Setting.DataPath)
	if err != nil {
		panic(err)
	}
	err = configs.DefaultFile(DefaultConfigFile)
	if err != nil {
		panic(err)
	}

	InitDefaultAssets(conf)

	log.Printf("初始化默认配置完成（%s）\n", conf.Setting.DataPath)
	log.Println("请设置配置后重新启动")
	os.Exit(0)
}

func InitDefaultAssets(conf *configs.Config) {
	utils.CopyDir(assets.Plugin, "plugin", xpath.Join(conf.Setting.DataPath, "plugin"), true, false)
}

func doExit() {
	log2.Infof("正在退出...")
	cancel()
	go func() {
		time.Sleep(5 * time.Second)
		os.Exit(0)
	}()
}

func printInfo() {
	fmt.Println(`--------------------------------------------------
    ___            _                   ______     
   /   |   ____   (_)____ ___   ___   / ____/____ 
  / /| |  / __ \ / // __ \__ \ / _ \ / / __ / __ \
 / ___ | / / / // // / / / / //  __// /_/ // /_/ /
/_/  |_|/_/ /_//_//_/ /_/ /_/ \___/ \____/ \____/
    `)
	fmt.Printf("AnimeGo %s\n", os.Getenv("ANIMEGO_VERSION"))
	fmt.Printf("AnimeGo config v%s\n", configs.ConfigVersion)
	fmt.Printf("%s\n", constant.AnimeGoGithub)
	fmt.Println("--------------------------------------------------")
}

func Main(ctx context.Context) {
	configFile = xpath.Abs(configFile)
	// 初始化默认配置、升级配置
	InitDefaultConfig()

	// 载入配置文件
	config := configs.Init(configFile)
	constant.Init(&constant.Options{
		DataPath: config.DataPath,
	})
	config.InitDir()

	// 释放资源
	InitDefaultAssets(config)

	// 初始化日志
	logger.Init(&logger.Options{
		File:    constant.LogFile,
		Debug:   debug,
		Context: ctx,
		WG:      &WG,
	})

	// 初始化request
	request.Init(&request.Options{
		UserAgent: fmt.Sprintf("%s/AnimeGo (%s)", os.Getenv("ANIMEGO_VERSION"), constant.AnimeGoGithub),
		Proxy:     config.Proxy(),
		Timeout:   config.Advanced.Request.TimeoutSecond,
		Retry:     config.Advanced.Request.RetryNum,
		RetryWait: config.Advanced.Request.RetryWaitSecond,
		Debug:     debug,
	})

	// 初始化插件-gpython
	gpython.Init()
	lib.Init()

	// 载入AnimeGo数据库（缓存）
	bolt := cache.NewBolt()
	bolt.Open(constant.CacheFile)

	// 载入Bangumi Archive数据库
	bangumiCache := cache.NewBolt()
	bangumiCache.Open(constant.BangumiCacheFile)

	// 初始化并启动定时任务
	scheduleVar := schedule.NewSchedule(&schedule.Options{
		WG: &WG,
	})
	scheduleVar.Add(&schedule.AddTaskOptions{
		Name:     "bangumi",
		StartRun: true,
		Task: task.NewBangumiTask(&task.BangumiOptions{
			Cache:      bangumiCache,
			CacheMutex: &BangumiCacheMutex,
		}),
	})
	scheduleUtils.AddTasks(scheduleVar, config.Plugin.Schedule)
	scheduleVar.Start(ctx)

	// 初始化并连接下载器
	downloader.Init(&downloader.Options{
		ConnectTimeoutSecond: config.Advanced.Client.ConnectTimeoutSecond,
		CheckTimeSecond:      config.Advanced.Client.CheckTimeSecond,
		RetryConnectNum:      config.Advanced.Client.RetryConnectNum,
		WG:                   &WG,
	})
	qbtConf := config.Setting.Client.QBittorrent
	qbt := qbittorrent.NewQBittorrent(qbtConf.Url, qbtConf.Username, qbtConf.Password)
	qbt.Start(ctx)

	// 初始化anisource配置
	anisource.Init(&anisource.Options{
		Options: &anidata.Options{
			Cache: bolt,
			CacheTime: map[string]int64{
				anidataMikan.Bucket:      int64(config.Advanced.Cache.MikanCacheHour * 60 * 60),
				anidataBangumi.Bucket:    int64(config.Advanced.Cache.BangumiCacheHour * 60 * 60),
				anidataThemoviedb.Bucket: int64(config.Advanced.Cache.ThemoviedbCacheHour * 60 * 60),
			},
			BangumiCache:     bangumiCache,
			BangumiCacheLock: &BangumiCacheMutex,
		},
		TMDBFailSkip:           config.Default.TMDBFailSkip,
		TMDBFailUseTitleSeason: config.Default.TMDBFailUseTitleSeason,
		TMDBFailUseFirstSeason: config.Default.TMDBFailUseFirstSeason,
	})

	// 初始化manager配置
	manager.Init(&manager.Options{
		Downloader: manager.Downloader{
			UpdateDelaySecond:      config.UpdateDelaySecond,
			DownloadPath:           config.DownloadPath,
			SavePath:               config.SavePath,
			Category:               config.Category,
			Tag:                    config.Tag,
			AllowDuplicateDownload: config.Download.AllowDuplicateDownload,
			SeedingTimeMinute:      config.Download.SeedingTimeMinute,
			IgnoreSizeMaxKb:        config.Download.IgnoreSizeMaxKb,
			Rename:                 config.Advanced.Download.Rename,
		},
		Filter: manager.Filter{
			MultiGoroutineMax:     config.Advanced.Feed.MultiGoroutine.GoroutineMax,
			MultiGoroutineEnabled: config.Advanced.Feed.MultiGoroutine.Enable,
			UpdateDelayMinute:     config.Advanced.Feed.UpdateDelayMinute,
			DelaySecond:           config.Advanced.Feed.DelaySecond,
		},
		WG: &WG,
	})

	// 初始化downloader manager
	downloadChan := make(chan *models.AnimeEntity, 10)
	downloaderManager := downloaderMgr.NewManager(qbt, bolt, downloadChan)

	// 初始化filter manager
	filterManager := filterMgr.NewManager(
		plugin.NewFilterPlugin(config.Plugin.Filter),
		feedRss.NewRss(config.Setting.Feed.Mikan.Url, config.Setting.Feed.Mikan.Name),
		mikan.Mikan{ThemoviedbKey: config.Setting.Key.Themoviedb},
		downloadChan)

	// 启动manager
	downloaderManager.Start(ctx)
	filterManager.Start(ctx)

	if webapi {
		// 初始化并运行Web API
		web.Init(&web.Options{
			Options: &api.Options{
				Ctx:                           ctx,
				AccessKey:                     config.WebApi.AccessKey,
				Cache:                         bolt,
				Config:                        config,
				BangumiCache:                  bangumiCache,
				BangumiCacheLock:              &BangumiCacheMutex,
				FilterManager:                 filterManager,
				DownloaderManagerCacheDeleter: downloaderManager,
			},
			Host:  config.WebApi.Host,
			Port:  config.WebApi.Port,
			WG:    &WG,
			Debug: debug,
		})
		web.Run(ctx)
	}

	// 等待运行结束
	WG.Wait()
}
