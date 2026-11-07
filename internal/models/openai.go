package models

type ChatCompletionRequest struct {
	Model         string                       `json:"model"`
	Messages      []Message                    `json:"messages"`
	Temperature   *float64                     `json:"temperature,omitempty"`
	MaxTokens     *int                         `json:"max_tokens,omitempty"`
	Stream        bool                         `json:"stream,omitempty"`
	StreamOptions *ChatCompletionStreamOptions `json:"stream_options,omitempty"`
}

// ChatCompletionStreamOptions controls optional streaming response behavior.
type ChatCompletionStreamOptions struct {
	// IncludeUsage asks the provider to send a final chunk containing token
	// usage statistics.
	IncludeUsage bool `json:"include_usage,omitempty"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatCompletionResponse is an OpenAI-compatible chat completion response.
type ChatCompletionResponse struct {
	// ID is the unique identifier for the completion.
	ID string `json:"id"`

	// Object is the response object type, always "chat.completion".
	Object string `json:"object"`

	// Created is the Unix timestamp in seconds when the completion was created.
	Created int64 `json:"created"`

	// Model is the model that generated the completion.
	Model string `json:"model"`

	// Choices contains the generated completion choices.
	Choices []ChatCompletionChoice `json:"choices"`

	// Usage reports token usage for the request.
	Usage ChatCompletionUsage `json:"usage"`
}

// ChatCompletionChoice is one generated completion choice.
type ChatCompletionChoice struct {
	// Index is the choice's position in the response.
	Index int64 `json:"index"`

	// Message is the assistant message generated for this choice.
	Message AssistantMessage `json:"message"`

	// FinishReason explains why generation stopped.
	FinishReason string `json:"finish_reason"`
}

// AssistantMessage is the assistant message included in a chat completion
// choice.
type AssistantMessage struct {
	// Role is always "assistant".
	Role string `json:"role"`

	// Content is the generated message text.
	Content string `json:"content"`
}

// ChatCompletionUsage reports token counts for a chat completion.
type ChatCompletionUsage struct {
	// PromptTokens is the number of tokens in the prompt.
	PromptTokens int64 `json:"prompt_tokens"`

	// CompletionTokens is the number of tokens in the generated completion.
	CompletionTokens int64 `json:"completion_tokens"`

	// TotalTokens is the sum of prompt and completion tokens.
	TotalTokens int64 `json:"total_tokens"`
}

// ChatCompletionChunk is one OpenAI-compatible streamed chat completion chunk.
type ChatCompletionChunk struct {
	// ID is the unique identifier shared by every chunk in the stream.
	ID string `json:"id"`

	// Object is the chunk object type, always "chat.completion.chunk".
	Object string `json:"object"`

	// Created is the Unix timestamp in seconds when the completion was created.
	Created int64 `json:"created"`

	// Model is the model that generated the completion.
	Model string `json:"model"`

	// Choices contains the incremental choices in this chunk.
	Choices []ChatCompletionChunkChoice `json:"choices"`

	// Usage is included only when the request asks for usage in the final
	// chunk.
	Usage *ChatCompletionUsage `json:"usage,omitempty"`
}

// ChatCompletionChunkChoice is one incremental choice in a streamed chunk.
type ChatCompletionChunkChoice struct {
	// Index is the choice's position in the stream.
	Index int64 `json:"index"`

	// Delta contains the incremental content for this choice.
	Delta ChatCompletionChunkDelta `json:"delta"`

	// FinishReason is null until the choice finishes, then holds the reason.
	FinishReason *string `json:"finish_reason"`
}

// ChatCompletionChunkDelta is the incremental message content in a streamed
// chunk.
type ChatCompletionChunkDelta struct {
	// Role is set on the first chunk of a choice.
	Role string `json:"role,omitempty"`

	// Content is the incremental text generated for this chunk.
	Content string `json:"content,omitempty"`
}
