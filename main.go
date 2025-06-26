package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// MCP Protocol Structures
type MCPRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

type MCPResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *MCPError   `json:"error,omitempty"`
}

type MCPError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type InitializeParams struct {
	ProtocolVersion string                 `json:"protocolVersion"`
	Capabilities    map[string]interface{} `json:"capabilities"`
	ClientInfo      map[string]interface{} `json:"clientInfo"`
}

type InitializeResult struct {
	ProtocolVersion string                 `json:"protocolVersion"`
	Capabilities    map[string]interface{} `json:"capabilities"`
	ServerInfo      map[string]interface{} `json:"serverInfo"`
}

type Tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}

type CallToolParams struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments,omitempty"`
}

type CallToolResult struct {
	Content []map[string]interface{} `json:"content"`
	IsError bool                     `json:"isError,omitempty"`
}

// Telnet Connection Manager
type TelnetConnection struct {
	conn     net.Conn
	host     string
	port     int
	timeout  time.Duration
	lastUsed time.Time
	mutex    sync.Mutex
}

type TelnetManager struct {
	connections map[string]*TelnetConnection
	mutex       sync.RWMutex
	logger      *logrus.Logger
}

// Security and Audit
type SecurityFilter struct {
	allowedCommands []string
	blockedCommands []string
	logger          *logrus.Logger
}

type AuditLogger struct {
	logger *logrus.Logger
	file   *os.File
}

// Global instances
var (
	telnetManager *TelnetManager
	securityFilter *SecurityFilter
	auditLogger   *AuditLogger
)

func init() {
	// Initialize logging
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)
	logger.SetFormatter(&logrus.JSONFormatter{})

	// Initialize audit logger
	auditFile, err := os.OpenFile("/tmp/telnet-mcp-audit.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Printf("Warning: Could not open audit log file: %v", err)
		auditFile = os.Stderr
	}

	auditLogger = &AuditLogger{
		logger: logger,
		file:   auditFile,
	}

	// Initialize telnet manager
	telnetManager = &TelnetManager{
		connections: make(map[string]*TelnetConnection),
		logger:      logger,
	}

	// Initialize security filter
	securityFilter = &SecurityFilter{
		allowedCommands: []string{
			"show version",
			"show config",
			"show running-config",
			"show interfaces",
			"show ip route",
			"show arp",
			"show users",
			"show processes",
			"show memory",
			"show cpu",
			"show log",
			"show status",
			"show system",
			"show vlan",
			"dir",
			"ls",
			"help",
			"?",
		},
		blockedCommands: []string{
			"configure",
			"config",
			"write",
			"copy",
			"erase",
			"delete",
			"reload",
			"reboot",
			"shutdown",
			"no ",
			"clear",
			"reset",
			"factory-reset",
			"format",
		},
		logger: logger,
	}
}

// Security Filter Implementation
func (sf *SecurityFilter) isCommandAllowed(command string) bool {
	command = strings.ToLower(strings.TrimSpace(command))
	
	// Check blocked commands first
	for _, blocked := range sf.blockedCommands {
		if strings.HasPrefix(command, blocked) {
			sf.logger.Warnf("Blocked command attempted: %s", command)
			return false
		}
	}

	// Check if command matches allowed patterns
	for _, allowed := range sf.allowedCommands {
		if strings.HasPrefix(command, allowed) {
			return true
		}
	}

	// Default deny for unknown commands
	sf.logger.Warnf("Unknown command denied: %s", command)
	return false
}

// Audit Logger Implementation
func (al *AuditLogger) logEvent(event string, details map[string]interface{}) {
	logEntry := map[string]interface{}{
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"event":     event,
		"details":   details,
	}
	
	jsonData, _ := json.Marshal(logEntry)
	fmt.Fprintln(al.file, string(jsonData))
	al.file.Sync()
}

// Telnet Manager Implementation
func (tm *TelnetManager) connect(host string, port int, timeout time.Duration) (*TelnetConnection, error) {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	connKey := fmt.Sprintf("%s:%d", host, port)
	
	// Check if connection already exists and is valid
	if conn, exists := tm.connections[connKey]; exists {
		conn.lastUsed = time.Now()
		return conn, nil
	}

	// Create new connection
	netConn, err := net.DialTimeout("tcp", connKey, timeout)
	if err != nil {
		tm.logger.Errorf("Failed to connect to %s:%d - %v", host, port, err)
		auditLogger.logEvent("connection_failed", map[string]interface{}{
			"host":  host,
			"port":  port,
			"error": err.Error(),
		})
		return nil, err
	}

	connection := &TelnetConnection{
		conn:     netConn,
		host:     host,
		port:     port,
		timeout:  timeout,
		lastUsed: time.Now(),
	}

	tm.connections[connKey] = connection
	tm.logger.Infof("Connected to %s:%d", host, port)
	
	auditLogger.logEvent("connection_established", map[string]interface{}{
		"host": host,
		"port": port,
	})

	return connection, nil
}

func (tm *TelnetManager) disconnect(host string, port int) error {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	connKey := fmt.Sprintf("%s:%d", host, port)
	
	if conn, exists := tm.connections[connKey]; exists {
		conn.conn.Close()
		delete(tm.connections, connKey)
		tm.logger.Infof("Disconnected from %s:%d", host, port)
		
		auditLogger.logEvent("connection_closed", map[string]interface{}{
			"host": host,
			"port": port,
		})
	}

	return nil
}

func (tc *TelnetConnection) executeCommand(command string) (string, error) {
	tc.mutex.Lock()
	defer tc.mutex.Unlock()

	// Security check
	if !securityFilter.isCommandAllowed(command) {
		auditLogger.logEvent("command_blocked", map[string]interface{}{
			"host":    tc.host,
			"command": command,
		})
		return "", fmt.Errorf("command not allowed: %s", command)
	}

	auditLogger.logEvent("command_executed", map[string]interface{}{
		"host":    tc.host,
		"command": command,
	})

	// Set read timeout
	tc.conn.SetReadDeadline(time.Now().Add(tc.timeout))
	tc.conn.SetWriteDeadline(time.Now().Add(tc.timeout))

	// Send command
	_, err := fmt.Fprintf(tc.conn, "%s\r\n", command)
	if err != nil {
		return "", err
	}

	// Read response
	var response strings.Builder
	scanner := bufio.NewScanner(tc.conn)
	
	// Read lines until we get a prompt or timeout
	for scanner.Scan() {
		line := scanner.Text()
		response.WriteString(line + "\n")
		
		// Simple prompt detection (customize based on router type)
		if matched, _ := regexp.MatchString(`[#>$]\s*$`, line); matched {
			break
		}
		
		// Prevent infinite reading
		if response.Len() > 10000 {
			break
		}
	}

	tc.lastUsed = time.Now()
	responseText := response.String()
	
	auditLogger.logEvent("command_response", map[string]interface{}{
		"host":         tc.host,
		"command":      command,
		"response_len": len(responseText),
	})

	return responseText, nil
}

// MCP Server Implementation
func handleRequest(request MCPRequest) MCPResponse {
	switch request.Method {
	case "initialize":
		return handleInitialize(request)
	case "tools/list":
		return handleToolsList(request)
	case "tools/call":
		return handleToolsCall(request)
	default:
		return MCPResponse{
			JSONRPC: "2.0",
			ID:      request.ID,
			Error: &MCPError{
				Code:    -32601,
				Message: "Method not found",
			},
		}
	}
}

func handleInitialize(request MCPRequest) MCPResponse {
	result := InitializeResult{
		ProtocolVersion: "2024-11-05",
		Capabilities: map[string]interface{}{
			"tools": map[string]interface{}{},
		},
		ServerInfo: map[string]interface{}{
			"name":    "telnet-mcp-server",
			"version": "1.0.0",
		},
	}

	return MCPResponse{
		JSONRPC: "2.0",
		ID:      request.ID,
		Result:  result,
	}
}

func handleToolsList(request MCPRequest) MCPResponse {
	tools := []Tool{
		{
			Name:        "telnet_connect",
			Description: "Connect to a router device via telnet",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"host": map[string]interface{}{
						"type":        "string",
						"description": "Router IP address",
					},
					"port": map[string]interface{}{
						"type":        "number",
						"description": "Telnet port (default: 23)",
						"default":     23,
					},
					"timeout": map[string]interface{}{
						"type":        "number",
						"description": "Connection timeout in seconds (default: 30)",
						"default":     30,
					},
				},
				"required": []string{"host"},
			},
		},
		{
			Name:        "telnet_execute_command",
			Description: "Execute a command on connected router",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"host": map[string]interface{}{
						"type":        "string",
						"description": "Router IP address",
					},
					"command": map[string]interface{}{
						"type":        "string",
						"description": "Command to execute",
					},
					"timeout": map[string]interface{}{
						"type":        "number",
						"description": "Command timeout in seconds (default: 10)",
						"default":     10,
					},
				},
				"required": []string{"host", "command"},
			},
		},
		{
			Name:        "telnet_disconnect",
			Description: "Disconnect from router device",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"host": map[string]interface{}{
						"type":        "string",
						"description": "Router IP address",
					},
					"port": map[string]interface{}{
						"type":        "number",
						"description": "Telnet port (default: 23)",
						"default":     23,
					},
				},
				"required": []string{"host"},
			},
		},
	}

	return MCPResponse{
		JSONRPC: "2.0",
		ID:      request.ID,
		Result: map[string]interface{}{
			"tools": tools,
		},
	}
}

func handleToolsCall(request MCPRequest) MCPResponse {
	var params CallToolParams
	paramData, _ := json.Marshal(request.Params)
	json.Unmarshal(paramData, &params)

	switch params.Name {
	case "telnet_connect":
		return handleTelnetConnect(request.ID, params.Arguments)
	case "telnet_execute_command":
		return handleTelnetExecuteCommand(request.ID, params.Arguments)
	case "telnet_disconnect":
		return handleTelnetDisconnect(request.ID, params.Arguments)
	default:
		return MCPResponse{
			JSONRPC: "2.0",
			ID:      request.ID,
			Error: &MCPError{
				Code:    -32601,
				Message: "Tool not found",
			},
		}
	}
}

func handleTelnetConnect(id interface{}, args map[string]interface{}) MCPResponse {
	host, ok := args["host"].(string)
	if !ok {
		return MCPResponse{
			JSONRPC: "2.0",
			ID:      id,
			Error: &MCPError{
				Code:    -32602,
				Message: "Invalid host parameter",
			},
		}
	}

	port := 23
	if p, ok := args["port"].(float64); ok {
		port = int(p)
	}

	timeout := 30 * time.Second
	if t, ok := args["timeout"].(float64); ok {
		timeout = time.Duration(t) * time.Second
	}

	_, err := telnetManager.connect(host, port, timeout)
	if err != nil {
		return MCPResponse{
			JSONRPC: "2.0",
			ID:      id,
			Error: &MCPError{
				Code:    -32603,
				Message: fmt.Sprintf("Connection failed: %v", err),
			},
		}
	}

	result := CallToolResult{
		Content: []map[string]interface{}{
			{
				"type": "text",
				"text": fmt.Sprintf("Successfully connected to %s:%d", host, port),
			},
		},
	}

	return MCPResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}
}

func handleTelnetExecuteCommand(id interface{}, args map[string]interface{}) MCPResponse {
	host, ok := args["host"].(string)
	if !ok {
		return MCPResponse{
			JSONRPC: "2.0",
			ID:      id,
			Error: &MCPError{
				Code:    -32602,
				Message: "Invalid host parameter",
			},
		}
	}

	command, ok := args["command"].(string)
	if !ok {
		return MCPResponse{
			JSONRPC: "2.0",
			ID:      id,
			Error: &MCPError{
				Code:    -32602,
				Message: "Invalid command parameter",
			},
		}
	}

	port := 23
	connKey := fmt.Sprintf("%s:%d", host, port)
	
	telnetManager.mutex.RLock()
	conn, exists := telnetManager.connections[connKey]
	telnetManager.mutex.RUnlock()

	if !exists {
		return MCPResponse{
			JSONRPC: "2.0",
			ID:      id,
			Error: &MCPError{
				Code:    -32603,
				Message: "Not connected to host. Please connect first.",
			},
		}
	}

	response, err := conn.executeCommand(command)
	if err != nil {
		return MCPResponse{
			JSONRPC: "2.0",
			ID:      id,
			Error: &MCPError{
				Code:    -32603,
				Message: fmt.Sprintf("Command execution failed: %v", err),
			},
		}
	}

	result := CallToolResult{
		Content: []map[string]interface{}{
			{
				"type": "text",
				"text": response,
			},
		},
	}

	return MCPResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}
}

func handleTelnetDisconnect(id interface{}, args map[string]interface{}) MCPResponse {
	host, ok := args["host"].(string)
	if !ok {
		return MCPResponse{
			JSONRPC: "2.0",
			ID:      id,
			Error: &MCPError{
				Code:    -32602,
				Message: "Invalid host parameter",
			},
		}
	}

	port := 23
	if p, ok := args["port"].(float64); ok {
		port = int(p)
	}

	err := telnetManager.disconnect(host, port)
	if err != nil {
		return MCPResponse{
			JSONRPC: "2.0",
			ID:      id,
			Error: &MCPError{
				Code:    -32603,
				Message: fmt.Sprintf("Disconnect failed: %v", err),
			},
		}
	}

	result := CallToolResult{
		Content: []map[string]interface{}{
			{
				"type": "text",
				"text": fmt.Sprintf("Disconnected from %s:%d", host, port),
			},
		},
	}

	return MCPResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}
}

// Main server loop
func main() {
	defer auditLogger.file.Close()

	log.Println("Telnet MCP Server starting...")
	
	auditLogger.logEvent("server_started", map[string]interface{}{
		"version": "1.0.0",
	})

	scanner := bufio.NewScanner(os.Stdin)
	
	for scanner.Scan() {
		line := scanner.Text()
		
		var request MCPRequest
		if err := json.Unmarshal([]byte(line), &request); err != nil {
			log.Printf("Invalid JSON received: %v", err)
			continue
		}

		response := handleRequest(request)
		
		responseJSON, err := json.Marshal(response)
		if err != nil {
			log.Printf("Failed to marshal response: %v", err)
			continue
		}

		fmt.Println(string(responseJSON))
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Error reading input: %v", err)
	}

	auditLogger.logEvent("server_stopped", map[string]interface{}{})
}