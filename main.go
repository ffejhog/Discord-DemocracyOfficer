package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/ian-kent/gofigure"
)

func main() {
	var cfg Config
	err := gofigure.Gofigure(&cfg)
	if err != nil {
		log.Fatal(err)
	}

	dBot := DBot{
		Config: cfg,
	}

	err = dBot.Connect()
	if err != nil {
		return
	}
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
	dBot.Discord.Close()
}
