// Package mcps is a minimal MCP (Model Context Protocol) server over
// stdio: newline-delimited JSON-RPC 2.0 with the initialize, tools/list
// and tools/call methods. Just enough for an agent to play a game; no
// external dependencies.
package mcps

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
)

// Tool is one callable exposed to the agent.
type Tool struct {
	Name        string
	Description string
	Schema      string // JSON Schema for the arguments, as a string
	Call        func(args map[string]any) (string, error)
}

type request struct {
	JSONRPC string           `json:"jsonrpc"`
	ID      *json.RawMessage `json:"id"`
	Method  string           `json:"method"`
	Params  json.RawMessage  `json:"params"`
}

// Serve reads JSON-RPC requests from stdin and answers on stdout until
// stdin closes. Blocks.
func Serve(name string, tools []Tool) {
	in := bufio.NewScanner(os.Stdin)
	in.Buffer(make([]byte, 0, 1024*1024), 1024*1024)
	out := bufio.NewWriter(os.Stdout)

	reply := func(id *json.RawMessage, result any) {
		if id == nil {
			return // notifications get no response
		}
		msg, _ := json.Marshal(map[string]any{
			"jsonrpc": "2.0", "id": id, "result": result,
		})
		fmt.Fprintf(out, "%s\n", msg)
		out.Flush()
	}
	replyErr := func(id *json.RawMessage, code int, text string) {
		if id == nil {
			return
		}
		msg, _ := json.Marshal(map[string]any{
			"jsonrpc": "2.0", "id": id,
			"error": map[string]any{"code": code, "message": text},
		})
		fmt.Fprintf(out, "%s\n", msg)
		out.Flush()
	}

	for in.Scan() {
		line := in.Bytes()
		if len(line) == 0 {
			continue
		}
		var req request
		if err := json.Unmarshal(line, &req); err != nil {
			continue
		}

		switch req.Method {
		case "initialize":
			reply(req.ID, map[string]any{
				"protocolVersion": "2024-11-05",
				"capabilities":    map[string]any{"tools": map[string]any{}},
				"serverInfo":      map[string]any{"name": name, "version": "0.1.0"},
			})

		case "tools/list":
			list := []map[string]any{}
			for _, t := range tools {
				list = append(list, map[string]any{
					"name":        t.Name,
					"description": t.Description,
					"inputSchema": json.RawMessage(t.Schema),
				})
			}
			reply(req.ID, map[string]any{"tools": list})

		case "tools/call":
			var params struct {
				Name      string         `json:"name"`
				Arguments map[string]any `json:"arguments"`
			}
			json.Unmarshal(req.Params, &params)
			var tool *Tool
			for i := range tools {
				if tools[i].Name == params.Name {
					tool = &tools[i]
					break
				}
			}
			if tool == nil {
				replyErr(req.ID, -32602, "unknown tool: "+params.Name)
				continue
			}
			text, err := tool.Call(params.Arguments)
			isErr := err != nil
			if isErr {
				text = err.Error()
			}
			reply(req.ID, map[string]any{
				"content": []map[string]any{{"type": "text", "text": text}},
				"isError": isErr,
			})

		case "ping":
			reply(req.ID, map[string]any{})

		default:
			// Unknown notifications are ignored; unknown requests error.
			replyErr(req.ID, -32601, "method not found: "+req.Method)
		}
	}
	if err := in.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "collider mcp: stdin error:", err)
	}
}
