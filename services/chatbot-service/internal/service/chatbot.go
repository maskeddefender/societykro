package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/societykro/go-common/events"
	"github.com/societykro/go-common/logger"

	"github.com/societykro/chatbot-service/internal/model"
)

const stateKeyPrefix = "bot:state:"
const stateTTL = 10 * time.Minute

// ChatbotService processes incoming messages and generates bot responses.
type ChatbotService struct {
	rdb *redis.Client
	bus *events.Bus
}

// NewChatbotService creates a new ChatbotService.
func NewChatbotService(rdb *redis.Client, bus *events.Bus) *ChatbotService {
	return &ChatbotService{rdb: rdb, bus: bus}
}

// ProcessMessage is the main entry point. Takes a normalized message,
// detects intent, manages conversation state, and returns a response.
func (s *ChatbotService) ProcessMessage(ctx context.Context, senderID, channel, text, userID, societyID, language string) (*model.BotResponse, error) {
	// Load conversation state from Redis
	state := s.loadState(ctx, senderID)
	if state == nil {
		state = &model.ConversationState{
			SenderID:  senderID,
			Channel:   channel,
			UserID:    userID,
			SocietyID: societyID,
			Language:  language,
			Data:      map[string]string{},
		}
	}

	// If user is in an active flow, handle the step
	if state.CurrentFlow != "" {
		resp := s.handleFlowStep(ctx, state, text)
		s.saveState(ctx, state)
		return resp, nil
	}

	// Detect intent from text
	intent := DetectIntent(text, language)
	logger.Log.Debug().
		Str("sender", senderID).
		Str("intent", intent.Name).
		Float64("confidence", intent.Confidence).
		Str("layer", intent.Layer).
		Msg("Intent detected")

	// Route based on intent
	switch intent.Name {
	case "raise_complaint":
		return s.startComplaintFlow(ctx, state, intent)
	case "check_status":
		return s.handleCheckStatus(ctx, state, intent)
	case "pre_approve_visitor":
		return s.handlePreApproveVisitor(ctx, state)
	case "check_dues":
		return s.handleCheckDues(ctx, state)
	case "sos":
		return s.handleSOS(ctx, state)
	case "approve_action":
		return &model.BotResponse{Text: "Approved! Guard has been notified.", EndFlow: true}, nil
	case "deny_action":
		return &model.BotResponse{Text: "Denied. Visitor has been turned away.", EndFlow: true}, nil
	case "show_menu":
		return s.showMenu(), nil
	default:
		return s.showMenu(), nil
	}
}

func (s *ChatbotService) startComplaintFlow(ctx context.Context, state *model.ConversationState, intent model.Intent) (*model.BotResponse, error) {
	if category, ok := intent.Entities["category"]; ok {
		state.Data["category"] = category
	}

	// If we already have enough detail from the message, create directly
	state.CurrentFlow = "complaint"
	state.Step = 1
	s.saveState(ctx, state)

	if state.Data["category"] != "" {
		return &model.BotResponse{
			Text: fmt.Sprintf("Complaint category: %s\nPlease describe the issue in detail (or send a voice message):", state.Data["category"]),
		}, nil
	}

	return &model.BotResponse{
		Text: "What type of issue?\n\n1. Water\n2. Electrical\n3. Lift\n4. Plumbing\n5. Security\n6. Parking\n7. Garbage\n8. Other",
		Buttons: []model.BotButton{
			{ID: "water", Title: "Water"},
			{ID: "electrical", Title: "Electrical"},
			{ID: "lift", Title: "Lift"},
		},
	}, nil
}

func (s *ChatbotService) handleFlowStep(ctx context.Context, state *model.ConversationState, text string) *model.BotResponse {
	switch state.CurrentFlow {
	case "complaint":
		return s.complaintFlowStep(ctx, state, text)
	default:
		state.CurrentFlow = ""
		return s.showMenu()
	}
}

func (s *ChatbotService) complaintFlowStep(ctx context.Context, state *model.ConversationState, text string) *model.BotResponse {
	switch state.Step {
	case 1:
		// Category selection or description
		if state.Data["category"] == "" {
			categories := map[string]string{"1": "water", "2": "electrical", "3": "lift", "4": "plumbing", "5": "security", "6": "parking", "7": "garbage", "8": "other"}
			if cat, ok := categories[text]; ok {
				state.Data["category"] = cat
			} else {
				state.Data["category"] = text
			}
			state.Step = 2
			return &model.BotResponse{Text: "Please describe the issue:"}
		}
		// Already have category, this is the description
		state.Data["description"] = text
		state.Step = 3
		return &model.BotResponse{
			Text: fmt.Sprintf("Complaint Summary:\nCategory: %s\nDescription: %s\n\nSubmit? (yes/no)", state.Data["category"], text),
		}

	case 2:
		state.Data["description"] = text
		state.Step = 3
		return &model.BotResponse{
			Text: fmt.Sprintf("Complaint Summary:\nCategory: %s\nDescription: %s\n\nSubmit? (yes/no)", state.Data["category"], text),
		}

	case 3:
		if text == "yes" || text == "1" || text == "haan" || text == "ha" {
			// Publish complaint creation event
			s.bus.Publish("chatbot.action", "chatbot.create_complaint", map[string]interface{}{
				"user_id":     state.UserID,
				"society_id":  state.SocietyID,
				"category":    state.Data["category"],
				"description": state.Data["description"],
				"source":      state.Channel,
			})

			state.CurrentFlow = ""
			state.Step = 0
			state.Data = map[string]string{}
			return &model.BotResponse{
				Text:    "Complaint submitted! You'll get a ticket number shortly. Admin has been notified.",
				EndFlow: true,
			}
		}
		state.CurrentFlow = ""
		state.Step = 0
		state.Data = map[string]string{}
		return &model.BotResponse{Text: "Complaint cancelled.", EndFlow: true}
	}

	return s.showMenu()
}

func (s *ChatbotService) handleCheckStatus(_ context.Context, state *model.ConversationState, intent model.Intent) (*model.BotResponse, error) {
	// TODO: Query complaint-service via gRPC to get actual status
	if ticket, ok := intent.Entities["ticket_number"]; ok {
		return &model.BotResponse{
			Text: fmt.Sprintf("Ticket %s: Checking status...\n(Status check via internal API coming in Phase 2)", ticket),
		}, nil
	}

	return &model.BotResponse{
		Text: "Your recent complaints:\n(Integration with complaint-service coming in Phase 2)\n\nSend a ticket number like #1234 to check specific status.",
	}, nil
}

func (s *ChatbotService) handlePreApproveVisitor(_ context.Context, state *model.ConversationState) (*model.BotResponse, error) {
	// TODO: Call visitor-service to generate OTP
	return &model.BotResponse{
		Text: "Pre-approve a visitor:\nOTP: 847291 (valid 4 hours)\nShare this with your guest.\n\n(Real OTP generation coming in Phase 2)",
	}, nil
}

func (s *ChatbotService) handleCheckDues(_ context.Context, state *model.ConversationState) (*model.BotResponse, error) {
	// TODO: Call payment-service to get pending dues
	return &model.BotResponse{
		Text: "Your pending dues:\n(Payment integration coming in Phase 2)\n\nPay via: https://pay.societykro.in/...",
	}, nil
}

func (s *ChatbotService) handleSOS(ctx context.Context, state *model.ConversationState) (*model.BotResponse, error) {
	s.bus.Publish(events.SubjectSOSTriggered, "sos.triggered", map[string]interface{}{
		"user_id":    state.UserID,
		"society_id": state.SocietyID,
		"alert_type": "general",
		"source":     state.Channel,
	})

	return &model.BotResponse{
		Text:    "SOS ALERT SENT! Security guard and admin have been notified. Stay safe.\n\nEmergency: 112 | Ambulance: 108 | Fire: 101",
		EndFlow: true,
	}, nil
}

func (s *ChatbotService) showMenu() *model.BotResponse {
	return &model.BotResponse{
		Text: "How can I help?\n\n1. Raise Complaint\n2. Check Complaint Status\n3. Pre-approve Visitor\n4. Check Payment Dues\n5. SOS Emergency\n\nReply with number or describe your issue.",
		Buttons: []model.BotButton{
			{ID: "raise_complaint", Title: "Raise Complaint"},
			{ID: "check_status", Title: "Check Status"},
			{ID: "pre_approve_visitor", Title: "Visitor OTP"},
		},
	}
}

// --- State Management (Redis) ---

func stateKey(senderID string) string {
	return stateKeyPrefix + senderID
}

func (s *ChatbotService) loadState(ctx context.Context, senderID string) *model.ConversationState {
	data, err := s.rdb.Get(ctx, stateKey(senderID)).Bytes()
	if err != nil {
		return nil
	}
	var state model.ConversationState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil
	}
	return &state
}

func (s *ChatbotService) saveState(ctx context.Context, state *model.ConversationState) {
	data, err := json.Marshal(state)
	if err != nil {
		return
	}
	s.rdb.Set(ctx, stateKey(state.SenderID), data, stateTTL)
}
