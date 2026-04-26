# Buddy Roadmap — post-v0.1

> 한 줄 요약: v0.1(Reliability wedge)이 dogfood 단계에 진입했고, 이후 M5(config + dogfood 마찰 fix) → M6(release prep) → v0.2(Control Plane) → v0.3(Orchestration) → v1.0(통합) 순으로 확장한다.
>
> 이 문서는 v0.1 이후 작업의 *Single Source of Truth* (SSoT). [`v0.1-spec.md`](./v0.1-spec.md) §8 마일스톤 표는 이 문서로 위임됨.

작성일: 2026-04-23 (최종 갱신: 2026-04-26 v0.1.0 release) / 상태: ACTIVE — M5 ✅ + M6 ✅ + v0.1.0 ✅ released. 이후 우선순위(v0.2 / v0.3)는 dogfood feedback 회수 시점에 따라 재정렬.

---

## 1. 현재 위치

### v0.1 status (2026-04-26 v0.1.0 release 기준)

| M | 내용 | 상태 |
|---|------|------|
| M1 (Go) | Schema + SQLite + migrations + outbox | DONE |
| M2 (Go) | Hook wrapper (cobra) + invariants ([spec §7.1](./v0.1-spec.md#71-hook-wrap--m2-implemented)) | DONE |
| M3 | Daemon (outbox → events → stats) + cli-wrapper hybrid | DONE |
| M4 | `buddy install/uninstall/doctor/stats/events` CLI | DONE |
| M5 | Config CLI / threshold tuning / dogfood 4-friction fix / purge / 페르소나 catalog | **DONE (PR #1)** |
| M6 | Release prep (cross-compile + GitHub Actions release + CHANGELOG) | **DONE (PR #2)** |
| **v0.1.0 release** | Tag `v0.1.0` + 4 binaries (linux/darwin × amd64/arm64) + SHA256SUMS GitHub Release | **DONE (2026-04-26)** |
| **Dogfood** | [`DOGFOOD.md`](../DOGFOOD.md) + [`dogfood-feedback-template.md`](./dogfood-feedback-template.md) — release binary로 본격 시작 가능 | READY |

### 향후 5개 마일스톤 한눈 표

| 버전 | wedge 추가 | 핵심 산출 | 의존 | 참조 갭 ([analysis](../../harness-engineering-analysis.md)) |
|------|-----------|----------|------|----------------------------------------------------|
| **M5** | Reliability+ | `buddy config`, `buddy purge`, dogfood 4-friction fix, 페르소나 catalog | dogfood feedback | D, F |
| **M6** | (release) | cross-compile matrix, GitHub Actions release, `--version` 임베드 | M5 | — |
| **v0.2** | + Control Plane | multi-session inventory, token/cost unified view | M6 | G |
| **v0.3** | + Orchestration | task DAG, wave executor, retry loop | M6 (M5 권장) | A, B |
| **v1.0** | 통합 | AGENTS.md auto-sync, plugin model, MCP server | v0.2 + v0.3 | (메타-패턴 #9) |

---

## 2. M5 — Config / Threshold tuning / dogfood 마찰 fix / Purge / 페르소나 polish ✅ DONE

> **상태:** 완료 (PR [#1](https://github.com/0xmhha/buddy/pull/1)). 14 commits, +4245/−76 LOC, 15 packages race-clean.
> 아래 §Tasks T1~T9는 *역사적 spec*으로 보존 — 실제 구현 위치는 코드의 `internal/config/`, `internal/purge/`, `internal/persona/`, `internal/install/`, `internal/db/`, `internal/diagnose/` 패키지 + `cmd/buddy/{config_cmd,loadconfig,purge_cmd,spawn,main}.go` 참조.
>
> **Lock-in (이번 M5에서 실제로 채택된 결정):**
> - T6: 후보 (a) — install이 `db.Open` 호출해 DB pre-create + 마이그레이션
> - T7: PID는 `cmd.Process.Release()` 전에 캡처 (`startAndDetach` helper)
> - T8: read-only `db.Open`이 missing parent/file 시 `ErrDBMissing` sentinel 반환, CLI가 친구 톤으로 변환
> - T9: 후보 (a) — `uninstall`이 자동 `daemon stop`, `--keep-daemon` escape hatch
> - Config 우선순위: 명시적 CLI flag > config 파일 > spec-locked default
> - 페르소나 catalog: `internal/persona/`, typed `Key` 상수, en→ko fallback (en map은 v0.2까지 비어있음)
> - DB busy_timeout: `_pragma=busy_timeout(5000)` 추가 (concurrent open race fix)
> - Daemon SIGTERM handler: PID 파일 publish 전에 설치 (race fix)

### Why

[Harness analysis](../../harness-engineering-analysis.md) §3 갭 D(silent failure 감지)·F(state schema 검증)는 v0.1이 일부 채웠지만, 사용자가 [spec §6.2](./v0.1-spec.md#62-hook-health-임계값-default--decided-2026-04-23)의 hard-coded threshold를 바꾸려면 recompile 외 방법이 없다. dogfood 회고가 들어오기 시작하면 *바로 조정할 수 있는 손잡이*가 필수.

또한 T1 dogfood validation 도중 발견된 4가지 마찰 포인트(아래 T6~T9)가 있어, config 작업과 함께 해소한다. M5 task 우선순위는 [`docs/dogfood-feedback-template.md`](./dogfood-feedback-template.md) 회수 결과에 따라 재정렬된다.

### Tasks

**Config infrastructure**

- **T1**: `~/.buddy/config.json` 스키마 + manual `Validate()` 메서드. 위치: 신규 `internal/config/` 패키지 (또는 기존 `internal/store/` 확장 검토). [spec §3](./v0.1-spec.md#3-기술-스택-결정-2026-04-24-pivot-ts--go) "struct + Validate() 메서드" 원칙 따름.
- **T2**: `buddy config get/set/unset/show` CLI. 위치: `internal/cli/config.go` + `cmd/buddy/main.go`에 cobra subcommand 등록.
- **T3**: doctor/aggregator/daemon이 config에서 threshold·poll·batch 값을 읽도록 변경. 영향 범위: `internal/aggregator`, `internal/daemon`, `internal/cli/doctor.go`. config 부재 시 [spec §6.2 DefaultThresholds](./v0.1-spec.md#62-hook-health-임계값-default--decided-2026-04-23)로 fallback.

**Data lifecycle**

- **T4**: `buddy purge --before <date>` — 오래된 `hook_events`·`hook_stats` 삭제. **`hook_outbox`는 절대 건드리지 않는다** ([spec §4 invariant 1](./v0.1-spec.md#4-아키텍처-m3-반영) — outbox는 sync write 대상이라 동시성 위험). 위치: `internal/cli/purge.go`. dry-run 플래그 (`--dry-run`) 기본 안전.

**Persona polish**

- **T5**: 페르소나 메시지 catalog 정리 — 현재 메시지 문자열이 코드 곳곳에 분산. 단일 카탈로그 (`internal/persona/messages.go`) + locale 키 기반 lookup으로 통합. i18n 토대 (locale=en fallback, 본격 ko/en 분리는 v0.2 이후). [spec §6.3](./v0.1-spec.md#63-buddy의-페르소나--말투--decided-2026-04-23) 원칙(침묵 default, 호들갑 X, 이모지 X) 보존.

**dogfood 4-friction fix** (T1 dogfood validation 중 발견)

- **T6**: `install` 직후 `doctor`가 raw SQL 에러를 뱉는 문제. 현재는 `daemon start`가 첫 마이그레이션을 트리거하므로, install→doctor 순서일 때 `no such table: hook_outbox`가 노출된다 ([DOGFOOD.md §2.1](../DOGFOOD.md) 참조).
  - Fix 후보 (택 1):
    - (a) `install`이 직접 마이그레이션을 실행해 buddy-dir·DB를 pre-create
    - (b) `doctor`가 schema 미존재를 감지해 친구 톤 메시지로 변환 ("DB가 아직 없어. `buddy daemon start` 한 번 띄워줘")
  - 결정 시점: M5 implementation kickoff. 기본 후보 (a) — install이 self-contained 동작이 우선.
  - 위치: `internal/cli/install.go` 또는 `internal/cli/doctor.go`.
- **T7**: `buddy daemon start`가 `pid -1`을 출력하는 문제 (`cmd.Process.Release` 후 PID race). Fix: detach 전에 PID를 캡처하거나, start 직후 PID 파일을 read해 보고. 위치: `internal/daemon/start.go` (또는 `internal/cli/daemon.go`의 fork 경로).
- **T8**: `--db` 부모 디렉터리가 없을 때 `out of memory (14)` cryptic 에러. Fix 후보 (택 1):
  - (a) `db.Open`이 write 경로일 때 부모 디렉터리를 `mkdir -p`로 보장 (명시적 `--db` 여부 무관)
  - (b) 부모 디렉터리 부재 시 친구 톤 에러로 fail-fast ("`<dir>` 가 없어. `mkdir -p` 먼저.")
  - 결정 시점: M5 implementation kickoff. 기본 후보 (a) — db.Open이 mkdir까지 책임지는 게 단순.
  - 위치: `internal/store/db.go` (`Open` 함수).
- **T9**: `uninstall`이 실행 중인 daemon을 안 끄는 문제 (orphan daemon). Fix 후보 (택 1):
  - (a) `uninstall` 내부에서 `daemon stop`을 자동 실행 (PID 파일 + flock 확인 후)
  - (b) `uninstall`이 daemon 실행을 감지하면 `--force` 없이 거부 + 친구 톤 안내
  - 결정 시점: M5 implementation kickoff. 기본 후보 (a) — uninstall이 daemon stop을 자동 호출, --keep-daemon 플래그로 escape hatch.
  - 위치: `internal/cli/uninstall.go` ([DOGFOOD.md §6](../DOGFOOD.md) "안전 종료" 가이드 참조).

### Acceptance

- dogfood feedback에서 발견된 모든 "이건 좀…" 사례가 (a) `buddy config set ... ...` 한 줄로 해결되거나, (b) T6~T9 중 하나로 종결됨.
- `buddy --help` 출력에 `config`·`purge`가 노출되며 각 subcommand가 자체 `--help`를 가짐.
- `install` → `doctor` 순서 (daemon 미실행) 실행 시 raw SQL error가 나오지 않는다 (T6 acceptance).
- `daemon start` 출력에 실제 PID가 표시된다 (T7 acceptance).

### Open questions (M5 종료 시점)

- ✅ **config 값 hot-reload?** — 1차안 *restart 요구* 적용됨. daemon은 시작 시점에 config 읽고, 변경 시 사용자가 `daemon stop`/`daemon start` 재시작. 실제 dogfood 사용 패턴 보고 v0.2에서 재검토.
- ⏳ **`purge`의 outbox 수동 cleanup 경로** — M5에서 "건드리지 않음" 정책 유지. outbox가 영영 드레인 안 되는 케이스(예: schema migration 실패) 발견 시 별도 명령 검토 — **현재 미발생, 데이터 들어오면 결정**.
- ⏳ **`config.ValidationError.Reason` i18n** — persona 카탈로그에 `KeyConfigReason*` 키들 declared & test 보호되어 있음. `translateConfigError` bullet 렌더링 wiring은 v0.2 i18n sweep으로.
- ⏳ **`queries.ErrInvalidLimit` / `ErrInvalidWindow`** — Korean 문자열을 sentinel error 값에 직접 보유. 카탈로그로 이전은 v0.2 sweep.
- ⏳ **English locale 카탈로그** — `internal/persona/en.go`는 빈 map placeholder. 본격 ko/en 분리는 v0.2.
- ⏳ **Subcommand-flag-aware locale** — root `PersistentPreRunE`가 현재 `config.DefaultPath()`만 읽음. 서브커맨드 `--config` 인지 locale 해석은 v0.2 polish.

---

## 3. M6 — Release prep ✅ DONE

> **상태:** 완료 (PR [#2](https://github.com/0xmhha/buddy/pull/2)). 6 commits, +394/−12 LOC. v0.1.0 tag publish 시 GitHub Actions 워크플로가 자동으로 4 binaries + SHA256SUMS 업로드 — 2026-04-26 13:12 UTC 동작 확인.
>
> **As-built (T1~T5 lock-in):**
> - T3: `var version = "0.1.0"` / `gitSHA = "dev"` / `buildDate = "unknown"` defaults + `versionString()` helper. Cobra version template override 로 `buddy version buddy 0.1.0 ...` double prefix 회피.
> - T1: `make release-binaries` → 4 platforms (linux/amd64, linux/arm64, darwin/amd64, darwin/arm64) `CGO_ENABLED=0` + `-trimpath` + ldflags + `dist/SHA256SUMS`. 파일명 `buddy_${RELEASE_VERSION}_${OS}_${ARCH}`.
> - T2: `.github/workflows/release.yml` — `v*` tag push → `make release-binaries` → `softprops/action-gh-release@v2` upload. Tag↔Makefile RELEASE_VERSION mismatch 시 fail-fast.
> - T4: `CHANGELOG.md` Keep a Changelog 1.1.0 + SemVer. v0.1.0 entry 가 M1~M5 user-facing surface 전체 + Deferred 항목 명시.
> - T5: README install 섹션 — release binary path 위 + `make build` 아래(`### Install — from source`). macOS Gatekeeper 노트 (`xattr -d com.apple.quarantine`).
>
> **단일 source-of-truth — `0.1.0` 토큰 5곳 (main.go, Makefile, CHANGELOG, README, release.yml). Drift 시 워크플로의 tag↔RELEASE_VERSION 검증으로 publish 전 차단.**

### Why

첫 release tag (`v0.1.0`) + 다른 사람이 install 가능한 형태. 현재는 `make build` + `--buddy-binary` 절대 경로 지정이 유일한 install 경로 ([DOGFOOD.md §1](../DOGFOOD.md)). 외부 사용자에게 "git clone + go build" 외 옵션을 제공하려면 binary 배포가 필요.

### Tasks

- **T1**: cross-compile matrix — `linux/amd64`, `linux/arm64`, `darwin/amd64`, `darwin/arm64`. Makefile에 `release-binaries` target 추가. `GOOS`/`GOARCH` × `CGO_ENABLED=0` (modernc.org/sqlite는 pure Go라 cgo 불요, [spec §3](./v0.1-spec.md#3-기술-스택-결정-2026-04-24-pivot-ts--go)).
- **T2**: GitHub Actions release workflow — `.github/workflows/release.yml`. tag (`v*`) push trigger → matrix build → binary attach + checksum (`SHA256SUMS`). 참고 패턴: `goreleaser` 또는 직접 `actions/upload-release-asset`.
- **T3**: `buddy --version` 출력에 commit SHA + build date 포함. ldflags로 주입: `go build -ldflags "-X main.gitSHA=$(git rev-parse --short HEAD) -X main.buildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ)"`. 위치: `cmd/buddy/main.go` (version cobra command).
- **T4**: `CHANGELOG.md` v0.1.0 entry — Keep a Changelog 포맷. M1~M4 + M5 + dogfood 마찰 fix 4건 정리.
- **T5**: README의 install 섹션에 release binary 경로 추가. 현재 `make build` 안내 ([README.md Quickstart](../README.md#quickstart))는 "from source" 섹션으로 분리, 상위에 `curl -L ... | tar` 또는 `gh release download` 안내. `curl … | sh` one-liner는 보안상 보수적 — 우선 명시적 다운로드 경로만 제공.

### Acceptance

- `git tag v0.1.0 && git push --tags` 한 번으로 GitHub release 페이지에 4개 binary + `SHA256SUMS` 자동 첨부.
- `buddy --version` 실행 시 `buddy 0.1.0 (sha=abc1234, built=2026-XX-XXT...Z)` 형식 출력.

### Open questions (M6 종료 시점)

- ⏳ **macOS notarization** — v0.1.0 미적용 (사용자가 `xattr -d com.apple.quarantine` 안내). v0.2 이후 검토.
- ⏳ **Third-party Actions SHA pinning** — `actions/checkout@v4`, `actions/setup-go@v5`, `softprops/action-gh-release@v2` 모두 mutable major tag. supply-chain 강화 시 commit SHA 핀+Dependabot. v0.2 polish.
- ⏳ **별도 `ci.yml`** — 현재 `release.yml`만 존재. tag 전 PR/push 단계에서 `go test/vet/fmt`를 강제하는 CI workflow 분리 필요.
- ⏳ **Single-source-of-truth `VERSION` 파일** — 현재 `0.1.0` 5곳에 분산되어 있고 tag↔Makefile sanity check 로 보호. release cadence가 잦아지면 파일 분리 또는 build-time embed 검토.

---

## 4. v0.2 — Control Plane (멀티-세션 dashboard)

### Why

[Harness analysis](../../harness-engineering-analysis.md) §3 갭 G "통합 observability" — 토큰·세션·비용·hook 상태가 별개 도구로 분산. v0.1의 `buddy doctor`/`stats`는 단일 머신·단일 세션 관점이지만, dashboard는 *동시에 떠 있는 여러 Claude Code 세션*을 한 번에 본다. recon이 PID→JSONL 매핑 패턴([analysis 메타-인사이트 #1](../../harness-engineering-analysis.md))을 이미 검증했으므로 차용 가능.

### Tasks

- **T1**: 활성 세션 발견 — recon 패턴 차용. `~/.claude/sessions/*.json` 또는 `~/.claude/projects/**/*.jsonl` (transcript) 스캔. 위치: 신규 `internal/sessions/` 패키지. PID는 transcript metadata 또는 process list 교차 참조.
- **T2**: Token usage parser — transcript JSONL의 `usage` 필드 추출. [spec §6.1 옵션 A](./v0.1-spec.md#61-hook-event-필드--decided-2026-04-23--옵션-a) `tokenUsage` 스키마(`inputTokens`/`outputTokens`/`cacheReadTokens`/`cacheCreateTokens`)를 그대로 재사용. 위치: `internal/sessions/transcript.go`.
- **T3**: 단일 세션 단위 stats를 *여러 세션 합산* 뷰로 확장. 기존 `buddy stats` 플래그에 `--all-sessions` 추가 또는 신규 `buddy sessions list/show` subcommand 검토. CLI 표면 디자인은 T4 결정 후 확정.
- **T4**: dashboard UI — TUI (`charmbracelet/bubbletea` + `lipgloss`) 또는 web (localhost HTTP, 포트 미정). **Open question (아래)**.
- **T5**: cost estimate — model 단가 테이블 (`internal/pricing/`)을 코드에 삽입(v1.0에서 외부 데이터로 분리 검토). `tokens × price`로 추정치 출력. 단가 변경 시 release로 업데이트.
- **T6**: i18n full split (en/ko strings, gettext-style or embed map). M5 T5의 forward-pointer 회수.

### Acceptance

- 동시에 떠 있는 Claude Code 세션 ≥2개 환경에서 `buddy <dashboard 진입 명령>`이 모든 세션의 token·cost·hook health를 한 화면에 보여준다.
- 임의 시점 cost estimate가 Anthropic console 실제 청구액과 ±5% 이내 (sanity check, 단가 테이블 정확도 의존).

### Open questions

- **TUI vs web** — 결정 보류. dogfood 결과 어떤 UX가 더 자연스러울지(터미널 안에서 끝내고 싶은 사용자 vs 브라우저 탭으로 분리하고 싶은 사용자)가 결정에 영향. M6 release 후 1~2주 사용자 피드백 수집 단계 권장.
- 멀티-머신 통합? — single-machine 유지가 [v0.1-spec.md §1 Non-goals](./v0.1-spec.md#1-non-goals-v01에서-안-하는-것)와 [§9 위험 매트릭스](./v0.1-spec.md#9-위험--open-questions)에 명시되어 있으나, "내 노트북 + 회사 워크스테이션" 케이스는 v0.2 dogfood에서 가장 흔한 요청 가능성. v1.0+ 검토.

---

## 5. v0.3 — Orchestration (task DAG executor)

### Why

[Harness analysis](../../harness-engineering-analysis.md) §3 갭 A(task dependency graph), B(retry/feedback loop). arc-reactor의 wave parallelization 패턴 + quality gate retry loop ([analysis 메타-인사이트 #2](../../harness-engineering-analysis.md))를 buddy의 신뢰성 위에 얹는다. 갭 A는 claude-squad·task-master가 부분 시도, 단일 자료구조로 통일된 적 없음 — buddy의 SQLite 인프라가 이 통일을 가능케 한다.

### Tasks

- **T1**: task DAG schema — SQLite 테이블 (`tasks`, `task_deps`) 또는 JSON 파일 backed. SQLite 우선 검토 (이미 인프라 보유, [spec §5](./v0.1-spec.md#5-데이터-모델-draft--사용자-결정-61-필요) 패턴 재사용). 위치: 신규 migration `internal/store/migrations/00X_tasks.sql`.
- **T2**: `buddy task add/list/run/status` CLI. 위치: `internal/cli/task.go`. 기존 cobra subcommand 트리에 등록.
- **T3**: wave 그룹화 — 의존성 없는 task를 자동 병렬 실행. Go goroutine + `errgroup.WithContext`로 구현. 동시 실행 상한은 config (`maxParallelTasks`, M5 T1 schema에 추가).
- **T4**: retry policy — exponential backoff + max attempts. quality gate(예: 외부 명령 exit code) 실패 시 자동 re-queue. spec과 동일하게 friend-tone error 메시지.
- **T5**: 외부 task tracker 통합? — Linear / Jira / GitHub Issues. **Open question (아래)**. 통합 시 위치는 `internal/integrations/` 신설.

### Acceptance

- `buddy task add` 로 등록한 task DAG이 `buddy task run` 한 번으로 wave 병렬 실행되고, 실패 시 retry policy에 따라 재시도되며, 모든 시도가 `hook_events`(또는 별도 `task_runs`)에 기록되어 사후 분석 가능.

### Open questions

- **buddy가 task tracker가 되어야 하는가, 기존 tracker의 *실행 엔진*만 되어야 하는가** — 후자가 scope가 작고 책임 분리 명확 (DB 동기화 부담 없음, tracker UI는 외부에 위임). 결정 보류 — v0.2 사용자 피드백 + buddy의 stats/observability와의 통합 가치를 함께 검토.
- DAG 시각화? — `buddy task graph --format dot`(graphviz) 정도가 합리적 선. 본격 시각화는 v0.2 dashboard에 통합.

---

## 6. v1.0 — 통합 (AGENTS.md auto-sync, plugin model, MCP server)

### Why

[Harness analysis](../../harness-engineering-analysis.md) 메타-패턴 "Agent knowledge is explicit metadata" — 모든 framework가 AGENTS.md/CLAUDE.md/manifest를 자동 동기화하지 않으면 agent가 capability를 발견하지 못한다. clawflows 패턴 ([메타-인사이트 #9](../../harness-engineering-analysis.md))을 차용. 동시에 MCP server를 노출하면 Claude Code가 buddy를 *직접 호출 가능한 도구*로 사용할 수 있어 통합 깊이가 한 단계 올라감.

### Tasks

- **T1**: AGENTS.md auto-sync — buddy가 본인 capability(stats·doctor·task·sessions)를 사용자 프로젝트의 AGENTS.md / CLAUDE.md에 자동 기록·갱신. 위치: 신규 `internal/agentsmd/` 패키지. install/uninstall에 hook (opt-in 플래그).
- **T2**: plugin model — 사용자 정의 hook health checker 등록. 외부 binary 또는 Go plugin (`-buildmode=plugin`은 cross-compile 친화적이지 않음 → 외부 binary + IPC 방식 권장). 권장 IPC: v1.0 T3에서 도입하는 MCP transport 재활용 (stdio JSON-RPC). 즉 buddy가 spawn한 plugin process가 buddy에게 MCP tool로 노출. 일관성 + 별도 transport 학습 부담 0. 인터페이스 정의 위치: 신규 `pkg/plugin/`.
- **T3**: MCP server — Claude Code가 buddy를 tool로 직접 호출 (예: `mcp__buddy__doctor`, `mcp__buddy__stats`, `mcp__buddy__task_run`). 참조 SDK: `modelcontextprotocol/go-sdk` (또는 그 시점의 표준). 위치: 신규 `cmd/buddy-mcp/` 또는 `buddy mcp serve` subcommand.
- **T4**: cross-harness 시도? — Codex/OpenCode 지원. [Harness analysis](../../harness-engineering-analysis.md)가 "각 harness hook 형식이 다름"을 갭 C로 지적했고, [v0.1-spec.md §1 Non-goals](./v0.1-spec.md#1-non-goals-v01에서-안-하는-것)는 의도적 비범위로 명시. v1.0 시점이면 buddy 자체가 단단해진 상태이므로 *시도 가치*만 검토 (실현 시 별도 spec 문서 필요).

### Acceptance

- 외부 사용자가 자기 프로젝트에 `buddy install` 한 번으로 (a) hook reliability, (b) multi-session dashboard, (c) task DAG, (d) AGENTS.md 자동 동기화, (e) Claude Code에서 MCP tool로 직접 호출 — 5개 모두 사용 가능.

### Open questions

- plugin model의 권한 경계 — 외부 binary가 buddy DB에 직접 쓸 수 있게 할지, IPC를 통해 buddy 프로세스에 위임할지. 후자가 안전하지만 latency/복잡도 부담. v0.3 사용 패턴 보고 결정.

---

## 7. 비범위 (의도적으로 안 할 것)

| 항목 | 사유 |
|------|------|
| Web UI 풀스택 (React/Vue 등) | CLI + 가벼운 dashboard(v0.2 T4 결정에 따라 TUI 또는 minimal HTTP)로 충분. 풀스택은 유지보수 부담 대비 가치 낮음. |
| 분산 시스템 (멀티-머신 동기화) | single-machine 가정 유지 ([v0.1-spec.md §9 위험 매트릭스](./v0.1-spec.md#9-위험--open-questions)). 동기화는 v1.0+ 검토. |
| 자체 LLM provider proxy | Claude Code 본체의 책임. buddy는 transcript를 *읽기*만 한다 ([v0.1-spec.md §1 Non-goals](./v0.1-spec.md#1-non-goals-v01에서-안-하는-것)). |
| Windows 지원 | macOS / Linux 우선 ([README.md "1차 타겟"](../README.md#1차-타겟)). v1.0+ 검토. PID file·flock·`os.Getppid` 등 POSIX 가정이 다수. |
| Cross-harness parity (Codex/OpenCode 동시 지원) | [Harness analysis](../../harness-engineering-analysis.md) 갭 C에서 "환상"으로 지적된 영역. v1.0 T4에서 *시도 가치*만 검토. |
| Privacy: 자동 텔레메트리 / 사용 통계 외부 전송 | buddy는 로컬 도구. 외부 전송 0이 default. opt-in도 v1.0+ 검토. |

---

## 8. 결정 의존성 그래프

```
┌──────────────────────────────────────────────────────┐
│  M5 ✅ + M6 ✅ + v0.1.0 release ✅  (현재 위치)      │
│  PR #1, PR #2, tag v0.1.0 (2026-04-26)               │
└──────────────────┬───────────────────────────────────┘
                   │
                   ▼
            ┌──────────────┐
            │  Dogfood     │  (release binary 로 본격 시작 가능)
            │  Feedback    │
            └──────┬───────┘
                   │ (v0.2 UX / v0.3 tracker 통합 여부 입력)
                   ▼
        ┌──────────────┐     ┌──────────────┐
        │  v0.2        │     │  v0.3        │
        │  Control     │ ◀┄┄ │  Orchestration│
        │  Plane       │  ┊  │  (task DAG)  │
        └──────┬───────┘  ┊  └──────┬───────┘
               │          ┊         │
               │   (순서 독립)      │
               ▼                    ▼
        ┌─────────────────────────────────┐
        │  v1.0  통합                      │
        │  AGENTS.md / plugin / MCP server│
        └─────────────────────────────────┘
```

**의존 규칙:**

- ~~M5 → M6~~ ✅ 충족.
- ~~M6 → v0.1.0 release~~ ✅ 충족 (tag-triggered workflow 동작 확인).
- v0.1.0 → dogfood: 외부 사용자 install 가능한 상태이므로 dogfood 가 본격 시작 가능.
- dogfood feedback → v0.2 / v0.3 우선순위 입력.
- **v0.2 ↔ v0.3 순서 무관**: 데이터(SQLite) 공유, 코드 경로 독립. dogfood 결과 어느 쪽 수요가 큰지로 결정.
- v1.0 = v0.2 + v0.3 통합 위에서만 의미.

**dogfood feedback의 영향 범위 (v0.1.0 release 이후):**

- v0.2: dashboard UX 형식 (TUI vs web) 결정에 영향.
- v0.3: tracker 통합 여부 결정에 영향.
- v0.2 i18n sweep: M5 deferred 항목(`ValidationError.Reason`, queries sentinels, en locale) + en locale 카탈로그 채우기.
- v0.2 release polish: SHA pinning, 별도 `ci.yml`, macOS notarization (M6 deferred).

---

## 부록: 참조 문서

- [`v0.1-spec.md`](./v0.1-spec.md) — v0.1 스펙 (M1~M4 결정 lock-in)
- [`DOGFOOD.md`](../DOGFOOD.md) — 첫 사용 가이드 (M5 4-friction 발견 출처)
- [`dogfood-feedback-template.md`](./dogfood-feedback-template.md) — 회고 템플릿 (M5 우선순위 입력)
- [`harness-engineering-analysis.md`](../../harness-engineering-analysis.md) — 갭 A/B/C/D/F/G 출처
- [`decision-1-schema-fields.md`](./decision-1-schema-fields.md) — 옵션 A schema 결정 근거
