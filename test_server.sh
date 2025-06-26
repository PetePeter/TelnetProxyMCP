#!/bin/bash

echo "Testing Telnet MCP Server..."

# Test 1: Initialize
echo "Test 1: Initialize"
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test-client","version":"1.0.0"}}}' | ./telnet-mcp-server &
SERVER_PID=$!
sleep 2
kill $SERVER_PID 2>/dev/null

echo ""

# Test 2: Tools List
echo "Test 2: Tools List"
echo '{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}' | timeout 5s ./telnet-mcp-server

echo ""

# Test 3: Invalid Method
echo "Test 3: Invalid Method"
echo '{"jsonrpc":"2.0","id":3,"method":"invalid_method","params":{}}' | timeout 5s ./telnet-mcp-server

echo ""
echo "MCP Server Tests Complete"