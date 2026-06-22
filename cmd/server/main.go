package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"thanhd/smart-cctv/cctv-notification/internal/config"
	"thanhd/smart-cctv/cctv-notification/internal/consumer"
	"thanhd/smart-cctv/cctv-notification/internal/notifier"
)

func main() {
	cfg := config.Load()

	var emailNotifier notifier.Notifier
	var telegramNotifier notifier.Notifier

	if cfg.BrevoAPIKey != "" && cfg.BrevoSenderEmail != "" {
		n, err := notifier.NewBrevoNotifier(cfg)
		if err != nil {
			log.Fatalf("init brevo notifier: %v", err)
		}
		emailNotifier = n
		log.Println("[startup] email notifier enabled (brevo)")
	}

	if cfg.TelegramToken != "" {
		n, err := notifier.NewTelegramNotifier(cfg)
		if err != nil {
			log.Fatalf("init telegram notifier: %v", err)
		}
		telegramNotifier = n
		log.Println("[startup] telegram notifier enabled")
	}

	if emailNotifier == nil && telegramNotifier == nil {
		log.Println("[startup] warning: no notifiers configured — set BREVO_API_KEY+BREVO_SENDER_EMAIL or TELEGRAM_TOKEN")
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	c := consumer.NewKafkaConsumer(cfg, emailNotifier, telegramNotifier)

	log.Printf("[startup] cctv-notification started — brokers=%v topic=%s group=%s",
		cfg.KafkaBrokers, cfg.KafkaTopic, cfg.KafkaGroupID)
	c.Run(ctx)
	log.Println("[shutdown] cctv-notification stopped")
}
