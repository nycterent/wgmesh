package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/atvirokodosprendimai/wgmesh/pkg/crypto"
	"github.com/atvirokodosprendimai/wgmesh/pkg/daemon"
	"github.com/atvirokodosprendimai/wgmesh/pkg/mesh"

	// Import discovery to register the DHT factory via init()
	_ "github.com/atvirokodosprendimai/wgmesh/pkg/discovery"
)

func main() {
	// Check for subcommands first
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "join":
			joinCmd()
			return
		case "init":
			initCmd()
			return
		case "status":
			statusCmd()
			return
		case "test-peer":
			testPeerCmd()
			return
		}
	}

	// Original CLI mode
	var (
		stateFile  = flag.String("state", "mesh-state.json", "Path to mesh state file")
		addNode    = flag.String("add", "", "Add node (format: hostname:ip:ssh_host[:ssh_port])")
		removeNode = flag.String("remove", "", "Remove node by hostname")
		list       = flag.Bool("list", false, "List all nodes")
		deploy     = flag.Bool("deploy", false, "Deploy configuration to all nodes")
		init       = flag.Bool("init", false, "Initialize new mesh")
		encrypt    = flag.Bool("encrypt", false, "Encrypt state file with password (asks for password)")
	)

	flag.Parse()

	// Handle encryption flag
	if *encrypt {
		var password string
		var err error

		if *init {
			// For init, ask for password twice
			password, err = crypto.ReadPasswordTwice("Enter encryption password: ")
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to read password: %v\n", err)
				os.Exit(1)
			}
		} else {
			// For other operations, ask once
			password, err = crypto.ReadPassword("Enter encryption password: ")
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to read password: %v\n", err)
				os.Exit(1)
			}
		}

		mesh.SetEncryptionPassword(password)
	}

	if *init {
		if err := mesh.Initialize(*stateFile); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to initialize mesh: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Mesh initialized successfully")
		return
	}

	m, err := mesh.Load(*stateFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load mesh state: %v\n", err)
		os.Exit(1)
	}

	switch {
	case *addNode != "":
		if err := m.AddNode(*addNode); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to add node: %v\n", err)
			os.Exit(1)
		}
		if err := m.Save(*stateFile); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to save state: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Node added successfully\n")

	case *removeNode != "":
		if err := m.RemoveNode(*removeNode); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to remove node: %v\n", err)
			os.Exit(1)
		}
		if err := m.Save(*stateFile); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to save state: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Node removed successfully\n")

	case *list:
		m.List()

	case *deploy:
		if err := m.Deploy(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to deploy: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Deployment completed successfully")

	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`wgmesh - WireGuard mesh network builder

SUBCOMMANDS (decentralized mode):
  init --secret              Generate a new mesh secret
  join --secret <SECRET>     Join a mesh network
  status --secret <SECRET>   Show mesh status

FLAGS (centralized mode):
  -state <file>    Path to mesh state file (default: mesh-state.json)
  -add <spec>      Add node (format: hostname:ip:ssh_host[:ssh_port])
  -remove <name>   Remove node by hostname
  -list            List all nodes
  -deploy          Deploy configuration to all nodes
  -init            Initialize new mesh state file
  -encrypt         Encrypt state file with password

EXAMPLES:
  # Decentralized mode (automatic peer discovery):
  wgmesh init --secret                          # Generate a new mesh secret
  wgmesh join --secret "wgmesh://v1/K7x2..."    # Join mesh on this node

  # Centralized mode (SSH-based deployment):
  wgmesh -init -encrypt                         # Initialize encrypted state
  wgmesh -add node1:10.99.0.1:192.168.1.10     # Add a node
  wgmesh -deploy                               # Deploy to all nodes`)
}

// initCmd handles the "init --secret" subcommand
func initCmd() {
	fs := flag.NewFlagSet("init", flag.ExitOnError)
	secretMode := fs.Bool("secret", false, "Generate a new mesh secret")
	fs.Parse(os.Args[2:])

	if *secretMode {
		secret, err := daemon.GenerateSecret()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to generate secret: %v\n", err)
			os.Exit(1)
		}

		uri := daemon.FormatSecretURI(secret)
		fmt.Println("Generated mesh secret:")
		fmt.Println()
		fmt.Println(uri)
		fmt.Println()
		fmt.Println("Share this secret with all nodes that should join the mesh.")
		fmt.Println("Run: wgmesh join --secret \"" + uri + "\"")
		return
	}

	fs.Usage()
	os.Exit(1)
}

// joinCmd handles the "join --secret" subcommand
func joinCmd() {
	fs := flag.NewFlagSet("join", flag.ExitOnError)
	secret := fs.String("secret", "", "Mesh secret (required)")
	advertiseRoutes := fs.String("advertise-routes", "", "Comma-separated list of routes to advertise")
	listenPort := fs.Int("listen-port", 51820, "WireGuard listen port")
	iface := fs.String("interface", "wg0", "WireGuard interface name")
	logLevel := fs.String("log-level", "info", "Log level (debug, info, warn, error)")
	fs.Parse(os.Args[2:])

	if *secret == "" {
		fmt.Fprintln(os.Stderr, "Error: --secret is required")
		fmt.Fprintln(os.Stderr, "Usage: wgmesh join --secret <SECRET>")
		os.Exit(1)
	}

	// Parse advertise routes
	var routes []string
	if *advertiseRoutes != "" {
		routes = strings.Split(*advertiseRoutes, ",")
		for i, r := range routes {
			routes[i] = strings.TrimSpace(r)
		}
	}

	// Create daemon config
	cfg, err := daemon.NewConfig(daemon.DaemonOpts{
		Secret:          *secret,
		InterfaceName:   *iface,
		WGListenPort:    *listenPort,
		AdvertiseRoutes: routes,
		LogLevel:        *logLevel,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create config: %v\n", err)
		os.Exit(1)
	}

	// Create and run daemon with DHT discovery
	d, err := daemon.NewDaemon(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create daemon: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Initializing mesh node with DHT discovery...")

	if err := d.RunWithDHTDiscovery(); err != nil {
		fmt.Fprintf(os.Stderr, "Daemon error: %v\n", err)
		os.Exit(1)
	}
}

// testPeerCmd tests direct peer exchange connectivity
func testPeerCmd() {
	fs := flag.NewFlagSet("test-peer", flag.ExitOnError)
	secret := fs.String("secret", "", "Mesh secret (required)")
	peerAddr := fs.String("peer", "", "Peer address to test (IP:PORT)")
	listenPort := fs.Int("port", 0, "Local port to listen on (0 = random)")
	fs.Parse(os.Args[2:])

	if *secret == "" || *peerAddr == "" {
		fmt.Fprintln(os.Stderr, "Usage: wgmesh test-peer --secret <SECRET> --peer <IP:PORT>")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "This tests direct UDP connectivity to another wgmesh node.")
		fmt.Fprintln(os.Stderr, "Run 'wgmesh join' on the peer first, note its exchange port,")
		fmt.Fprintln(os.Stderr, "then test with: wgmesh test-peer --secret <SECRET> --peer <PEER_IP>:<EXCHANGE_PORT>")
		os.Exit(1)
	}

	cfg, err := daemon.NewConfig(daemon.DaemonOpts{Secret: *secret})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create config: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Testing peer exchange with %s\n", *peerAddr)
	fmt.Printf("Network ID: %x\n", cfg.Keys.NetworkID[:8])

	// Create UDP socket
	addr := &net.UDPAddr{Port: *listenPort}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to bind UDP: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	fmt.Printf("Listening on port %d\n", conn.LocalAddr().(*net.UDPAddr).Port)

	// Resolve peer
	peerUDP, err := net.ResolveUDPAddr("udp", *peerAddr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to resolve peer: %v\n", err)
		os.Exit(1)
	}

	// Create and send test message
	announcement := crypto.CreateAnnouncement("test-pubkey", "10.0.0.1", "test:51820", nil, nil)
	data, err := crypto.SealEnvelope(crypto.MessageTypeHello, announcement, cfg.Keys.GossipKey)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create message: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Sending HELLO to %s (%d bytes)...\n", *peerAddr, len(data))
	_, err = conn.WriteToUDP(data, peerUDP)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to send: %v\n", err)
		os.Exit(1)
	}

	// Wait for response
	fmt.Println("Waiting for response (10s timeout)...")
	conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	buf := make([]byte, 65536)
	n, from, err := conn.ReadFromUDP(buf)
	if err != nil {
		fmt.Fprintf(os.Stderr, "No response: %v\n", err)
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Possible issues:")
		fmt.Fprintln(os.Stderr, "- Peer not running or wrong port")
		fmt.Fprintln(os.Stderr, "- Firewall blocking UDP")
		fmt.Fprintln(os.Stderr, "- Different secrets (different gossip keys)")
		os.Exit(1)
	}

	fmt.Printf("Received %d bytes from %s\n", n, from.String())

	// Try to decrypt
	envelope, reply, err := crypto.OpenEnvelope(buf[:n], cfg.Keys.GossipKey)
	if err != nil {
		fmt.Printf("Failed to decrypt (wrong secret?): %v\n", err)
		os.Exit(1)
	}

	fmt.Println("SUCCESS! Peer exchange working!")
	fmt.Printf("  Message type: %s\n", envelope.MessageType)
	fmt.Printf("  Peer pubkey: %s\n", reply.WGPubKey)
	fmt.Printf("  Peer mesh IP: %s\n", reply.MeshIP)
}

// statusCmd handles the "status --secret" subcommand
func statusCmd() {
	fs := flag.NewFlagSet("status", flag.ExitOnError)
	secret := fs.String("secret", "", "Mesh secret (required)")
	iface := fs.String("interface", "wg0", "WireGuard interface name")
	fs.Parse(os.Args[2:])

	if *secret == "" {
		fmt.Fprintln(os.Stderr, "Error: --secret is required")
		fmt.Fprintln(os.Stderr, "Usage: wgmesh status --secret <SECRET>")
		os.Exit(1)
	}

	// Create config to derive keys
	cfg, err := daemon.NewConfig(daemon.DaemonOpts{
		Secret:        *secret,
		InterfaceName: *iface,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create config: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Mesh Status\n")
	fmt.Printf("===========\n")
	fmt.Printf("Interface: %s\n", cfg.InterfaceName)
	fmt.Printf("Network ID: %x\n", cfg.Keys.NetworkID[:8])
	fmt.Printf("Mesh Subnet: 10.%d.0.0/16\n", cfg.Keys.MeshSubnet[0])
	fmt.Printf("Gossip Port: %d\n", cfg.Keys.GossipPort)
	fmt.Println()

	// TODO: Query actual WireGuard interface for peer status
	fmt.Println("(Run 'wg show' to see connected peers)")
}
