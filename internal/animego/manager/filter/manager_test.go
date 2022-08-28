package filter

import (
	"AnimeGo/internal/animego/anisource/mikan"
	mikanRss "AnimeGo/internal/animego/feed/mikan"
	"AnimeGo/internal/logger"
	"AnimeGo/internal/store"
	"context"
	"fmt"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	fmt.Println("begin")
	logger.Init()
	defer logger.Flush()

	store.Init(nil)

	m.Run()
	fmt.Println("end")
}

func TestManager_UpdateFeed(t *testing.T) {
	rss := mikanRss.NewRss(store.Config.RssMikan().Url, store.Config.RssMikan().Name)
	mk := mikan.MikanAdapter{ThemoviedbKey: store.Config.KeyTmdb()}
	m := NewManager(nil, rss, mk, nil)

	exit := make(chan bool)
	ctx, cancel := context.WithCancel(context.Background())
	m.Start(ctx)

	go func() {
		time.Sleep(30 * time.Second)
		cancel()
		exit <- true
	}()

	//time.Sleep(120 * time.Second)

	<-exit
}
