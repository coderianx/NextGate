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

const (
	chatCompletionObject      = "chat.completion"
	chatCompletionChunkObject = "chat.completion.chunk"
)

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

	response, err := client.Chat.Completions.New(ctx, newChatCompletionParams(req))
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
		Usage:   mapUsage(response.Usage),
	}, nil
}

// ChatStream sends a streaming chat completion request to an OpenAI-compatible
// provider and calls yield for every normalized chunk. It stops and returns the
// first error from the provider, the stream, or yield.
func ChatStream(
	ctx context.Context,
	apiKey string,
	req models.ChatCompletionRequest,
	baseURL string,
	yield func(models.ChatCompletionChunk) error,
) error {
	if yield == nil {
		return fmt.Errorf("ai: stream yield function is nil")
	}

	client := openai.NewClient(
		option.WithAPIKey(apiKey),
		option.WithBaseURL(baseURL),
	)

	params := newChatCompletionParams(req)

	if req.StreamOptions != nil {
		params.StreamOptions = openai.ChatCompletionStreamOptionsParam{
			IncludeUsage: openai.Bool(req.StreamOptions.IncludeUsage),
		}
	}

	stream := client.Chat.Completions.NewStreaming(ctx, params)
	defer stream.Close()

	var (
		fallbackID      string
		fallbackCreated int64
		sentRole        = make(map[int64]bool)
	)

	for stream.Next() {
		chunk := stream.Current()

		id := chunk.ID
		if id == "" {
			if fallbackID == "" {
				fallbackID = newCompletionID()
			}

			id = fallbackID
		}

		created := chunk.Created
		if created == 0 {
			if fallbackCreated == 0 {
				fallbackCreated = time.Now().Unix()
			}

			created = fallbackCreated
		}

		model := chunk.Model
		if model == "" {
			model = req.Model
		}

		choices := make([]models.ChatCompletionChunkChoice, 0, len(chunk.Choices))

		for _, choice := range chunk.Choices {
			delta := models.ChatCompletionChunkDelta{
				Role:    choice.Delta.Role,
				Content: choice.Delta.Content,
			}

			if !sentRole[choice.Index] {
				if delta.Role == "" {
					delta.Role = "assistant"
				}

				sentRole[choice.Index] = true
			}

			var finishReason *string

			if choice.FinishReason != "" {
				reason := choice.FinishReason
				finishReason = &reason
			}

			choices = append(choices, models.ChatCompletionChunkChoice{
				Index:        choice.Index,
				Delta:        delta,
				FinishReason: finishReason,
			})
		}

		var usage *models.ChatCompletionUsage

		if chunk.JSON.Usage.Valid() {
			mappedUsage := mapUsage(chunk.Usage)
			usage = &mappedUsage
		}

		if err := yield(models.ChatCompletionChunk{
			ID:      id,
			Object:  chatCompletionChunkObject,
			Created: created,
			Model:   model,
			Choices: choices,
			Usage:   usage,
		}); err != nil {
			return err
		}
	}

	return stream.Err()
}

// newChatCompletionParams converts a NextGate request into the provider SDK
// parameters shared by streaming and non-streaming requests.
func newChatCompletionParams(req models.ChatCompletionRequest) openai.ChatCompletionNewParams {
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

	return params
}

// mapUsage converts provider token usage into the public response model.
func mapUsage(usage openai.CompletionUsage) models.ChatCompletionUsage {
	return models.ChatCompletionUsage{
		PromptTokens:     usage.PromptTokens,
		CompletionTokens: usage.CompletionTokens,
		TotalTokens:      usage.TotalTokens,
	}
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
