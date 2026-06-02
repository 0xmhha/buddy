# Buddy Roadmap — Historical Milestone Log

> **Archived 2026-05-21** — M1~M6 all delivered (v0.1.0 released 2026-04-26). 잔여 작업의 SSoT는 [`../BACKLOG.md`](../BACKLOG.md).

> **Scope**: M1 ~ M6 + v0.1.0 release 까지의 **역사적 milestone log**. *현재의 잔여 작업과 우선순위* 는 [`../BACKLOG.md`](../BACKLOG.md) 가 단일 SSoT.
>
> v0.4.1 시점 (2026-05-11) 의 §4 v0.2 Control Plane / §5 v0.3 Orchestration / §6 v1.0 통합 outline 은 본 cycle 에서 *ADR-002 + ADR-005 의 부분 supersede* + 본 backlog 통합으로 **`BACKLOG.md` 로 위임**. 향후 v0.x → v1.0 trigger 는 [`HANDOFF.md`](../HANDOFF.md) §1 + [`BACKLOG.md`](../BACKLOG.md) §1.
>
> 작성일: 2026-04-23 / 최종 갱신: 2026-05-19 (BACKLOG.md 통합 + 슬림화) / 상태: HISTORICAL

---

## 1. M1 ~ M6 + v0.1.0 milestone log

> 모두 ✅ Done — *역사적 사실* 로 보존. 향후 *왜 이 구조였는지* 의 reference.

| M | 내용 | 상태 |
|---|------|------|
| M1 (Go) | Schema + SQLite + migrations + outbox | ✅ DONE |
| M2 (Go) | Hook wrapper (cobra) + invariants (spec §7.1) | ✅ DONE |
| M3 | Daemon (outbox → events → stats) + cli-wrapper hybrid | ✅ DONE |
| M4 | `buddy install/uninstall/doctor/stats/events` CLI | ✅ DONE |
| M5 | Config CLI + threshold tuning + dogfood 4-friction fix + purge + 페르소나 catalog | ✅ DONE (PR #1) |
| M6 | Release prep — cross-compile + GitHub Actions release + CHANGELOG | ✅ DONE (PR #2) |
| **v0.1.0 release** | Tag `v0.1.0` + 4 binaries (linux/darwin × amd64/arm64) + SHA256SUMS | ✅ **2026-04-26** |

### M5 lock-in 결정 (역사적 보존)

- T6: install 이 `db.Open` 호출하여 DB pre-create + 마이그레이션
- T7: PID 는 `cmd.Process.Release()` 전에 캡처 (`startAndDetach` helper)
- T8: read-only `db.Open` 이 missing parent/file 시 `ErrDBMissing` sentinel
- T9: `uninstall` 이 자동 `daemon stop`, `--keep-daemon` escape hatch
- Config 우선순위: 명시적 CLI flag > config 파일 > spec-locked default
- 페르소나 catalog: `internal/persona/`, typed `Key` 상수, en↔ko parity ([`BACKLOG.md`](../BACKLOG.md) Wave 2-3 = ✅ Done, en/ko 57/57)
- DB busy_timeout: `_pragma=busy_timeout(5000)` (concurrent open race fix)
- Daemon SIGTERM handler: PID 파일 publish 전 설치 (race fix)

### M6 lock-in 결정 (역사적 보존)

- T3: `var version = "0.1.0"` + `versionString()` helper. Cobra version template override 로 double prefix 회피.
- T1: `make release-binaries` → 4 platforms `CGO_ENABLED=0` + `-trimpath` + ldflags + `dist/SHA256SUMS`
- T2: `.github/workflows/release.yml` — `v*` tag push → matrix build → `softprops/action-gh-release@v2` upload
- T4: `CHANGELOG.md` Keep a Changelog 1.1.0 + SemVer
- T5: README install — release binary path 위 + `make build` 아래

> **Single source-of-truth** — version 토큰 5 곳 (main.go, Makefile, CHANGELOG, README, release.yml). drift 시 workflow 의 tag↔RELEASE_VERSION 검증으로 publish 전 차단. `make verify-versions` + `make verify-go-version` 으로 보호 (v0.6.2~).

---

## 2. v0.1 이후 진행 요약 (2026-04-26 ~ 2026-05-19)

v0.1.0 release 이후 본 cycle 까지의 publish:

| Tag | Date | 요지 |
|-----|------|------|
| v0.2.0 | 2026-05-11 | Plugin track version reset (ADR-004) + B6 PROCEDURE template / lint |
| v0.3.0 | 2026-05-11 | analytics-mcp W4-2.1~2.6 ship (7 MCP tools + SQLite adapter + 10 skill integration) |
| v0.4.0 | 2026-05-11 | cli buddy W3-3 agent runtime + background scheduler |
| v0.4.1 | 2026-05-11 | release.yml: include `buddy-mcp_*` in publish pattern |
| v0.5.0 | 2026-05-12 | cli buddy W3-4 partial: PROCEDURE output parser |
| v0.6.0 | 2026-05-12 | W3-4 follow-on: conditional next-phase branches |
| v0.6.1 | 2026-05-12 | actions/checkout@v6, setup-go@v6, action-gh-release@v3 (Node 24 호환) |
| v0.6.2 | 2026-05-12 | release pipeline hardening — `make verify-go-version`/`verify-versions` + `ci.yml` test gate + `buddy agent log` + `Store.LatestRun` |
| v0.6.3 | 2026-05-12 | Tier 1.4 exponential backoff + W3-5: `cmd/buddy/main.go` 688→147 lines split (6 sibling) |
| v0.6.4 | 2026-05-12 | Tier 1.5 streaming log capture + verify-quality F1-F4 cleanup |
| v0.6.5 | 2026-05-12 | Tier 1.6 scheduler live refresh — `RefreshInterval` polling + `tracked` map diff |
| v0.6.6 | 2026-05-12 | Tier 1.8 webhook output target + W3-6 reference webtoon agent |
| v0.7.0 | 2026-05-17 | W3-2 minimum-viable TUI + 3 follow-on (detail/scheduler-preview/in-app delete) + F5 log retention |
| v0.7.1 | 2026-05-18 | W3-2 follow-on finish (log tail / in-app edit / create form) + W3-4 chain control (continue_on_fail + auto_cascade) |
| v0.7.2 | 2026-05-18 | TUI hook-stats pane (A-3.2 W3-5 follow-on) |
| v0.7.3 | 2026-05-19 | ADR-006/007 + HANDOFF sync |
| v0.7.4 | 2026-05-19 | ADR-008 feature-management-mcp scope lock-in |
| v0.7.5 | 2026-05-19 | dogfood guide v0.7.x |

→ **cli buddy spec §9 W3-1~W3-6 모두 Done**. Plugin v1.0.0 entry conditions B-1/B-3/B-4 closed, **B-2 (production dogfood) 만 잔여** — [`BACKLOG.md`](../BACKLOG.md) Wave 1.

---

## 3. 의도적 비범위 (모두 유지)

| 항목 | 사유 |
|------|------|
| Web UI 풀스택 | CLI + 경량 TUI 로 충분 |
| 분산 시스템 (멀티-머신) | single-machine 가정 ([`v0.1-spec.md`](./v0.1-spec.md) §1 Non-goal) |
| 자체 LLM provider proxy | Claude Code 본체 책임 |
| Windows | macOS/Linux 우선 (POSIX 가정 다수) |
| Cross-harness parity (Codex/OpenCode 동시) | harness-analysis 갭 C "환상" — v1.0+ 검토만 ([`BACKLOG.md`](../BACKLOG.md) W6-2) |
| 자동 telemetry / 사용 통계 외부 전송 | 로컬 도구, opt-in 도 v1.0+ |

---

## 4. v0.2 / v0.3 / v1.0 outline — 본 문서에서 *제거*

원래 §4 v0.2 Control Plane, §5 v0.3 Orchestration, §6 v1.0 통합 outline 이 있었음. 본 cycle (2026-05-19) 에 *ADR-002 + ADR-005 부분 supersede* + cli buddy spec §9 W3-1~W3-6 cascade 로 통합되어 **`BACKLOG.md` 로 위임**. 역사적 outline 은 `git log --before=2026-05-19 -- docs/roadmap.md` 로 추적.

---

## 5. 참조

- [`BACKLOG.md`](../BACKLOG.md) — **단일 잔여 작업 SSoT** + 진행률
- [`v0.1-spec.md`](./v0.1-spec.md) — M1~M5 LOCKED spec (역사적)
- [`cli-buddy-spec.md`](./cli-buddy-spec.md) — cli buddy spec (Accepted, ADR-005)
- [`two-tracks-charter.md`](../two-tracks-charter.md) — 두 트랙 책임 경계 SSoT
- [`HANDOFF.md`](../HANDOFF.md) — 세션 인계 가이드
- [`superpowers/decisions/README.md`](../superpowers/decisions/README.md) — ADR Index (8 건 Accepted)
- [`decision-1-schema-fields.md`](./decision-1-schema-fields.md) — 옵션 A schema 결정 근거
