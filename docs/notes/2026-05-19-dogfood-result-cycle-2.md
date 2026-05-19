# Buddy Dogfood Cycle 2 — 2026-05-19 ~ (진행 중)

> **Guide**: [`docs/dogfood-guide.md`](../dogfood-guide.md) 따름. cycle-1 ([`./2026-05-10-dogfood-result-cycle-1.md`](./2026-05-10-dogfood-result-cycle-1.md)) 의 후속.
>
> **Baseline**: v0.7.5 (released 2026-05-19). plugin v1.0.0 entry conditions 중 B-2 (production dogfood) 가 *유일 unclosed*. 본 cycle 의 목적 = B-2 *진척 + 가능하면 충족*.

---

## §A. Pre-flight (2026-05-19 자동 수행)

| 체크 | 결과 |
|------|------|
| `buddy --version` | `buddy 0.7.5 (sha=dev, built=unknown)` ✅ |
| `/tmp/dogfood.db` 4 agents 존재 | ✅ (cascade-demo / daily-report / hourly-cleanup / hello-world) |
| `buddy stats --window 1h` (seeded 3 hook_stats rows) | ✅ PreToolUse 170/p95 350ms / Stop 7/p95 8ms |
| `buddy --help` 16 subcommands 모두 표시 | ✅ |
| `buddy tui --help` 7 modes 언급 | ✅ |
| `make verify-versions` (5 SSoT) | ✅ 0.7.5 agree |
| `go test -race -count=1 ./...` (23 packages) | ✅ all green |
| `make test-skill-form --strict` | ✅ 148 / 62 allowlist / 86 pass / 0 deviate |

**Quick smoke (guide §2.1) PASS** — 5분 격리 검증 통과.

---

## §B. Short cycle (1 day) — 사용자 직접 진행 필요

### 진행 권장 시나리오

다음 3개 path 를 각자 1회 이상 완주 후 finding 기록.

#### B.1 Plugin path

```
# Claude Code 세션 안에서
/buddy:concretize-idea
  ↳ 본인의 진짜 아이디어 1개 입력 → forcing question 6개 답변
/buddy:define-features
  ↳ concretize 결과로 actor/use-case/feature backlog 도출
/buddy:autoplan
  ↳ 4-mode review (scope/eng/design/devex) 통과 시도
```

**무엇을 보나** (guide §1.1):
- skill 의 self-check 가 작동
- AskUserQuestion 폭주 없음
- 한국어 friend-tone 일관성
- (특히) `/buddy:concretize-idea` 의 6 forcing question 이 실제로 *demand reality* 를 분리해주는가

#### B.2 cli buddy agent path

```bash
# 실 claude CLI 가 있는 환경에서
DB=/tmp/dogfood.db
buddy agent --db "$DB" run hello-world
# → 실 claude subprocess spawn → define-features 실행 → result JSON 출력 확인

# 또는 cascade-demo (auto_cascade: max_depth: 3 + continue_on_fail)
buddy agent --db "$DB" run cascade-demo
# → 의도된 cascade 발화 확인 (depth 0 → 1 → 2 → 3 STOP)
```

**무엇을 보나** (guide §1.2):
- subprocess spawn 의 stdout/stderr 가 `buddy agent log` 에 line-by-line 보임
- continue_on_fail 동작
- auto_cascade depth 제한
- TUI 의 7 mode 키 입력 모두 정상 (이전 확인 — section confirmed working)

#### B.3 Hook monitor path

```bash
# 평소 Claude Code 사용을 며칠 지속
buddy install      # 이미 했다면 skip
buddy daemon start # 이미 돌고 있으면 skip
# 24h 후
buddy stats --window 24h
buddy events --limit 100
buddy tui --db ~/.buddy/buddy.db   # H 키 → 같은 데이터 TUI 검증
```

**무엇을 보나** (guide §1.3):
- hook 호출이 누락 없이 outbox 로 기록
- p50/p95 latency 합리적
- TUI hook-stats pane 과 CLI 결과 일치

---

## §C. Findings (발견 즉시 추가)

> 각 finding 은 guide §3.2 의 4-field format 따름.

### B1 — (template — 실 finding 으로 교체)

**Surface**: plugin | cli-buddy | hook-monitor | cross-surface
**Severity**: blocker | high | medium | low
**Repro**: (3-5 줄)
**Expected vs Actual**:
- expected: ...
- actual: ...
**Recommendation**: bug-fix | UX-fix | new-feature | ADR | documentation | not-fixable

---

## §D. Deferred / Out-of-scope

cycle 진행 중 "지금은 안 한다" 결정된 것은 여기 기록 + trigger 명시.

---

## §E. Cycle close (cycle 종료 시 채움)

| 항목 | 값 |
|------|---|
| Cycle 시작 | 2026-05-19 |
| Cycle 종료 | (TBD) |
| Surface 별 사용 시간 | plugin __ h / cli-buddy __ h / hook-monitor __ days |
| Findings 분류 | blocker __ / high __ / medium __ / low __ |
| Cycle 안에 fix 된 것 | __ |
| Deferred | __ |
| ADR escalated | __ |
| 다음 cycle trigger | (e.g., "v0.8.0 출시 후 1주") |

---

## §F. 즉시 가능한 action (사용자가 진행)

1. **Plugin path**: Claude Code 세션 안에서 `/buddy:concretize-idea` 로 본인 아이디어 1개 실행. 도중 마찰 finding 즉시 §C 에 추가.
2. **cli buddy path**: 위 §B.2 의 `buddy agent run hello-world` 실행. claude CLI 환경 필요. spawn 실패 시 그게 first finding.
3. **Hook monitor path**: 며칠간 `buddy stats` 매일 capture. 일주일 후 cycle close.

각 path 는 독립적으로 진행 가능. 한 path 만 해도 partial cycle 인정 (cycle close 시 미진행 path 는 deferred 로 명시).
