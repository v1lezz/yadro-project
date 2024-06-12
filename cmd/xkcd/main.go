package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os/signal"
	"sync"
	"syscall"
	"time"
	"yadro-project/internal/adapters/handler"
	"yadro-project/internal/adapters/index"
	"yadro-project/internal/adapters/repository"
	"yadro-project/internal/config"
	"yadro-project/internal/core/services"
	"yadro-project/pkg/words"
	"yadro-project/pkg/xkcd"
)

func main() {
	var cfgPath string
	flag.StringVar(&cfgPath, "c", "", "parse file path config")
	flag.Parse()
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()
	cfg, err := config.NewConfig(cfgPath)
	if err != nil {
		log.Fatal(err)
	}

	db, err := repository.NewPostgresConn(ctx, cfg.DbCFG)
	if err != nil {
		log.Fatal(err)
	}
	idx, err := index.NewPostgresConn(ctx, cfg.DbCFG)
	if err != nil {
		log.Fatal(err)
	}
	stemmer := words.NewSnowBallStem()
	parser := xkcd.NewXkcdParse(cfg.AppCFG.SourceURL, cfg.AppCFG.Parallel, stemmer)
	cSVC := services.NewComicsService(db, parser, idx, stemmer)

	authDB, err := repository.NewAuthJSONRepository("users.json")
	if err != nil {
		log.Fatal(err)
	}
	aSVC := services.NewAuthService(authDB, cfg.AuthCFG.TokenMaxTime)
	lSVC := services.NewLimitService(cfg.SrvCFG.RateLimit, cfg.SrvCFG.ConcurrencyLimit)
	mutex := &sync.Mutex{}
	srv := NewServer(ctx, *cSVC, *lSVC, *aSVC, mutex, fmt.Sprintf(":%d", cfg.SrvCFG.Port))
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		if err = srv.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()
	<-ctx.Done()
	//ждем завершения последнего апдейта
	mutex.Lock()
	ctx, cancel = context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	if err = srv.Shutdown(ctx); err != nil {
		log.Fatal(err)
	}

}

func NewServer(ctx context.Context, cSVC services.ComicsService, lSVC services.LimitService, aSVC services.AuthService, mutex *sync.Mutex, addr string) *http.Server {
	router := http.NewServeMux()
	c := handler.NewComicsHandler(cSVC, mutex)
	l := handler.NewLimitHandler(lSVC, aSVC)
	a := handler.NewAuthHandler(aSVC)
	router.Handle("GET /pics", a.AuthMiddleware(http.HandlerFunc(c.GetComics)))
	router.Handle("POST /update", l.LimitingMiddleware(http.HandlerFunc(c.UpdateComics)))
	router.HandleFunc("POST /login", a.LoginHandler)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Hour * 24):
				if mutex.TryLock() {
					_, _ = cSVC.UpdateComics(ctx)
					mutex.Unlock()
				}
			}
		}
	}()
	return &http.Server{
		Addr:    addr,
		Handler: router,
		BaseContext: func(_ net.Listener) context.Context {
			return ctx
		},
	}

}
