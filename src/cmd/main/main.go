package main

import (
	"finance-tracker/src/config"
	"finance-tracker/src/controller"
	"finance-tracker/src/database"
	"finance-tracker/src/repository"
	"finance-tracker/src/server"
	"log"
)

// @title           Finance Tracker API
// @version         0.0.1
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

	expenseRepository := repository.NewExpenseRepository(db.Conn())

	s := server.New(":"+c.Server.Port,
		controller.NewDocsController(),
		controller.NewHealthController(),
		controller.NewCreateExpenseController(expenseRepository),
	)

	return s.Start()
}
