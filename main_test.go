package main

import (
	"testing"

	"github.com/sirupsen/logrus"
)

func TestSecurityFilter_isCommandAllowed(t *testing.T) {
	// Initialize test security filter
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel) // Reduce noise during tests
	
	sf := &SecurityFilter{
		blockedCommands: []string{
			"configure",
			"config",
			"write memory",
			"write",
			"wr",
			"copy running-config startup-config",
			"delete",
			"del",
			"erase",
			"format",
			"factory-reset",
			"factory-default",
			"reload",
			"reboot",
			"restart",
			"shutdown",
			"halt",
			"reset",
			"boot ",
			"upgrade",
			"update",
		},
		logger: logger,
	}

	// Test cases for commands that should be ALLOWED
	allowedCommands := []string{
		"cat /etc/config/wireless",
		"cat /etc/passwd",
		"ls -la",
		"head -n 10 /var/log/messages",
		"tail -f /var/log/syslog",
		"grep error /var/log/messages",
		"find /etc -name '*.conf'",
		"ps aux",
		"top",
		"df -h",
		"free -m",
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
		"help",
		"?",
		"pwd",
		"whoami",
		"id",
		"uptime",
		"date",
		"uname -a",
		"ping 8.8.8.8",
		"traceroute google.com",
		"nslookup google.com",
		"dig google.com",
		"netstat -an",
		"ifconfig",
		"ip addr show",
		"route -n",
		"arp -a",
		"lsof",
		"which ssh",
		"whereis python",
		"file /bin/ls",
		"stat /etc/passwd",
		"wc -l /etc/passwd",
		"sort /etc/passwd",
		"uniq /etc/passwd",
		"cut -d: -f1 /etc/passwd",
		"awk '{print $1}' /etc/passwd",
		"sed 's/root/admin/g' /etc/passwd",
		"more /etc/hosts",
		"less /etc/hosts",
		"vi /tmp/test.txt", // Not dangerous as it's vi (read-only usage)
		"nano /tmp/test.txt", // Not dangerous in /tmp
	}

	for _, cmd := range allowedCommands {
		t.Run("Allow_"+cmd, func(t *testing.T) {
			if !sf.isCommandAllowed(cmd) {
				t.Errorf("Command '%s' should be allowed but was blocked", cmd)
			}
		})
	}

	// Test cases for commands that should be BLOCKED
	blockedCommands := []string{
		"configure terminal",
		"config t",
		"write memory",
		"write",
		"wr",
		"copy running-config startup-config",
		"delete flash:config.txt",
		"del /etc/config/wireless",
		"erase startup-config",
		"format flash:",
		"factory-reset",
		"factory-default",
		"reload",
		"reboot",
		"restart",
		"shutdown now",
		"halt",
		"reset",
		"boot system flash:image.bin",
		"upgrade firmware.bin",
		"update system",
		"CONFIGURE TERMINAL", // Test case insensitivity
		"Config Terminal",    // Test case insensitivity
		"WRITE MEMORY",       // Test case insensitivity
		"DELETE file.txt",    // Test case insensitivity
	}

	for _, cmd := range blockedCommands {
		t.Run("Block_"+cmd, func(t *testing.T) {
			if sf.isCommandAllowed(cmd) {
				t.Errorf("Command '%s' should be blocked but was allowed", cmd)
			}
		})
	}

	// Test edge cases
	edgeCases := []struct {
		command string
		allowed bool
		desc    string
	}{
		{"", true, "empty command should be allowed"},
		{"   ", true, "whitespace-only command should be allowed"},
		{"cat", true, "single word command should be allowed"},
		{"configure", false, "exact match of blocked command should be blocked"},
		{"configures", false, "command starting with blocked word should be blocked"},
		{"preconfigure", true, "command containing but not starting with blocked word should be allowed"},
		{"write file.txt", false, "write command with arguments should be blocked"},
		{"writeonly", false, "command starting with 'write' should be blocked"},
		{"copyright notice", true, "command containing 'copy' but not starting with it should be allowed"},
		{"boot", true, "boot without space should be allowed"},
		{"boot system", false, "boot with space should be blocked"},
		{"bootstrap", true, "command starting with 'boot' but not 'boot ' should be allowed"},
	}

	for _, tc := range edgeCases {
		t.Run("Edge_"+tc.desc, func(t *testing.T) {
			result := sf.isCommandAllowed(tc.command)
			if result != tc.allowed {
				t.Errorf("Command '%s' - expected allowed=%v, got allowed=%v (%s)", 
					tc.command, tc.allowed, result, tc.desc)
			}
		})
	}
}

func TestSecurityFilter_CaseInsensitivity(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	
	sf := &SecurityFilter{
		blockedCommands: []string{"configure", "delete", "reboot"},
		logger:          logger,
	}

	testCases := []struct {
		command string
		allowed bool
	}{
		{"configure", false},
		{"CONFIGURE", false},
		{"Configure", false},
		{"ConFiGuRe", false},
		{"delete file", false},
		{"DELETE FILE", false},
		{"Delete File", false},
		{"reboot now", false},
		{"REBOOT NOW", false},
		{"Reboot Now", false},
	}

	for _, tc := range testCases {
		t.Run("CaseTest_"+tc.command, func(t *testing.T) {
			result := sf.isCommandAllowed(tc.command)
			if result != tc.allowed {
				t.Errorf("Command '%s' - expected allowed=%v, got allowed=%v", 
					tc.command, tc.allowed, result)
			}
		})
	}
}

func TestSecurityFilter_WhitespaceHandling(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	
	sf := &SecurityFilter{
		blockedCommands: []string{"configure", "delete"},
		logger:          logger,
	}

	testCases := []struct {
		command string
		allowed bool
	}{
		{"  configure  ", false},
		{"\tconfigure\t", false},
		{"\nconfigure\n", false},
		{"  delete file  ", false},
		{"\tdelete\tfile\t", false},
		{"  cat file  ", true},
		{"\tls\t-la\t", true},
	}

	for _, tc := range testCases {
		t.Run("WhitespaceTest_"+tc.command, func(t *testing.T) {
			result := sf.isCommandAllowed(tc.command)
			if result != tc.allowed {
				t.Errorf("Command '%s' - expected allowed=%v, got allowed=%v", 
					tc.command, tc.allowed, result)
			}
		})
	}
}

// Benchmark tests to ensure performance is acceptable
func BenchmarkSecurityFilter_AllowedCommand(b *testing.B) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	
	sf := &SecurityFilter{
		blockedCommands: []string{
			"configure", "config", "write memory", "write", "wr",
			"copy running-config startup-config", "delete", "del",
			"erase", "format", "factory-reset", "factory-default",
			"reload", "reboot", "restart", "shutdown", "halt",
			"reset", "boot ", "upgrade", "update",
		},
		logger: logger,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sf.isCommandAllowed("cat /etc/config/wireless")
	}
}

func BenchmarkSecurityFilter_BlockedCommand(b *testing.B) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	
	sf := &SecurityFilter{
		blockedCommands: []string{
			"configure", "config", "write memory", "write", "wr",
			"copy running-config startup-config", "delete", "del",
			"erase", "format", "factory-reset", "factory-default",
			"reload", "reboot", "restart", "shutdown", "halt",
			"reset", "boot ", "upgrade", "update",
		},
		logger: logger,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sf.isCommandAllowed("configure terminal")
	}
}