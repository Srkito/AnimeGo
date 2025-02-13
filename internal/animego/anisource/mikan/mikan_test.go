package mikan_test

import (
	"fmt"
	"sync"
	"testing"

	"github.com/wetor/AnimeGo/internal/animego/anidata"
	"github.com/wetor/AnimeGo/internal/animego/anisource"
	"github.com/wetor/AnimeGo/internal/animego/anisource/mikan"
	"github.com/wetor/AnimeGo/internal/constant"
	"github.com/wetor/AnimeGo/internal/models"
	"github.com/wetor/AnimeGo/internal/utils"
	"github.com/wetor/AnimeGo/pkg/cache"
	"github.com/wetor/AnimeGo/pkg/json"
	"github.com/wetor/AnimeGo/pkg/log"
	"github.com/wetor/AnimeGo/pkg/request"
	"github.com/wetor/AnimeGo/third_party/gpython"
)

const ThemoviedbKey = "d3d8430aefee6c19520d0f7da145daf5"

func TestMain(m *testing.M) {
	fmt.Println("begin")
	_ = utils.CreateMutiDir("data")
	constant.PluginPath = "../../../../assets/plugin"
	log.Init(&log.Options{
		File:  "data/log.log",
		Debug: true,
	})
	b := cache.NewBolt()
	b.Open("data/bolt.db")
	anisource.Init(&anisource.Options{
		Options: &anidata.Options{
			Cache: b,
			CacheTime: map[string]int64{
				"mikan":      int64(7 * 24 * 60 * 60),
				"bangumi":    int64(3 * 24 * 60 * 60),
				"themoviedb": int64(14 * 24 * 60 * 60),
			},
		},
		TMDBFailSkip:           false,
		TMDBFailUseTitleSeason: true,
		TMDBFailUseFirstSeason: true,
	})
	bangumiCache := cache.NewBolt()
	bangumiCache.Open("../../../../test/testdata/bolt_sub.bolt")
	bangumiCache.Add("bangumi_sub")
	mutex := sync.Mutex{}
	anidata.Init(&anidata.Options{
		Cache:            b,
		BangumiCache:     bangumiCache,
		BangumiCacheLock: &mutex,
	})

	gpython.Init()

	request.Init(&request.Options{
		Proxy: "http://127.0.0.1:7890",
	})
	m.Run()
	b.Close()
	fmt.Println("end")
}

func TestParseMikan(t *testing.T) {
	type args struct {
		name string
		url  string
	}
	tests := []struct {
		name      string
		args      args
		wantAnime *models.AnimeEntity
	}{
		// TODO: Add test cases.
		{
			name: "海贼王",
			args: args{
				url:  "https://mikanani.me/Home/Episode/18b60d48a72c603b421468aade7fdd0868ff2f2f",
				name: "OPFans枫雪动漫][ONE PIECE 海贼王][第1029话][1080p][周日版][MP4][简体] [299.5MB]",
			},
			wantAnime: &models.AnimeEntity{
				MikanID:      228,
				ThemoviedbID: 37854,
				ID:           975,
				Ep:           1029,
				Season:       1,
			},
		},
		{
			name: "欢迎来到实力至上主义的教室 第二季",
			args: args{
				url:  "https://mikanani.me/Home/Episode/8849c25e05d6e2623b5333bc78d3a489a9b1cc59",
				name: "[ANi] Classroom of the Elite S2 - 欢迎来到实力至上主义的教室 第二季 - 07 [1080P][Baha][WEB-DL][AAC AVC][CHT][MP4] [254.26 MB]",
			},
			wantAnime: &models.AnimeEntity{
				MikanID:      2775,
				ThemoviedbID: 72517,
				ID:           371546,
				Ep:           7,
				Season:       2,
			},
		},
	}

	m := mikan.Mikan{ThemoviedbKey: ThemoviedbKey}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotAnime := m.Parse(&models.AnimeParseOptions{
				Name: tt.args.name,
				Url:  tt.args.url,
			})
			data1, _ := json.Marshal(gotAnime)
			t.Log(string(data1))
			if gotAnime.MikanID != tt.wantAnime.MikanID {
				t.Errorf("Parse().MikanID = %v, want %v", gotAnime.MikanID, tt.wantAnime.MikanID)
			}
			if gotAnime.ID != tt.wantAnime.ID {
				t.Errorf("Parse().ID = %v, want %v", gotAnime.ID, tt.wantAnime.ID)
			}
			if gotAnime.ThemoviedbID != tt.wantAnime.ThemoviedbID {
				t.Errorf("Parse().ThemoviedbID = %v, want %v", gotAnime.ThemoviedbID, tt.wantAnime.ThemoviedbID)
			}
			if gotAnime.Ep != tt.wantAnime.Ep {
				t.Errorf("Parse().Ep = %v, want %v", gotAnime.Ep, tt.wantAnime.Ep)
			}
			if gotAnime.Season != tt.wantAnime.Season {
				t.Errorf("Parse().Season = %v, want %v", gotAnime.Season, tt.wantAnime.Season)
			}

		})
	}
}
