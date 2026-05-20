# Buddy Dogfood Cycle 3 — 2026-05-20 ~ (진행 중)

> **Guide**: [`docs/dogfood-guide.md`](../dogfood-guide.md) 따름.
> **선행 cycles**:
> - cycle-1 ([`./2026-05-10-dogfood-result-cycle-1.md`](./2026-05-10-dogfood-result-cycle-1.md)) — minimum-viable subset 통과로 close (2026-05-10).
> - cycle-2 ([`./2026-05-19-dogfood-result-cycle-2.md`](./2026-05-19-dogfood-result-cycle-2.md)) — baseline v0.7.5 가 1 일 만에 stale (v0.13.0 까지 6 release ship) 로 *superseded* close (2026-05-20). user-paced 실행 시작 전 baseline 이 무효화된 lesson 이 본 cycle 의 baseline-as-of convention 으로 promote.
>
> **Baseline-as-of**: v0.13.0 / 2026-05-20 / commit `d5f2755`
> 이 cycle 은 이 baseline 이 stale 되기 전 종료 의무. 다음 cycle 은 변화된 baseline + 변경 이유를 명시 (cycle-2 lesson promote).
>
> **Function (cycle 의 역할)**: plugin v1.0.0 entry conditions 9 개 중 **8/9 closed**, 마지막 unclosed = **B-2 production dogfood** 단 1 개 ([`docs/BACKLOG.md`](../BACKLOG.md) §1 / §2.E 참조). 본 cycle 의 진척 = 사실상 v1.0.0 ship gate. B-2 충족 또는 명시적 partial-cycle close 둘 중 하나로 끝남.
>
> **Surface 확장 (cycle-2 → cycle-3)**: cycle-2 의 3 path (plugin / cli-buddy agent / hook-monitor) 에 W7 신규 5 surface (`buddy session` / `usage` / `knowledge` / `advise` / `notify`) 가 추가. 총 8 surface. §B 는 3 primary + 5 add-on (dependency DAG) 으로 tiered.

---

## §A. Pre-flight (cycle 시작 시 실행)

cycle 시작 시 아래 표를 채움. baseline-as-of 마커가 stale 되기 전 (= 새 release ship 전) §A 확정.

| 체크 | 명령 | 결과 |
|------|------|------|
| binary 버전 | `buddy --version` | (TBD — `buddy 0.13.0` 기대) |
| 5 SSoT version agree | `make verify-versions` | (TBD) |
| 23 packages race-clean | `go test -race -count=1 ./...` | (TBD) |
| 148 skill PROCEDURE lint | `make test-skill-form --strict` | (TBD) |
| `~/.buddy/buddy.db` 존재 | `ls -la ~/.buddy/buddy.db` | (TBD) |
| daemon 가동 | `buddy doctor` | (TBD) |
| `buddy --help` 16 subcommand + W7 5 (`session/usage/knowledge/advise/notify`) 노출 | `buddy --help` | (TBD) |
| TUI 7 modes 정상 | `buddy tui` 진입 후 `j/k/Enter/t/e/c/d/s/H` 키 1 회씩 | (TBD) |

각 항목 ✅ / ⚠️ / ❌ 로 표기. quick smoke (guide §2.1) 5 분 안에 통과해야 §B 진입.

---

## §B. Short cycle (1 day) — 사용자 직접 진행

cycle-2 함정 (8 surface flat list 시도 시 baseline stale) 방지 위해 *tiered*:

- **Primary 3** (§B.1 ~ §B.3): cycle-2 와 동일. user-paced 1-day 안에 모두 완주 권장. partial 도 cycle close 시 인정.
- **Add-on 5** (§B.4.a ~ §B.4.e): W7 신규 surface. dependency DAG 순서로 진행 — 상위가 ingest/persist 안 된 상태에선 하위 surface 가 empty. 1-day 안에 ≥1 add-on 완주 권장.

8/8 다 못 끝내도 partial-cycle 인정 (cycle close 시 미진행 path 는 §D deferred 로 명시).

### B.1 Plugin path (primary)

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
- 특히 `/buddy:concretize-idea` 의 6 forcing question 이 실제로 *demand reality* 를 분리해주는가

### B.2 cli buddy agent path (primary)

```bash
DB=/tmp/dogfood.db
buddy agent --db "$DB" run hello-world
# → 실 claude subprocess spawn → define-features 실행 → result JSON 출력 확인

# 또는 cascade-demo (auto_cascade: max_depth: 3 + continue_on_fail)
buddy agent --db "$DB" run cascade-demo
# → 의도된 cascade 발화 확인 (depth 0 → 1 → 2 → 3 STOP)
```

**무엇을 보나** (guide §1.2):
- subprocess spawn 의 stdout/stderr 가 `buddy agent log` 에 line-by-line
- continue_on_fail 동작
- auto_cascade depth 제한
- TUI 7 mode 키 입력 정상

### B.3 Hook monitor path (primary)

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
- hook 호출 누락 없이 outbox 기록
- p50/p95 latency 합리적
- TUI hook-stats pane 과 CLI 결과 일치

### B.4 W7 surface add-ons (optional, dependency DAG 순서)

```
session ──┬── usage ──────┐
          │               ├── advise ── notify
          └── knowledge ──┘
```

#### B.4.a `buddy session` (foundation — 타 W7 surface 에 transcripts 공급, ADR-012)

```bash
buddy session list                  # active sessions
buddy session list --include-ended  # 종료된 것도
buddy session show <id>             # 단일 세션 상세
buddy session purge --before 7d --dry-run  # 7일 이전 정리 (dry-run)
```

**무엇을 보나**:
- hybrid hook+fsLister 가 hook 미수신 세션도 잡는가
- daemon poll (default 60s) 이 새 세션을 add 하는가
- `show` 의 turn count / token usage 가 실제 transcript 와 일치하는가

#### B.4.b `buddy usage` (sessions 의존, ADR-013)

```bash
buddy usage today        # 24h 토큰/세션
buddy usage trend --days 7
buddy usage top --limit 10
```

**무엇을 보나**:
- sessions 가 empty 면 7 metric 모두 0 (foundation 확인)
- trend graph 가 시간 축 정렬되는가
- top N 의 token 합산이 session show 와 일치

#### B.4.c `buddy knowledge` (sessions 의존 — ingest 필요, ADR-014)

```bash
buddy knowledge ingest             # 모든 session transcript chunk
buddy knowledge ingest --embed     # 임베딩까지 (python script 필요)
buddy knowledge stats              # chunk 수 / embedding coverage / last ingest
buddy knowledge query "내 질문"    # BM25 (default)
buddy knowledge query "내 질문" --mode hybrid  # BM25 + vector
```

**무엇을 보나**:
- ingest 가 idempotent (재실행 시 dedup)
- BM25 retrieval 가 의미 있는 chunk 반환
- hybrid 가 BM25 단독보다 정밀도 향상 체감되는가
- embedding 스크립트 미설치 시 friend-tone fallback

#### B.4.d `buddy advise` (usage+knowledge 의존, ADR-015 + ADR-017 drift rule)

```bash
buddy advise                        # 현재 advisory (실시간 평가, persist X)
buddy advise --persist              # advisories 테이블에 기록
buddy advise --history --since 7d   # 기록된 advisory 조회
buddy advise --mute-kind <kind>     # 특정 kind 전체 mute
```

**무엇을 보나**:
- 5 rule (cost / hour / kind / drift / TBD) 가 실제 trigger 되는가
- friend-tone 한국어 메시지
- KindGoalDrift (ADR-017) 가 cosine 임계값 넘을 때만 fire
- usage/knowledge empty 면 "지금은 특별히 알릴 조언이 없어." 출력 확인

#### B.4.e `buddy notify` (advise → dispatch, ADR-016)

```bash
# 4 channel 각각 synthetic dispatch
buddy notify test --channel desktop
buddy notify test --channel webhook
buddy notify test --channel tui-banner
buddy notify test --channel shell

buddy notify status --limit 20      # notification_log v8 조회
```

**무엇을 보나**:
- 4 channel 모두 dispatch 성공 (env 설정 시; 미설정 채널은 friend-tone error)
- per-channel severity floor 가 적용
- dedup (동일 advisory 2회 dispatch 안 됨)
- daemon auto-dispatch 가 background 에서 advise rule fire 시 자동 호출

---

## §C. Findings (발견 즉시 추가)

> 각 finding 은 guide §3.2 의 4-field format. ID 는 `B<surface>-<n>` (e.g., `B1-1` plugin 첫 finding / `B4a-1` session 첫 finding).

### B?-? — (template — 실 finding 으로 교체)

**Surface**: plugin | cli-buddy | hook-monitor | session | usage | knowledge | advise | notify | cross-surface
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
| Cycle 시작 | 2026-05-20 |
| Cycle 종료 | (TBD) |
| Baseline-as-of (close 시점) | v0.13.0 / commit `d5f2755` — stale 여부: (TBD) |
| Primary surface 별 사용 시간 | plugin __ h / cli-buddy __ h / hook-monitor __ days |
| Add-on surface 완주 | session __ / usage __ / knowledge __ / advise __ / notify __ |
| Findings 분류 | blocker __ / high __ / medium __ / low __ |
| Cycle 안에 fix 된 것 | __ |
| Deferred | __ |
| ADR escalated | __ |
| **v1.0.0 ship gate (B-2)** | (충족 / partial / 미충족 — 결정 사유) |
| 다음 cycle trigger | __ |

**Baseline promotion 의무**: cycle close 시 새 release (v0.14.0+) 가 ship 되어 baseline 이 stale 됐다면, 다음 cycle 의 baseline-as-of frontmatter 에 *변경된 commit + 변경 이유* 를 반드시 명시 (cycle-2 lesson canonical).

---

## §F. 즉시 가능한 action (사용자가 진행)

1. **Pre-flight (§A)**: 위 8 항목 5 분 안에 통과 → §B 진입.
2. **Primary path (§B.1 ~ §B.3)**: 각자 1 회 이상 완주. 마찰 finding 즉시 §C.
3. **Add-on (§B.4)**: DAG 순서로 ≥1 surface 완주. session → usage → knowledge → advise → notify 가 자연스러운 흐름.
4. **Baseline stale watch**: cycle 진행 중 `git log --oneline v0.13.0..HEAD` 가 비지 않으면 baseline 이 stale — *즉시 cycle close 결정* (cycle-2 함정 재발 방지).

각 path 는 독립적으로 진행 가능. 한 path 만 해도 partial cycle 인정.
