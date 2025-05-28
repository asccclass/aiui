package main

import(
	"os"
	"io"
	"fmt"
	"time"
	"bytes"
	"net/http"
	"encoding/json"
)

// CallToolRequestParams MCP 工具調用請求參數
type CallToolRequestParams struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

// CallToolRequest MCP 工具調用請求
type CallToolRequest struct {
	Params CallToolRequestParams `json:"params"`
}

// CallToolResultContent MCP 工具調用結果內容
type CallToolResultContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// CallToolResult MCP 工具調用結果
type CallToolResult struct {
	Content []CallToolResultContent `json:"content"`
	IsError *bool                   `json:"isError,omitempty"`
}

// callMCPTool 調用 MCP Server 的工具
func callMCPTool(toolName string, args map[string]interface{}) (string, error) {
	request := CallToolRequest{
		Params: CallToolRequestParams{
			Name:      toolName,
			Arguments: args,
		},
	}
	serverURL := os.Getenv("MCPSrv") // MCP Server URL
	if serverURL == "" {
		return "", fmt.Errorf("MCPSrv environment variable not set")
	}
	jsonData, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("marshal request: %s", err.Error())
	}
	hClient := &http.Client {
	   Timeout: 60 * time.Second,
	}
	resp, err := hClient.Post(serverURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("make request: %s", err.Error())
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %s", err.Error())
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("server error (status %d): %s", resp.StatusCode, string(body))
	}
	var result CallToolResult
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("unmarshal response: %s", err.Error())
	}
	if result.IsError != nil && *result.IsError {
		return "", fmt.Errorf("tool error: %s", result.Content[0].Text)
	}
	if len(result.Content) > 0 {
		return result.Content[0].Text, nil
	}
	return "操作完成", nil
}