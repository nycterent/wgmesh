package daemon

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// localNodeState is the persisted state for a local node
type localNodeState struct {
	WGPubKey     string `json:"wg_pubkey"`
	WGPrivateKey string `json:"wg_private_key"`
}

// loadLocalNode loads the local node state from a file
func loadLocalNode(path string) (*LocalNode, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var state localNodeState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, err
	}

	return &LocalNode{
		WGPubKey:     state.WGPubKey,
		WGPrivateKey: state.WGPrivateKey,
	}, nil
}

// saveLocalNode saves the local node state to a file
func saveLocalNode(path string, node *LocalNode) error {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	state := localNodeState{
		WGPubKey:     node.WGPubKey,
		WGPrivateKey: node.WGPrivateKey,
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}

	// Write with secure permissions
	return os.WriteFile(path, data, 0600)
}

// interfaceExists checks if a network interface exists
func interfaceExists(name string) bool {
	switch runtime.GOOS {
	case "linux":
		_, err := os.Stat("/sys/class/net/" + name)
		return err == nil
	case "darwin":
		cmd := exec.Command("ifconfig", name)
		return cmd.Run() == nil
	default:
		return false
	}
}

// createInterface creates a WireGuard interface
func createInterface(name string) error {
	switch runtime.GOOS {
	case "linux":
		cmd := exec.Command("ip", "link", "add", "dev", name, "type", "wireguard")
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to create interface: %s: %w", string(output), err)
		}
		return nil
	case "darwin":
		// On macOS, wireguard-go creates the interface when started
		// We'll use a userspace implementation
		return nil
	default:
		return fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}
}

// configureInterface configures a WireGuard interface with private key and port
func configureInterface(name, privateKey string, listenPort int) error {
	// Write private key to temp file
	tmpFile, err := os.CreateTemp("", "wg-key-*")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(privateKey); err != nil {
		tmpFile.Close()
		return fmt.Errorf("failed to write private key: %w", err)
	}
	tmpFile.Close()

	// Configure interface
	args := []string{"set", name, "private-key", tmpFile.Name(), "listen-port", fmt.Sprintf("%d", listenPort)}
	cmd := exec.Command("wg", args...)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to configure interface: %s: %w", string(output), err)
	}

	return nil
}

// setInterfaceAddress sets the IP address on an interface
func setInterfaceAddress(name, address string) error {
	switch runtime.GOOS {
	case "linux":
		// Remove existing addresses first
		exec.Command("ip", "addr", "flush", "dev", name).Run()

		cmd := exec.Command("ip", "addr", "add", address, "dev", name)
		if output, err := cmd.CombinedOutput(); err != nil {
			// Ignore "file exists" error (address already set)
			if !strings.Contains(string(output), "File exists") {
				return fmt.Errorf("failed to set address: %s: %w", string(output), err)
			}
		}
		return nil
	case "darwin":
		// Extract IP and netmask
		parts := strings.Split(address, "/")
		if len(parts) != 2 {
			return fmt.Errorf("invalid address format: %s", address)
		}
		ip := parts[0]

		cmd := exec.Command("ifconfig", name, "inet", ip, ip, "alias")
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to set address: %s: %w", string(output), err)
		}
		return nil
	default:
		return fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}
}

// setInterfaceUp brings an interface up
func setInterfaceUp(name string) error {
	switch runtime.GOOS {
	case "linux":
		cmd := exec.Command("ip", "link", "set", "dev", name, "up")
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to bring interface up: %s: %w", string(output), err)
		}
		return nil
	case "darwin":
		cmd := exec.Command("ifconfig", name, "up")
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to bring interface up: %s: %w", string(output), err)
		}
		return nil
	default:
		return fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}
}

// setInterfaceDown brings an interface down
func setInterfaceDown(name string) error {
	switch runtime.GOOS {
	case "linux":
		cmd := exec.Command("ip", "link", "set", "dev", name, "down")
		cmd.Run() // Ignore errors - interface might not be up
		return nil
	case "darwin":
		cmd := exec.Command("ifconfig", name, "down")
		cmd.Run() // Ignore errors
		return nil
	default:
		return nil
	}
}

// resetInterface resets an existing interface for reconfiguration
func resetInterface(name string) error {
	// Bring interface down first
	setInterfaceDown(name)

	switch runtime.GOOS {
	case "linux":
		// Flush all addresses
		exec.Command("ip", "addr", "flush", "dev", name).Run()
		// Remove all peers
		exec.Command("wg", "set", name, "peer", "remove").Run()
		return nil
	case "darwin":
		return nil
	default:
		return nil
	}
}

// isPortInUse checks if a UDP port is already bound
func isPortInUse(port int) bool {
	addr := fmt.Sprintf(":%d", port)
	conn, err := net.ListenPacket("udp", addr)
	if err != nil {
		return true // Port is in use
	}
	conn.Close()
	return false
}

// findAvailablePort finds an available UDP port starting from the given port
func findAvailablePort(startPort int) int {
	for port := startPort; port < startPort+100; port++ {
		if !isPortInUse(port) {
			return port
		}
	}
	return 0 // No available port found
}

// getWGInterfacePort gets the listen port of a WireGuard interface (0 if not set)
func getWGInterfacePort(name string) int {
	cmd := exec.Command("wg", "show", name, "listen-port")
	output, err := cmd.Output()
	if err != nil {
		return 0
	}
	var port int
	fmt.Sscanf(strings.TrimSpace(string(output)), "%d", &port)
	return port
}
