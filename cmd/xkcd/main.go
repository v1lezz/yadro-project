package main

import (
	"flag"
	"log"
	"yadro-project/internal/app"
	"yadro-project/internal/config"
	"yadro-project/pkg/database"
	"yadro-project/pkg/index"
	"yadro-project/pkg/words"
	"yadro-project/pkg/xkcd"
)

func main() {
	var cfgPath, s string
	var i bool
	flag.StringVar(&cfgPath, "c", "", "parse file path config")
	flag.StringVar(&s, "s", "", "parse string for search")
	flag.BoolVar(&i, "i", false, "use index for search")
	flag.Parse()
	cfg, err := config.NewConfig(cfgPath)
	if err != nil {
		log.Fatal(err)
	}

	db, err := database.NewJsonDB(cfg.DBCfg.DBFile)
	if err != nil {
		log.Fatal(err)
	}
	idx, err := index.NewFileIndex(cfg.IndexCfg)
	if err != nil {
		log.Fatal(err)
	}
	a := app.NewApp(xkcd.NewXkcdParse(cfg.AppCFG.SourceURL, cfg.AppCFG.Parallel),
		words.NewSnowBallStem(),
		db, idx)
	if err = a.Run(s, i); err != nil {
		log.Fatal(err)
	}
}
