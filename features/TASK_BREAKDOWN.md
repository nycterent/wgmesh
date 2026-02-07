# Bootstrap Feature - Task Breakdown

This file provides a detailed task breakdown for GitHub Projects or similar tracking tools.

## Phase 1: Foundation & Registry Bootstrap

### Epic 1.1: Registry Rendezvous Discovery

**Priority:** P0 (Critical Path)  
**Estimate:** 1.5 weeks  
**Dependencies:** None

#### Tasks

- [ ] **Task 1.1.1:** Create `pkg/discovery/registry.go` skeleton
  - Assignee: TBD
  - Estimate: 2 hours
  - Description: Define RendezvousRegistry struct and interfaces
  
- [ ] **Task 1.1.2:** Implement GitHub Issue search API
  - Assignee: TBD
  - Estimate: 4 hours
  - Description: HTTP client for GitHub search API, handle rate limits
  - Acceptance: Can search for issues by title
  
- [ ] **Task 1.1.3:** Implement encrypted peer list format
  - Assignee: TBD
  - Estimate: 3 hours
  - Description: Base64(Encrypt(gossip_key, json([PeerInfo, ...])))
  - Acceptance: Can encrypt/decrypt peer lists
  
- [ ] **Task 1.1.4:** Implement issue creation (first node)
  - Assignee: TBD
  - Estimate: 4 hours
  - Description: Create issue via API, handle auth
  - Acceptance: First node creates registry entry
  
- [ ] **Task 1.1.5:** Implement issue update (append peer)
  - Assignee: TBD
  - Estimate: 3 hours
  - Description: Edit issue to add new peer
  - Acceptance: Subsequent nodes update registry
  
- [ ] **Task 1.1.6:** Add retry and error handling
  - Assignee: TBD
  - Estimate: 2 hours
  - Description: Exponential backoff, fallback to DHT
  
- [ ] **Task 1.1.7:** Write unit tests
  - Assignee: TBD
  - Estimate: 4 hours
  - Description: Mock HTTP client, test all paths
  - Acceptance: 80%+ coverage
  
- [ ] **Task 1.1.8:** Write integration test
  - Assignee: TBD
  - Estimate: 3 hours
  - Description: Real GitHub API test (with test token)
  - Acceptance: Can create and find test issue

### Epic 1.2: Enhanced Key Derivation

**Priority:** P0 (Critical Path)  
**Estimate:** 0.5 weeks  
**Dependencies:** None

#### Tasks

- [ ] **Task 1.2.1:** Add RendezvousID derivation
  - Assignee: TBD
  - Estimate: 1 hour
  - Description: SHA256(secret || "rv")[0:8]
  
- [ ] **Task 1.2.2:** Add MembershipKey derivation
  - Assignee: TBD
  - Estimate: 1 hour
  - Description: HKDF(secret, "wgmesh-membership-v1", 32)
  
- [ ] **Task 1.2.3:** Add EpochSeed derivation
  - Assignee: TBD
  - Estimate: 1 hour
  - Description: HKDF(secret, "wgmesh-epoch-v1", 32)
  
- [ ] **Task 1.2.4:** Implement rotating network IDs
  - Assignee: TBD
  - Estimate: 2 hours
  - Description: CurrentNetworkIDs(secret) returns current + previous hour
  - Acceptance: IDs rotate at hour boundary
  
- [ ] **Task 1.2.5:** Write unit tests
  - Assignee: TBD
  - Estimate: 2 hours
  - Description: Test determinism, rotation
  - Acceptance: 100% coverage for crypto package
  
- [ ] **Task 1.2.6:** Update DerivedKeys struct
  - Assignee: TBD
  - Estimate: 1 hour
  - Description: Add new fields, update documentation

### Epic 1.3: Membership Token Authentication

**Priority:** P0 (Critical Path)  
**Estimate:** 1 week  
**Dependencies:** Epic 1.2

#### Tasks

- [ ] **Task 1.3.1:** Create `pkg/crypto/membership.go`
  - Assignee: TBD
  - Estimate: 1 hour
  
- [ ] **Task 1.3.2:** Implement GenerateMembershipToken()
  - Assignee: TBD
  - Estimate: 2 hours
  - Description: HMAC(membershipKey, pubkey || hourBucket)
  
- [ ] **Task 1.3.3:** Implement ValidateMembershipToken()
  - Assignee: TBD
  - Estimate: 2 hours
  - Description: Check current + previous hour, constant-time compare
  
- [ ] **Task 1.3.4:** Add token to PeerExchange HELLO message
  - Assignee: TBD
  - Estimate: 2 hours
  - Description: Update exchange.go protocol
  
- [ ] **Task 1.3.5:** Add token validation to HandlePeerExchange()
  - Assignee: TBD
  - Estimate: 3 hours
  - Description: Silent reject on invalid token
  - Acceptance: Unauthenticated peers rejected
  
- [ ] **Task 1.3.6:** Write unit tests
  - Assignee: TBD
  - Estimate: 3 hours
  - Description: Test generation, validation, clock skew
  - Acceptance: 100% coverage
  
- [ ] **Task 1.3.7:** Write integration test
  - Assignee: TBD
  - Estimate: 2 hours
  - Description: Two nodes exchange with tokens

### Epic 1.4: Bootstrap Discovery Chain

**Priority:** P0 (Critical Path)  
**Estimate:** 1 week  
**Dependencies:** Epics 1.1, 1.2, 1.3

#### Tasks

- [ ] **Task 1.4.1:** Implement Bootstrap() in daemon.go
  - Assignee: TBD
  - Estimate: 3 hours
  - Description: Layer 0 → Layer 2 fallback chain
  
- [ ] **Task 1.4.2:** Implement peer deduplication
  - Assignee: TBD
  - Estimate: 2 hours
  - Description: Deduplicate by pubkey
  
- [ ] **Task 1.4.3:** Add retry with exponential backoff
  - Assignee: TBD
  - Estimate: 2 hours
  
- [ ] **Task 1.4.4:** Add metrics/logging
  - Assignee: TBD
  - Estimate: 2 hours
  - Description: Log discovery source for each peer
  
- [ ] **Task 1.4.5:** Update daemon Run() to call Bootstrap()
  - Assignee: TBD
  - Estimate: 1 hour
  
- [ ] **Task 1.4.6:** Write integration test
  - Assignee: TBD
  - Estimate: 4 hours
  - Description: Two nodes, registry → DHT fallback
  - Acceptance: First node creates registry, second finds it

**Milestone 1:** Registry Bootstrap Complete (End of Week 3)

---

## Phase 1.5: LAN Multicast Discovery

### Epic 1.5: LAN Multicast

**Priority:** P0 (Critical Path)  
**Estimate:** 1 week  
**Dependencies:** Phase 1 complete

#### Tasks

- [ ] **Task 1.5.1:** Create `pkg/discovery/lan.go` skeleton
  - Assignee: TBD
  - Estimate: 1 hour
  
- [ ] **Task 1.5.2:** Implement multicast group join
  - Assignee: TBD
  - Estimate: 3 hours
  - Description: Use golang.org/x/net/ipv4
  - Acceptance: Can join multicast group
  
- [ ] **Task 1.5.3:** Implement encrypted announcement format
  - Assignee: TBD
  - Estimate: 2 hours
  - Description: Encrypt(gossip_key, PeerInfo)
  
- [ ] **Task 1.5.4:** Implement Announce() - send multicast
  - Assignee: TBD
  - Estimate: 2 hours
  - Description: Broadcast every 5 seconds
  
- [ ] **Task 1.5.5:** Implement Listen() - receive multicast
  - Assignee: TBD
  - Estimate: 3 hours
  - Description: Decrypt and validate announcements
  
- [ ] **Task 1.5.6:** Add IPv6 support (ff02::1)
  - Assignee: TBD
  - Estimate: 2 hours
  
- [ ] **Task 1.5.7:** Integrate with daemon bootstrap
  - Assignee: TBD
  - Estimate: 2 hours
  - Description: Add Layer 1 to bootstrap chain
  
- [ ] **Task 1.5.8:** Write unit tests
  - Assignee: TBD
  - Estimate: 3 hours
  - Description: Mock UDP connections
  
- [ ] **Task 1.5.9:** Write integration test
  - Assignee: TBD
  - Estimate: 3 hours
  - Description: Two nodes on same network
  - Acceptance: Discovery in <5 seconds

**Milestone 2:** Complete Discovery (End of Week 4)

---

## Phase 2: Privacy Enhancements

### Epic 2.1: Rotating DHT Infohash

**Priority:** P1  
**Estimate:** 0.5 weeks  
**Dependencies:** Task 1.2.4 (rotating IDs implemented)

#### Tasks

- [ ] **Task 2.1.1:** Update dht.go to use CurrentNetworkIDs()
  - Assignee: TBD
  - Estimate: 2 hours
  
- [ ] **Task 2.1.2:** Modify announceLoop() - announce to current only
  - Assignee: TBD
  - Estimate: 1 hour
  
- [ ] **Task 2.1.3:** Modify queryLoop() - query current + previous
  - Assignee: TBD
  - Estimate: 2 hours
  
- [ ] **Task 2.1.4:** Add logging for infohash rotation
  - Assignee: TBD
  - Estimate: 1 hour
  
- [ ] **Task 2.1.5:** Write integration test
  - Assignee: TBD
  - Estimate: 3 hours
  - Description: Test hour boundary transition

### Epic 2.2: Epoch-Based Relay Selection

**Priority:** P1  
**Estimate:** 1 week  
**Dependencies:** Task 1.2.3 (EpochSeed)

#### Tasks

- [ ] **Task 2.2.1:** Create `pkg/daemon/epoch.go`
  - Assignee: TBD
  - Estimate: 1 hour
  
- [ ] **Task 2.2.2:** Define Epoch struct
  - Assignee: TBD
  - Estimate: 1 hour
  
- [ ] **Task 2.2.3:** Implement RotateEpoch()
  - Assignee: TBD
  - Estimate: 3 hours
  - Description: Deterministic peer selection using HMAC
  
- [ ] **Task 2.2.4:** Add epoch rotation timer (10 minutes)
  - Assignee: TBD
  - Estimate: 2 hours
  
- [ ] **Task 2.2.5:** Handle small meshes (<2 peers)
  - Assignee: TBD
  - Estimate: 1 hour
  
- [ ] **Task 2.2.6:** Add epoch to daemon state
  - Assignee: TBD
  - Estimate: 1 hour
  
- [ ] **Task 2.2.7:** Write unit tests
  - Assignee: TBD
  - Estimate: 3 hours

### Epic 2.3: Dandelion++ Implementation

**Priority:** P1  
**Estimate:** 1.5 weeks  
**Dependencies:** Epic 2.2

#### Tasks

- [ ] **Task 2.3.1:** Create `pkg/privacy/dandelion.go`
  - Assignee: TBD
  - Estimate: 1 hour
  
- [ ] **Task 2.3.2:** Define DandelionAnnounce struct
  - Assignee: TBD
  - Estimate: 1 hour
  
- [ ] **Task 2.3.3:** Implement HandleDandelionAnnounce()
  - Assignee: TBD
  - Estimate: 4 hours
  - Description: Stem/fluff decision logic
  
- [ ] **Task 2.3.4:** Implement ShouldFluff() - 10% probability
  - Assignee: TBD
  - Estimate: 1 hour
  
- [ ] **Task 2.3.5:** Implement RelayToStem()
  - Assignee: TBD
  - Estimate: 2 hours
  - Description: Forward to epoch relay peer
  
- [ ] **Task 2.3.6:** Implement FluffToDHT()
  - Assignee: TBD
  - Estimate: 2 hours
  - Description: Announce using relay's IP
  
- [ ] **Task 2.3.7:** Add max hops enforcement (4 hops)
  - Assignee: TBD
  - Estimate: 1 hour
  
- [ ] **Task 2.3.8:** Integrate with daemon announce flow
  - Assignee: TBD
  - Estimate: 3 hours
  
- [ ] **Task 2.3.9:** Add --privacy CLI flag
  - Assignee: TBD
  - Estimate: 1 hour
  
- [ ] **Task 2.3.10:** Write unit tests
  - Assignee: TBD
  - Estimate: 4 hours
  
- [ ] **Task 2.3.11:** Write integration test
  - Assignee: TBD
  - Estimate: 4 hours
  - Description: 3+ nodes, verify origin IP not in DHT

**Milestone 3:** Privacy Features Complete (End of Week 7)

---

## Phase 3: In-Mesh Gossip

### Epic 3.1: Gossip Protocol

**Priority:** P0 (Critical Path)  
**Estimate:** 2 weeks  
**Dependencies:** Phase 2 complete

#### Tasks

- [ ] **Task 3.1.1:** Create `pkg/discovery/gossip.go`
  - Assignee: TBD
  - Estimate: 1 hour
  
- [ ] **Task 3.1.2:** Define MeshGossip struct
  - Assignee: TBD
  - Estimate: 1 hour
  
- [ ] **Task 3.1.3:** Implement UDP socket binding (derived port)
  - Assignee: TBD
  - Estimate: 2 hours
  
- [ ] **Task 3.1.4:** Define gossip message format
  - Assignee: TBD
  - Estimate: 2 hours
  - Description: Encrypted peer list
  
- [ ] **Task 3.1.5:** Implement Start() - start listeners
  - Assignee: TBD
  - Estimate: 2 hours
  
- [ ] **Task 3.1.6:** Implement GossipLoop()
  - Assignee: TBD
  - Estimate: 3 hours
  - Description: Every 10s, random peer exchange
  
- [ ] **Task 3.1.7:** Implement ExchangeWithRandomPeer()
  - Assignee: TBD
  - Estimate: 3 hours
  
- [ ] **Task 3.1.8:** Implement message handling
  - Assignee: TBD
  - Estimate: 3 hours
  - Description: Receive, decrypt, merge into peerstore
  
- [ ] **Task 3.1.9:** Add transitive discovery support
  - Assignee: TBD
  - Estimate: 2 hours
  - Description: Include known_peers in exchange
  
- [ ] **Task 3.1.10:** Integrate with daemon
  - Assignee: TBD
  - Estimate: 2 hours
  
- [ ] **Task 3.1.11:** Write unit tests
  - Assignee: TBD
  - Estimate: 4 hours
  
- [ ] **Task 3.1.12:** Write integration test
  - Assignee: TBD
  - Estimate: 4 hours
  - Description: A knows B, B knows C → A discovers C
  - Acceptance: Transitive discovery works

**Milestone 4:** Gossip & Convergence Complete (End of Week 9)

---

## Phase 4: Advanced Features

### Epic 4.1: QR Code Generation

**Priority:** P2  
**Estimate:** 0.5 weeks  
**Dependencies:** None

#### Tasks

- [ ] **Task 4.1.1:** Add go-qrcode dependency
  - Assignee: TBD
  - Estimate: 0.5 hours
  
- [ ] **Task 4.1.2:** Implement qrCmd() in main.go
  - Assignee: TBD
  - Estimate: 2 hours
  
- [ ] **Task 4.1.3:** Add UTF-8 terminal output
  - Assignee: TBD
  - Estimate: 2 hours
  
- [ ] **Task 4.1.4:** Add PNG file output option
  - Assignee: TBD
  - Estimate: 1 hour
  
- [ ] **Task 4.1.5:** Integrate with init command
  - Assignee: TBD
  - Estimate: 1 hour
  - Description: wgmesh init --secret --qr

### Epic 4.2: Collision Resolution

**Priority:** P2  
**Estimate:** 1 week  
**Dependencies:** None

#### Tasks

- [ ] **Task 4.2.1:** Create `pkg/daemon/collision.go`
  - Assignee: TBD
  - Estimate: 1 hour
  
- [ ] **Task 4.2.2:** Implement DetectCollision() in peerstore
  - Assignee: TBD
  - Estimate: 2 hours
  
- [ ] **Task 4.2.3:** Implement DeterministicWinner()
  - Assignee: TBD
  - Estimate: 2 hours
  - Description: Lexicographic comparison
  
- [ ] **Task 4.2.4:** Implement ResolveCollision() in daemon
  - Assignee: TBD
  - Estimate: 3 hours
  - Description: Loser re-derives with nonce
  
- [ ] **Task 4.2.5:** Add collision detection to reconciliation loop
  - Assignee: TBD
  - Estimate: 2 hours
  
- [ ] **Task 4.2.6:** Write unit tests
  - Assignee: TBD
  - Estimate: 3 hours
  
- [ ] **Task 4.2.7:** Write integration test
  - Assignee: TBD
  - Estimate: 3 hours
  - Description: Force collision, verify resolution

### Epic 4.3: Systemd Service

**Priority:** P2  
**Estimate:** 1 week  
**Dependencies:** None

#### Tasks

- [ ] **Task 4.3.1:** Create `pkg/daemon/systemd.go`
  - Assignee: TBD
  - Estimate: 1 hour
  
- [ ] **Task 4.3.2:** Implement service file generation
  - Assignee: TBD
  - Estimate: 3 hours
  
- [ ] **Task 4.3.3:** Implement install-service command
  - Assignee: TBD
  - Estimate: 2 hours
  
- [ ] **Task 4.3.4:** Implement service enable/start
  - Assignee: TBD
  - Estimate: 2 hours
  
- [ ] **Task 4.3.5:** Implement uninstall command
  - Assignee: TBD
  - Estimate: 1 hour
  
- [ ] **Task 4.3.6:** Write integration test
  - Assignee: TBD
  - Estimate: 2 hours

### Epic 4.4: Persistent Peer Cache

**Priority:** P2  
**Estimate:** 0.5 weeks  
**Dependencies:** None

#### Tasks

- [ ] **Task 4.4.1:** Create `pkg/daemon/cache.go`
  - Assignee: TBD
  - Estimate: 1 hour
  
- [ ] **Task 4.4.2:** Implement LoadPeerCache()
  - Assignee: TBD
  - Estimate: 2 hours
  
- [ ] **Task 4.4.3:** Implement SavePeerCache()
  - Assignee: TBD
  - Estimate: 2 hours
  
- [ ] **Task 4.4.4:** Add periodic save (5 minutes)
  - Assignee: TBD
  - Estimate: 1 hour
  
- [ ] **Task 4.4.5:** Implement cache expiration (24 hours)
  - Assignee: TBD
  - Estimate: 2 hours
  
- [ ] **Task 4.4.6:** Integrate with daemon startup
  - Assignee: TBD
  - Estimate: 1 hour

### Epic 4.5: Secret Rotation

**Priority:** P3  
**Estimate:** 1 week  
**Dependencies:** Epic 3.1 (needs gossip)

#### Tasks

- [ ] **Task 4.5.1:** Create `pkg/crypto/rotation.go`
  - Assignee: TBD
  - Estimate: 1 hour
  
- [ ] **Task 4.5.2:** Define rotation message format
  - Assignee: TBD
  - Estimate: 2 hours
  
- [ ] **Task 4.5.3:** Implement rotation announcement
  - Assignee: TBD
  - Estimate: 3 hours
  
- [ ] **Task 4.5.4:** Implement dual-secret mode
  - Assignee: TBD
  - Estimate: 4 hours
  
- [ ] **Task 4.5.5:** Implement coordinated switchover
  - Assignee: TBD
  - Estimate: 4 hours
  
- [ ] **Task 4.5.6:** Implement rotate-secret command
  - Assignee: TBD
  - Estimate: 2 hours
  
- [ ] **Task 4.5.7:** Write integration test
  - Assignee: TBD
  - Estimate: 4 hours

**Milestone 5:** Production Ready (End of Week 12)

---

## Phase 5: Testing & Documentation

### Epic 5.1: Integration Testing

**Priority:** P0  
**Estimate:** 1 week  
**Dependencies:** All features complete

#### Tasks

- [ ] **Task 5.1.1:** Two-node bootstrap test
- [ ] **Task 5.1.2:** LAN discovery test
- [ ] **Task 5.1.3:** Three-node transitive test
- [ ] **Task 5.1.4:** Privacy mode test
- [ ] **Task 5.1.5:** Collision resolution test
- [ ] **Task 5.1.6:** 10-node convergence test
- [ ] **Task 5.1.7:** Hour boundary rotation test
- [ ] **Task 5.1.8:** Service restart test

### Epic 5.2: Documentation

**Priority:** P0  
**Estimate:** Week 13  
**Dependencies:** Testing complete

#### Tasks

- [ ] **Task 5.2.1:** Update README.md
- [ ] **Task 5.2.2:** Create tutorial: Your First Mesh
- [ ] **Task 5.2.3:** Write Privacy Guide
- [ ] **Task 5.2.4:** Write Deployment Guide
- [ ] **Task 5.2.5:** Create ARCHITECTURE.md
- [ ] **Task 5.2.6:** Update CONTRIBUTING.md
- [ ] **Task 5.2.7:** Create API.md
- [ ] **Task 5.2.8:** Update CHANGELOG.md

---

## Summary Statistics

**Total Epics:** 16  
**Total Tasks:** 150+  
**Total Estimate:** 13 weeks  
**Critical Path:** Phase 1 → 1.5 → 3 (Registry → LAN → Gossip)

**Priority Distribution:**
- P0 (Critical): 8 epics
- P1 (High): 3 epics
- P2 (Medium): 4 epics
- P3 (Low): 1 epic

**Dependencies Map:**
```
Phase 1 (Foundation)
  ├─ 1.1 Registry
  ├─ 1.2 Key Derivation
  ├─ 1.3 Membership Tokens (depends on 1.2)
  └─ 1.4 Bootstrap Chain (depends on 1.1, 1.2, 1.3)

Phase 1.5 (LAN)
  └─ 1.5 LAN Multicast (depends on Phase 1)

Phase 2 (Privacy)
  ├─ 2.1 Rotating Infohash (depends on 1.2.4)
  ├─ 2.2 Epoch (depends on 1.2.3)
  └─ 2.3 Dandelion (depends on 2.2)

Phase 3 (Gossip)
  └─ 3.1 Gossip Protocol (depends on Phase 2)

Phase 4 (Advanced)
  ├─ 4.1 QR Code (no dependencies)
  ├─ 4.2 Collision (no dependencies)
  ├─ 4.3 Systemd (no dependencies)
  ├─ 4.4 Cache (no dependencies)
  └─ 4.5 Rotation (depends on 3.1)

Phase 5 (Testing)
  ├─ 5.1 Integration Tests (depends on all features)
  └─ 5.2 Documentation (depends on 5.1)
```
