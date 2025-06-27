package providers

import (
	"fmt"
	"log"
	"strconv"

	"base_lara_go_project/app/core"
	"base_lara_go_project/config"

	"github.com/joho/godotenv"
	"gopkg.in/gomail.v2"
)

func RegisterMailer() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	// Get mail configuration from config package
	mailConfig := config.MailConfig()
	defaultMailer := mailConfig["default"].(string)
	mailers := mailConfig["mailers"].(map[string]interface{})
	mailerConfig := mailers[defaultMailer].(map[string]interface{})
	fromConfig := mailConfig["from"].(map[string]interface{})

	host := mailerConfig["host"].(string)
	portStr := mailerConfig["port"].(string)
	username := mailerConfig["username"].(string)
	password := mailerConfig["password"].(string)
	from := fromConfig["address"].(string)
	fromName := fromConfig["name"].(string)

	// Convert port string to integer
	port, err := strconv.Atoi(portStr)
	if err != nil {
		log.Fatalf("Invalid MAIL_PORT: %s", portStr)
	}

	// Create mail configuration
	mailConfigInstance := &core.MailConfig{
		Host:     host,
		Port:     port,
		Username: username,
		Password: password,
		From:     from,
		FromName: fromName,
	}

	// Create mailer dialer
	mailer := gomail.NewDialer(host, port, username, password)

	// Create mail provider and set global instance
	mailProvider := core.NewMailProvider(mailConfigInstance, mailer)
	core.SetMailService(mailProvider)

	fmt.Printf("Mailer configured for %s:%d\n", host, port)
}
