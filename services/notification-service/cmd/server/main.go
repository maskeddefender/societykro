package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/societykro/go-common/config"
	"github.com/societykro/go-common/database"
	"github.com/societykro/go-common/events"
	"github.com/societykro/go-common/logger"

	"github.com/societykro/notification-service/internal/service"
)

func main() {
	cfg := config.Load()

	logger.Init(cfg.App.Env)
	log := &logger.Log
	log.Info().Str("service", "notification-service").Msg("Starting")

	// PostgreSQL (read-only, for querying notification targets)
	pool, err := database.NewPostgresPool(cfg.Database.DSN(), cfg.Database.MaxConnections)
	if err != nil {
		log.Fatal().Err(err).Msg("PostgreSQL connection failed")
	}
	defer pool.Close()

	// NATS
	bus, err := events.NewBus(cfg.NatsURL)
	if err != nil {
		log.Fatal().Err(err).Msg("NATS connection failed")
	}
	defer bus.Close()

	if err := bus.EnsureStream(events.StreamSocietyKro, events.AllSubjects()); err != nil {
		log.Fatal().Err(err).Msg("NATS stream creation failed")
	}

	// Dispatcher with stub senders (replace with real implementations later)
	dispatcher := service.NewDispatcher(
		&service.StubPushSender{},
		&service.StubWhatsAppSender{},
		&service.StubTelegramSender{},
		&service.StubSMSSender{},
	)

	handler := service.NewEventHandler(pool, dispatcher)

	// Subscribe to ALL event subjects
	subjects := []string{
		"complaint.*",
		"visitor.*",
		"payment.*",
		"notice.*",
		"sos.*",
	}

	for _, subj := range subjects {
		durableName := "notification-" + subj
		if err := bus.Subscribe(subj, durableName, handler.Handle); err != nil {
			log.Fatal().Err(err).Str("subject", subj).Msg("NATS subscription failed")
		}
	}

	log.Info().Msg("notification-service running — listening for events")

	// Block until signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info().Msg("Shutting down notification-service...")
}
