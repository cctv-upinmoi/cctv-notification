package consumer

import (
	"context"
	"encoding/json"
	"log"

	"github.com/IBM/sarama"
	"thanhd/smart-cctv/cctv-notification/internal/config"
	"thanhd/smart-cctv/cctv-notification/internal/model"
	"thanhd/smart-cctv/cctv-notification/internal/notifier"
)

type KafkaConsumer struct {
	cfg      *config.Config
	email    notifier.Notifier
	telegram notifier.Notifier
}

func NewKafkaConsumer(cfg *config.Config, email, telegram notifier.Notifier) *KafkaConsumer {
	return &KafkaConsumer{cfg: cfg, email: email, telegram: telegram}
}

func (c *KafkaConsumer) Run(ctx context.Context) {
	saramaCfg := sarama.NewConfig()
	saramaCfg.Consumer.Offsets.Initial = sarama.OffsetNewest
	saramaCfg.Consumer.Return.Errors = true
	saramaCfg.Version = sarama.V3_0_0_0

	client, err := sarama.NewConsumerGroup(c.cfg.KafkaBrokers, c.cfg.KafkaGroupID, saramaCfg)
	if err != nil {
		log.Fatalf("[kafka] failed to create consumer group: %v", err)
	}
	defer client.Close()

	handler := &kafkaHandler{email: c.email, telegram: c.telegram}
	log.Printf("[kafka] connected — topic=%s group=%s", c.cfg.KafkaTopic, c.cfg.KafkaGroupID)

	for {
		if ctx.Err() != nil {
			return
		}
		if err := client.Consume(ctx, []string{c.cfg.KafkaTopic}, handler); err != nil {
			log.Printf("[kafka] consume error: %v", err)
		}
	}
}

// kafkaHandler implements sarama.ConsumerGroupHandler
type kafkaHandler struct {
	email    notifier.Notifier
	telegram notifier.Notifier
}

func (h *kafkaHandler) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (h *kafkaHandler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }

func (h *kafkaHandler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		var event model.NotificationDispatchEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			log.Printf("[kafka] invalid message: %v", err)
			sess.MarkMessage(msg, "")
			continue
		}

		log.Printf("[kafka] received alert: camera=%s zone=%s recipients=%d",
			event.CameraName, event.ZoneName, len(event.Recipients))

		for _, r := range event.Recipients {
			for _, ch := range r.Channels {
				switch ch {
				case "EMAIL":
					if r.Email != "" && h.email != nil {
						if err := h.email.Send(event, r.Email); err != nil {
							log.Printf("[email] send to %s failed: %v", r.Email, err)
						}
					}
				case "TELEGRAM":
					if r.TelegramChatID != "" && h.telegram != nil {
						if err := h.telegram.Send(event, r.TelegramChatID); err != nil {
							log.Printf("[telegram] send to %s failed: %v", r.TelegramChatID, err)
						}
					}
				}
			}
		}

		sess.MarkMessage(msg, "")
	}
	return nil
}
