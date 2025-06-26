# Telnet MCP Server

A Model Context Protocol (MCP) server that provides secure telnet client capabilities for Claude Code and other LLM applications. Designed for defensive security analysis of network devices.

## Features

- **Secure Telnet Client**: Connect to router devices with command filtering
- **MCP Protocol Support**: Full JSON-RPC 2.0 compliance for Claude Code integration
- **Security Features**: Command filtering, audit logging, session management
- **Router Management**: Support for multiple concurrent router connections
- **Comprehensive Logging**: Detailed audit trail for compliance and security

## Quick Start

### Prerequisites

- Go 1.21 or later
- Claude Code or compatible MCP client
- Network access to target router devices

### Installation

1. Clone and build:
```bash
git clone <repository>
cd telnet-mcp-server
go mod tidy
go build -o telnet-mcp-server
```

2. Configure Claude Code:
```bash
# Copy example configuration
cp claude_config_example.json ~/.config/claude-code/config.json
```

3. Run with Claude Code:
```bash
claude-code
```

## Configuration

### Claude Code Integration

Add to your Claude Code configuration:

```json
{
  "mcpServers": {
    "telnet-proxy": {
      "transport": "stdio",
      "command": "./telnet-mcp-server",
      "env": {
        "ROUTER_IPS": "10.98.1.131,10.98.1.171",
        "LOG_LEVEL": "info"
      }
    }
  }
}
```

### Environment Variables

- `ROUTER_IPS`: Comma-separated list of allowed router IP addresses
- `LOG_LEVEL`: Logging level (debug, info, warn, error)
- `AUDIT_LOG_PATH`: Path to audit log file (default: /tmp/telnet-mcp-audit.log)

## Available Tools

### telnet_connect
Connect to a router device.

**Parameters:**
- `host` (required): Router IP address
- `port` (optional): Telnet port (default: 23)
- `timeout` (optional): Connection timeout in seconds (default: 30)

### telnet_execute_command
Execute a command on connected router.

**Parameters:**
- `host` (required): Router IP address
- `command` (required): Command to execute
- `timeout` (optional): Command timeout in seconds (default: 10)

### telnet_disconnect
Disconnect from router device.

**Parameters:**
- `host` (required): Router IP address
- `port` (optional): Telnet port (default: 23)

## Security Features

### Command Filtering

The server uses a **blacklist security model** that allows most commands while blocking only dangerous operations:

**Security Approach:**
- ✅ **Default Allow**: All commands are permitted by default
- ❌ **Minimal Blacklist**: Only dangerous commands are blocked
- 🔍 **Full Audit**: All commands are logged regardless of status

**Allowed Commands (Examples):**
- File operations: `cat`, `head`, `tail`, `more`, `less`, `grep`, `find`
- System info: `ls`, `ps`, `top`, `df`, `free`, `uptime`, `whoami`
- Network tools: `ping`, `traceroute`, `nslookup`, `dig`, `netstat`
- Router commands: `show version`, `show config`, `show interfaces`, etc.
- Text processing: `awk`, `sed`, `sort`, `uniq`, `wc`, `cut`

**Blocked Commands:**
- Configuration: `configure`, `config`, `write memory`, `wr`
- Destructive: `delete`, `del`, `erase`, `format`, `factory-reset`
- System control: `reload`, `reboot`, `restart`, `shutdown`, `halt`, `reset`
- Firmware: `boot ` (with space), `upgrade`, `update`

### Audit Logging

All activities are logged to the audit file with timestamps:
- Connection events
- Command executions
- Security violations
- Error conditions

## Usage Examples

### Basic Router Analysis

```bash
# In Claude Code
Connect to router 10.98.1.131 and analyze its configuration for security vulnerabilities.
```

### Network Discovery

```bash  
# In Claude Code
Map the network topology by connecting to both routers and examining their routing tables.
```

### Security Audit

```bash
# In Claude Code
Perform a comprehensive security audit of router 10.98.1.171, checking for default passwords, insecure protocols, and misconfigurations.
```

## Development

### Building

```bash
go build -o telnet-mcp-server
```

### Testing

```bash
go test ./...
```

### Running in Development

```bash
go run main.go
```

## Security Considerations

- This tool is designed for **defensive security analysis only**
- All communications are logged for audit purposes
- Command filtering prevents destructive operations
- Use only on trusted internal networks
- Review audit logs regularly for security monitoring

## Troubleshooting

### Connection Issues

1. Verify network connectivity to target devices
2. Check firewall rules and port accessibility
3. Ensure telnet service is running on target devices
4. Review timeout settings in configuration

### Command Execution Problems

1. Check command syntax for your router type
2. Verify command is in allowed list
3. Review audit logs for blocked commands
4. Ensure proper session state

### Claude Code Integration

1. Verify MCP server configuration path
2. Check environment variables
3. Review Claude Code logs for MCP errors
4. Ensure server binary has execute permissions

## License

MIT License - See LICENSE file for details.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## Support

For issues and questions:
- Check the troubleshooting section
- Review audit logs
- Open an issue on GitHub