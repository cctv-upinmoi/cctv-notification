package notifier

import (
	"fmt"
	"strconv"
	"strings"
	"thanhd/smart-cctv/cctv-notification/internal/config"
	"thanhd/smart-cctv/cctv-notification/internal/model"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type TelegramNotifier struct {
	bot *tgbotapi.BotAPI
}

func NewTelegramNotifier(cfg *config.Config) (*TelegramNotifier, error) {
	bot, err := tgbotapi.NewBotAPI(cfg.TelegramToken)
	if err != nil {
		return nil, fmt.Errorf("init telegram bot: %w", err)
	}
	return &TelegramNotifier{bot: bot}, nil
}

func (n *TelegramNotifier) Name() string { return "telegram" }

func (n *TelegramNotifier) Send(event model.NotificationDispatchEvent, chatIDStr string) error {
	chatID, err := strconv.ParseInt(chatIDStr, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid telegram chat ID %q: %w", chatIDStr, err)
	}

	zoneLine := ""
	if event.ZoneName != "" {
		zoneLine = fmt.Sprintf("📍 *Khu vực:* %s\n", escapeMarkdown(event.ZoneName))
	}
	text := fmt.Sprintf(
		"%s *%s*\n\n"+
			"📷 *Camera:* %s\n"+
			"%s"+
			"👤 *Số người:* %d\n"+
			"📊 *Độ tin cậy:* %.1f%%\n"+
			"🕐 *Thời gian:* %s",
		alertEmoji(event.AlertType),
		escapeMarkdown(alertLabel(event.AlertType)),
		escapeMarkdown(event.CameraName),
		zoneLine,
		event.PersonCount,
		event.Confidence*100,
		event.DetectedAt.Format("02/01/2006 15:04:05"),
	)

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = tgbotapi.ModeMarkdown
	if _, err := n.bot.Send(msg); err != nil {
		return fmt.Errorf("send message to %d: %w", chatID, err)
	}

	if event.ImageURL != "" {
		photo := tgbotapi.NewPhoto(chatID, tgbotapi.FileURL(event.ImageURL))
		if _, err := n.bot.Send(photo); err != nil {
			// non-fatal: Telegram servers may not reach internal imageUrl
			fmt.Printf("[telegram] warn: send photo to %d: %v\n", chatID, err)
		}
	}

	return nil
}

func escapeMarkdown(s string) string {
	return strings.NewReplacer("_", "\\_", "*", "\\*", "[", "\\[", "`", "\\`").Replace(s)
}
