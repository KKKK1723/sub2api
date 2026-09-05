package domain

import "time"

// GroupModelsListConfig controls the optional custom /v1/models response list.
type GroupModelsListConfig struct {
	Enabled      bool       `json:"enabled"`
	Models       []string   `json:"models,omitempty"`
	LastSyncedAt *time.Time `json:"last_synced_at,omitempty"`
	SyncError    string     `json:"sync_error,omitempty"`
}
