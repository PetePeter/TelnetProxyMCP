# Telnet-Client-Proxy MCP Server Q&A

## Overview

This document provides comprehensive guidance for creating a telnet-client-proxy MCP (Model Context Protocol) server that enables Claude Code and other LLMs to interact with router devices via telnet for defensive security analysis.

## Architecture & Design

### What is the purpose of this MCP server?

The telnet-client-proxy MCP server acts as a secure bridge between LLMs (like Claude Code) and network devices accessible via telnet. It enables:

- **Automated security audits** of router configurations
- **Vulnerability discovery** through systematic analysis
- **Network topology mapping** and documentation
- **Configuration compliance checking** against security standards
- **Defensive security analysis** of network infrastructure

### How does the architecture work?

```
Claude Code (MCP Client) ←→ Telnet-Proxy MCP Server ←→ Router Devices
                                       ↓
                              Security Layer (Filtering, Logging)
```

**Core Components:**
1. **MCP Protocol Handler** - JSON-RPC 2.0 interface for Claude Code
2. **Telnet Connection Manager** - Handles persistent connections to routers
3. **Command Queue System** - Safely queues and executes commands
4. **Security Wrapper** - Input sanitization and output filtering
5. **Session State Management** - Tracks command history and device states
6. **Audit Logger** - Records all commands and responses for security review

## Programming Language Recommendation

### Why Go is the optimal choice?

**Go** is recommended for this implementation because:

1. **Excellent Network Libraries**
   - Native telnet support with TLS capabilities (TELNETS)
   - Robust error handling for network operations
   - Built-in concurrency for managing multiple router connections

2. **MCP Protocol Support**
   - Strong JSON-RPC 2.0 library ecosystem
   - Excellent JSON marshaling/unmarshaling
   - WebSocket and HTTP SSE support for transport

3. **Security Features**
   - Memory safety
   - Strong typing prevents many common vulnerabilities
   - Excellent standard library for cryptographic operations

4. **Deployment Advantages**
   - Single binary deployment
   - Cross-platform compilation
   - No runtime dependencies

**Alternative Options:**
- **Python**: Good with telnetlib3/Exscript, but requires runtime
- **Node.js**: Strong MCP SDK support, but async complexity
- **C#**: Excellent MCP SDK, but .NET runtime requirement

## Security Considerations

### Is this approach secure for defensive security analysis?

**Yes, when properly implemented.** This is a legitimate defensive security use case:

✅ **Approved Use Cases:**
- Router configuration analysis for vulnerabilities
- Network topology discovery
- Security compliance auditing
- Defensive security research

⚠️ **Security Safeguards Required:**
- Command filtering to prevent destructive operations
- Session timeouts and connection limits
- Comprehensive audit logging
- Input sanitization and output filtering
- Rate limiting to prevent abuse

### What are the telnet security implications?

**Telnet Protocol Risks:**
- **Unencrypted communication** - All data transmitted in plain text
- **No authentication verification** - Vulnerable to man-in-the-middle attacks
- **Data integrity issues** - No guarantees of data accuracy

**Mitigation Strategies:**
- Use only on trusted internal networks
- Implement TLS wrapper (TELNETS) where possible
- Monitor all communications
- Restrict to read-only operations where feasible

## MCP Server Implementation

### What tools should the MCP server expose?

**Recommended MCP Tools:**

```json
{
  "tools": [
    {
      "name": "telnet_connect",
      "description": "Connect to a router device",
      "parameters": {
        "host": "string",
        "port": "number (default: 23)",
        "timeout": "number (default: 30)"
      }
    },
    {
      "name": "telnet_execute_command",
      "description": "Execute a command on connected router",
      "parameters": {
        "command": "string",
        "timeout": "number (default: 10)"
      }
    },
    {
      "name": "telnet_get_config",
      "description": "Retrieve router configuration",
      "parameters": {
        "config_type": "string"
      }
    },
    {
      "name": "telnet_discover_interfaces",
      "description": "Discover all network interfaces",
      "parameters": {}
    },
    {
      "name": "telnet_show_routes",
      "description": "Display routing table",
      "parameters": {}
    },
    {
      "name": "telnet_disconnect",
      "description": "Disconnect from router",
      "parameters": {}
    }
  ]
}
```

### What security filtering should be implemented?

**Command Filtering Rules:**
```go
// Allowed commands (read-only operations)
var allowedCommands = []string{
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
}

// Blocked commands (destructive operations)
var blockedCommands = []string{
    "configure",
    "write",
    "erase",
    "delete",
    "reload",
    "shutdown",
    "no ",
    "clear",
}
```

## Claude Code Integration

### How to configure Claude Code for MCP server?

**Claude Code Configuration (`claude_desktop_config.json`):**

```json
{
  "mcpServers": {
    "telnet-proxy": {
      "transport": "stdio",
      "command": "./telnet-mcp-server",
      "args": [],
      "env": {
        "ROUTER_IPS": "10.98.1.131,10.98.1.171",
        "LOG_LEVEL": "info",
        "AUDIT_LOG_PATH": "/var/log/telnet-mcp-audit.log"
      }
    }
  }
}
```

**Project-level Configuration:**
```json
{
  "mcpServers": {
    "telnet-proxy": {
      "transport": "stdio", 
      "command": "go",
      "args": ["run", "main.go"],
      "cwd": "/path/to/telnet-mcp-server",
      "env": {
        "ROUTER_IPS": "10.98.1.131,10.98.1.171"
      }
    }
  }
}
```

## LLM Prompt Strategies

### What system prompts work best for router analysis?

**Primary System Prompt:**
```
You are a network security analyst with access to router telnet commands through an MCP server. Your mission is to conduct defensive security analysis of network infrastructure.

Target Devices:
- Router A: 10.98.1.131
- Router B: 10.98.1.171

Available Capabilities:
- Execute read-only router commands
- Analyze configurations for vulnerabilities
- Map network topology
- Document security findings

Guidelines:
1. Always explain your reasoning before executing commands
2. Focus on security assessment and vulnerability identification
3. Document all findings systematically
4. Suggest remediation steps for identified issues
5. Maintain audit trail of all activities

Never attempt destructive operations or configuration changes.
```

### What task-specific prompts are most effective?

**Security Assessment Prompt:**
```
Conduct a comprehensive security audit of the router at {IP}. Follow this methodology:

1. **Device Identification**: Determine make, model, firmware version
2. **Configuration Analysis**: Review running configuration for security issues
3. **Interface Assessment**: Identify all interfaces and their security status
4. **Access Control Review**: Check user accounts and access permissions
5. **Service Analysis**: Identify running services and protocols
6. **Vulnerability Assessment**: Cross-reference findings with known vulnerabilities

Provide detailed findings and remediation recommendations.
```

**Network Discovery Prompt:**
```
Map the network topology connected to router {IP}:

1. **Route Table Analysis**: Examine all routes and connected networks
2. **Interface Discovery**: Document all physical and logical interfaces
3. **ARP Table Review**: Identify connected devices
4. **VLAN Analysis**: Document VLAN configurations if present
5. **Topology Mapping**: Create visual representation of network structure

Generate comprehensive network documentation.
```

### How to handle command execution and error recovery?

**Command Execution Pattern:**
```
Before executing any command:
1. Explain the purpose and expected outcome
2. Verify the command is safe and read-only
3. Execute using appropriate MCP tool
4. Analyze results and extract relevant information
5. Document findings and suggest next steps

If errors occur:
1. Identify the error type and cause
2. Attempt alternative commands if available
3. Document the limitation for future reference
4. Continue with alternative analysis approaches
```

## Implementation Examples

### Basic Usage Flow

1. **Connection Establishment**
   ```
   User: "Analyze the security configuration of router 10.98.1.131"
   
   LLM: "I'll connect to the router and perform a security analysis..."
   → telnet_connect(host: "10.98.1.131")
   → telnet_execute_command(command: "show version")
   ```

2. **Configuration Analysis**
   ```
   → telnet_get_config(config_type: "running")
   → telnet_execute_command(command: "show users")
   → telnet_execute_command(command: "show ip route")
   ```

3. **Security Assessment**
   ```
   → Analysis of configuration for:
     - Default passwords
     - Insecure protocols (HTTP, Telnet, SNMP v1/v2)
     - Access control lists
     - Encryption settings
     - Firmware vulnerabilities
   ```

### Advanced Analysis Scenarios

**Multi-Router Comparison:**
```
Prompt: "Compare security configurations between both routers and identify inconsistencies"

Process:
1. Connect to both routers sequentially
2. Extract standardized configuration data
3. Perform comparative analysis
4. Identify security gaps and inconsistencies
5. Recommend configuration standardization
```

**Automated Vulnerability Scanning:**
```
Prompt: "Perform automated vulnerability assessment on all accessible routers"

Process:
1. Enumerate all target devices
2. Identify device types and firmware versions
3. Check against known vulnerability databases
4. Test for common security misconfigurations
5. Generate prioritized remediation report
```

## Deployment Considerations

### How to deploy and run the MCP server?

**Development Setup:**
```bash
# Clone and build
git clone <repository>
cd telnet-mcp-server
go mod init telnet-mcp-server
go build -o telnet-mcp-server

# Run with Claude Code
./telnet-mcp-server
```

**Production Deployment:**
```bash
# Build for production
go build -ldflags="-s -w" -o telnet-mcp-server

# Run with systemd
sudo systemctl start telnet-mcp-server
sudo systemctl enable telnet-mcp-server
```

### What monitoring and logging should be implemented?

**Audit Logging Requirements:**
- All telnet connections and disconnections
- Every command executed with timestamp
- All responses received from devices
- Error conditions and failures
- Security events and anomalies

**Monitoring Metrics:**
- Connection success/failure rates
- Command execution times
- Error frequency by device
- Security event counts
- Performance metrics

## Troubleshooting

### Common Issues and Solutions

**Connection Problems:**
- Verify network connectivity to target devices
- Check firewall rules and port accessibility
- Validate device telnet service status
- Review timeout settings

**Authentication Issues:**
- Confirm telnet access doesn't require authentication
- Check for banner/prompt matching problems
- Verify character encoding settings

**Command Execution Failures:**
- Review command syntax for device type
- Check for command filtering restrictions
- Verify session state and context
- Examine response parsing logic

### Best Practices

1. **Security First**: Always implement comprehensive security measures
2. **Audit Everything**: Maintain detailed logs for compliance
3. **Error Handling**: Implement robust error recovery mechanisms
4. **Rate Limiting**: Prevent abuse through connection limits
5. **Documentation**: Keep comprehensive documentation of capabilities
6. **Testing**: Regularly test against target devices
7. **Updates**: Keep libraries and dependencies current

## Conclusion

This telnet-client-proxy MCP server enables powerful defensive security analysis capabilities while maintaining appropriate security safeguards. The Go-based implementation provides robust performance and security features necessary for enterprise network security assessment.

The integration with Claude Code opens new possibilities for automated security analysis, vulnerability discovery, and network documentation that would be difficult to achieve manually.