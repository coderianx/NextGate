package ai

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/coderianx/NextGate/internal/models"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

const chatCompletionObject = "chat.completion"

// Chat sends a non-streaming chat completion request to an OpenAI-compatible
// provider and returns a normalized OpenAI-compatible response.
func Chat(
	ctx context.Context,
	apiKey string,
	req models.ChatCompletionRequest,
	baseURL string,
) (*models.ChatCompletionResponse, error) {
	client := openai.NewClient(
		option.WithAPIKey(apiKey),
		option.WithBaseURL(baseURL),
	)

	messages := make([]openai.ChatCompletionMessageParamUnion, 0, len(req.Messages))

	for _, msg := range req.Messages {
		switch msg.Role {
		case "user":
			messages = append(messages, openai.UserMessage(msg.Content))
		case "assistant":
			messages = append(messages, openai.AssistantMessage(msg.Content))
		case "system", "developer":
			messages = append(messages, openai.SystemMessage(msg.Content))
		}
	}

	params := openai.ChatCompletionNewParams{
		Model:    req.Model,
		Messages: messages,
	}

	if req.Temperature != nil {
		params.Temperature = openai.Float(*req.Temperature)
	}

	if req.MaxTokens != nil {
		params.MaxTokens = openai.Int(int64(*req.MaxTokens))
	}

	response, err := client.Chat.Completions.New(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("ai: chat request failed: %w", err)
	}

	if len(response.Choices) == 0 {
		return nil, fmt.Errorf("ai: returned no response choices")
	}

	choices := make([]models.ChatCompletionChoice, 0, len(response.Choices))

	for _, choice := range response.Choices {
		finishReason := choice.FinishReason
		if finishReason == "" {
			finishReason = "stop"
		}

		choices = append(choices, models.ChatCompletionChoice{
			Index: choice.Index,
			Message: models.AssistantMessage{
				Role:    "assistant",
				Content: choice.Message.Content,
			},
			FinishReason: finishReason,
		})
	}

	id := response.ID
	if id == "" {
		id = newCompletionID()
	}

	created := response.Created
	if created == 0 {
		created = time.Now().Unix()
	}

	model := response.Model
	if model == "" {
		model = req.Model
	}

	return &models.ChatCompletionResponse{
		ID:      id,
		Object:  chatCompletionObject,
		Created: created,
		Model:   model,
		Choices: choices,
		Usage: models.ChatCompletionUsage{
			PromptTokens:     response.Usage.PromptTokens,
			CompletionTokens: response.Usage.CompletionTokens,
			TotalTokens:      response.Usage.TotalTokens,
		},
	}, nil
}

// newCompletionID returns a fallback completion identifier for providers that
// do not include one in their response.
func newCompletionID() string {
	random := make([]byte, 12)

	if _, err := rand.Read(random); err != nil {
		return fmt.Sprintf("chatcmpl-%d", time.Now().UnixNano())
	}

	return "chatcmpl-" + hex.EncodeToString(random)
}
