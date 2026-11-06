package models

import "time"

// AddNewProvider is the JSON body accepted when creating a provider.
type AddNewProvider struct {
	// Name is the provider name, such as "groq".
	Name string `json:"name"`

	// BaseURL is the provider API base URL. Known provider names may supply a
	// default when this field is empty.
	BaseURL string `json:"base_url"`

	// APIKey is the plaintext provider credential to encrypt before storage.
	APIKey string `json:"api_key"`
}

// Providers describes a provider returned by the provider list endpoint.
type Providers struct {
	// ID is the provider's database identifier.
	ID int64 `json:"id"`

	// Name is the provider name.
	Name string `json:"name"`

	// BaseURL is the provider API base URL.
	BaseURL string `json:"base_url"`

	// CreatedAt is the time at which the provider was stored.
	CreatedAt time.Time `json:"created_at"`
}
