# Vortyx Backend Development Roadmap

## Project Overview
The Vortyx backend is a unified monolith platform for MSPs and MSSPs, comprising 16 distinct services across MSP (Managed Services) and MSSP (Managed Security Services) modules, plus a distributed VORT agent.

---

## Phase 1: Foundation & Configuration Fixes
**Timeline**: Week 1 (Days 1-5)
**Priority**: Critical

### Objectives
1. Fix sqlc.yaml path mismatches preventing database code generation
2. Verify build integrity across all modules
3. Establish working development environment

### Deliverables
- Corrected sqlc.yaml configuration
- Verified `go build ./...` passes
- Verified `sqlc generate` works for all 16 services
- Updated AGENTS.md with corrected paths

### Task Breakdown

| Task ID | Description | Dependencies | Effort |
|---------|-------------|--------------|--------|
| 1.1 | Analyze current sqlc.yaml vs actual directory structure | None | 1 hr |
| 1.2 | Update sqlc.yaml paths for MSP services (pulse, pilot, nexus, horizon, control, optic, grid) | 1.1 | 1 hr |
| 1.3 | Update sqlc.yaml paths for MSSP services (radar, guard, shield, mind, probe, reflex, sonar, signal) | 1.1 | 1 hr |
| 1.4 | Update sqlc.yaml path for platform service | 1.1 | 30 min |
| 1.5 | Run sqlc generate and verify all 16 services generate correctly | 1.2, 1.3, 1.4 | 1 hr |
| 1.6 | Run go build ./... and fix any compilation errors | 1.5 | 2 hr |
| 1.7 | Update AGENTS.md documentation with correct paths | 1.6 | 30 min |

### Success Criteria
- [ ] sqlc generate completes without errors for all 16 services
- [ ] go build ./... exits with code 0
- [ ] All generated files are in correct locations
- [ ] Documentation reflects actual paths

### Risk Mitigation
- **Risk**: Breaking existing generated code
  - **Mitigation**: Backup current generated files before regeneration
- **Risk**: Import path conflicts after changes
  - **Mitigation**: Verify all imports resolve correctly

### Resources Required
- 1 Developer
- Go 1.24+
- sqlc CLI

---

## Phase 2: Architecture Improvements
**Timeline**: Week 2 (Days 6-10)
**Priority**: High

### Objectives
1. Create shared utilities package to reduce code duplication
2. Add transaction support for database operations
3. Implement centralized error handling
4. Remove dead code

### Deliverables
- New `internal/pkg/common/` package
- Transaction helper methods in services
- Error handling framework
- Cleaned up server.go

### Task Breakdown

| Task ID | Description | Dependencies | Effort |
|---------|-------------|--------------|--------|
| 2.1 | Create internal/pkg/common/ with validation helpers | Phase 1 complete | 2 hr |
| 2.2 | Create internal/pkg/errors/ with custom error types | Phase 1 complete | 2 hr |
| 2.3 | Add transaction helper to base service struct | 2.2 | 1 hr |
| 2.4 | Refactor api/api.go to use new error framework | 2.2 | 1 hr |
| 2.5 | Remove dead code (platformPkg reference in server.go) | Phase 1 complete | 15 min |
| 2.6 | Externalize hardcoded config (auth domain) | Phase 1 complete | 1 hr |
| 2.7 | Add WithTx method to all 16 services | 2.3 | 2 hr |
| 2.8 | Run tests and verify build | All above | 1 hr |

### Success Criteria
- [ ] Shared package used by at least 3 services
- [ ] Transaction support available in all services
- [ ] Consistent error handling across API layer
- [ ] No dead code in server.go

### Risk Mitigation
- **Risk**: Breaking existing imports
  - **Mitigation**: Run incremental builds after each change

---

## Phase 3: Service Implementation
**Timeline**: Weeks 3-6 (Days 11-30)
**Priority**: High

### Objectives
1. Implement business logic for all MSP services
2. Implement business logic for all MSSP services
3. Add request validation
4. Implement per-service health checks

### Deliverables
- Functional MSP services (7 services)
- Functional MSSP services (8 services)
- Request validation layer
- Service-specific health endpoints

### Task Breakdown

#### Week 3: MSP Core Services
| Task ID | Description | Dependencies | Effort |
|---------|-------------|--------------|--------|
| 3.1 | Implement Pulse service (RMM - metrics, monitoring) | Phase 2 complete | 4 hr |
| 3.2 | Implement Pilot service (PSA - ticketing) | Phase 2 complete | 4 hr |
| 3.3 | Implement Nexus service (CMDB - asset tracking) | Phase 2 complete | 4 hr |

#### Week 4: MSP Extended + Platform
| Task ID | Description | Dependencies | Effort |
|---------|-------------|--------------|--------|
| 3.4 | Implement Horizon service (vCIO - planning) | Phase 2 complete | 4 hr |
| 3.5 | Implement Control service (SaaS management) | Phase 2 complete | 4 hr |
| 3.6 | Implement Optic service (CCTV surveillance) | Phase 2 complete | 4 hr |
| 3.7 | Implement Grid service (Network orchestration) | Phase 2 complete | 4 hr |

#### Week 5: MSSP Core Services
| Task ID | Description | Dependencies | Effort |
|---------|-------------|--------------|--------|
| 3.8 | Implement Radar service (SIEM) | Phase 2 complete | 4 hr |
| 3.9 | Implement Guard service (EDR) | Phase 2 complete | 4 hr |
| 3.10 | Implement Shield service (GRC) | Phase 2 complete | 4 hr |

#### Week 6: MSSP Extended + Validation
| Task ID | Description | Dependencies | Effort |
|---------|-------------|--------------|--------|
| 3.11 | Implement Mind service (Training) | Phase 2 complete | 4 hr |
| 3.12 | Implement Probe service (Vulnerability scanner) | Phase 2 complete | 4 hr |
| 3.13 | Implement Reflex service (SOAR) | Phase 2 complete | 4 hr |
| 3.14 | Implement Sonar service (NDR) | Phase 2 complete | 4 hr |
| 3.15 | Implement Signal service (CTI) | Phase 2 complete | 4 hr |
| 3.16 | Add request validation to all services | Phase 2 complete | 4 hr |
| 3.17 | Add per-service health endpoints | Phase 2 complete | 2 hr |

### Success Criteria
- [ ] All 15 services (7 MSP + 8 MSSP) have CRUD operations
- [ ] Each service has at least 3 working RPCs
- [ ] Request validation rejects invalid input
- [ ] Health endpoints return service-specific status

---

## Phase 4: VORT Agent Integration & Testing
**Timeline**: Weeks 7-8 (Days 31-40)
**Priority**: High

### Objectives
1. Define protobuf contracts for agent-backend communication
2. Implement ConnectRPC handlers for agent
3. Implement end-to-end integration tests
4. Performance testing

### Deliverables
- Agent protobuf definitions
- ConnectRPC handlers in vort module
- Integration test suite
- Performance benchmarks

### Task Breakdown

| Task ID | Description | Dependencies | Effort |
|---------|-------------|--------------|--------|
| 4.1 | Define proto/vort/agent/v1/service.proto | Phase 3 complete | 4 hr |
| 4.2 | Generate Go client/server code | 4.1 | 1 hr |
| 4.3 | Implement agent registration handler | 4.2 | 2 hr |
| 4.4 | Implement agent command dispatch | 4.2 | 4 hr |
| 4.5 | Implement agent telemetry ingestion | 4.2 | 4 hr |
| 4.6 | Create integration test suite | 4.3, 4.4, 4.5 | 8 hr |
| 4.7 | Run load testing | 4.6 | 4 hr |
| 4.8 | Document API contracts | All above | 2 hr |

### Success Criteria
- [ ] Agent can register via ConnectRPC
- [ ] Commands dispatch to agent successfully
- [ ] Telemetry stored in database
- [ ] Integration tests pass
- [ ] Load test handles 1000+ concurrent agents

---

## Phase 5: Production Readiness
**Timeline**: Weeks 9-10 (Days 41-50)
**Priority**: Medium

### Objectives
1. Logging and observability
2. Rate limiting
3. API documentation
4. Deployment scripts

### Deliverables
- Structured logging throughout
- Rate limiting middleware
- OpenAPI/Swagger docs
- Docker Compose for production
- Kubernetes manifests (optional)

### Task Breakdown

| Task ID | Description | Dependencies | Effort |
|---------|-------------|--------------|--------|
| 5.1 | Add structured logging to all services | Phase 4 complete | 4 hr |
| 5.2 | Implement rate limiting | Phase 4 complete | 3 hr |
| 5.3 | Generate OpenAPI docs | Phase 4 complete | 2 hr |
| 5.4 | Create production Docker Compose | Phase 4 complete | 2 hr |
| 5.5 | Add Prometheus metrics | Phase 4 complete | 3 hr |
| 5.6 | Create deployment runbook | All above | 2 hr |

---

## Summary Timeline

| Phase | Duration | Focus |
|-------|----------|-------|
| Phase 1 | Week 1 | Foundation & Configuration |
| Phase 2 | Week 2 | Architecture Improvements |
| Phase 3 | Weeks 3-6 | Service Implementation |
| Phase 4 | Weeks 7-8 | Agent Integration & Testing |
| Phase 5 | Weeks 9-10 | Production Readiness |

**Total Estimated Timeline**: 10 weeks (50 days)
**Total Estimated Effort**: ~200 developer hours
