package entity

// DebugInfo contains system debug information
type DebugInfo struct {
	DatabaseStatus string            `json:"database_status"`
	Version        string            `json:"version"`
	Uptime         string            `json:"uptime"`
	Config         map[string]string `json:"config"`
}
