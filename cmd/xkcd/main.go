package main

import (
	"log"
	"yadro-project/internal/app"
	"yadro-project/internal/config"
	"yadro-project/pkg/database"
	"yadro-project/pkg/words"
	"yadro-project/pkg/xkcd"
)

func main() {
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatal(err)
	}

	db, err := database.NewJsonDB(cfg.DBCfg.DBFile)
	if err != nil {
		log.Fatal(err)
	}
	a := app.NewApp(xkcd.NewXkcdParse(cfg.AppCFG.SourceURL, cfg.AppCFG.Parallel),
		words.NewSnowBallStem(),
		db)
	if err = a.Run(); err != nil {
		log.Fatal(err)
	}
}
