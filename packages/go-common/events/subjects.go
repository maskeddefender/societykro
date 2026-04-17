package events

// Stream names
const (
	StreamSocietyKro = "SOCIETYKRO"
)

// All event subjects used across services.
const (
	// Complaint events
	SubjectComplaintCreated   = "complaint.created"
	SubjectComplaintAssigned  = "complaint.assigned"
	SubjectComplaintResolved  = "complaint.resolved"
	SubjectComplaintEscalated = "complaint.escalated"
	SubjectComplaintClosed    = "complaint.closed"

	// Visitor events
	SubjectVisitorLogged   = "visitor.logged"
	SubjectVisitorApproved = "visitor.approved"
	SubjectVisitorDenied   = "visitor.denied"

	// Payment events
	SubjectPaymentGenerated = "payment.generated"
	SubjectPaymentReceived  = "payment.received"
	SubjectPaymentOverdue   = "payment.overdue"

	// Notice events
	SubjectNoticePosted = "notice.posted"

	// SOS events
	SubjectSOSTriggered = "sos.triggered"

	// User events
	SubjectUserCreated      = "user.created"
	SubjectUserJoinedSociety = "user.joined_society"
)

// AllSubjects returns all subjects for stream creation.
func AllSubjects() []string {
	return []string{
		"complaint.*",
		"visitor.*",
		"payment.*",
		"notice.*",
		"sos.*",
		"user.*",
	}
}
