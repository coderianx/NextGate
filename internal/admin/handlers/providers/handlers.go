// Package providers provides authenticated HTTP handlers for managing AI
// providers.
package providers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/coderianx/NextGate/internal/config"
	"github.com/coderianx/NextGate/internal/models"
	"github.com/coderianx/NextGate/internal/secret"
	"github.com/coderianx/NextGate/internal/store"
	"github.com/go-chi/chi/v5"
)

var (
	// GroqBaseURL is the default base URL used when adding a provider named
	// "groq".
	GroqBaseURL string = "https://api.groq.com/openai/v1"
)

// AddNewProvider validates a provider request, encrypts its API key with the
// keyring-backed AES-256 key, and stores the provider in the database. It
// responds with 201 on success and reports validation or persistence failures
// as JSON errors.
func AddNewProvider(
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

	var req models.AddNewProvider

	err := json.NewDecoder(r.Body).Decode(&req)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "invalid_request",
		})

		return
	}

	if req.Name == "" {
		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "name_is_required",
		})

		return
	}

	switch req.Name {
	case "groq":
		req.BaseURL = GroqBaseURL
	}

	if req.BaseURL == "" {
		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "base_url_is_required",
		})

		return
	}

	if req.APIKey == "" {
		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "api_key_is_required",
		})

		return
	}

	key, err := config.GetAESKey()

	if err != nil {
		log.Println("[ADMIN] ", err)

		w.WriteHeader(500)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "aes_256_gcm_get_secret_key",
		})

		return
	}

	encryptAPIKey, err := secret.Encrypt(key, req.APIKey)

	if err != nil {
		log.Println("[ADMIN] ", err)

		w.WriteHeader(500)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "encrypt_api_key_error",
		})

		return
	}

	_, err = store.DB.Exec(
		`
		INSERT INTO providers (name,base_url,api_key)
		VALUES ($1,$2,$3)
		`,
		req.Name,
		req.BaseURL,
		encryptAPIKey,
	)

	if err != nil {
		w.WriteHeader(500)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "database_insert_error",
		})

		return
	}

	w.WriteHeader(201)

	json.NewEncoder(w).Encode(map[string]any{
		"message": "new_provider_added_successfully",
	})
}

// ListMyProviders returns all configured providers without exposing their
// encrypted API keys.
func ListMyProviders(
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

	rows, err := store.DB.Query(
		`
		SELECT id,name,base_url,created_at
		FROM providers
		`,
	)

	if err != nil {
		w.WriteHeader(500)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "database_query_error",
		})

		return
	}

	defer rows.Close()

	var providers []models.Providers

	for rows.Next() {
		var provider models.Providers

		err := rows.Scan(
			&provider.ID,
			&provider.Name,
			&provider.BaseURL,
			&provider.CreatedAt,
		)

		if err != nil {
			w.WriteHeader(500)

			json.NewEncoder(w).Encode(map[string]any{
				"error": "scan_providers_error",
			})

			return
		}

		providers = append(providers, provider)
	}

	if err = rows.Err(); err != nil {
		w.WriteHeader(500)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "scan_providers_error",
		})

		return
	}

	w.WriteHeader(200)

	json.NewEncoder(w).Encode(map[string]any{
		"providers": providers,
	})
}

// DeleteProvider deletes the provider identified by the {id} URL parameter. It
// returns 400 for an invalid id and 404 when no matching provider exists.
func DeleteProvider(
	w http.ResponseWriter,
	r *http.Request,
) {
	w.Header().Set("Content-Type", "application/json")

	if store.DB == nil {
		w.WriteHeader(http.StatusInternalServerError)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "database_not_connected",
		})

		return
	}

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || id <= 0 {
		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "invalid_provider_id",
		})

		return
	}

	result, err := store.DB.Exec(
		`
		DELETE FROM providers
		WHERE id = $1
		`,
		id,
	)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "database_delete_error",
		})

		return
	}

	deletedRows, err := result.RowsAffected()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "database_delete_error",
		})

		return
	}

	if deletedRows == 0 {
		w.WriteHeader(http.StatusNotFound)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "provider_not_found",
		})

		return
	}

	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(map[string]any{
		"message": "provider_deleted_successfully",
	})
}
