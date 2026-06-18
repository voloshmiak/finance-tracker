package main

import (
	"finance-tracker/src/assistant"
	"finance-tracker/src/assistant/tool"
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
// @version         0.1.0
// @description		This is theAPI documentation for the Finance Tracker application
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

	transactions := repository.NewTransactions(db.Conn())
	cards := repository.NewCards(db.Conn())
	parsers := parser.NewRegistry()

	readDB, err := database.New(c.Database.ReadOnlyDSN())
	if err != nil {
		return err
	}
	defer readDB.Conn().Close()

	sql := tool.NewSQL(readDB.Conn())

	gemini, err := assistant.New(c.Assistant.GeminiAPIKey)
	if err != nil {
		return err
	}
	gemini.Register(sql)

	reminder := usecase.NewReminder(client, cards, c.Server.AuthToken)
	webhook := usecase.NewWebhook(parsers, transactions, client, client, c.Telegram.ChatID)
	assistant := usecase.NewAssistant(gemini)

	auth := server.BasicAuth(c.Assistant.Username, c.Assistant.Password)
	s := server.New(":"+c.Server.Port, auth)
	s.Register(
		controller.NewDocs(),
		controller.NewHealth(),
		controller.NewReminder(reminder),
		controller.NewWebhook(webhook, c.Telegram.WebhookSecret),
		controller.NewAssistantPage(assistant),
		controller.NewAssistantAsk(assistant),
	)

	return s.Start()
}
