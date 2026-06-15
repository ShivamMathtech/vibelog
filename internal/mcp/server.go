package mcp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/vibelog/vibelog/internal/store"
	"github.com/vibelog/vibelog/pkg/types"
)

type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type JSONRPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id,omitempty"`
	Result  interface{} `json:"result,omitempty"`
	Error   *RPCError   `json:"error,omitempty"`
}

type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type Server struct {
	store  *store.Store
	reader *bufio.Reader
	writer *bufio.Writer
}

func NewServer(s *store.Store) *Server {
	return &Server{
		store:  s,
		reader: bufio.NewReader(os.Stdin),
		writer: bufio.NewWriter(os.Stdout),
	}
}

func (s *Server) Run() error {
	for {
		line, err := s.reader.ReadString('
')
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var req JSONRPCRequest
		if err := json.Unmarshal([]byte(line), &req); err != nil {
			s.writeError(nil, -32700, "Parse error")
			continue
		}

		resp := s.handleRequest(req)
		if resp != nil {
			if err := s.writeResponse(resp); err != nil {
				return err
			}
		}
	}
}

func (s *Server) handleRequest(req JSONRPCRequest) *JSONRPCResponse {
	switch req.Method {
	case "initialize":
		return s.handleInitialize(req)
	case "initialized":
		return nil
	case "tools/list":
		return s.handleToolsList(req)
	case "tools/call":
		return s.handleToolCall(req)
	default:
		return s.makeError(req.ID, -32601, fmt.Sprintf("Method not found: %s", req.Method))
	}
}

func (s *Server) handleInitialize(req JSONRPCRequest) *JSONRPCResponse {
	result := map[string]interface{}{
		"protocolVersion": "2024-11-05",
		"capabilities": map[string]interface{}{
			"tools": map[string]interface{}{},
		},
		"serverInfo": map[string]interface{}{
			"name":    "vibelog",
			"version": "0.1.0",
		},
	}
	return s.makeResult(req.ID, result)
}

func (s *Server) handleToolsList(req JSONRPCRequest) *JSONRPCResponse {
	tools := []map[string]interface{}{
		{
			"name":        "search_sessions",
			"description": "Search past coding sessions by keyword, file, or decision.",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"query": map[string]interface{}{
						"type":        "string",
						"description": "Search query to find in session history",
					},
					"file_path": map[string]interface{}{
						"type":        "string",
						"description": "Optional: filter to sessions affecting this file",
					},
					"since": map[string]interface{}{
						"type":        "string",
						"description": "Optional: ISO date to search from (e.g., 2026-06-01)",
					},
				},
				"required": []string{"query"},
			},
		},
		{
			"name":        "get_decisions",
			"description": "Retrieve tagged decisions from past sessions.",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"session_id": map[string]interface{}{
						"type":        "string",
						"description": "Optional: specific session ID",
					},
					"tags": map[string]interface{}{
						"type":        "string",
						"description": "Optional: comma-separated tags to filter",
					},
				},
			},
		},
		{
			"name":        "get_file_history",
			"description": "Get history of changes related to a specific file.",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"file_path": map[string]interface{}{
						"type":        "string",
						"description": "Relative path to the file",
					},
				},
				"required": []string{"file_path"},
			},
		},
		{
			"name":        "summarize_session",
			"description": "Generate a structured summary of a specific session.",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"session_id": map[string]interface{}{
						"type":        "string",
						"description": "Session ID to summarize",
					},
				},
				"required": []string{"session_id"},
			},
		},
	}
	return s.makeResult(req.ID, map[string]interface{}{"tools": tools})
}

func (s *Server) handleToolCall(req JSONRPCRequest) *JSONRPCResponse {
	var params struct {
		Name      string          `json:"name"`
		Arguments json.RawMessage `json:"arguments"`
	}
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return s.makeError(req.ID, -32602, "Invalid params")
	}

	var result interface{}
	var err error

	switch params.Name {
	case "search_sessions":
		result, err = s.toolSearchSessions(params.Arguments)
	case "get_decisions":
		result, err = s.toolGetDecisions(params.Arguments)
	case "get_file_history":
		result, err = s.toolGetFileHistory(params.Arguments)
	case "summarize_session":
		result, err = s.toolSummarizeSession(params.Arguments)
	default:
		return s.makeError(req.ID, -32602, fmt.Sprintf("Unknown tool: %s", params.Name))
	}

	if err != nil {
		return s.makeError(req.ID, -32603, err.Error())
	}

	return s.makeResult(req.ID, result)
}

func (s *Server) toolSearchSessions(args json.RawMessage) (interface{}, error) {
	var req struct {
		Query    string `json:"query"`
		FilePath string `json:"file_path,omitempty"`
		Since    string `json:"since,omitempty"`
	}
	if err := json.Unmarshal(args, &req); err != nil {
		return nil, err
	}

	results, err := s.store.Search(req.Query)
	if err != nil {
		return nil, err
	}

	if req.FilePath != "" {
		var filtered []types.SearchResult
		for _, r := range results {
			if strings.Contains(r.Content, req.FilePath) {
				filtered = append(filtered, r)
			}
		}
		results = filtered
	}

	if req.Since != "" {
		since, err := time.Parse("2006-01-02", req.Since)
		if err == nil {
			var filtered []types.SearchResult
			for _, r := range results {
				if r.CreatedAt.After(since) {
					filtered = append(filtered, r)
				}
			}
			results = filtered
		}
	}

	return map[string]interface{}{"results": results}, nil
}

func (s *Server) toolGetDecisions(args json.RawMessage) (interface{}, error) {
	var req struct {
		SessionID string `json:"session_id,omitempty"`
		Tags      string `json:"tags,omitempty"`
	}
	if err := json.Unmarshal(args, &req); err != nil {
		return nil, err
	}

	var decisions []types.Decision
	var err error

	if req.SessionID != "" {
		decisions, err = s.store.GetDecisions(req.SessionID)
	} else {
		rows, err := s.store.DB().Query("SELECT id, event_id, session_id, title, context, tags, created_at FROM decisions ORDER BY created_at DESC LIMIT 50")
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		for rows.Next() {
			var d types.Decision
			if err := rows.Scan(&d.ID, &d.EventID, &d.SessionID, &d.Title, &d.Context, &d.Tags, &d.CreatedAt); err != nil {
				return nil, err
			}
			decisions = append(decisions, d)
		}
		err = rows.Err()
	}
	if err != nil {
		return nil, err
	}

	if req.Tags != "" {
		tagList := strings.Split(req.Tags, ",")
		var filtered []types.Decision
		for _, d := range decisions {
			for _, t := range tagList {
				if strings.Contains(d.Tags, strings.TrimSpace(t)) {
					filtered = append(filtered, d)
					break
				}
			}
		}
		decisions = filtered
	}

	return map[string]interface{}{"decisions": decisions}, nil
}

func (s *Server) toolGetFileHistory(args json.RawMessage) (interface{}, error) {
	var req struct {
		FilePath string `json:"file_path"`
	}
	if err := json.Unmarshal(args, &req); err != nil {
		return nil, err
	}

	pattern := fmt.Sprintf("%%%s%%", req.FilePath)
	rows, err := s.store.DB().Query(`
		SELECT e.session_id, e.type, e.content, e.created_at, s.branch
		FROM events e
		JOIN sessions s ON s.id = e.session_id
		WHERE e.metadata LIKE ? OR e.content LIKE ?
		ORDER BY e.created_at DESC
		LIMIT 50
	`, pattern, pattern)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []types.SearchResult
	for rows.Next() {
		var r types.SearchResult
		if err := rows.Scan(&r.SessionID, &r.EventType, &r.Content, &r.CreatedAt, &r.Branch); err != nil {
			return nil, err
		}
		results = append(results, r)
	}
	return map[string]interface{}{"history": results}, rows.Err()
}

func (s *Server) toolSummarizeSession(args json.RawMessage) (interface{}, error) {
	var req struct {
		SessionID string `json:"session_id"`
	}
	if err := json.Unmarshal(args, &req); err != nil {
		return nil, err
	}

	session, err := s.store.GetSession(req.SessionID)
	if err != nil {
		return nil, err
	}

	events, err := s.store.GetEvents(req.SessionID)
	if err != nil {
		return nil, err
	}

	decisions, err := s.store.GetDecisions(req.SessionID)
	if err != nil {
		return nil, err
	}

	var summary strings.Builder
	summary.WriteString(fmt.Sprintf("Session: %s
", session.ID))
	summary.WriteString(fmt.Sprintf("Branch: %s
", session.Branch))
	summary.WriteString(fmt.Sprintf("Started: %s

", session.StartedAt.Format(time.RFC3339)))

	summary.WriteString("Key Decisions:
")
	for _, d := range decisions {
		summary.WriteString(fmt.Sprintf("- %s: %s
", d.Title, d.Context))
	}

	summary.WriteString("
Events:
")
	for _, e := range events {
		summary.WriteString(fmt.Sprintf("[%s] %s: %s
", e.CreatedAt.Format("15:04"), e.Type, truncate(e.Content, 100)))
	}

	return map[string]interface{}{
		"summary": summary.String(),
		"session": session,
		"stats": map[string]interface{}{
			"total_events":    len(events),
			"total_decisions": len(decisions),
		},
	}, nil
}

func (s *Server) makeResult(id interface{}, result interface{}) *JSONRPCResponse {
	return &JSONRPCResponse{JSONRPC: "2.0", ID: id, Result: result}
}

func (s *Server) makeError(id interface{}, code int, message string) *JSONRPCResponse {
	return &JSONRPCResponse{JSONRPC: "2.0", ID: id, Error: &RPCError{Code: code, Message: message}}
}

func (s *Server) writeError(id interface{}, code int, message string) {
	s.writeResponse(s.makeError(id, code, message))
}

func (s *Server) writeResponse(resp *JSONRPCResponse) error {
	data, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	if _, err := s.writer.Write(data); err != nil {
		return err
	}
	if _, err := s.writer.Write([]byte("
")); err != nil {
		return err
	}
	return s.writer.Flush()
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}