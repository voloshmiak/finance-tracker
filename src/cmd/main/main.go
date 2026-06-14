package main

import (
	"context"
	"finance-tracker/src/assistant"
	"finance-tracker/src/config"
	"finance-tracker/src/controller"
	"finance-tracker/src/database"
	"finance-tracker/src/parser"
	"finance-tracker/src/repository"
	"finance-tracker/src/server"
	"finance-tracker/src/telegram"
	"finance-tracker/src/usecase"
	"log"
)

// @title           Finance Tracker API
// @version         0.0.5
// @description		This is the API documentation for the Finance Tracker application
// @BasePath        /api
func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	c, err := config.New()
	if err != nil {
		return err
	}

	db, err := database.New(c.Database.DSN())
	if err != nil {
		return err
	}
	defer db.Conn().Close()

	if err = db.Migrate(); err != nil {
		return err
	}

	client := telegram.NewClient(c.Telegram.Token, c.Telegram.Host, c.Telegram.ChatID)

	transactionRepository := repository.NewTransactionRepository(db.Conn())
	cardRepository := repository.NewCardRepository(db.Conn())
	parserRegistry := parser.NewRegistry()

	// The assistant's SQL tool runs under a dedicated read-only connection.
	readDB, err := database.New(c.Database.ReadOnlyDSN())
	if err != nil {
		return err
	}
	defer readDB.Conn().Close()

	ai, err := assistant.New(context.Background(),
		c.Assistant.GeminiAPIKey, c.Assistant.Model,
		assistant.NewSQLTool(readDB.Conn()),
	)
	if err != nil {
		return err
	}

	s := server.New(":"+c.Server.Port,
		controller.NewDocsController(),
		controller.NewHealthController(),
		controller.NewReminderController(
			usecase.NewReminder(client, cardRepository, c.Server.AuthToken),
		),
		controller.NewWebhookController(
			usecase.NewWebhook(parserRegistry, transactionRepository, client, client, c.Telegram.ChatID),
			c.Telegram.WebhookSecret,
		),
		controller.NewAssistantController(
			usecase.NewAssistant(ai),
			c.Assistant.Username, c.Assistant.Password,
		),
	)

	return s.Start()
}
