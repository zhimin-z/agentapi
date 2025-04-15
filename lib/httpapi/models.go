package httpapi

type MessageType string

const (
	MessageTypeUser MessageType = "user"
	MessageTypeRaw  MessageType = "raw"
)

// Message represents a message
type Message struct {
	Content string `json:"content" example:"Hello world" doc:"Message content"`
	Role    string `json:"role" example:"user" doc:"Role of the message author"`
}

// StatusResponse represents the server status
type StatusResponse struct {
	Body struct {
		Status string `json:"status" example:"running" doc:"Current server status"`
	}
}

// MessagesResponse represents the list of messages
type MessagesResponse struct {
	Body struct {
		Messages []Message `json:"messages" doc:"List of messages"`
	}
}

type MessageRequestBody struct {
	Content string      `json:"content" example:"Hello, server!" doc:"Message content"`
	Type    MessageType `json:"type" enum:"user,raw" example:"user" doc:"Type of the message"`
}

// MessageRequest represents a request to create a new message
type MessageRequest struct {
	Body MessageRequestBody `json:"body"`
}

// MessageResponse represents a newly created message
type MessageResponse struct {
	Body struct {
		Ok bool `json:"ok" doc:"Whether the message was sent successfully"`
	}
}
