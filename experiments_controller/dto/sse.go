package dto

// SSEMessageData represents data prepared for Server-Sent Events
type SSEMessageData struct {
	ID          string      // Message ID
	Data        interface{} // Message data
	IsCompleted bool        // Whether this is a completion message
}
