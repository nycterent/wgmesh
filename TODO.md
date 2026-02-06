wgmeshbuilder: Token-Based Mesh
Autodiscovery
Problem Statement
Currently wgmeshbuilder requires a centralized operator who manually --add nodes,
maintains mesh-state.json , and runs --deploy over SSH. The goal is to enable a fully
decentralized mode where any node with the same shared secret automatically
discovers peers and forms a WireGuard mesh within ~60 seconds, with zero pre-
configuration beyond the secret itself.
Target UX:
# Generate a mesh secret (once, by anyone)
wgmesh init --secret
# outputs: wgmesh://K7x2mP9qR4... (also renderable as QR code)
# On every node that should join:
wgmesh join --secret "K7x2mP9qR4..."
# That's it. Mesh forms automatically.
Architecture Overview
The system uses a layered discovery approach. Each layer operates independently and
results are merged, meaning the mesh works on a LAN with no internet, on the internet
with no LAN, or both.
┌─────────────────────────────────────────────────┐
│ Shared Secret │
│ (only input, encodes everything) │
└──────────┬──────────────┬───────────────┬────────┘
│ │ │
┌─────▼─────┐ ┌─────▼──────┐ ┌──────▼───────┐
│ Layer 1: │ │ Layer 2: │ │ Layer 3: │
│ LAN │ │ BitTorrent │ │ In-mesh │
│ Multicast │ │ DHT │ │ Gossip │
│ (local) │ │ (global) │ │ (post-conn) │
└─────┬─────┘ └─────┬──────┘ └──────┬───────┘
│ │ │
┌─────▼──────────────▼───────────────▼────────┐
│ Peer Merge & Diff Engine │
│ (existing wg set / route diff logic) │
└─────────────────────┬───────────────────────┘
│
┌──────▼──────┐
│ WireGuard │
│ wg0 iface │
└─────────────┘
Key Derivation from Shared Secret
Everything is derived from the single secret. No other configuration needed.
secret (arbitrary string, min 16 chars)
│
├─ network_id = SHA256(secret)[0:20] → DHT infohash (20 bytes)
├─ gossip_key = HKDF(secret, salt="wgmesh-gossip-v1", 32 bytes) → symmetric encryption
├─ mesh_subnet = HKDF(secret, salt="wgmesh-subnet-v1", 2 bytes) → deterministic /16 (10
├─ multicast_id = HKDF(secret, salt="wgmesh-mcast-v1", 4 bytes) → multicast group discr
└─ psk = HKDF(secret, salt="wgmesh-wg-psk-v1", 32 bytes) → WireGuard PresharedKe
Mesh IP allocation: Each node’s mesh IP is derived deterministically from its WG public
key:
mesh_ip = mesh_subnet_base + uint16(SHA256(wg_pubkey || secret)[0:2])
Collision probability is low for small meshes (<100 nodes in a /16). On collision detection
(duplicate mesh IP with different pubkey), the node with the lexicographically lower
pubkey wins, the other re-derives with a nonce.
Layer 1: LAN Multicast Discovery
Purpose: Sub-second discovery on local networks. Works offline, no internet required.
Mechanism:
Bind UDP socket to a well-known multicast group (e.g., 239.192.77.69:51821 )
Periodically (every 5s) broadcast an encrypted announcement:
ANNOUNCE = Encrypt(gossip_key, {
wg_pubkey: base64,
mesh_ip: "10.x.y.z",
wg_endpoint: "192.168.1.50:51820",
routable_nets: ["192.168.10.0/24"],
timestamp: unix_epoch,
nonce: random_bytes(12)
})
Only nodes with the same secret can decrypt — others see random bytes
On receiving a valid announcement: add/update peer via wg set
Go implementation notes:
Use golang.org/x/net/ipv4 for multicast group join
Also send on IPv6 link-local multicast ff02::1 for dual-stack environments
Include a protocol version byte so future changes don’t break old nodes
Time to mesh (LAN): 1-5 seconds.
Layer 2: BitTorrent Mainline DHT (Internet Discovery)
Purpose: Global peer discovery with zero infrastructure. The public BitTorrent DHT has
millions of nodes and has been running for ~20 years.
Why this is the right choice:
No server to run, no account to create, no API to pay for
Proven at planetary scale (tens of millions of concurrent nodes)
Go library exists: github.com/anacrolix/dht/v2
Bootstrap nodes are well-known and highly available ( router.bittorrent.com:6881 ,
router.utorrent.com:6881 , dht.transmissionbt.com:6881 )
Mechanism:
1. Bootstrap into Mainline DHT (UDP, standard BEP 5 protocol)
2. announce_peer(network_id, wg_listen_port)
- This publishes our IP:port on the DHT under our network_id
3. get_peers(network_id) → returns list of IP:port pairs
- These are other nodes that announced the same network_id
4. For each discovered IP:port, initiate encrypted peer exchange (see below)
5. Re-announce every 15 minutes (DHT tokens rotate every 5-10 min)
6. Re-query get_peers every 30 seconds until mesh is stable, then every 60s
Encrypted Peer Exchange Protocol (runs over UDP or TCP to discovered peers):
After DHT gives us an IP:port, we don’t yet know the node’s WG pubkey. We need a
small side-channel handshake:
Client → Server: HELLO || nonce_c(12) || Encrypt(gossip_key, nonce_c, {
protocol: "wgmesh-v1",
wg_pubkey: <our pubkey>,
mesh_ip: <our mesh ip>,
wg_endpoint: <our best guess at public endpoint>,
routable_networks: [...],
timestamp: <unix>
})
Server → Client: REPLY || nonce_s(12) || Encrypt(gossip_key, nonce_s, {
protocol: "wgmesh-v1",
wg_pubkey: <their pubkey>,
mesh_ip: <their mesh ip>,
wg_endpoint: <their endpoint>,
routable_networks: [...],
known_peers: [<list of other peers they know about>],
timestamp: <unix>
})
The known_peers field enables transitive discovery — even if the DHT is slow, once two
nodes connect they share their full peer lists, accelerating mesh convergence.
Security considerations:
The DHT infohash (network_id) is publicly visible. An observer can see that some IPs
are participating in some network, but cannot determine it’s WireGuard or decrypt
the peer exchange.
To mitigate: rotate network_id periodically by including a time component:
SHA256(secret || floor(unix_time / 3600)) . Nodes check both current and
previous hour’s IDs during transition.
DHT announce requires outbound UDP. Nodes behind strict corporate firewalls or
CGNAT may not be able to announce, but can still query get_peers and initiate
connections to public nodes.
Time to mesh (internet): 15-60 seconds typical. First get_peers response usually
arrives within 5-15s.
Layer 3: In-Mesh Gossip
Purpose: Once WireGuard tunnels are up, use them for faster, more reliable peer
exchange.
Mechanism:
Each node runs a small gossip protocol inside the WG mesh (over the wg0 interface)
UDP port derived from secret: gossip_port = 51821 + (uint16(HKDF(secret,
"gossip-port")) % 1000)
Every 10 seconds, pick a random known peer and exchange full peer lists
Encrypted with gossip_key (belt-and-suspenders; WG already encrypts, but this
authenticates mesh membership at the application layer)
Why this layer matters:
DHT has inherent lag (minutes). Gossip converges in seconds once any two nodes
are connected.
Handles the “third node joins” case: Node C finds Node A via DHT. Node A tells Node
C about Node B via gossip. Node C connects to Node B without needing a separate
DHT lookup.
Handles routable
_networks changes propagating quickly across the mesh.
Daemon Mode Implementation
The existing wgmeshbuilder has two modes: imperative CLI ( --add , --deploy over
SSH) and the new daemon mode.
New CLI Surface
# Initialize and print a new mesh secret
wgmesh init --secret
# Output: wgmesh://K7x2mP9qR4sT8vW1xY3zA5bC7dE9fG0hI2jK4lM6n
# Join a mesh (runs as daemon)
wgmesh join --secret "K7x2mP9qR4..."
# Join with additional options
wgmesh join --secret "K7x2mP9qR4..." \
--advertise-routes "192.168.10.0/24,10.0.0.0/8" \
--listen-port 51820 \
--interface wg0 \
--log-level debug
# Show mesh status
wgmesh status --secret "K7x2mP9qR4..."
# Generate QR code for the secret
wgmesh qr --secret "K7x2mP9qR4..."
# Existing SSH-based mode still works for centralized management
wgmesh --add node1:10.99.0.1:192.168.1.10
wgmesh --deploy
Daemon Loop (pseudo-code)
func RunDaemon(secret string, opts DaemonOpts) {
// Derive all keys/params from secret
cfg := DeriveConfig(secret)
// Generate or load WG keypair from local state
localNode := LoadOrCreateLocalNode(cfg)
// Start WG interface
EnsureWGInterface(cfg.InterfaceName, localNode)
// Start all discovery layers concurrently
peers := NewPeerStore()
go LANMulticastDiscovery(cfg, localNode, peers)
go DHTDiscovery(cfg, localNode, peers)
go InMeshGossip(cfg, localNode, peers)
// Main reconciliation loop
ticker := time.NewTicker(5 * time.Second)
for range ticker.C {
desired := peers.GetAll()
current := ReadCurrentWGConfig(cfg.InterfaceName)
diff := ComputeDiff(current, desired) // reuse existing diff logic
if diff.HasChanges() {
ApplyWGChanges(cfg.InterfaceName, diff) // wg set commands
ApplyRouteChanges(cfg.InterfaceName, diff) // ip route commands
UpdatePersistentConfig(cfg, desired) // write wg0.conf for reboot survival
}
}
}
Peer Store Design
Central data structure that all three discovery layers write into. The reconciliation loop
reads from it.
type PeerInfo struct {
WGPubKey string
MeshIP string
Endpoint string // best known endpoint (ip:port)
RoutableNetworks []string
LastSeen time.Time
DiscoveredVia []string // ["lan", "dht", "gossip"]
Latency *time.Duration // measured via WG handshake
}
type PeerStore struct {
mu sync.RWMutex
peers map[string]*PeerInfo // keyed by WG pubkey
}
// Merge logic: newest timestamp wins for mutable fields (endpoint, routable_networks)
// A peer is considered dead after 5 minutes of no updates from any layer
// Dead peers are removed from WG config after 10 minutes grace period
Security Model
Threat: Secret Compromise
If the shared secret leaks, an attacker can join the mesh. Mitigations:
WireGuard PSK (derived from secret) provides forward secrecy per-session
Implement wgmesh rotate-secret that coordinates a secret rotation across the mesh
via in-mesh gossip (all nodes switch simultaneously after a grace period)
Consider supporting short-lived secrets: wgmesh join --secret "..." --expires
24h
Threat: DHT Surveillance
An observer watching the DHT can see which IPs announce the same infohash.
Mitigations:
Rotate infohash hourly (include time component in derivation)
Use Tor-style onion routing through the DHT (overkill for most cases)
For high-security deployments, disable DHT layer and use LAN + pre-seeded peers
only
Threat: Replay Attacks on Peer Exchange
An attacker captures and replays encrypted peer exchange messages. Mitigations:
Timestamp in every message, reject messages older than 60 seconds
Nonce ensures each message is unique
Gossip_key + AES-256-GCM provides authentication
Threat: Mesh IP Collision
Two nodes derive the same mesh IP. Mitigations:
Detection via gossip (two different pubkeys claiming same IP)
Deterministic resolution: lower pubkey (lexicographic) wins
Loser re-derives with nonce++ until unique
/16 gives 65534 usable addresses; collision probability < 0.1% for meshes under 50
nodes
Wire Format: QR Code / Token
The QR code / token encodes a URI:
wgmesh://v1/<base64url-encoded-secret>
That’s it. Everything else is derived.
Optional extensions (appended as query params, all optional):
wgmesh://v1/<secret>?routes=192.168.10.0/24&port=51820&name=my-mesh
QR code generation:
echo "wgmesh://v1/$(head -c 32 /dev/urandom | base64url)" | qrencode -t UTF8
File Structure Changes
wgmeshbuilder/
├── main.go # Add: join, status, qr subcommands
├── pkg/
│ ├── mesh/ # Existing: centralized mesh management
│ │ ├── types.go
│ │ ├── mesh.go
│ │ └── deploy.go
│ ├── daemon/ # NEW: decentralized daemon mode
│ │ ├── daemon.go # Main daemon loop + reconciliation
│ │ ├── config.go # Key derivation from secret
│ │ └── peerstore.go # Thread-safe peer store with merge logic
│ ├── discovery/ # NEW: all discovery layers
│ │ ├── lan.go # Layer 1: UDP multicast
│ │ ├── dht.go # Layer 2: BitTorrent Mainline DHT
│ │ ├── gossip.go # Layer 3: in-mesh gossip
│ │ └── exchange.go # Encrypted peer exchange protocol
│ ├── crypto/ # NEW: key derivation + message encryption
│ │ ├── derive.go # HKDF-based derivation from secret
│ │ └── envelope.go # Encrypt/decrypt gossip messages (AES-256-GCM)
│ ├── wireguard/ # Existing: WG config management
│ │ ├── keys.go
│ │ ├── config.go
│ │ ├── apply.go
│ │ └── convert.go
│ └── ssh/ # Existing: remote SSH operations
│ ├── client.go
│ └── wireguard.go
└── mesh-state.json # Existing (for centralized mode)
New Go Dependencies
github.com/anacrolix/dht/v2 # Mainline DHT implementation
golang.org/x/net/ipv4 # Multicast group management
golang.org/x/crypto/hkdf # Key derivation
github.com/skip2/go-qrcode # QR code generation (optional, for wgmesh qr)
Phased Implementation Plan
Phase 1: Foundation (daemon mode + LAN discovery)
Key derivation from secret ( pkg/crypto/derive.go )
Encrypted message envelope ( pkg/crypto/envelope.go )
PeerStore with merge logic ( pkg/daemon/peerstore.go )
Daemon main loop + reconciliation using existing diff engine ( pkg/daemon/daemon.go )
LAN multicast discovery ( pkg/discovery/lan.go )
CLI: wgmesh join --secret , wgmesh status , wgmesh init --secret
Milestone: Two nodes on the same LAN mesh automatically within 5 seconds
Phase 2: Internet discovery (DHT)
DHT bootstrap + announce + get_peers ( pkg/discovery/dht.go )
Encrypted peer exchange protocol ( pkg/discovery/exchange.go )
Transitive peer sharing (known_peers in exchange)
DHT infohash rotation (hourly)
Milestone: Two nodes on different networks mesh automatically within 60 seconds
Phase 3: Convergence optimization (gossip)
In-mesh gossip protocol ( pkg/discovery/gossip.go )
Routable networks propagation via gossip
Dead peer detection + cleanup
QR code generation ( wgmesh qr )
Milestone: 10-node mesh fully converges within 90 seconds
Phase 4: Hardening
Secret rotation protocol
Mesh IP collision resolution
Systemd unit generation ( wgmesh install-service )
Persistent peer cache (survive daemon restart without re-discovering)
Metrics/health endpoint
Open Questions for Discussion
1. DHT port sharing with WireGuard: The DHT needs a UDP port. Should it share
WG’s listen port (complex, requires multiplexing) or use a separate port (simpler,
one more port to open)?
2. STUN integration: Should we add STUN (e.g., stun.l.google.com:19302 ) for nodes
behind NAT to discover their public endpoint? This would improve the quality of
endpoint information shared via DHT. The pion/stun Go library makes this
straightforward.
3. TURN/relay fallback: For nodes behind strict CGNAT where direct WG connections
fail, should there be a relay mode? EasyTier does this, but it adds significant
complexity and requires relay infrastructure.
4. Backwards compatibility: Should daemon mode be able to read/write the existing
mesh-state.json format? This would allow mixed-mode operation where some
nodes are managed centrally and others self-discover.
5. Maximum mesh size: The /16 subnet and collision-based IP allocation work well up
to ~100 nodes. For larger meshes, consider explicit DHCP-like allocation via gossip
consensus. Is >100 nodes a target?
References & Prior Art
Project Approach Relevance
EasyTier Shared name+secret, OSPF routing,
public relay nodes
Closest to target UX; uses relay
servers though
wgautomesh Gossip + LAN broadcast, shared
secret
Layer 1 + 3 reference
implementation
wiresmith Consul KV for peer registry KV-based discovery pattern
wireguard-
dynamic Token → KV store → auto-join Simple token-based UX
STUNMESH-
go
STUN + Curve25519 + pluggable KV STUN integration pattern
NetBird WebRTC ICE + STUN + Signal server Full NAT traversal reference
KadNode Mainline DHT for P2P DNS DHT-as-rendezvous pattern
Weaveworks
Mesh Gossip with shared-secret auth Go gossip library
anacrolix/dht Go Mainline DHT implementation Direct dependency candidate
