# Bootstrap Feature - Quick Reference

**Full Plan:** See [IMPLEMENTATION_PLAN.md](./IMPLEMENTATION_PLAN.md)  
**Specification:** See [bootstrap.md](./bootstrap.md)

## Phase Summary

| Phase | Duration | Focus | Status |
|-------|----------|-------|--------|
| 1 | Weeks 1-3 | Registry + Foundation | ⏸️ Not Started |
| 1.5 | Week 4 | LAN Multicast | ⏸️ Not Started |
| 2 | Weeks 5-7 | Privacy Features | ⏸️ Not Started |
| 3 | Weeks 8-9 | In-Mesh Gossip | ⏸️ Not Started |
| 4 | Weeks 10-12 | Advanced Features | ⏸️ Not Started |
| 5 | Week 13 | Testing & Release | ⏸️ Not Started |

## Implementation Status

### ✅ Already Implemented (Pre-Work)

- Core daemon infrastructure
- Key derivation (basic)
- DHT discovery (Layer 2)
- Encrypted peer exchange
- CLI commands: join, init, status, test-peer
- WireGuard integration

### 🚧 In Progress

(None - waiting for Phase 1 start)

### ❌ Pending Implementation

#### Phase 1 (Critical Path)

- [ ] **registry.go** - GitHub Issues discovery (Layer 0)
- [ ] **membership.go** - Token generation/validation
- [ ] **derive.go enhancements** - Rotating infohash, new keys
- [ ] **daemon.go** - Bootstrap chain integration
- [ ] **lan.go** - Multicast discovery (Layer 1)

#### Phase 2 (Privacy)

- [ ] **dandelion.go** - Stem/fluff relay
- [ ] **epoch.go** - Relay rotation
- [ ] DHT modifications for rotating infohash

#### Phase 3 (Convergence)

- [ ] **gossip.go** - In-mesh protocol (Layer 3)

#### Phase 4 (Polish)

- [ ] **qr command** - QR code generation
- [ ] **collision.go** - IP collision handling
- [ ] **systemd.go** - Service installation
- [ ] **cache.go** - Persistent peer cache
- [ ] **rotation.go** - Secret rotation

## Files to Create/Modify

### New Files (16)

```
pkg/
├── crypto/
│   ├── membership.go          # Phase 1
│   └── rotation.go            # Phase 4
├── daemon/
│   ├── epoch.go               # Phase 2
│   ├── collision.go           # Phase 4
│   ├── cache.go               # Phase 4
│   └── systemd.go             # Phase 4
├── discovery/
│   ├── registry.go            # Phase 1
│   ├── lan.go                 # Phase 1.5
│   └── gossip.go              # Phase 3
└── privacy/
    └── dandelion.go           # Phase 2
```

### Files to Modify (3)

```
pkg/
├── crypto/
│   └── derive.go              # Phase 1 - Add rotating IDs
├── daemon/
│   └── daemon.go              # Phase 1 - Bootstrap chain
└── discovery/
    └── dht.go                 # Phase 2 - Rotating infohash
```

## Critical Dependencies

### Go Packages (New)

- **Phase 4 only:** `github.com/skip2/go-qrcode`
- All other features use stdlib or existing dependencies

### System Requirements

- Go 1.23+
- WireGuard kernel support
- Root/sudo for testing
- (Optional) GITHUB_TOKEN for registry creation

## Key Milestones

### M1: Registry Bootstrap ✓ (End Week 3)
- Registry discovery working
- First/second node can bootstrap
- Works behind firewalls

### M2: Complete Discovery ✓ (End Week 4)
- All 4 discovery layers operational
- LAN + DHT + Registry working
- Offline LAN mesh functional

### M3: Privacy Features ✓ (End Week 7)
- Dandelion++ relay working
- Origin IP not visible in DHT
- Epoch rotation implemented

### M4: Gossip & Convergence ✓ (End Week 9)
- Transitive discovery works
- 10-node mesh converges <90s
- Routable networks propagate

### M5: Production Ready ✓ (End Week 12)
- All features complete
- 80%+ test coverage
- Documentation complete

## Testing Checklist

### Unit Tests (Per Phase)

- [ ] Phase 1: registry, membership, derive
- [ ] Phase 2: dandelion, epoch
- [ ] Phase 3: gossip
- [ ] Phase 4: collision, cache, qr

### Integration Tests

- [ ] Two-node bootstrap (registry)
- [ ] LAN discovery (<5s)
- [ ] Three-node transitive (gossip)
- [ ] Privacy mode (no direct DHT)
- [ ] Collision resolution

### Manual Tests

- [ ] First-node creates registry
- [ ] Second-node finds registry
- [ ] LAN discovery same subnet
- [ ] DHT discovery different networks
- [ ] Mesh convergence (3+ nodes)
- [ ] Hour boundary (infohash rotation)
- [ ] Service install/restart
- [ ] QR code generation

## Quick Commands

```bash
# Build
go build -o wgmesh

# Test all
go test ./...

# Test with coverage
go test -cover ./... -coverprofile=coverage.out
go tool cover -html=coverage.out

# Lint
golangci-lint run

# Integration tests (needs root)
sudo go test -tags=integration ./pkg/discovery/...
```

## Development Workflow

1. **Create branch:** `git checkout -b feature/registry-discovery`
2. **Implement:** Follow IMPLEMENTATION_PLAN.md Phase details
3. **Test:** Unit tests + integration tests
4. **Document:** Update godoc comments
5. **Review:** Self-review checklist
6. **PR:** Link to implementation plan
7. **Merge:** After CI passes and review approved

## Design Decisions (Quick Ref)

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Registry | GitHub Issues | Zero infra, works behind FW |
| Privacy default | Opt-in → Default | Stability first, then security |
| DHT port | Separate | Simpler, safer |
| Max mesh size | 100 nodes | Collision rate acceptable |
| STUN | Optional | Add if needed |
| Garlic | No | Overkill for VPN |
| Floodfill | No | Registry + DHT sufficient |

## Risk Mitigation

### Top 3 Risks

1. **DHT unreliable** → Mitigation: Registry + LAN fallback
2. **NAT traversal fails** → Mitigation: Document STUN config
3. **Rate limits (GitHub)** → Mitigation: Cache peers, use other layers

## Performance Targets

- First node bootstrap: <10s
- Second node discovery: <60s
- LAN discovery: <5s
- 10-node convergence: <90s
- CPU idle: <1%
- Memory: <50MB
- DHT traffic: <100KB/hour
- Gossip traffic: <10KB/minute

## Success Criteria (Release)

- [ ] All 4 discovery layers working
- [ ] Privacy features operational
- [ ] 80%+ test coverage
- [ ] Zero critical security issues
- [ ] Documentation complete
- [ ] Performance targets met
- [ ] 10+ manual test scenarios pass

## Weekly Progress Tracking

```markdown
### Week 1
- [ ] Registry discovery implementation
- [ ] Unit tests for registry
- [ ] GitHub API integration working

### Week 2
- [ ] Enhanced key derivation
- [ ] Membership token implementation
- [ ] Token validation tests

### Week 3
- [ ] Bootstrap chain integration
- [ ] End-to-end registry test
- [ ] Milestone 1 complete

(Continue for all 13 weeks...)
```

## Resources

- **Spec:** [bootstrap.md](./bootstrap.md) - 1305 lines, comprehensive
- **Plan:** [IMPLEMENTATION_PLAN.md](./IMPLEMENTATION_PLAN.md) - This document
- **DevSwarm:** https://github.com/HackrsValv/devswarm - Registry inspiration
- **Dandelion++:** https://github.com/bitcoin/bips/blob/master/bip-0156.mediawiki
- **DHT Library:** https://github.com/anacrolix/dht

---

**Last Updated:** 2026-02-07  
**Status:** Ready for Phase 1 Start  
**Next Review:** Start of Phase 1
