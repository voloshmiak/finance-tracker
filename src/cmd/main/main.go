package main

import (
	"finance-tracker/src/config"
	"finance-tracker/src/controller"
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
	s := server.New(":"+c.Server.Port,
		controller.NewDocsController(),
		controller.NewHealthController(),
	)
	return s.Start()
}
