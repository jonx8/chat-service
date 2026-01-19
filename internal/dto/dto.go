package dto

type CreateChatRequest struct {
	Title string `json:"title"`
}

type CreateMessageRequest struct {
	Text string `json:"text"`
}
