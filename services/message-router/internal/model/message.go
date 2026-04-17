package model

import "time"

// IncomingMessage is the normalized internal representation of a message
// received from any channel (WhatsApp, Telegram).
type IncomingMessage struct {
	ID          string            `json:"id"`
	Source      string            `json:"source"`       // "whatsapp" or "telegram"
	SenderID    string            `json:"sender_id"`    // phone for WA, chat_id for TG
	SenderPhone string            `json:"sender_phone"` // always phone if available
	MessageType string            `json:"message_type"` // "text", "voice", "image", "location"
	Text        string            `json:"text,omitempty"`
	VoiceURL    string            `json:"voice_url,omitempty"`
	ImageURL    string            `json:"image_url,omitempty"`
	Language    string            `json:"language,omitempty"` // detected by Bhashini LID
	UserID      string            `json:"user_id,omitempty"` // resolved from phone lookup
	SocietyID   string            `json:"society_id,omitempty"`
	FlatID      string            `json:"flat_id,omitempty"`
	Role        string            `json:"role,omitempty"`
	Timestamp   time.Time         `json:"timestamp"`
	RawPayload  map[string]interface{} `json:"raw_payload,omitempty"`
}

// OutgoingMessage is the response to send back to the user.
type OutgoingMessage struct {
	Destination string            `json:"destination"` // phone or chat_id
	Channel     string            `json:"channel"`     // "whatsapp" or "telegram"
	Text        string            `json:"text"`
	Buttons     []MessageButton   `json:"buttons,omitempty"`
	ImageURL    string            `json:"image_url,omitempty"`
}

// MessageButton represents an interactive button in WA/TG messages.
type MessageButton struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

// WhatsAppWebhookPayload is the incoming webhook body from Meta Cloud API.
type WhatsAppWebhookPayload struct {
	Object string `json:"object"`
	Entry  []struct {
		ID      string `json:"id"`
		Changes []struct {
			Value struct {
				MessagingProduct string `json:"messaging_product"`
				Metadata         struct {
					DisplayPhoneNumber string `json:"display_phone_number"`
					PhoneNumberID      string `json:"phone_number_id"`
				} `json:"metadata"`
				Contacts []struct {
					Profile struct {
						Name string `json:"name"`
					} `json:"profile"`
					WaID string `json:"wa_id"`
				} `json:"contacts"`
				Messages []struct {
					From      string `json:"from"`
					ID        string `json:"id"`
					Timestamp string `json:"timestamp"`
					Type      string `json:"type"`
					Text      *struct {
						Body string `json:"body"`
					} `json:"text,omitempty"`
					Audio *struct {
						ID       string `json:"id"`
						MimeType string `json:"mime_type"`
					} `json:"audio,omitempty"`
					Image *struct {
						ID       string `json:"id"`
						MimeType string `json:"mime_type"`
						Caption  string `json:"caption"`
					} `json:"image,omitempty"`
					Interactive *struct {
						Type        string `json:"type"`
						ButtonReply *struct {
							ID    string `json:"id"`
							Title string `json:"title"`
						} `json:"button_reply,omitempty"`
						ListReply *struct {
							ID    string `json:"id"`
							Title string `json:"title"`
						} `json:"list_reply,omitempty"`
					} `json:"interactive,omitempty"`
				} `json:"messages"`
			} `json:"value"`
		} `json:"changes"`
	} `json:"entry"`
}

// TelegramUpdate is the incoming webhook body from Telegram Bot API.
type TelegramUpdate struct {
	UpdateID int `json:"update_id"`
	Message  *struct {
		MessageID int `json:"message_id"`
		From      struct {
			ID        int64  `json:"id"`
			FirstName string `json:"first_name"`
			LastName  string `json:"last_name"`
			Username  string `json:"username"`
		} `json:"from"`
		Chat struct {
			ID   int64  `json:"id"`
			Type string `json:"type"`
		} `json:"chat"`
		Date  int    `json:"date"`
		Text  string `json:"text,omitempty"`
		Voice *struct {
			FileID   string `json:"file_id"`
			Duration int    `json:"duration"`
		} `json:"voice,omitempty"`
		Photo []struct {
			FileID string `json:"file_id"`
			Width  int    `json:"width"`
			Height int    `json:"height"`
		} `json:"photo,omitempty"`
	} `json:"message"`
	CallbackQuery *struct {
		ID      string `json:"id"`
		Data    string `json:"data"`
		Message *struct {
			Chat struct {
				ID int64 `json:"id"`
			} `json:"chat"`
		} `json:"message"`
	} `json:"callback_query"`
}
