package main

import (
	"bufio"
	"log"
	"os"
)

func main() {
	app := NewApp(NewFlagParse(), NewSnowBallStem(), bufio.NewWriter(os.Stdout))
	defer app.Close()
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
