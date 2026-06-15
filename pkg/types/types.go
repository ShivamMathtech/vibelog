package types

import "time"

type Session struct {
	ID        string    `json:"id"`
	RepoPath  string    `json:"repo_path"`
	Branch    string    `json:"branch"`
	StartedAt time.Time `json:"started_at"`
	EndedAt   *time.Time `json:"ended_at,omitempty"`
	Summary   string    `json:"summary"`
}

type Event struct {
	ID        string    `json:"id"`
	SessionID string    `json:"session_id"`
	Type      string    `json:"type"` // prompt, response, file_change, decision, error
	Content   string    `json:"content"`
	Metadata  string    `json:"metadata"` // JSON
	CreatedAt time.Time `json:"created_at"`
}

type Decision struct {
	ID        string    `json:"id"`
	EventID   string    `json:"event_id"`
	SessionID string    `json:"session_id"`
	Title     string    `json:"title"`
	Context   string    `json:"context"`
	Tags      string    `json:"tags"`
	CreatedAt time.Time `json:"created_at"`
}

type SearchResult struct {
	SessionID   string    `json:"session_id"`
	EventType   string    `json:"event_type"`
	Content     string    `json:"content"`
	CreatedAt   time.Time `json:"created_at"`
	Branch      string    `json:"branch"`
	Rank        float64   `json:"rank"`
}
