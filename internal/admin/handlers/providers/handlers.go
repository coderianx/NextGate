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

func AddNewModel(
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

	providerID, err := strconv.ParseInt(
		chi.URLParam(r, "id"), 10, 64,
	)

	if err != nil && providerID == 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"error": "invalid_provider_id",
		})
		return
	}

	var providerIsExists bool

	err = store.DB.QueryRow(
		`
		SELECT EXISTS(
		    SELECT 1
		    FROM providers
		    WHERE id = ?
		);
		`,
		providerID,
	).Scan(&providerIsExists)

	if err != nil {
		w.WriteHeader(500)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "database_error",
		})

		return
	}

	if !providerIsExists {
		w.WriteHeader(404)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "provider_not_found",
		})

		return
	}

	var req models.AddNewModel

	err = json.NewDecoder(r.Body).Decode(&req)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "invalid_request",
		})

		return
	}

	if req.ModelID == "" {
		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "model_id_is_required",
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

	_, err = store.DB.Exec(
		`
		INSERT INTO models (provider_id,model_id,name)
		VALUES (?,?,?)
		`,
		providerID,
		req.ModelID,
		req.Name,
	)

	if err != nil {
		if store.IsUniqueError(err) {
			w.WriteHeader(409)

			json.NewEncoder(w).Encode(map[string]any{
				"error": "you_are_already_exists_have_this_model",
			})

			return
		}

		w.WriteHeader(500)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "database_error",
		})

		return
	}

	w.WriteHeader(201)

	json.NewEncoder(w).Encode(map[string]any{
		"message": "new_model_added",
	})
}

// ListModels returns all models configured for the provider identified by the
// {id} URL parameter.
func ListModels(
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

	providerID, err := strconv.ParseInt(
		chi.URLParam(r, "id"), 10, 64,
	)

	if err != nil || providerID <= 0 {
		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "invalid_provider_id",
		})

		return
	}

	var providerExists bool

	err = store.DB.QueryRow(
		`
		SELECT EXISTS(
			SELECT 1
			FROM providers
			WHERE id = ?
		)
		`,
		providerID,
	).Scan(&providerExists)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "database_error",
		})

		return
	}

	if !providerExists {
		w.WriteHeader(http.StatusNotFound)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "provider_not_found",
		})

		return
	}

	rows, err := store.DB.Query(
		`
		SELECT provider_id, model_id, name, enabled
		FROM models
		WHERE provider_id = ?
		ORDER BY id
		`,
		providerID,
	)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "database_query_error",
		})

		return
	}

	defer rows.Close()

	providerModels := make([]models.Models, 0)

	for rows.Next() {
		var providerModel models.Models

		err := rows.Scan(
			&providerModel.ProviderID,
			&providerModel.ModelID,
			&providerModel.Name,
			&providerModel.Enabled,
		)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)

			json.NewEncoder(w).Encode(map[string]any{
				"error": "scan_models_error",
			})

			return
		}

		providerModels = append(providerModels, providerModel)
	}

	if err := rows.Err(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "scan_models_error",
		})

		return
	}

	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(map[string]any{
		"models": providerModels,
	})
}

// SetModelEnabled updates whether the model identified by the {model_id} URL
// parameter is enabled for the provider identified by the {id} URL parameter.
func SetModelEnabled(
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

	providerID, err := strconv.ParseInt(
		chi.URLParam(r, "id"), 10, 64,
	)

	if err != nil || providerID <= 0 {
		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "invalid_provider_id",
		})

		return
	}

	modelID := chi.URLParam(r, "model_id")

	if modelID == "" {
		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "model_id_is_required",
		})

		return
	}

	var req models.SetModelEnabled

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "invalid_request",
		})

		return
	}

	result, err := store.DB.Exec(
		`
		UPDATE models
		SET enabled = ?
		WHERE provider_id = ? AND model_id = ?
		`,
		req.Enabled,
		providerID,
		modelID,
	)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "database_update_error",
		})

		return
	}

	updatedRows, err := result.RowsAffected()

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "database_update_error",
		})

		return
	}

	if updatedRows == 0 {
		w.WriteHeader(http.StatusNotFound)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "model_not_found",
		})

		return
	}

	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(map[string]any{
		"message": "model_enabled_updated",
		"enabled": req.Enabled,
	})
}

// DeleteModel deletes the model identified by the {model_id} URL parameter
// from the provider identified by the {id} URL parameter.
func DeleteModel(
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

	providerID, err := strconv.ParseInt(
		chi.URLParam(r, "id"), 10, 64,
	)

	if err != nil || providerID <= 0 {
		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "invalid_provider_id",
		})

		return
	}

	modelID := chi.URLParam(r, "model_id")

	if modelID == "" {
		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "model_id_is_required",
		})

		return
	}

	result, err := store.DB.Exec(
		`
		DELETE FROM models
		WHERE provider_id = ? AND model_id = ?
		`,
		providerID,
		modelID,
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
			"error": "model_not_found",
		})

		return
	}

	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(map[string]any{
		"message": "model_deleted_successfully",
	})
}
