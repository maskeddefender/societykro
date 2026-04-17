package model

// Intent represents a detected user intent from their message.
type Intent struct {
	Name       string            `json:"name"`       // e.g. "raise_complaint", "check_status", "approve_visitor"
	Confidence float64           `json:"confidence"`  // 0.0 to 1.0
	Entities   map[string]string `json:"entities"`    // extracted entities (category, ticket_number, etc.)
	Layer      string            `json:"layer"`       // "keyword", "pattern", "llm"
}

// ConversationState tracks multi-turn chatbot conversations in Redis.
type ConversationState struct {
	SenderID    string            `json:"sender_id"`
	Channel     string            `json:"channel"`     // "whatsapp" or "telegram"
	CurrentFlow string            `json:"current_flow"` // "", "complaint", "visitor", "payment"
	Step        int               `json:"step"`
	Data        map[string]string `json:"data"`        // accumulated data during flow
	UserID      string            `json:"user_id"`
	SocietyID   string            `json:"society_id"`
	Language    string            `json:"language"`
}

// BotResponse is what the chatbot sends back to the user.
type BotResponse struct {
	Text    string       `json:"text"`
	Buttons []BotButton  `json:"buttons,omitempty"`
	EndFlow bool         `json:"end_flow"` // true = clear conversation state
}

// BotButton represents an interactive button option.
type BotButton struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}
