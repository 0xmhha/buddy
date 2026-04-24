# Plan — Dogfood 준비 + Roadmap 문서

> 목적: v0.1 첫 dogfood를 매끄럽게 시작하기 위한 가이드를 만들고, M5 이후
> 작업 리스트를 한 문서에 정리한다. 실제 dogfood 사용은 사용자 본인의 행동.

---

## Task 1 — DOGFOOD.md 가이드

**Goal**: 사용자가 본인 머신에서 buddy v0.1을 안전하게 켜고, 며칠 사용하며
feedback 데이터를 모을 수 있는 *완전한 step-by-step 가이드*.

**Files to create:**
- `DOGFOOD.md` — repo 루트
- (선택) `docs/dogfood-feedback-template.md` — 사용자가 며칠 후 작성할 회고 템플릿

**가이드가 다뤄야 할 것:**

### 0. 안전 장치 (사용자 안전 우선)
- `~/.claude/settings.json` 백업 절차 (buddy install이 자동 백업하지만 *수동
  복사*도 안내 — 신뢰성 1순위)
- `--with-cliwrap` 사용 시 cliwrap-agent / cliwrap binary 설치 여부 확인
  (`go install github.com/0xmhha/cli-wrapper/cmd/cliwrap@latest`)
- 첫 사용은 **--with-cliwrap 없이** 시작 권장 (실패 격리 단순화)

### 1. 설치
```bash
# 1) 빌드
cd /path/to/buddy
make build

# 2) 절대 경로 확인
BUDDY_BIN="$PWD/bin/buddy"
echo "$BUDDY_BIN"

# 3) install (default — without cliwrap)
$BUDDY_BIN install --buddy-binary "$BUDDY_BIN"

# 4) settings.json 백업 위치 확인
ls -la ~/.claude/settings.json.buddy.bak

# 5) wrapping 결과 일별
diff ~/.claude/settings.json.buddy.bak ~/.claude/settings.json
```

### 2. daemon 시작
```bash
$BUDDY_BIN daemon start
$BUDDY_BIN daemon status   # running (pid …)
```

### 3. 사용 — Claude Code 평소처럼 쓰기
- 며칠 동안 평소처럼 Claude Code 사용
- 매일 한 번 정도 `buddy doctor`, `buddy stats --window 24h --by-tool`
- 무엇이든 어색하면 `buddy events --limit 50` 으로 raw 확인

### 4. 트러블슈팅 (자주 발생할 만한 것 미리 정리)
- `daemon이 실행 중이 아니야` → `buddy daemon status` 후 stop/start
- outbox backlog 알림 → daemon 재시작 또는 batch 크기 늘려보기
  (`buddy daemon run --batch 2000` 임시 실행)
- hook이 평소보다 느려진 느낌 → `buddy stats --by-tool` 으로 어떤 tool인지
- settings.json 손상 의심 → `cp ~/.claude/settings.json.buddy.bak ~/.claude/settings.json`

### 5. 며칠 후 회고
- `dogfood-feedback-template.md` 사용 (제공)
- 데이터 수집 항목:
  - 매일의 hook 호출 수 (`buddy stats --window 24h | wc -l`)
  - 가장 자주 트리거된 hook 이름
  - p95 시간 변화 (느려지는 추세인가?)
  - silent failure 발견 케이스 (있으면 events에 기록 남음)
  - buddy 자체에 대한 마찰 포인트 (UX/문구/명령 부족 등)

### 6. 안전 종료
```bash
$BUDDY_BIN daemon stop
$BUDDY_BIN uninstall   # 백업에서 자동 복원
```

**Tone**: 친구 톤 페르소나와 일관 (호들갑 X, 차분한 안내). 코드 블록은
copy-paste 가능하게 정확히. 한국어 주.

**Subagent 임무:**
1. `DOGFOOD.md` 와 `docs/dogfood-feedback-template.md` 작성
2. 각 명령이 실제로 *현재 buddy 코드*에서 동작함을 *임시 디렉토리에서 dry-run*
   하여 검증 (e.g., `--claude-dir /tmp/dogfood-claude` 로 가짜 settings.json
   만들고 install→doctor→uninstall 한 번 돌려보고 friend-tone 메시지가
   가이드의 인용과 일치하는지 확인)
3. 발견된 마찰 포인트가 있으면 가이드의 "트러블슈팅" 섹션에 추가
4. README.md에 `→ 처음 켜는 사람은 [DOGFOOD.md](./DOGFOOD.md) 부터` 한 줄 추가

**Subagent가 결과 commit + 보고**

---

## Task 2 — docs/roadmap.md

**Goal**: M5 이후 작업을 한 문서에 정리. 우선순위·근거·acceptance criteria
포함. 미래 자아 또는 다른 contributor가 읽고 *왜 이 순서인지* 이해 가능.

**Files to create:**
- `docs/roadmap.md`

**구조:**

### 1. 한 줄 요약 + 현재 위치
- v0.1 status (M1~M4 완료)
- 향후 5개 마일스톤 한 표

### 2. M5 — Config / Threshold tuning / Purge / 페르소나 polish
**Why**: dogfood feedback이 들어오기 시작하면 *바로 조정*할 수 있는 손잡이
필요. config 명령 없이는 사용자가 spec §6.2 hard-coded threshold를 바꾸려면
recompile.

**Tasks:**
- T1: `~/.buddy/config.json` 스키마 + Zod-style validate (Go에서는 manual
  Validate 메서드)
- T2: `buddy config get/set/unset/show` CLI
- T3: doctor/aggregator/daemon이 config에서 threshold/poll/batch 읽도록
- T4: `buddy purge --before <date>` (오래된 events/stats 삭제, outbox는 건드리지
  않음)
- T5: 페르소나 메시지 catalog 정리 (i18n 토대 — locale=en fallback)

**Acceptance**: dogfood feedback에서 발견된 모든 "이건 좀…" 사례가 config 한
줄로 해결 가능.

### 3. M6 — Release prep
**Why**: 첫 release tag (v0.1.0) + 다른 사람이 install 가능한 형태.

**Tasks:**
- T1: cross-compile matrix (`linux/amd64`, `linux/arm64`, `darwin/amd64`,
  `darwin/arm64`) — Makefile target
- T2: GitHub Actions release workflow (tag → binary attach)
- T3: `buddy --version` 출력에 commit SHA + build date 포함
- T4: `CHANGELOG.md` v0.1.0 entry
- T5: README의 install 섹션을 `curl … | sh` 패턴으로 교체

**Acceptance**: `git tag v0.1.0 && git push --tags` 한 번에 GitHub release
페이지에 4개 binary 첨부.

### 4. v0.2 — Control Plane (멀티-세션 dashboard)
**Why**: 분석 보고서 §3 갭 G "통합 observability". buddy doctor/stats는
단일 세션 단위, dashboard는 *여러 Claude Code 세션 동시*.

**Tasks (큰 묶음):**
- T1: Recon 패턴 차용 — `~/.claude/sessions/*.json` 스캔으로 활성 세션 발견
- T2: Token usage parser — transcript JSONL에서 `usage` 추출 (already in
  schema option A)
- T3: 단일 세션 관점 stats를 *여러 세션 합산*으로 확장
- T4: TUI dashboard (charmbracelet/bubbletea?) 또는 web (port 9090?) — 결정
  필요
- T5: cost estimate (model price × tokens)

**Open question**: TUI vs web. dogfood 결과 어떤 UX가 더 자연스러울지가
결정에 영향.

### 5. v0.3 — Orchestration (task DAG executor)
**Why**: 분석 보고서 §3 갭 A,B. arc-reactor의 wave parallelization 패턴
+ retry loop.

**Tasks:**
- T1: task DAG schema + storage (SQLite or JSON-backed)
- T2: `buddy task add/list/run/status` CLI
- T3: wave 그룹화 (independent tasks 자동 병렬)
- T4: retry policy (exponential backoff, quality gate)
- T5: 외부 task tracker 통합? (Linear/Jira/GitHub Issues — 검토 필요)

**Open question**: buddy가 task tracker가 되어야 하는가, *기존* tracker의
*실행 엔진*만 되어야 하는가. 후자가 scope 작아 매력.

### 6. v1.0 — 통합 (AGENTS.md auto-sync, plugin model, MCP server)
**Why**: 분석 보고서 메타-패턴 "agent knowledge is explicit metadata".
clawflows 패턴 차용.

**Tasks:**
- T1: AGENTS.md auto-sync — buddy가 본인 capability를 AGENTS.md에 자동 기록
- T2: plugin model — 사용자 정의 hook health checker 등록
- T3: MCP server — Claude Code가 buddy를 tool로 직접 호출 (`mcp__buddy__doctor` 등)
- T4: cross-harness 시도? (Codex/OpenCode 지원) — 분석 보고서가 "환상"이라
  경고했지만 v1.0 시점이면 검토 가치

### 7. 비범위 (의도적으로 안 할 것)
- Web UI 풀스택 (CLI + 가벼운 dashboard로 충분)
- 분산 시스템 (single-machine 가정 유지)
- 자체 LLM provider proxy (Claude Code의 책임)
- Windows 지원 — v1.0+ 검토

### 8. 결정 의존성 그래프
```
M5 (config)         ──▶ M6 (release)         ──▶ v0.2 (dashboard)
                                                   │
                            v0.3 (orchestration) ◀┘
                                  │
                                  ▼
                            v1.0 (AGENTS.md/MCP/plugin)
```

dogfood feedback은 M5 task 우선순위에 직접 영향. v0.2와 v0.3은 독립적이라
순서는 dogfood 결과에 따라 결정.

**Subagent 임무:**
1. `docs/roadmap.md` 작성 (위 구조 그대로 살을 붙이되 친구 톤은 적당히 —
   roadmap은 정보 밀도 우선)
2. spec `v0.1-spec.md` §8 마일스톤 표를 roadmap 링크로 짧게 갈음
3. README의 "역사적 메모" 아래에 `다음 작업: docs/roadmap.md` 한 줄 추가
4. commit + 보고

---

## Workflow per task

1. Baseline SHA 기록
2. Implementation subagent dispatch (위 task 정의 그대로 전달)
3. HEAD SHA 기록
4. code-reviewer subagent dispatch — 문서 review (정확성, 빠진 항목, 친구 톤
   일관성, 명령어 실제 동작 일치 여부)
5. Critical/Important fix 후 다음 task

## After all tasks
- spec/README final consistency check
- finishing-a-development-branch
