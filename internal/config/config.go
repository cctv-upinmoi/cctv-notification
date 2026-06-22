package config

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	KafkaBrokers []string
	KafkaTopic   string
	KafkaGroupID string

	BrevoAPIKey      string
	BrevoSenderEmail string
	BrevoSenderName  string

	TelegramToken string
}

func Load() *Config {
	_ = godotenv.Load()

	brokersRaw := getEnv("KAFKA_BROKERS", "localhost:9092")
	brokers := strings.Split(brokersRaw, ",")
	for i, b := range brokers {
		brokers[i] = strings.TrimSpace(b)
	}

	return &Config{
		KafkaBrokers: brokers,
		KafkaTopic:   getEnv("KAFKA_TOPIC", "notification-dispatch"),
		KafkaGroupID: getEnv("KAFKA_GROUP_ID", "cctv-notification"),

		BrevoAPIKey:      getEnv("BREVO_API_KEY", ""),
		BrevoSenderEmail: getEnv("BREVO_SENDER_EMAIL", ""),
		BrevoSenderName:  getEnv("BREVO_SENDER_NAME", "Smart CCTV"),

		TelegramToken: getEnv("TELEGRAM_TOKEN", ""),
	}
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getEnvInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
		log.Printf("invalid int for %s, using default %d", key, def)
	}
	return def
}
