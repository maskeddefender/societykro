package service

import "github.com/societykro/notification-service/internal/model"

// Templates maps event types to notification templates.
// In production, these would be loaded from DB and support i18n.
var Templates = map[string]model.NotificationTemplate{
	"complaint.created": {
		EventType: "complaint.created",
		TitleTpl:  "New Complaint: {{.Title}}",
		BodyTpl:   "{{.RaisedByName}} raised a {{.Category}} complaint: {{.Title}}",
		Channels:  []string{"push", "whatsapp"},
		Priority:  "normal",
	},
	"complaint.assigned": {
		EventType: "complaint.assigned",
		TitleTpl:  "Complaint Assigned",
		BodyTpl:   "Complaint #{{.TicketNumber}} has been assigned to a vendor",
		Channels:  []string{"push"},
		Priority:  "normal",
	},
	"complaint.resolved": {
		EventType: "complaint.resolved",
		TitleTpl:  "Complaint Resolved",
		BodyTpl:   "Complaint #{{.TicketNumber}} has been resolved. Please confirm.",
		Channels:  []string{"push", "whatsapp"},
		Priority:  "normal",
	},
	"complaint.emergency": {
		EventType: "complaint.emergency",
		TitleTpl:  "EMERGENCY: {{.Title}}",
		BodyTpl:   "URGENT: {{.RaisedByName}} reported an emergency - {{.Title}}",
		Channels:  []string{"push", "whatsapp", "sms"},
		Priority:  "high",
	},
	"visitor.logged": {
		EventType: "visitor.logged",
		TitleTpl:  "Visitor at Gate",
		BodyTpl:   "{{.VisitorName}} is at the gate for your flat. Approve or deny.",
		Channels:  []string{"push", "whatsapp"},
		Priority:  "high",
	},
	"visitor.approved": {
		EventType: "visitor.approved",
		TitleTpl:  "Visitor Approved",
		BodyTpl:   "{{.VisitorName}} has been approved to enter.",
		Channels:  []string{"push"},
		Priority:  "normal",
	},
	"payment.generated": {
		EventType: "payment.generated",
		TitleTpl:  "Maintenance Bill Generated",
		BodyTpl:   "Your maintenance bill of Rs {{.Amount}} for {{.Month}} is ready. Due: {{.DueDate}}",
		Channels:  []string{"push", "whatsapp"},
		Priority:  "normal",
	},
	"payment.received": {
		EventType: "payment.received",
		TitleTpl:  "Payment Received",
		BodyTpl:   "Payment of Rs {{.Amount}} received. Receipt: {{.ReceiptURL}}",
		Channels:  []string{"push", "whatsapp"},
		Priority:  "normal",
	},
	"payment.overdue": {
		EventType: "payment.overdue",
		TitleTpl:  "Payment Overdue",
		BodyTpl:   "Your maintenance payment of Rs {{.Amount}} is overdue. Please pay now to avoid late fees.",
		Channels:  []string{"push", "whatsapp", "sms"},
		Priority:  "high",
	},
	"notice.posted": {
		EventType: "notice.posted",
		TitleTpl:  "New Notice: {{.Title}}",
		BodyTpl:   "{{.CreatedByName}} posted: {{.Title}}",
		Channels:  []string{"push"},
		Priority:  "normal",
	},
	"sos.triggered": {
		EventType: "sos.triggered",
		TitleTpl:  "SOS ALERT",
		BodyTpl:   "EMERGENCY SOS triggered by {{.TriggeredByName}}: {{.AlertType}}",
		Channels:  []string{"push", "whatsapp", "sms"},
		Priority:  "high",
	},
}
