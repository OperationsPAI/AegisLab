package dto

// SSEMessageData represents data prepared for Server-Sent Events
type SSEMessageData struct {
	ID          string // Message ID
	Data        any    // Message data
	IsCompleted bool   // Whether this is a completion message
}
