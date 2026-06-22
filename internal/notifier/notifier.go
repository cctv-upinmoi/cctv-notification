package notifier

import "thanhd/smart-cctv/cctv-notification/internal/model"

// Notifier dispatches an alert to a single contact address.
type Notifier interface {
	Send(event model.NotificationDispatchEvent, address string) error
	Name() string
}
