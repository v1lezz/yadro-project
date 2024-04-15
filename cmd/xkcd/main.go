package main

import (
	"log"
	"os"
	"yadro-project/internal/app"
	"yadro-project/internal/config"
	"yadro-project/pkg/database"
	"yadro-project/pkg/flag_parser"
	"yadro-project/pkg/words"
	"yadro-project/pkg/xkcd"
)

func main() {
	cfg, err := config.NewConfig("config.yaml")
	if err != nil {
		log.Fatal(err)
	}
	parallel := 10
	a := app.NewApp(xkcd.NewXkcdParse(cfg.AppCFG.SourceURL, parallel),
		words.NewSnowBallStem(),
		database.NewJsonDB(cfg.DBCfg.DBFile),
		flag_parser.NewFlagParser(),
		os.Stdout)
	defer a.Close()
	if err = a.Run(); err != nil {
		log.Fatal(err)
	}
}
