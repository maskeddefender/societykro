package service

import (
	"regexp"
	"strings"

	"github.com/societykro/chatbot-service/internal/model"
)

// DetectIntent runs a 3-layer intent detection pipeline:
// Layer 1: Keyword matching (handles ~60% of queries)
// Layer 2: Regex pattern matching (handles ~20%)
// Layer 3: LLM fallback (handles ~15%, placeholder for now)
// If all fail, returns "unknown" intent for menu display.
func DetectIntent(text, language string) model.Intent {
	lower := strings.ToLower(strings.TrimSpace(text))

	// Layer 1: Keyword matching
	if intent := matchKeywords(lower); intent.Name != "" {
		return intent
	}

	// Layer 2: Pattern matching
	if intent := matchPatterns(lower); intent.Name != "" {
		return intent
	}

	// Layer 3: LLM fallback (placeholder)
	// TODO: Call Gemma 2 9B via vLLM for ambiguous queries

	return model.Intent{Name: "unknown", Confidence: 0.0, Layer: "none"}
}

func matchKeywords(text string) model.Intent {
	// Complaint keywords (Hindi + English)
	complaintWords := []string{
		"problem", "complaint", "issue", "broken", "not working", "kharab",
		"toot", "bigad", "leak", "leaking", "paani nahi", "bijli nahi",
		"lift", "pump", "overflow", "block", "smell", "badbu", "gandagi",
		"parking", "noise", "shor", "cockroach", "pest", "mosquito",
	}
	for _, w := range complaintWords {
		if strings.Contains(text, w) {
			// Try to detect category
			category := detectCategory(text)
			entities := map[string]string{}
			if category != "" {
				entities["category"] = category
			}
			return model.Intent{Name: "raise_complaint", Confidence: 0.85, Entities: entities, Layer: "keyword"}
		}
	}

	// Status check
	statusWords := []string{"status", "update", "kya hua", "kab hoga", "check", "track", "mera complaint"}
	for _, w := range statusWords {
		if strings.Contains(text, w) {
			return model.Intent{Name: "check_status", Confidence: 0.85, Layer: "keyword"}
		}
	}

	// Visitor/guest
	visitorWords := []string{"guest", "visitor", "mehmaan", "aane wala", "aa rahe", "gate", "entry"}
	for _, w := range visitorWords {
		if strings.Contains(text, w) {
			return model.Intent{Name: "pre_approve_visitor", Confidence: 0.80, Layer: "keyword"}
		}
	}

	// Payment
	paymentWords := []string{"pay", "dues", "pending", "baaki", "paisa", "maintenance", "bill", "kitna"}
	for _, w := range paymentWords {
		if strings.Contains(text, w) {
			return model.Intent{Name: "check_dues", Confidence: 0.80, Layer: "keyword"}
		}
	}

	// Emergency / SOS
	sosWords := []string{"sos", "emergency", "help", "bachao", "fire", "aag", "chor", "thief", "ambulance"}
	for _, w := range sosWords {
		if strings.Contains(text, w) {
			return model.Intent{Name: "sos", Confidence: 0.95, Layer: "keyword"}
		}
	}

	// Menu / Help
	menuWords := []string{"menu", "help", "hi", "hello", "namaste", "start", "options"}
	for _, w := range menuWords {
		if text == w || strings.HasPrefix(text, w+" ") {
			return model.Intent{Name: "show_menu", Confidence: 0.90, Layer: "keyword"}
		}
	}

	// Approval responses (inline buttons)
	if text == "1" || text == "approve" || text == "yes" || text == "haan" {
		return model.Intent{Name: "approve_action", Confidence: 0.90, Entities: map[string]string{"action": "approve"}, Layer: "keyword"}
	}
	if text == "2" || text == "deny" || text == "no" || text == "nahi" {
		return model.Intent{Name: "deny_action", Confidence: 0.90, Entities: map[string]string{"action": "deny"}, Layer: "keyword"}
	}

	return model.Intent{}
}

var ticketPattern = regexp.MustCompile(`(?i)(?:#|C-?)(\d{3,6})`)

func matchPatterns(text string) model.Intent {
	// Ticket number pattern
	if matches := ticketPattern.FindStringSubmatch(text); len(matches) > 1 {
		return model.Intent{
			Name:       "check_status",
			Confidence: 0.90,
			Entities:   map[string]string{"ticket_number": "C-" + matches[1]},
			Layer:      "pattern",
		}
	}

	return model.Intent{}
}

func detectCategory(text string) string {
	categories := map[string][]string{
		"water":      {"water", "paani", "tank", "pump", "leaking", "leak", "pipe", "tap", "nalni"},
		"electrical": {"electricity", "bijli", "light", "switch", "wire", "mcb", "short circuit"},
		"lift":       {"lift", "elevator"},
		"plumbing":   {"plumbing", "drain", "nala", "gutter", "toilet", "bathroom", "flush", "basin"},
		"security":   {"security", "guard", "cctv", "camera", "gate", "theft", "chor", "suspicious"},
		"parking":    {"parking", "car", "bike", "vehicle", "gaadi"},
		"garbage":    {"garbage", "kachra", "dustbin", "waste", "gandagi", "smell", "badbu"},
		"pest_control": {"cockroach", "rat", "mosquito", "pest", "termite", "ant", "makhi", "keeda"},
		"noise":      {"noise", "shor", "loud", "music", "party", "construction"},
		"generator":  {"generator", "inverter", "power", "backup", "dg set"},
		"intercom":   {"intercom", "bell", "doorbell"},
	}

	for category, keywords := range categories {
		for _, kw := range keywords {
			if strings.Contains(text, kw) {
				return category
			}
		}
	}
	return "other"
}
