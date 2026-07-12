//go:build e2e

package e2e

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"testing"
)

type jsonrpcRequest struct {
	JSONRPC string `json:"jsonrpc"`
	ID      int    `json:"id,omitempty"`
	Method  string `json:"method"`
	Params  any    `json:"params,omitempty"`
}

type jsonrpcResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int             `json:"id,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// mcpSession wraps a running wikipedia-mcp process and exchanges JSON-RPC messages.
type mcpSession struct {
	t      *testing.T
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	reader *bufio.Reader
}

func startMCPSession(t *testing.T) *mcpSession {
	t.Helper()
	cmd := exec.Command(mcpBinary)
	cmd.Env = append(os.Environ(), "WIKIPEDIA_API_BASE="+apiBase)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		t.Fatalf("stdin pipe: %v", err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatalf("stdout pipe: %v", err)
	}
	cmd.Stderr = nil

	if err := cmd.Start(); err != nil {
		t.Fatalf("start MCP server: %v", err)
	}

	s := &mcpSession{
		t:      t,
		cmd:    cmd,
		stdin:  stdin,
		reader: bufio.NewReader(stdout),
	}
	s.initialize()
	return s
}

func (s *mcpSession) send(req jsonrpcRequest) {
	s.t.Helper()
	data, err := json.Marshal(req)
	if err != nil {
		s.t.Fatalf("marshal request: %v", err)
	}
	if _, err := s.stdin.Write(append(data, '\n')); err != nil {
		s.t.Fatalf("write request: %v", err)
	}
}

// readResponse reads the next JSON-RPC response (skipping log lines if any).
func (s *mcpSession) readResponse() jsonrpcResponse {
	s.t.Helper()
	for {
		line, err := s.reader.ReadString('\n')
		if err != nil {
			s.t.Fatalf("read response: %v", err)
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Skip lines that are not JSON-RPC responses (e.g., server logs)
		if !strings.HasPrefix(line, "{") {
			continue
		}
		var resp jsonrpcResponse
		if err := json.Unmarshal([]byte(line), &resp); err != nil {
			s.t.Fatalf("unmarshal response %q: %v", line, err)
		}
		return resp
	}
}

func (s *mcpSession) initialize() {
	s.send(jsonrpcRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
		Params: map[string]any{
			"protocolVersion": "2024-11-05",
			"capabilities":    map[string]any{},
			"clientInfo": map[string]any{
				"name":    "e2e-test",
				"version": "1.0.0",
			},
		},
	})

	resp := s.readResponse()
	if resp.Error != nil {
		s.t.Fatalf("initialize error: %s", resp.Error.Message)
	}
	if !strings.Contains(string(resp.Result), "protocolVersion") {
		s.t.Fatalf("initialize result missing protocolVersion: %s", resp.Result)
	}

	// Notify initialized
	s.send(jsonrpcRequest{
		JSONRPC: "2.0",
		Method:  "notifications/initialized",
	})
}

func (s *mcpSession) close() {
	s.stdin.Close()
	s.cmd.Wait()
}

func TestMCPServerInitialize(t *testing.T) {
	s := startMCPSession(t)
	defer s.close()
	// initialize already exercised in startMCPSession; verify protocolVersion present
}

func TestMCPToolsList(t *testing.T) {
	s := startMCPSession(t)
	defer s.close()

	s.send(jsonrpcRequest{
		JSONRPC: "2.0",
		ID:      2,
		Method:  "tools/list",
		Params:  map[string]any{},
	})

	resp := s.readResponse()
	if resp.Error != nil {
		t.Fatalf("tools/list error: %s", resp.Error.Message)
	}

	var list struct {
		Tools []struct {
			Name string `json:"name"`
		} `json:"tools"`
	}
	if err := json.Unmarshal(resp.Result, &list); err != nil {
		t.Fatalf("unmarshal tools/list result: %v", err)
	}

	wantTools := map[string]bool{
		"wikipedia_search":      false,
		"wikipedia_get_page":    false,
		"wikipedia_get_summary": false,
		"wikipedia_random":      false,
	}
	for _, tool := range list.Tools {
		if _, ok := wantTools[tool.Name]; ok {
			wantTools[tool.Name] = true
		}
	}
	for name, found := range wantTools {
		if !found {
			t.Errorf("tool %q not found in tools/list response", name)
		}
	}
}

func TestMCPToolCallSearch(t *testing.T) {
	s := startMCPSession(t)
	defer s.close()

	s.send(jsonrpcRequest{
		JSONRPC: "2.0",
		ID:      3,
		Method:  "tools/call",
		Params: map[string]any{
			"name": "wikipedia_search",
			"arguments": map[string]any{
				"query": "Go (programming language)",
				"limit": 5,
			},
		},
	})

	resp := s.readResponse()
	if resp.Error != nil {
		t.Fatalf("tools/call error: %s", resp.Error.Message)
	}

	var call struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
		IsError bool `json:"isError"`
	}
	if err := json.Unmarshal(resp.Result, &call); err != nil {
		t.Fatalf("unmarshal tools/call result: %v", err)
	}
	if call.IsError {
		t.Fatal("tools/call returned isError=true")
	}
	if len(call.Content) == 0 {
		t.Fatal("expected content in tools/call result")
	}

	found := false
	for _, c := range call.Content {
		if strings.Contains(c.Text, "Go (programming language)") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("search result text does not mention 'Go (programming language)':\n%v", call.Content)
	}
}

func TestMCPToolCallSummary(t *testing.T) {
	s := startMCPSession(t)
	defer s.close()

	s.send(jsonrpcRequest{
		JSONRPC: "2.0",
		ID:      4,
		Method:  "tools/call",
		Params: map[string]any{
			"name": "wikipedia_get_summary",
			"arguments": map[string]any{
				"title": "Go (programming language)",
			},
		},
	})

	resp := s.readResponse()
	if resp.Error != nil {
		t.Fatalf("tools/call error: %s", resp.Error.Message)
	}

	var call struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
		IsError bool `json:"isError"`
	}
	if err := json.Unmarshal(resp.Result, &call); err != nil {
		t.Fatalf("unmarshal tools/call result: %v", err)
	}
	if call.IsError {
		t.Fatal("tools/call returned isError=true")
	}

	found := false
	for _, c := range call.Content {
		if strings.Contains(c.Text, "Go (programming language)") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("summary does not mention 'Go (programming language)':\n%v", call.Content)
	}
}

var _ = fmt.Sprintf
