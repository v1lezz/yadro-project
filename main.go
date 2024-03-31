package main

import (
	"log"
	"os"
)

func main() {
	app := NewApp(NewFlagParse(), NewSnowBallStem(), os.Stdout)
	defer app.Close()
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
