package plugin

import (
	"fmt"
	"testing"

	"go.uber.org/zap"

	"github.com/wetor/AnimeGo/internal/animego/feed"
	mikanRss "github.com/wetor/AnimeGo/internal/animego/feed/rss"
	"github.com/wetor/AnimeGo/internal/models"
	"github.com/wetor/AnimeGo/internal/plugin/python/lib"
	"github.com/wetor/AnimeGo/internal/utils"
	"github.com/wetor/AnimeGo/third_party/gpython"
)

func TestMain(m *testing.M) {
	fmt.Println("begin")
	logger, _ := zap.NewDevelopment()
	zap.ReplaceGlobals(logger)
	m.Run()
	fmt.Println("end")
}

func TestJavaScript_Filter(t *testing.T) {
	_ = utils.CreateMutiDir("data")
	feed.Init(&feed.Options{
		TempPath: "data",
	})
	rss := mikanRss.NewRss("", "")
	items := rss.Parse("testdata/Mikan.xml")
	fmt.Println(len(items))
	js := NewFilterPlugin([]models.Plugin{
		{
			Enable: true,
			Type:   "js",
			File:   "testdata/test.js",
		},
	})
	result := js.Filter(items)
	fmt.Println(len(result))
	for _, r := range result {
		fmt.Println(r.Name)
	}

}

func TestJavaScript_Filter2(t *testing.T) {
	list := []*models.FeedItem{
		{
			Name: "0000",
		},
		{
			Name: "1108011",
		},
		{
			Name: "2222",
		},
		{
			Name: "3333",
		},
	}
	js := NewFilterPlugin([]models.Plugin{
		{
			Enable: true,
			Type:   "js",
			File:   "testdata/regexp.js",
		},
	})
	result := js.Filter(list)
	fmt.Println(len(result))
	for _, r := range result {
		fmt.Println(r.Name)
	}
}

func TestPython_Filter(t *testing.T) {
	gpython.Init()
	lib.InitLog()
	_ = utils.CreateMutiDir("data")
	feed.Init(&feed.Options{
		TempPath: "data",
	})
	rss := mikanRss.NewRss("", "")
	items := rss.Parse("testdata/Mikan.xml")
	fmt.Println(len(items))
	js := NewFilterPlugin([]models.Plugin{
		{
			Enable: true,
			Type:   "py",
			File:   "testdata/filter.py",
		},
	})
	result := js.Filter(items)
	fmt.Println(len(result))
	for _, r := range result {
		fmt.Println(r.Name)
	}

}

func TestPython_Filter2(t *testing.T) {
	gpython.Init()
	lib.InitLog()
	list := []*models.FeedItem{
		{
			Name: "0000",
		},
		{
			Name: "1108011",
		},
		{
			Name: "2222",
		},
		{
			Name: "3333",
		},
	}
	js := NewFilterPlugin([]models.Plugin{
		{
			Enable: true,
			Type:   "py",
			File:   "testdata/test_re.py",
		},
	})
	result := js.Filter(list)
	fmt.Println(len(result))
	for _, r := range result {
		fmt.Println(r.Name)
	}
}
