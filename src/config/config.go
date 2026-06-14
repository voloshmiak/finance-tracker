package config

import (
	"fmt"
	"log"

	"github.com/caarlos0/env/v6"
	"github.com/joho/godotenv"
)

type Config struct {
	Server    Server
	Database  Database
	Telegram  Telegram
	Assistant Assistant
}

type Server struct {
	Port      string `env:"PORT" envDefault:"8080"`
	AuthToken string `env:"SERVER_AUTH_TOKEN,required"`
}

type Database struct {
	Host     string `env:"DB_HOST,required"`
	Port     string `env:"DB_PORT" envDefault:"5432"`
	Name     string `env:"DB_NAME,required"`
	User     string `env:"DB_USER,required"`
	Password string `env:"DB_PASSWORD,required"`
	SSLMode  string `env:"DB_SSLMODE" envDefault:"disable"`

	// ReadOnlyUser/Password are a restricted role (GRANT SELECT only) used by
	// the AI assistant's SQL tool. Falls back to the main user if unset.
	ReadOnlyUser     string `env:"DB_READONLY_USER"`
	ReadOnlyPassword string `env:"DB_READONLY_PASSWORD"`
}

func (d Database) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		d.User, d.Password, d.Host, d.Port, d.Name, d.SSLMode)
}

// ReadOnlyDSN builds a connection string for the restricted read-only role.
func (d Database) ReadOnlyDSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		d.ReadOnlyUser, d.ReadOnlyPassword, d.Host, d.Port, d.Name, d.SSLMode)
}

type Assistant struct {
	GeminiAPIKey string `env:"GEMINI_API_KEY,required"`
	Model        string `env:"ASSISTANT_MODEL" envDefault:"gemini-2.5-flash"`
	Username     string `env:"ASSISTANT_USERNAME,required"`
	Password     string `env:"ASSISTANT_PASSWORD,required"`
}

type Telegram struct {
	Token         string `env:"TELEGRAM_TOKEN,required"`
	ChatID        int    `env:"TELEGRAM_CHAT_ID,required"`
	Host          string `env:"TELEGRAM_HOST" envDefault:"https://api.telegram.org"`
	WebhookSecret string `env:"TELEGRAM_WEBHOOK_SECRET,required"`
}

func New() (*Config, error) {
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("No .env file found, relying on environment variables")
	}

	cfg := new(Config)
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
