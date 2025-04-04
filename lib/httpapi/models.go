package httpapi

// Message represents a message
type Message struct {
	ID      string `json:"id" example:"msg123" doc:"Unique message identifier"`
	Content string `json:"content" example:"Hello world" doc:"Message content"`
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

// MessageRequest represents a request to create a new message
type MessageRequest struct {
	Body struct {
		Content string `json:"content" maxLength:"1000" example:"Hello, server!" doc:"Message content"`
	}
}

// MessageResponse represents a newly created message
type MessageResponse struct {
	Body struct {
		Message Message `json:"message" doc:"The created message"`
	}
}
