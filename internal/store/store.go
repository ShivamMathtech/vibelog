package store

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/vibelog/vibelog/pkg/types"
	_ "modernc.org/sqlite"
)

type Store struct {
	db *sql.DB
}

func New(dbPath string) (*Store, error) {
	db, err := sql.Open("sqlite", dbPath+"?_pragma=foreign_keys(1)")
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Store) migrate() error {
	schema := `
CREATE TABLE IF NOT EXISTS sessions (
    id TEXT PRIMARY KEY,
    repo_path TEXT NOT NULL,
    branch TEXT NOT NULL,
    started_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    ended_at DATETIME,
    summary TEXT
);

CREATE TABLE IF NOT EXISTS events (
    id TEXT PRIMARY KEY,
    session_id TEXT REFERENCES sessions(id),
    type TEXT NOT NULL CHECK(type IN ('prompt', 'response', 'file_change', 'decision', 'error')),
    content TEXT NOT NULL,
    metadata TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE VIRTUAL TABLE IF NOT EXISTS events_fts USING fts5(
    content, 
    content='events',
    content_rowid='rowid'
);

CREATE TABLE IF NOT EXISTS decisions (
    id TEXT PRIMARY KEY,
    event_id TEXT REFERENCES events(id),
    session_id TEXT REFERENCES sessions(id),
    title TEXT NOT NULL,
    context TEXT,
    tags TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TRIGGER IF NOT EXISTS events_ai AFTER INSERT ON events BEGIN
    INSERT INTO events_fts(rowid, content) VALUES (new.rowid, new.content);
END;

CREATE TRIGGER IF NOT EXISTS events_ad AFTER DELETE ON events BEGIN
    INSERT INTO events_fts(events_fts, rowid, content) VALUES('delete', old.rowid, old.content);
END;

CREATE TRIGGER IF NOT EXISTS events_au AFTER UPDATE ON events BEGIN
    INSERT INTO events_fts(events_fts, rowid, content) VALUES('delete', old.rowid, old.content);
    INSERT INTO events_fts(rowid, content) VALUES (new.rowid, new.content);
END;

CREATE INDEX IF NOT EXISTS idx_events_session ON events(session_id);
CREATE INDEX IF NOT EXISTS idx_events_type ON events(type);
CREATE INDEX IF NOT EXISTS idx_decisions_session ON decisions(session_id);
`
	_, err := s.db.Exec(schema)
	return err
}

func (s *Store) CreateSession(repoPath, branch string) (*types.Session, error) {
	sess := &types.Session{
		ID:        uuid.New().String(),
		RepoPath:  repoPath,
		Branch:    branch,
		StartedAt: time.Now(),
	}
	_, err := s.db.Exec(
		"INSERT INTO sessions (id, repo_path, branch, started_at) VALUES (?, ?, ?, ?)",
		sess.ID, sess.RepoPath, sess.Branch, sess.StartedAt,
	)
	return sess, err
}

func (s *Store) EndSession(id string) error {
	_, err := s.db.Exec("UPDATE sessions SET ended_at = ? WHERE id = ?", time.Now(), id)
	return err
}

func (s *Store) AddEvent(sessionID, eventType, content, metadata string) (*types.Event, error) {
	ev := &types.Event{
		ID:        uuid.New().String(),
		SessionID: sessionID,
		Type:      eventType,
		Content:   content,
		Metadata:  metadata,
		CreatedAt: time.Now(),
	}
	_, err := s.db.Exec(
		"INSERT INTO events (id, session_id, type, content, metadata, created_at) VALUES (?, ?, ?, ?, ?, ?)",
		ev.ID, ev.SessionID, ev.Type, ev.Content, ev.Metadata, ev.CreatedAt,
	)
	return ev, err
}

func (s *Store) AddDecision(eventID, sessionID, title, context, tags string) (*types.Decision, error) {
	d := &types.Decision{
		ID:        uuid.New().String(),
		EventID:   eventID,
		SessionID: sessionID,
		Title:     title,
		Context:   context,
		Tags:      tags,
		CreatedAt: time.Now(),
	}
	_, err := s.db.Exec(
		"INSERT INTO decisions (id, event_id, session_id, title, context, tags, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)",
		d.ID, d.EventID, d.SessionID, d.Title, d.Context, d.Tags, d.CreatedAt,
	)
	return d, err
}

func (s *Store) GetSessions() ([]types.Session, error) {
	rows, err := s.db.Query("SELECT id, repo_path, branch, started_at, ended_at, summary FROM sessions ORDER BY started_at DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []types.Session
	for rows.Next() {
		var sess types.Session
		var endedAt sql.NullTime
		if err := rows.Scan(&sess.ID, &sess.RepoPath, &sess.Branch, &sess.StartedAt, &endedAt, &sess.Summary); err != nil {
			return nil, err
		}
		if endedAt.Valid {
			sess.EndedAt = &endedAt.Time
		}
		sessions = append(sessions, sess)
	}
	return sessions, rows.Err()
}

func (s *Store) GetEvents(sessionID string) ([]types.Event, error) {
	rows, err := s.db.Query("SELECT id, session_id, type, content, metadata, created_at FROM events WHERE session_id = ? ORDER BY created_at", sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []types.Event
	for rows.Next() {
		var ev types.Event
		if err := rows.Scan(&ev.ID, &ev.SessionID, &ev.Type, &ev.Content, &ev.Metadata, &ev.CreatedAt); err != nil {
			return nil, err
		}
		events = append(events, ev)
	}
	return events, rows.Err()
}

func (s *Store) Search(query string) ([]types.SearchResult, error) {
	rows, err := s.db.Query(`
		SELECT e.session_id, e.type, e.content, e.created_at, s.branch, rank
		FROM events_fts
		JOIN events e ON e.rowid = events_fts.rowid
		JOIN sessions s ON s.id = e.session_id
		WHERE events_fts MATCH ?
		ORDER BY rank
		LIMIT 50
	`, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []types.SearchResult
	for rows.Next() {
		var r types.SearchResult
		if err := rows.Scan(&r.SessionID, &r.EventType, &r.Content, &r.CreatedAt, &r.Branch, &r.Rank); err != nil {
			return nil, err
		}
		results = append(results, r)
	}
	return results, rows.Err()
}

func (s *Store) GetDecisions(sessionID string) ([]types.Decision, error) {
	rows, err := s.db.Query("SELECT id, event_id, session_id, title, context, tags, created_at FROM decisions WHERE session_id = ? ORDER BY created_at DESC", sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var decisions []types.Decision
	for rows.Next() {
		var d types.Decision
		if err := rows.Scan(&d.ID, &d.EventID, &d.SessionID, &d.Title, &d.Context, &d.Tags, &d.CreatedAt); err != nil {
			return nil, err
		}
		decisions = append(decisions, d)
	}
	return decisions, rows.Err()
}

func (s *Store) GetSession(id string) (*types.Session, error) {
	var sess types.Session
	var endedAt sql.NullTime
	err := s.db.QueryRow("SELECT id, repo_path, branch, started_at, ended_at, summary FROM sessions WHERE id = ?", id).
		Scan(&sess.ID, &sess.RepoPath, &sess.Branch, &sess.StartedAt, &endedAt, &sess.Summary)
	if err != nil {
		return nil, err
	}
	if endedAt.Valid {
		sess.EndedAt = &endedAt.Time
	}
	return &sess, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) DB() *sql.DB {
	return s.db
}
