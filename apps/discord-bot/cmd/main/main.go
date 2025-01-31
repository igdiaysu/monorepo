package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"

	_ "time/tzdata"

	"github.com/bwmarrin/discordgo"
	"github.com/senchabot-opensource/monorepo/apps/discord-bot/internal/handler"
	"github.com/senchabot-opensource/monorepo/apps/discord-bot/internal/service"
	twsrvc "github.com/senchabot-opensource/monorepo/service/twitch"
)

func main() {
	twsrvc.InitTwitchOAuth2Token()

	discordClient, _ := discordgo.New("Bot " + os.Getenv("TOKEN"))

	var wg sync.WaitGroup

	service := service.New()
	handler := handler.New(discordClient, service)

	handler.InitBotEventHandlers()

	go func() {
		err := discordClient.Open()
		if err != nil {
			log.Fatal("Cannot open the session: ", err)
		}
		defer discordClient.Close()

		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt)
		<-stop
		wg.Done()

		log.Println("Graceful shutdown")
	}()

	go func() {
		log.Println("Starting HTTP server...")
		mux := http.NewServeMux()
		handler.InitHttpHandlers(mux)

		error := http.ListenAndServe(":8080", mux)
		if error != nil {
			log.Fatal("ListenAndServe Error:", error)
		}
	}()

	select {}
}
