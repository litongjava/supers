package events

import (
	"bytes"
	"encoding/json"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"net/http"

	"github.com/litongjava/supers/utils"
)

// WebhookHandler sends events to configured webhook URLs.
type WebhookHandler struct {
	URLs []string
}

// NewWebhookHandler creates a handler using the config.
func NewWebhookHandler() *WebhookHandler {
	if utils.CONFIG.Events == nil {
		return &WebhookHandler{}
	}
	return &WebhookHandler{URLs: utils.CONFIG.Events.Webhooks}
}

// Handle marshals the event and posts to each URL.
func (w *WebhookHandler) Handle(e Event) {
	payload, _ := json.Marshal(e)
	for _, url := range w.URLs {
		_, err := http.Post(url, "application/json", bytes.NewBuffer(payload))
		if err != nil {
			hlog.Error(err)
		}
	}
}
