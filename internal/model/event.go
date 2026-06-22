package model

import "time"

type NotificationDispatchEvent struct {
	EventID     string          `json:"eventId"`
	CameraID    string          `json:"cameraId"`
	CameraName  string          `json:"cameraName"`
	ZoneName    string          `json:"zoneName"`
	AlertType   string          `json:"alertType"`
	Confidence  float64         `json:"confidence"`
	PersonCount int             `json:"personCount"`
	DetectedAt  time.Time       `json:"detectedAt"`
	ImageURL    string          `json:"imageUrl"`
	Recipients  []RecipientInfo `json:"recipients"`
}

type RecipientInfo struct {
	UserID         string   `json:"userId"`
	Email          string   `json:"email"`
	TelegramChatID string   `json:"telegramChatId"`
	Channels       []string `json:"channels"`
}
