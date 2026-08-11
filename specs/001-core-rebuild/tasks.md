# Tasks: Core Rebuild

**Input**: Design documents from `/specs/001-core-rebuild/`

**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, contracts/

**Tests**: Test tasks are included — the specification requires verifiable acceptance scenarios, and the constitution (§I Quality Gate) mandates tests + lint + security on every change.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1–US5)
- Exact file paths are included in every description

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Verify the baseline state and project infrastructure before any work begins

- [ ] T001 Verify baseline: `go build ./...`, `make check` (golangci-lint + gosec + go test) passes on the current committed state
- [ ] T002 [P] Confirm the gitignore contract covers runtime artifacts (`.buenosaires/`, `plugins/*/assets/`, `*.log`) in `.gitignore` and remove stray tracked files (`probe.log`, `.tmp`) from the repository
- [ ] T003 [P] Confirm toolchain fidelity: `go.mod` directive `go 1.24.3`, `.golangci.yml` v2 format with `govet.inline` disabled, CI pins golangci-lint v2.12.2 + gosec v2.28.0 in `.github/workflows/ci.yml`

**Checkpoint**: Baseline verified green — implementation phases can begin.

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story work

**⚠️ CRITICAL**: No user story work can begin until this phase is complete and verified.

- [ ] T004 Implement config validation per `contracts/config-contract.md` in `internal/config/config.go`: non-empty `user`/`branch`/`repository_url`, `gui.port` range 1–65535 (warn + default 9099 when invalid)
- [ ] T005 [P] Harden `LoadStatus` corrupt-JSON handling in `internal/status/status.go`: unparsable `status.json` must NOT silently wipe other state — log + preserve backup or start empty with warning (per decision documented in status-contract)
- [ ] T006 [P] Add config precedence tests in `internal/config/config_test.go`: repo `user`/`log_dir`/`allow_sudo` override global when set; empty repo values fall through to global; `folder_to_scan` defaults to plugin name
- [ ] T007 Add `internal/status/status_test.go`: load/save round-trip, nil `scripts` map guard, traversal rejection via `sanitizeRepoPath`, corrupt-JSON recovery path, file permissions (0600 status, 0750 dir)
- [ ] T008 Fix `Duration` JSON round-trip in `plugins/shell/duration.go`: string form only (`1.2s`); numeric JSON input must error, never be interpreted as nanoseconds (finding I1); add round-trip test in `plugins/shell/shell_test.go`

**Checkpoint**: Foundation ready — config validation, status resilience, and duration contract verified by tests.

---

## Phase 3: User Story 1 - Operator installs and configures the tool (Priority: P1) 🎯 MVP

**Goal**: Guided setup writes a valid global configuration; the monitor fails clearly when configuration is missing/invalid

**Independent Test**: Run the setup flow on a fresh HOME and verify `~/.buenosaires/config.toml` contents; start the monitor without config and observe a clear error.

### Tests for User Story 1

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [ ] T009 [P] [US1] Test install command in `cmd/install_test.go`: mocked stdin prompts → `SaveGlobalConfig` called with expected values; GUI "y" + invalid port defaults to 9099
- [ ] T010 [P] [US1] Test startup error path in `cmd/run_test.go`: missing/empty `~/.buenosaires/config.toml` produces a clear, actionable error message

### Implementation for User Story 1

- [ ] T011 [US1] Surface stdin read errors in `cmd/install.go` (currently silently ignored via `_`) and validate port range (depends on T009)
- [ ] T012 [US1] Improve `cmd/run.go` startup: validate loaded global config per T004 and emit actionable error before the monitor loop (depends on T004, T010)

**Checkpoint**: Setup + clear-error behavior fully functional and independently testable.

---

## Phase 4: User Story 2 - Operator monitors a repository branch (Priority: P1)

**Goal**: Continuous monitoring with worktree-independent change detection, initial sync, and correct insert/modify/delete processing

**Independent Test**: Set up a test repo, run the monitor, push insert/modify/delete/invalid/fix commits, and verify each is processed exactly as specified.

### Tests for User Story 2

- [ ] T013 [P] [US2] Unit tests for `processChanges` in `cmd/run_test.go`: insert/modify/delete dispatch, rename (delete+insert pair), missing-base-commit fallback to initial sync
- [ ] T014 [P] [US2] Extend `shouldProcessScript` tests in `cmd/run_test.go` for rename-pair semantics (insert side of a rename of a failed script re-processes)
- [ ] T015 [US2] Lifecycle integration test in `plugins/shell/shell_test.go`: materialized script → `LintAndValidate` → `Run` → asset JSON written with generation increment, status, commit_hash, duration (depends on T008)

### Implementation for User Story 2

- [ ] T016 [US2] Clean up `processScript` in `cmd/run.go`: unify log-dir resolution (single precedence expression), ensure status saved as `pending` before processing and final state after
- [ ] T017 [US2] Add signal handling + graceful shutdown for the run loop in `cmd/run.go` (context + `os.Interrupt`/`SIGTERM`; stop polling, close web server) — closes finding U3
- [ ] T018 [US2] Add retry/backoff for repeated fetch failures in `cmd/run.go` (exponential backoff capped, reset on success) — closes finding U5

**Checkpoint**: Monitor loop robust under dirty worktree, failed fetch/pull, and restart; lifecycle verified by tests.

---

## Phase 5: User Story 3 - Operator inspects execution results (Priority: P2)

**Goal**: Accurate per-script status, logs, and asset metadata — visible via log files and the web interface

**Independent Test**: Process a failing and a succeeding script, then verify logs, status entries, and web UI rendering for both.

### Tests for User Story 3

- [ ] T019 [P] [US3] Log format contract test: per-script log contains `--- LINT OUTPUT ---` and `--- EXECUTION OUTPUT ---` sections (extend shell lifecycle test T015)
- [ ] T020 [P] [US3] Extend `internal/web/server_test.go`: asset lookup for nested paths (folder-prefixed scripts), 400 on traversal/unsafe names, security headers on asset + CSS routes

### Implementation for User Story 3

- [ ] T021 [US3] Verify log file writes use mode 0600 and `MkdirAll(0750)` in `cmd/run.go`; align any divergence
- [ ] T022 [US3] Resolve status/asset semantic inconsistency: document that shell plugin's test phase is `skipped` and asset `tests_passed` mirrors lint outcome (comment in `plugins/shell/shell.go` + README §Status Tracking) — closes findings B1/I2
- [ ] T023 [US3] Replace README placeholder screenshot URL (`i.imgur.com/example.png`) and add a Web Interface section screenshot or remove the reference — closes finding I3

**Checkpoint**: Logs, status, and web UI independently inspectable and accurate.

---

## Phase 6: User Story 4 - Developer deploys via git push (Priority: P2)

**Goal**: GitOps workflow — pushed scripts auto-deploy; fixes re-deploy; unchanged successful scripts never re-run

**Independent Test**: Push a valid script, then a broken version, then a fix; verify execution outcomes; restart the monitor and verify no re-run of succeeded scripts.

### Tests for User Story 4

- [ ] T024 [P] [US4] Resume test in `cmd/run_test.go`: status with `last_commit` set → diff vs new commit processes only changed scripts; succeeded inserts skipped (SC-004)
- [ ] T025 [US4] End-to-end scenario test (bash-driven or go-git fixture): insert → execute, modify → re-execute, delete → status cleanup, invalid → never executed

### Implementation for User Story 4

- [ ] T026 [US4] Fix any gaps found by T024/T025 in `cmd/run.go` change detection (insert-dedupe, modify-always, delete-cleanup) — no new behavior, verify existing contract

**Checkpoint**: Git push → deploy loop proven; restart resume verified.

---

## Phase 7: User Story 5 - Security-conscious operator controls privilege (Priority: P3)

**Goal**: Sudo executes if and only if global AND repo configs both allow it

**Independent Test**: Run scripts under all four `allow_sudo` combinations and verify privilege behavior (SC-003).

### Tests for User Story 5

- [ ] T027 [P] [US5] Two-gate test in `cmd/run_test.go`: `allowSudo = global.AllowSudo && repo.AllowSudo` — assert all four combinations (false/false, false/true, true/false → no sudo; true/true → sudo)

### Implementation for User Story 5

- [ ] T028 [US5] Document the sudo trust model in README §Configuration (scripts execute with the effective user's privileges; repo alone cannot escalate) — depends on T027

**Checkpoint**: Privilege model verified for all combinations and documented.

---

## Phase 8: Polish & Cross-Cutting Concerns

**Purpose**: Knowledge sync, full validation, and final hygiene

- [ ] T029 [P] Update `memory/application_overview.md` with ratified decisions: test-phase semantics (skipped/tests_passed mirrors lint), config validation rules, signal handling, fetch backoff (constitution §IV)
- [ ] T030 Run `quickstart.md` end-to-end validation (install → initial sync → lifecycle → two-gate → resilience → web UI → container)
- [ ] T031 [P] Final gate: `make check` green locally AND CI pipeline green on the feature branch (test + lint + security + build-and-push)
- [ ] T032 [P] Cleanup: remove stray `.tmp` and `probe.log`; confirm `git status` contains only intentional changes

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — can start immediately
- **Foundational (Phase 2)**: Depends on Setup — BLOCKS all user stories
- **User Stories (Phase 3–7)**: All depend on Foundational
  - US1 (P1) and US2 (P1) can proceed in parallel after Phase 2
  - US3 (P2), US4 (P2), US5 (P3) can proceed in parallel after US2's change-detection contract is verified (they exercise `processScript`/monitor behavior)
- **Polish (Phase 8)**: Depends on all stories

### User Story Dependencies

- **US1 (P1)**: after Foundational — independent of other stories
- **US2 (P1)**: after Foundational — no dependency on US1
- **US3 (P2)**: after Foundational; integrates US2 processing outputs (logs/status/assets)
- **US4 (P2)**: after Foundational; depends on US2 change detection being correct
- **US5 (P3)**: after Foundational; independent of US1–US4

### Within Each User Story

- Tests MUST be written and FAIL before implementation (TDD — T009/T010, T013–T015, T019/T020, T024/T025, T027)
- Models/contracts before services; services before integration
- Story complete before moving to next priority

### Parallel Opportunities

- Phase 1 T002/T003 run in parallel
- Phase 2 T005/T006/T007/T008 run in parallel (different files)
- US1 tests (T009/T010) in parallel; US2 tests (T013/T014/T015) in parallel
- US3–US5 test tasks in parallel (distinct test files)
- Phase 8 T029/T031/T032 in parallel (distinct files)

---

## Implementation Strategy

### MVP First (US1 + US2)

1. Complete Phase 1: Setup verification
2. Complete Phase 2: Foundational (CRITICAL)
3. Complete Phase 3: US1 → independent test → demo
4. Complete Phase 4: US2 → independent test → demo (this is the MVP: install + monitor)
5. **STOP and VALIDATE** before continuing to P2/P3 stories

### Incremental Delivery

1. Setup + Foundational → foundation ready
2. US1 + US2 → MVP (guided setup + monitoring with validation/execution/logs)
3. US3 → observability (logs + status + web UI)
4. US4 → GitOps push-deploy verification
5. US5 → privilege model verification
6. Phase 8 → knowledge sync + full validation

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to the spec's user story for traceability
- Each user story is independently completable and testable
- Verify tests fail before implementing; commit after each task or logical group
- Constitution gates apply at every commit: `make check` must pass, runtime artifacts never committed, `#nosec` exceptions carry justifications
- Avoid: vague tasks, same-file conflicts, cross-story dependencies that break independence
