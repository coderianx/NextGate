package openai

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/coderianx/NextGate/internal/ai"
	"github.com/coderianx/NextGate/internal/config"
	"github.com/coderianx/NextGate/internal/middlewares"
	"github.com/coderianx/NextGate/internal/models"
	"github.com/coderianx/NextGate/internal/secret"
	"github.com/coderianx/NextGate/internal/store"
)

func ChatCompletion(
	w http.ResponseWriter,
	r *http.Request,
) {
	w.Header().Set("Content-Type", "application/json")

	if store.DB == nil {
		w.WriteHeader(500)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "database_not_connected",
		})

		return
	}

	if _, ok := middlewares.GetAPIKey(r); !ok {
		w.WriteHeader(http.StatusUnauthorized)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "unauthorized",
		})
		return
	}

	var req models.ChatCompletionRequest

	err := json.NewDecoder(r.Body).Decode(&req)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "invalid_request",
		})

		return
	}

	if req.Model == "" {
		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{
				"message": "Missing required parameter: model",
				"type":    "invalid_request_error",
				"param":   "model",
				"code":    "missing_required_parameter",
			},
		})
		return
	}

	if len(req.Messages) == 0 {
		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{
				"message": "Messages cannot be empty",
				"type":    "invalid_request_error",
				"param":   "messages",
				"code":    "missing_required_parameter",
			},
		})
		return
	}

	for _, msg := range req.Messages {
		if msg.Role == "" {
			w.WriteHeader(http.StatusBadRequest)

			json.NewEncoder(w).Encode(map[string]any{
				"error": map[string]any{
					"message": "Message role is required",
					"type":    "invalid_request_error",
					"param":   "messages",
					"code":    "invalid_value",
				},
			})
			return
		}

		switch msg.Role {
		case "system", "developer", "user", "assistant", "tool":
		default:
			w.WriteHeader(http.StatusBadRequest)

			json.NewEncoder(w).Encode(map[string]any{
				"error": map[string]any{
					"message": "Invalid message role: " + msg.Role,
					"type":    "invalid_request_error",
					"param":   "messages",
					"code":    "invalid_value",
				},
			})
			return
		}

		if msg.Content == "" && msg.Role != "assistant" && msg.Role != "tool" {
			w.WriteHeader(http.StatusBadRequest)

			json.NewEncoder(w).Encode(map[string]any{
				"error": map[string]any{
					"message": "Message content cannot be empty",
					"type":    "invalid_request_error",
					"param":   "messages",
					"code":    "invalid_value",
				},
			})
			return
		}
	}

	if req.Temperature != nil && (*req.Temperature < 0 || *req.Temperature > 2) {
		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{
				"message": "Temperature must be between 0 and 2",
				"type":    "invalid_request_error",
				"param":   "temperature",
				"code":    "invalid_value",
			},
		})
		return
	}

	if req.MaxTokens != nil && *req.MaxTokens <= 0 {
		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{
				"message": "max_tokens must be greater than 0",
				"type":    "invalid_request_error",
				"param":   "max_tokens",
				"code":    "invalid_value",
			},
		})
		return
	}

	var (
		modelProviderID int64
		baseURL         string
		encryptedAPIKey string
	)

	err = store.DB.QueryRow(
		`
		SELECT provider_id
		FROM models
		WHERE model_id = ? AND enabled = ?;
		`,
		req.Model,
		true,
	).Scan(
		&modelProviderID,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(404)

			json.NewEncoder(w).Encode(map[string]any{
				"error": map[string]any{
					"message": fmt.Sprintf("The model '%s' does not exist or you do not have access to it.", req.Model),
					"type":    "invalid_request_error",
					"param":   "model",
					"code":    "model_not_found",
				},
			})

			return
		}

		w.WriteHeader(http.StatusInternalServerError)

		json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{
				"message": "An internal server error occurred.",
				"type":    "server_error",
				"param":   nil,
				"code":    "internal_error",
			},
		})

		return
	}

	err = store.DB.QueryRow(
		`
		SELECT base_url,api_key
		FROM providers
		WHERE id = ?;
		`,
		modelProviderID,
	).Scan(
		&baseURL,
		&encryptedAPIKey,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusInternalServerError)

			json.NewEncoder(w).Encode(map[string]any{
				"error": map[string]any{
					"message": "The provider associated with this model could not be found.",
					"type":    "server_error",
					"param":   nil,
					"code":    "provider_not_found",
				},
			})

			return
		}

		log.Printf("database error while fetching provider: %v", err)

		w.WriteHeader(http.StatusInternalServerError)

		json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{
				"message": "An internal server error occurred.",
				"type":    "server_error",
				"param":   nil,
				"code":    "internal_error",
			},
		})

		return
	}

	aesKey, err := config.GetAESKey()
	if err != nil {
		log.Printf("get AES key: %v", err)

		w.WriteHeader(http.StatusInternalServerError)

		json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{
				"message": "An internal server error occurred.",
				"type":    "server_error",
				"param":   nil,
				"code":    "internal_error",
			},
		})

		return
	}

	providerAPIKey, err := secret.Decrypt(aesKey, encryptedAPIKey)
	if err != nil {
		log.Printf("decrypt provider API key: %v", err)

		w.WriteHeader(http.StatusInternalServerError)

		json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{
				"message": "An internal server error occurred.",
				"type":    "server_error",
				"param":   nil,
				"code":    "internal_error",
			},
		})

		return
	}

	if req.Stream {
		streamChatCompletion(w, r, providerAPIKey, req, baseURL)
		return
	}

	resp, err := ai.Chat(
		r.Context(),
		providerAPIKey,
		req,
		baseURL,
	)
	if err != nil {
		log.Printf("chat completion request failed: %v", err)

		w.WriteHeader(http.StatusBadGateway)

		json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{
				"message": "The upstream provider request failed.",
				"type":    "server_error",
				"param":   nil,
				"code":    "upstream_error",
			},
		})

		return
	}

	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("encode chat completion response: %v", err)
	}
}

// streamChatCompletion writes an OpenAI-compatible server-sent event stream.
// The server write deadline is cleared for the response so long-lived streams
// are not cut off by the API server's default timeout.
func streamChatCompletion(
	w http.ResponseWriter,
	r *http.Request,
	apiKey string,
	req models.ChatCompletionRequest,
	baseURL string,
) {
	controller := http.NewResponseController(w)

	if err := controller.SetWriteDeadline(time.Time{}); err != nil &&
		!errors.Is(err, http.ErrNotSupported) {
		log.Printf("disable write deadline for stream: %v", err)
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)

	if err := controller.Flush(); err != nil {
		log.Printf("flush stream headers: %v", err)
		return
	}

	err := ai.ChatStream(
		r.Context(),
		apiKey,
		req,
		baseURL,
		func(chunk models.ChatCompletionChunk) error {
			payload, err := json.Marshal(chunk)
			if err != nil {
				return err
			}

			if _, err := fmt.Fprintf(w, "data: %s\n\n", payload); err != nil {
				return err
			}

			return controller.Flush()
		},
	)
	if err != nil {
		if r.Context().Err() != nil ||
			errors.Is(err, context.Canceled) ||
			errors.Is(err, context.DeadlineExceeded) {
			return
		}

		log.Printf("chat completion stream failed: %v", err)

		payload, marshalErr := json.Marshal(map[string]any{
			"error": map[string]any{
				"message": "The upstream provider request failed.",
				"type":    "server_error",
				"param":   nil,
				"code":    "upstream_error",
			},
		})
		if marshalErr == nil {
			_, _ = fmt.Fprintf(w, "data: %s\n\n", payload)
			_ = controller.Flush()
		}

		return
	}

	_, _ = fmt.Fprint(w, "data: [DONE]\n\n")
	_ = controller.Flush()
}
