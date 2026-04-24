# M4 — User-facing CLI Surface

Goal: complete the v0.1 user-facing CLI on top of M1–M3 infrastructure
(`internal/db`, `internal/aggregator`, `internal/daemon`, `internal/cliwrapcfg`).

## References (read these first)
- `docs/v0.1-spec.md` §7 (CLI surface) and §6 (decisions: thresholds + friend persona)
- `cmd/buddy/main.go` for the existing cobra wiring pattern
- `README.md` for project tone/positioning ("친구")

## Persona reminder (Decision 3 lock-in)
- 침묵이 default: 정상 상태는 출력 없음
- 친구 톤 한국어 메시지 (이모지 X, 호들갑 X)
- critical 알림도 절제된 친구 톤
- debug/log는 구조적, 친구 톤 X

Decision 2 default thresholds (config로 변경 가능, 일단 hardcoded const):
```
hookTimeoutMs:    30000
hookSlowMs:        5000
hookFailRatePct:     20
outboxBacklog:     1000
notifyChannel: 'stderr'
```

---

## Task 1: `buddy install` / `buddy uninstall`

**Goal**: Claude Code의 `~/.claude/settings.json`에 hook을 자동 wrapping하고, `--with-cliwrap` 옵션 시 cliwrap.yaml도 생성. `uninstall`은 원복.

**Files to create:**
- `internal/install/install.go` — settings.json read/parse/modify/write, idempotent
- `internal/install/install_test.go` — temp HOME 기반 통합 테스트
- `cmd/buddy/main.go`에 `install` / `uninstall` 서브커맨드 추가

**Requirements:**
1. `~/.claude/settings.json`을 읽어 PreToolUse/PostToolUse/Stop 등 hook 정의를 buddy hook-wrap으로 감싸기
2. **Idempotent**: 이미 wrapping된 hook은 다시 wrapping하지 않음 (식별: command가 `buddy hook-wrap`으로 시작)
3. **원본 보존**: 첫 install 시 원본 settings.json을 `settings.json.buddy.bak`으로 백업
4. `--with-cliwrap` 플래그: `cliwrapcfg.Render()`로 `~/.buddy/cliwrap.yaml` 생성
5. `uninstall`: 백업에서 원복하거나, 각 hook의 buddy wrapping을 제거
6. **--db / --buddy-binary 플래그**: 테스트 가능하도록 path injection 가능하게

**Tests:**
- 깨끗한 settings.json에 install → 모든 hook이 wrapping됨
- 이미 wrapping된 settings.json에 install → 변화 없음 (idempotent)
- install → uninstall → 원본과 동일
- `--with-cliwrap` 시 cliwrap.yaml 생성됨
- 친구 톤 메시지 ("등록 완료. 이제 옆에서 보고 있을게.")

**Edge cases:**
- settings.json 자체가 없을 때 → 깔끔한 친구 톤 에러
- 백업 파일이 이미 있을 때 → 덮어쓰지 않음 (사용자 작업 보호)

---

## Task 2: `buddy doctor`

**Goal**: hook health 즉시 진단. 1회 스냅샷, daemon 의존 X (read-only).

**Files to create:**
- `internal/diagnose/doctor.go` — 진단 로직 (Status struct + Check() 함수)
- `internal/diagnose/doctor_test.go` — 다양한 상태 케이스 테스트
- `cmd/buddy/main.go`에 `doctor` 서브커맨드 추가

**Check 항목 (각 항목당 친구 톤 메시지):**
1. **Daemon 상태**: `daemon.CheckStatus()` 호출
2. **Outbox backlog**: pending row 수 > 1000이면 경고
3. **Slow hooks**: 최근 5min p95 > 5000ms 인 hook 나열
4. **High failure rate**: 최근 5min 실패율 > 20% 인 hook 나열
5. **DB 접근 가능 여부**: 읽기 실패면 첫 줄에 보고

**Output 형식:**
```
$ buddy doctor
모두 정상이야.                        # 정상
또는
어, 몇 가지 봐줄 게 있어.

  • 'pre-commit' hook이 좀 느려졌어. p95가 8.2초 (기준 5초).
  • 'lint' hook 실패율이 35% 야. 최근 20번 중 7번 실패.
  • outbox에 1,247개 쌓였어. daemon 한 번 봐줘.
  • daemon이 실행 중이 아니야. 'buddy daemon start'로 띄울 수 있어.
```

**Exit code**: 정상이면 0, 경고 있으면 1.

**Tests:**
- 빈 DB → "모두 정상이야"
- backlog 초과 시 경고 메시지
- slow p95 시 경고
- 높은 failure rate 시 경고
- daemon 안 떠있을 때 메시지

---

## Task 3: `buddy stats`

**Goal**: hook_stats 조회 + 사람이 읽기 좋은 형식 출력.

**Files to create:**
- `internal/queries/stats.go` — hook_stats SELECT + format
- `internal/queries/stats_test.go`
- `cmd/buddy/main.go`에 `stats` 서브커맨드 추가

**Flags:**
- `--window <5m|1h|24h>` (default 1h, 매핑: 5m→5, 1h→60, 24h→1440)
- `--by-tool` (있으면 hook×tool 단위 row, 없으면 hook 단위 합산)
- `--hook <name>` (필터)
- `--db <path>`

**Output 예시:**
```
지난 1시간 hook 통계 (—by-tool 없을 때)

  hook            count   p50    p95    실패율
  pre-commit       234   12.3s  31.0s   8%
  lint             567    0.1s   0.2s   0%

지난 1시간 hook 통계 (—by-tool 있을 때)

  hook       tool     count   p50    p95    실패율
  pre-commit Bash      234   12.3s  31.0s   8%
  pre-commit Read     1892    0.05s  0.08s   0%
  ...
```

ms는 사람 친화적으로 (1234ms → 1.2s, 12345ms → 12.3s).

**Tests:**
- 빈 DB → "아직 기록된 hook 통계가 없어."
- 윈도우 필터 정확
- by-tool 분리 정확
- ms → human format 정확 (table-driven)

---

## Task 4: `buddy events`

**Goal**: hook_events tail. 디버그용. 기본 마지막 N개, `--follow`로 실시간.

**Files to create:**
- `internal/queries/events.go` — SELECT + format
- `internal/queries/events_test.go`
- `cmd/buddy/main.go`에 `events` 서브커맨드 추가

**Flags:**
- `--limit <n>` (default 20)
- `--hook <name>` (필터)
- `--follow` 또는 `-f` (poll mode, default 1s)
- `--db <path>`

**Output 형식 (한 줄):**
```
2026-04-24T11:23:45Z  pre-commit  PostToolUse  Bash  exit=0  dur=1.2s  sess=abc1234
```

**Persona: 이 명령은 debug 용이라 친구 톤 X**, 구조적 한 줄 output.

다만 `--follow`로 시작/종료 시 친구 톤 한 줄:
- 시작: `(따라가는 중. Ctrl-C로 멈춰.)`
- Ctrl-C: `(끝.)`

**Tests:**
- limit 정확
- hook 필터 정확
- format 정확 (table-driven)
- follow는 통합 테스트 어려우니 unit으로 polling loop만 테스트

---

## Workflow per task

1. 새 git branch 안 만든다 — 모두 main에 직접 commit (이 repo는 v0.1 PoC, branch 부담 X)
2. 각 task 시작 시: 현재 HEAD를 BASE_SHA로 기록
3. subagent가 구현 + 테스트 + commit
4. 완료 후 HEAD_SHA 기록
5. code-reviewer subagent로 BASE..HEAD diff 리뷰
6. Critical/Important issue 있으면 fix subagent 디스패치
7. 다음 task

## After all tasks

- 통합 sanity test: buddy install (temp HOME) → hook-wrap 흉내 → daemon run → stats / doctor / events 모두 동작 확인
- spec/README 업데이트
- finishing-a-development-branch
