package apikeys

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/coderianx/NextGate/internal/middlewares"
	"github.com/coderianx/NextGate/internal/models"
	"github.com/coderianx/NextGate/internal/store"
	"github.com/go-chi/chi/v5"
)

func CreateNewAPIKey(
	w http.ResponseWriter,
	r *http.Request,
) {
	w.Header().Set("Content-Type", "application/json")

	if store.DB == nil {
		w.WriteHeader(500)

		json.NewEncoder(w).Encode(map[any]string{
			"error": "database_not_connected",
		})

		return
	}

	claims, ok := middlewares.GetClaims(r)

	if !ok {
		w.WriteHeader(http.StatusUnauthorized)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "unauthorized",
		})

		return
	}

	var req models.CreateNewAPIKey

	err := json.NewDecoder(r.Body).Decode(&req)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "invalid_request",
		})

		return
	}

	if req.Title == "" {
		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "title_is_required",
		})

		return
	}

	apiKey, err := generateAPIKey()

	if err != nil {
		w.WriteHeader(500)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "generate_api_key_error",
		})

		return
	}

	apiKeyHash := sha256.Sum256([]byte(apiKey))

	_, err = store.DB.Exec(
		`
		INSERT INTO apikeys (user_id, api_key, title)
		VALUES (?,?,?)
		`,
		claims.UserID,
		hex.EncodeToString(apiKeyHash[:]),
		req.Title,
	)

	if err != nil {
		if store.IsUniqueError(err) {
			w.WriteHeader(409)

			json.NewEncoder(w).Encode(map[string]any{
				"error": "title_already_exists_in_your_api_keys",
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
		"message": "api_key_created",
		"apiKey":  apiKey,
	})
}

// ListMyAPIKeys returns the id and title of every API key owned by the
// authenticated user.
func ListMyAPIKeys(
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

	claims, ok := middlewares.GetClaims(r)
	if !ok || claims == nil {
		w.WriteHeader(http.StatusUnauthorized)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "unauthorized",
		})

		return
	}

	rows, err := store.DB.Query(
		`
		SELECT id, title
		FROM apikeys
		WHERE user_id = $1
		`,
		claims.UserID,
	)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "database_query_error",
		})

		return
	}
	defer rows.Close()

	apiKeys := make([]models.ApiKeySummary, 0)

	for rows.Next() {
		var apiKey models.ApiKeySummary

		if err := rows.Scan(&apiKey.ID, &apiKey.Title); err != nil {
			w.WriteHeader(http.StatusInternalServerError)

			json.NewEncoder(w).Encode(map[string]any{
				"error": "scan_api_keys_error",
			})

			return
		}

		apiKeys = append(apiKeys, apiKey)
	}

	if err := rows.Err(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "scan_api_keys_error",
		})

		return
	}

	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(map[string]any{
		"apiKeys": apiKeys,
	})

}

// DeleteAPIKey deletes the API key identified by the {id} URL parameter. Only
// keys owned by the authenticated user can be deleted. It returns 400 for an
// invalid id and 404 when no matching key exists.
func DeleteAPIKey(
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

	claims, ok := middlewares.GetClaims(r)
	if !ok || claims == nil {
		w.WriteHeader(http.StatusUnauthorized)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "unauthorized",
		})

		return
	}

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || id <= 0 {
		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "invalid_api_key_id",
		})

		return
	}

	result, err := store.DB.Exec(
		`
		DELETE FROM apikeys
		WHERE id = $1 AND user_id = $2
		`,
		id,
		claims.UserID,
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
			"error": "api_key_not_found",
		})

		return
	}

	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(map[string]any{
		"message": "api_key_deleted_successfully",
	})
}
