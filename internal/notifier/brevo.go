package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"

	"thanhd/smart-cctv/cctv-notification/internal/config"
	"thanhd/smart-cctv/cctv-notification/internal/model"
)

const brevoAPIURL = "https://api.brevo.com/v3/smtp/email"

// alertLabel returns a Vietnamese display label for a given alertType.
func alertLabel(alertType string) string {
	switch alertType {
	case "INTRUSION":
		return "Xâm nhập vùng cấm"
	case "NO_HARDHAT":
		return "Không đội mũ bảo hiểm"
	case "NO_SAFETY_VEST":
		return "Không mặc áo bảo hộ"
	case "NO_MASK":
		return "Không đeo khẩu trang"
	default:
		return alertType
	}
}

func alertEmoji(alertType string) string {
	switch alertType {
	case "INTRUSION":
		return "🚨"
	default:
		return "⚠️"
	}
}

const emailHTML = `<!DOCTYPE html>
<html><head><meta charset="UTF-8"></head>
<body style="font-family:Arial,sans-serif;max-width:600px;margin:0 auto;padding:20px;color:#333">
  <h2 style="color:#c53030;border-bottom:2px solid #fed7d7;padding-bottom:8px">{{.Emoji}} {{.AlertLabel}}</h2>
  <table style="width:100%;border-collapse:collapse;margin:16px 0">
    <tr><td style="padding:8px;font-weight:bold;width:140px">Camera</td>
        <td style="padding:8px;border-bottom:1px solid #eee">{{.CameraName}}</td></tr>
    {{if .ZoneName}}
    <tr><td style="padding:8px;font-weight:bold">Khu vực</td>
        <td style="padding:8px;border-bottom:1px solid #eee">{{.ZoneName}}</td></tr>
    {{end}}
    <tr><td style="padding:8px;font-weight:bold">Loại</td>
        <td style="padding:8px;border-bottom:1px solid #eee">{{.AlertLabel}}</td></tr>
    <tr><td style="padding:8px;font-weight:bold">Số người</td>
        <td style="padding:8px;border-bottom:1px solid #eee">{{.PersonCount}}</td></tr>
    <tr><td style="padding:8px;font-weight:bold">Độ tin cậy</td>
        <td style="padding:8px;border-bottom:1px solid #eee">{{.ConfidencePct}}%</td></tr>
    <tr><td style="padding:8px;font-weight:bold">Thời gian</td>
        <td style="padding:8px">{{.DetectedAtFmt}}</td></tr>
  </table>
  {{if .ImageURL}}
  <a href="{{.ImageURL}}" style="display:inline-block;padding:10px 20px;background:#3182ce;color:#fff;text-decoration:none;border-radius:4px">
    Xem ảnh
  </a>
  {{end}}
</body></html>`

type emailData struct {
	CameraName    string
	ZoneName      string
	AlertLabel    string
	Emoji         string
	PersonCount   int
	ConfidencePct string
	DetectedAtFmt string
	ImageURL      string
}

type brevoSender struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type brevoRecipient struct {
	Email string `json:"email"`
}

type brevoPayload struct {
	Sender      brevoSender      `json:"sender"`
	To          []brevoRecipient `json:"to"`
	Subject     string           `json:"subject"`
	HTMLContent string           `json:"htmlContent"`
}

type BrevoNotifier struct {
	cfg    *config.Config
	tmpl   *template.Template
	client *http.Client
}

func NewBrevoNotifier(cfg *config.Config) (*BrevoNotifier, error) {
	tmpl, err := template.New("brevo").Parse(emailHTML)
	if err != nil {
		return nil, fmt.Errorf("parse brevo template: %w", err)
	}
	return &BrevoNotifier{cfg: cfg, tmpl: tmpl, client: &http.Client{}}, nil
}

func (n *BrevoNotifier) Name() string { return "brevo" }

func (n *BrevoNotifier) Send(event model.NotificationDispatchEvent, address string) error {
	var buf bytes.Buffer
	if err := n.tmpl.Execute(&buf, emailData{
		CameraName:    event.CameraName,
		ZoneName:      event.ZoneName,
		AlertLabel:    alertLabel(event.AlertType),
		Emoji:         alertEmoji(event.AlertType),
		PersonCount:   event.PersonCount,
		ConfidencePct: fmt.Sprintf("%.1f", event.Confidence*100),
		DetectedAtFmt: event.DetectedAt.Format("02/01/2006 15:04:05"),
		ImageURL:      event.ImageURL,
	}); err != nil {
		return fmt.Errorf("render brevo email: %w", err)
	}

	payload := brevoPayload{
		Sender:      brevoSender{Name: n.cfg.BrevoSenderName, Email: n.cfg.BrevoSenderEmail},
		To:          []brevoRecipient{{Email: address}},
		Subject:     fmt.Sprintf("%s [Smart CCTV] %s — %s", alertEmoji(event.AlertType), alertLabel(event.AlertType), event.CameraName),
		HTMLContent: buf.String(),
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal brevo payload: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, brevoAPIURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create brevo request: %w", err)
	}
	req.Header.Set("api-key", n.cfg.BrevoAPIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := n.client.Do(req)
	if err != nil {
		return fmt.Errorf("brevo send to %s: %w", address, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("brevo API error (status %d): %s", resp.StatusCode, string(respBody))
	}
	return nil
}
