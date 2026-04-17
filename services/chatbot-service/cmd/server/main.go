package main

import (
	"context"
	"encoding/json"
	"os"
	"os/signal"
	"syscall"

	"github.com/societykro/go-common/config"
	"github.com/societykro/go-common/database"
	"github.com/societykro/go-common/events"
	"github.com/societykro/go-common/logger"

	"github.com/societykro/chatbot-service/internal/service"
)

func main() {
	cfg := config.Load()

	logger.Init(cfg.App.Env)
	log := &logger.Log
	log.Info().Str("service", "chatbot-service").Msg("Starting")

	// Redis (for conversation state)
	rdb, err := database.NewRedisClient(cfg.Redis.Addr(), cfg.Redis.Password, cfg.Redis.DB)
	if err != nil {
		log.Fatal().Err(err).Msg("Redis connection failed")
	}
	defer rdb.Close()

	// NATS
	bus, err := events.NewBus(cfg.NatsURL)
	if err != nil {
		log.Fatal().Err(err).Msg("NATS connection failed")
	}
	defer bus.Close()

	bus.EnsureStream("MESSAGES", []string{"message.*"})
	bus.EnsureStream("CHATBOT_ACTIONS", []string{"chatbot.*"})

	// Chatbot service
	chatbot := service.NewChatbotService(rdb, bus)

	// Subscribe to message.received events from message-router
	err = bus.Subscribe("message.received", "chatbot-consumer", func(evt events.Event) error {
		var msg struct {
			SenderID  string `json:"sender_id"`
			Source    string `json:"source"`
			Text      string `json:"text"`
			UserID    string `json:"user_id"`
			SocietyID string `json:"society_id"`
			Language  string `json:"language"`
		}

		if err := json.Unmarshal(evt.Payload, &msg); err != nil {
			log.Error().Err(err).Msg("Failed to unmarshal message event")
			return nil // don't retry
		}

		if msg.Text == "" {
			log.Debug().Str("sender", msg.SenderID).Msg("Empty text message, skipping")
			return nil
		}

		resp, err := chatbot.ProcessMessage(
			context.Background(),
			msg.SenderID, msg.Source, msg.Text,
			msg.UserID, msg.SocietyID, msg.Language,
		)
		if err != nil {
			log.Error().Err(err).Str("sender", msg.SenderID).Msg("Chatbot processing failed")
			return nil
		}

		// Publish response for message-router to send back
		bus.Publish("chatbot.response", "chatbot.response", map[string]interface{}{
			"destination": msg.SenderID,
			"channel":     msg.Source,
			"text":        resp.Text,
			"buttons":     resp.Buttons,
		})

		log.Info().Str("sender", msg.SenderID).Str("response", resp.Text[:min(50, len(resp.Text))]).Msg("Bot response generated")
		return nil
	})
	if err != nil {
		log.Fatal().Err(err).Msg("NATS subscription failed")
	}

	log.Info().Msg("chatbot-service running — listening for messages")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info().Msg("Shutting down chatbot-service...")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
