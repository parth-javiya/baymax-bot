package whatsappclient

import (
	whatsapp "github.com/Rhymen/go-whatsapp"
)

// Message content body whatsapp text message
type Message struct {
	Type          string
	From          string
	ID            string
	Info          whatsapp.MessageInfo
	TextSource    whatsapp.TextMessage
	ImageSource   whatsapp.ImageMessage
	StickerSource whatsapp.StickerMessage
	Actions       GroupAction
}

type GroupAction struct {
	Action       string
	ActionBy     string
	Participants map[string]interface{}
}
