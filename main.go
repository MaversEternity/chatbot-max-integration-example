package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	maxbot "github.com/max-messenger/max-bot-api-client-go"

	"max-chat/bot"
	"max-chat/mockserver"
)

func main() {
	token := os.Getenv("TOKEN")
	if token == "" {
		token = "test-token"
	}
	mock := os.Getenv("MOCK") == "true" || os.Getenv("MOCK") == "1"

	if mock {
		// Start mock server in background
		srv := mockserver.New()
		go func() {
			if err := srv.Start(":8081"); err != nil {
				log.Fatalf("Mock server error: %v", err)
			}
		}()
		time.Sleep(100 * time.Millisecond) // let server start
		log.Println("Mock mode enabled")
	}

	// Build API options
	var opts []maxbot.Option
	if mock {
		opts = append(opts, maxbot.WithBaseURL("http://localhost:8081"))
	}
	if !mock {
		opts = append(opts, maxbot.WithDebugMode())
	}

	api, err := maxbot.New(token, opts...)
	if err != nil {
		fmt.Printf("Failed to create API client: %v\n", err)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		exit := make(chan os.Signal, 1)
		signal.Notify(exit, os.Kill, os.Interrupt)
		<-exit
		cancel()
	}()

	// Clear any existing webhook subscriptions so long polling works
	if !mock {
		subs, err := api.Subscriptions.GetSubscriptions(ctx)
		if err != nil {
			log.Printf("Failed to get subscriptions: %v", err)
		} else {
			log.Printf("Found %d webhook subscriptions", len(subs.Subscriptions))
			for _, s := range subs.Subscriptions {
				res, err := api.Subscriptions.Unsubscribe(ctx, s.Url)
				if err != nil {
					log.Printf("Failed to unsubscribe %s: %v", s.Url, err)
				} else {
					log.Printf("Removed webhook subscription: %s (success=%v)", s.Url, res.Success)
				}
			}
		}
	}

	info, err := api.Bots.GetBot(ctx)
	if err != nil {
		fmt.Printf("Failed to get bot info: %v\n", err)
		return
	}
	fmt.Printf("Bot started: %s\n", info.Name)

	// Listen for API errors in background
	go func() {
		for err := range api.GetErrors() {
			log.Printf("API error: %v", err)
		}
	}()

	// Debug: verify API connectivity
	log.Println("Starting long polling for updates...")
	log.Printf("Bot ID: %d, Username: %s", info.UserId, info.Username)

	// Debug: test raw updates endpoint
	if !mock {
		go func() {
			resp, err := http.Get(fmt.Sprintf("https://platform-api.max.ru/updates?access_token=%s&timeout=5&limit=1", token))
			if err != nil {
				log.Printf("DEBUG raw updates call failed: %v", err)
				return
			}
			defer resp.Body.Close()
			body, _ := io.ReadAll(resp.Body)
			log.Printf("DEBUG raw updates response: status=%d body=%s", resp.StatusCode, string(body))
		}()
	}

	tasks := bot.NewTaskStore()
	handler := bot.NewHandler(api, tasks)
	handler.Run(ctx)
}
