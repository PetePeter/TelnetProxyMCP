# Usage Examples for Telnet MCP Server

## Basic Setup

1. **Start the MCP server with Claude Code**:
   ```bash
   # Ensure configuration is in place
   cp claude_config_example.json ~/.config/claude-code/config.json
   
   # Start Claude Code
   claude-code
   ```

2. **Verify MCP server is loaded**:
   In Claude Code, you should see the telnet-proxy server listed in available tools.

## Router Analysis Examples

### 1. Initial Router Connection and Discovery

**Prompt to Claude Code:**
```
Connect to router 10.98.1.131 and perform an initial discovery. I want to understand what type of device this is and its basic configuration.
```

**Expected Flow:**
1. Claude Code uses `telnet_connect` to establish connection
2. Executes `show version` to identify device type
3. Runs `show interfaces` to discover network interfaces
4. Provides summary of device capabilities

### 2. Security Configuration Audit

**Prompt to Claude Code:**
```
Perform a comprehensive security audit of router 10.98.1.131. Check for:
- Default passwords or weak authentication
- Insecure protocols (HTTP, Telnet, SNMPv1/v2)
- Access control lists
- Logging configuration
- Firmware vulnerabilities

Provide a detailed security report with remediation recommendations.
```

**Expected Flow:**
1. Connect to router
2. Execute security-focused commands:
   - `show running-config` - Full configuration review
   - `show users` - User account analysis
   - `show version` - Firmware version check
   - `show log` - Logging configuration
3. Analyze results for security issues
4. Generate detailed security report

### 3. Network Topology Mapping

**Prompt to Claude Code:**
```
Map the complete network topology for both routers (10.98.1.131 and 10.98.1.171). I need to understand:
- All connected networks
- Routing relationships
- Interface configurations
- VLAN setup if present

Create a visual representation of the network structure.
```

**Expected Flow:**
1. Connect to first router
2. Execute topology discovery commands:
   - `show ip route` - Routing table analysis  
   - `show interfaces` - Interface details
   - `show arp` - Connected devices
   - `show vlan` - VLAN configuration
3. Repeat for second router
4. Cross-reference information
5. Generate network topology diagram

### 4. Configuration Comparison

**Prompt to Claude Code:**
```
Compare the configurations between routers 10.98.1.131 and 10.98.1.171. Identify:
- Configuration differences
- Inconsistent security settings
- Different firmware versions
- Varying access control policies

Recommend standardization improvements.
```

**Expected Flow:**
1. Connect to both routers sequentially
2. Extract comparable configuration data
3. Perform side-by-side analysis
4. Identify discrepancies
5. Recommend standardization steps

### 5. Interface and Performance Analysis

**Prompt to Claude Code:**
```
Analyze the network interfaces on router 10.98.1.131 for:
- Interface utilization
- Error rates
- Speed/duplex settings
- Administrative status
- Physical connectivity

Identify any performance issues or misconfigurations.
```

**Expected Flow:**
1. Connect to router
2. Execute interface analysis commands:
   - `show interfaces` - Detailed interface information
   - `show processes` - CPU utilization
   - `show memory` - Memory usage
3. Analyze performance metrics
4. Identify bottlenecks or issues

## Advanced Usage Scenarios

### Automated Security Scanning

```
Create an automated security scan script that:
1. Connects to both routers
2. Performs standardized security checks
3. Generates compliance reports
4. Tracks security posture over time

Run this scan weekly and alert on any changes.
```

### Vulnerability Assessment

```
Check both routers for known vulnerabilities:
1. Identify exact firmware versions
2. Cross-reference with CVE databases
3. Check for default credentials
4. Scan for open services
5. Assess patch requirements

Prioritize remediation based on risk levels.
```

### Configuration Backup and Documentation

```
Create comprehensive documentation for both routers:
1. Backup all configurations
2. Document network topology
3. Create change management procedures
4. Establish configuration baselines

Generate standard operating procedures for maintenance.
```

## Error Handling Examples

### Connection Failures

```
If connection to router 10.98.1.131 fails:
1. Verify network connectivity
2. Check if telnet service is running
3. Test with alternative connection methods
4. Document network accessibility issues
```

### Command Execution Errors

```
When commands fail or return unexpected results:
1. Try alternative command syntax
2. Check device-specific command variations
3. Verify user permissions
4. Document device-specific behaviors
```

## Best Practices

### 1. Always Start with Device Identification
```
Before any analysis, run:
- show version
- show system (if available)
- help or ? (to understand available commands)
```

### 2. Use Systematic Approach
```
Follow this order for comprehensive analysis:
1. Device identification
2. Interface discovery
3. Configuration review
4. Security assessment
5. Performance analysis
6. Documentation generation
```

### 3. Document Everything
```
For each session:
1. Record device information
2. Save configuration outputs
3. Note any anomalies
4. Track changes over time
5. Maintain security baseline
```

### 4. Security First
```
Always prioritize security analysis:
1. Check for default passwords
2. Verify encryption settings
3. Review access controls
4. Assess logging configuration
5. Validate firmware currency
```

## Sample Session Log

```
User: Connect to router 10.98.1.131 and show me its basic information

Claude Code: I'll connect to the router and gather basic information about the device.

[Using telnet_connect with host: 10.98.1.131]
✓ Connected successfully to 10.98.1.131:23

[Using telnet_execute_command with command: "show version"]
Router Output:
Cisco IOS Software, C2900 Software (C2900-UNIVERSALK9-M), Version 15.1(4)M4
Technical Support: http://www.cisco.com/techsupport
Copyright (c) 1986-2012 by Cisco Systems, Inc.
...

[Using telnet_execute_command with command: "show interfaces"]
Router Output:
FastEthernet0/0 is up, line protocol is up
  Hardware is Lance, address is 0013.1986.0a60
  Internet address is 192.168.1.1/24
...

Analysis:
- Device: Cisco 2900 series router
- IOS Version: 15.1(4)M4 (may need security updates)
- Primary Interface: FastEthernet0/0 (192.168.1.1/24)
- Status: Operational

Recommendations:
1. Check for IOS security updates
2. Review interface security settings
3. Verify access control configurations
```

## Troubleshooting Commands

### Test MCP Server Functionality
```bash
# Test tools list
echo '{"jsonrpc":"2.0","id":1,"method":"tools/list"}' | ./telnet-mcp-server

# Test connection (will fail without real router)
echo '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"telnet_connect","arguments":{"host":"10.98.1.131"}}}' | ./telnet-mcp-server
```

### Debug Connection Issues
```bash
# Check network connectivity
ping 10.98.1.131

# Test telnet manually
telnet 10.98.1.131 23

# Check audit logs
tail -f /tmp/telnet-mcp-audit.log
```

These examples demonstrate the comprehensive capabilities of the telnet MCP server for defensive security analysis and network device management.