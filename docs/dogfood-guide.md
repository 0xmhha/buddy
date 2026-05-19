# Buddy Dogfood Guide — v0.7.x

> **이 문서의 목적**: dogfood 실행자(현재는 본인, 향후는 contributor 도)가 *어떻게 dogfood 를 돌리고*, *결과를 어떻게 기록하고*, *기록을 어떻게 코드 fix 또는 ADR 로 연결하는지* 의 단일 가이드.
>
> Plugin v1.0.0 entry condition B-2 ("plugin+cli 통합 dogfood production-proven — 외부 SaaS 1건 이상 end-to-end") 의 *유일한 실행 절차*.
>
> v0.1 시절의 hook-only 가이드는 [`DOGFOOD.md`](../DOGFOOD.md) 가 historical reference 로 보존.

---

## 0. 사전 조건

### 0.1 binary / plugin 설치

```bash
# binary 최신 release (v0.7.4 이상)
gh release download -p 'buddy_*_darwin_arm64' -D /tmp -R 0xmhha/buddy
mv /tmp/buddy_*_darwin_arm64 /usr/local/bin/buddy
chmod +x /usr/local/bin/buddy
buddy --version    # buddy 0.7.4+ 확인

# plugin (Claude Code 안)
/plugin install buddy
/reload-plugins
# 확인: /plugin 목록에 "buddy 0.7.4 enabled"
```

### 0.2 안전 장치

```bash
# settings.json 백업 (hook install 사용 예정 시)
cp ~/.claude/settings.json ~/.claude/settings.json.pre-dogfood.$(date +%Y%m%d)

# 첫 dogfood 는 cliwrap 없이 시작 (실패 모드 격리)
# DOGFOOD.md §0.2 참조
```

### 0.3 dogfood-result 기록 위치

이번 dogfood cycle 의 결과는 `docs/notes/{YYYY-MM-DD}-dogfood-result-cycle-{N}.md` 에 기록한다 (cycle 1 의 형식: [`docs/notes/2026-05-10-dogfood-result-cycle-1.md`](./notes/2026-05-10-dogfood-result-cycle-1.md) 참조). 새 cycle 시작 시 N+1 증가.

---

## 1. Three dogfood surfaces

cli buddy 의 *unified control plane* 약속을 검증하려면 세 surface 모두 통과해야 한다. 각각 독립적으로 실행 가능.

### 1.1 Plugin (Claude Code session 안)

**무엇**: 148 skill + 99 slash command + MCP tools (feature_*, analytics_query_*) 가 실제 Claude Code 세션에서 동작하는지.

**시나리오**:
1. 새 아이디어로 `/buddy:concretize-idea` 실행 → 9-phase orchestrator chain (concretize → define-features → design-system → plan-build → build-feature → verify-quality → ship-release → iterate-product → manage-lifecycle) 진입
2. 도중 자연 발화로 stage skill 직접 호출 (e.g., `/buddy:review-engineering`, `/buddy:design-billing-system`)
3. autoplan: `/buddy:autoplan` 으로 4-mode review (scope/eng/design/devex) 통과
4. feature registry: `/buddy:query-feature-registry` 로 기존 feature 검색
5. hook stats: 작업 중 `/buddy:status` 나 `buddy stats` 로 hook 신뢰성 관찰

**무엇을 보나**:
- skill 의 self-check 가 의도된대로 작동하는지
- next-phase cascade 가 자연스러운지 (auto_cascade 미적용 시 사용자가 명시적 호출)
- AskUserQuestion 의 horizontal-questions 가 *질문 폭주* 없이 한 번에 묶이는지
- friend-tone 페르소나 (silent default / Korean / no emoji) 유지

### 1.2 cli buddy agent runtime

**무엇**: `buddy agent create/run/scheduler` + `buddy tui` 7 modes 가 실제 multi-step 자동화에 사용 가능한지.

**시나리오**:
1. agent spec 작성 (e.g., `examples/webtoon-agent/spec.yaml` 참고)
2. `buddy agent create my-spec.yaml` → 등록 확인
3. on-demand 실행: `buddy agent run <id>` → `claude` subprocess spawn → chain 진행
4. scheduler 모드 (별 shell): `buddy agent scheduler` (cron tick + live refresh)
5. TUI 검증:
   - `buddy tui` 진입
   - `j/k` 네비, `Enter` 상세, `t` 로그 tail, `e` 편집, `c` 생성, `d` 삭제, `s` scheduler preview, `H` hook stats
   - 각 mode 의 esc/h 복귀, q 종료 검증

**무엇을 보나**:
- subprocess spawn 의 stdout/stderr 처리 (실 `claude` CLI 와 buddy 의 LogSink streaming)
- continue_on_fail / auto_cascade 의 의도된 의미
- scheduler tick 의 overlap 방지
- TUI 의 키 무반응 / 화면 깜빡임 / 색상 / 한글 깨짐
- 에러 메시지가 friend-tone (한국어 + 친구톤) 유지

### 1.3 Hook reliability monitor

**무엇**: v0.1.0 의 hook-wrap + daemon + aggregator + stats pipeline 이 실제 long-running 사용에서 신뢰성 데이터를 정확히 모으는지.

**시나리오**:
1. `buddy install` 로 hook wrapping 진입 (settings.json 편집)
2. `buddy daemon start` (배경 daemon)
3. 평소처럼 Claude Code 며칠~일주일 사용
4. `buddy stats --window 1h` / `--window 24h` 로 hook count/failures/p50/p95 확인
5. `buddy events --follow` 로 실시간 hook event 관찰
6. `buddy tui` 의 `H` 키 (ModeHookStats) 에서 같은 데이터를 TUI 로 확인 — CLI 와 TUI 결과 일치 여부 검증

**무엇을 보나**:
- hook 호출이 *누락 없이* outbox 로 기록되는지
- p50/p95 latency 가 합리적 수치인지 (예: PreToolUse Bash p95 > 1초면 의심)
- daemon crash / restart 시 outbox 가 *유실 없이* 복구되는지
- `buddy purge` 로 오래된 데이터 정리 시 *현 작업 중 데이터* 가 안 다치는지

---

## 2. 실행 시나리오 매트릭스

| 시나리오 | 소요 시간 | 검증 대상 | B-2 인정 |
|---------|----------|----------|---------|
| **Quick smoke** | 5 분 | 3 surface 각각 기본 동작 (binary launch / TUI 7 modes / `buddy stats` 출력) | ❌ (개발자 sanity check) |
| **Short cycle** | 1 day | 단일 plugin skill chain 1회 완주 + 단일 agent 1 회 실행 + 24h hook stats | ❌ (cycle 종료용 checkpoint) |
| **Production cycle** | 1 week+ | 실 외부 SaaS 프로젝트에서 plugin+cli 통합 end-to-end (9-phase 중 ≥3 phase 사용, agent ≥1 회 schedule 실행 완주, hook stats 누락 0) | ✅ **B-2 인정** |

---

## 3. 피드백 수집

### 3.1 Where to record

새 cycle 시작 시:

```bash
cp docs/dogfood-feedback-template.md docs/notes/$(date +%Y-%m-%d)-dogfood-result-cycle-N.md
# N = 다음 순번. 최근 cycle 의 N+1 (cycle-1 = 2026-05-10).
```

작성 중에는 *모든 finding* 을 발견 즉시 기록 (나중 회상 시 손실 큼).

### 3.2 Per-finding format

각 finding 은 다음 4 필드로 기록:

```markdown
### B{N} — {category} — {one-line title}

**Surface**: plugin | cli-buddy | hook-monitor | cross-surface
**Severity**: blocker | high | medium | low
**Repro**: (3-5 줄, 누가 봐도 재현 가능한 단계)
**Expected vs Actual**:
- expected: ...
- actual: ...

**Recommendation**: bug-fix | UX-fix | new-feature | ADR | documentation | not-fixable
```

`B{N}` 은 cycle 내 finding 번호 (cycle 1 의 B1-B7 같은 형식 유지).

### 3.3 Severity scale

| Severity | 정의 | SLA |
|----------|------|------|
| **blocker** | 핵심 사용 경로가 막힘 (e.g., `buddy agent run` 즉시 crash, plugin install 실패) | 다음 cycle 안에 fix + 별 release |
| **high** | 정상 동작이지만 안전한 default 가 깨짐 (e.g., 에러 메시지가 영어/한국어 혼재, data loss 가능성) | 다음 release 안에 fix |
| **medium** | UX 마찰 / 가독성 문제 (e.g., 색상 안 보임, 키 hint 헷갈림) | 다음 minor bump 에 묶음 |
| **low** | nice-to-have / 미관 | deferred items 에 추가 |

---

## 4. Improvement workflow

발견된 finding 을 코드 fix / ADR / deferred 중 어디로 보낼지 결정하는 분기 룰.

### 4.1 Triage rules

```
finding 발견
    │
    ▼
재현 가능?
    │ NO → finding 노트에 "non-repro" 표시, 추가 dogfood 에서 재발화 대기
    │ YES
    ▼
1줄로 root cause 가 명확한가?
    │ NO → /buddy:investigate (systematic-debugging) 진입 — 4-phase 분석 후 root cause 명확화
    │ YES
    ▼
recommendation 분기:
  ├─ bug-fix: 즉시 fix → /buddy:tdd → commit → 다음 patch release
  ├─ UX-fix: severity 에 따라 (high → 즉시, medium → 묶음, low → deferred)
  ├─ new-feature: 적당 시 ADR 작성 (§4.5) → impl → cycle close
  ├─ ADR (decision-only): /buddy:write-adr 또는 직접 작성 → README ADR Index 갱신
  ├─ documentation: 단일 docs commit + 다음 release 의 doc-and-decisions patch 에 묶음
  └─ not-fixable: deferred items 에 명시 (§4.6)
```

### 4.2 Bug → fix cycle (즉시)

severity = blocker / high:

1. `git checkout -b fix/{short-desc}` (worktree 권장: `git worktree add ...`)
2. `/buddy:tdd` → RED: failing test 추가 (bug 를 expose)
3. fix 코드 작성 → GREEN: test PASS
4. REFACTOR (필요 시)
5. `make verify-versions` + `go build / vet / test -race` + `make test-skill-form --strict` 4-gate PASS
6. `feat(...)` 또는 `fix(...)` commit (conventional commits)
7. `git push` + 새 cycle 종료 시 patch release (v0.7.x+1)

### 4.3 UX issue → cycle decision

severity = medium:

1. dogfood-result note 에 UX finding 기록
2. *같은 영역 finding 2개 이상 누적* 시 → high 로 escalate → 즉시 fix
3. 단일 finding 이면 다음 minor 에 묶음 (UX-fix-batch commit)

### 4.4 Missing feature → ADR 절차

dogfood 중 "이 기능이 있으면 좋겠다" 발견 시:

1. **즉시 코딩 금지**. ADR 먼저.
2. 기존 ADR Index 확인 — 동일 영역 ADR 이 있으면 *amending ADR* 작성
3. ADR 양식: `docs/superpowers/decisions/README.md` §"ADR 양식 (표준 7 섹션)" 따름
4. ADR 작성 후 사용자 confirm → Accepted 전환 → impl cycle 진입
5. impl 완료 시 ADR 의 §5 Verification 채움

### 4.5 Deferred → cycle-handoff §4.x

당장 fix 안 하기로 결정한 finding 은 명시적으로 deferred 처리:

1. `docs/notes/{date}-dogfood-result-cycle-N.md` 의 §"deferred" 섹션에 finding 추가
2. trigger 조건 명시 (e.g., "trigger: 동일 finding 2회 이상 재발 시 escalate")
3. HANDOFF.md 의 deferred 표에도 mirror (cross-reference)

### 4.6 Cycle close

dogfood cycle 종료 시:

1. dogfood-result note 의 모든 finding 분류 통계 (blocker/high/medium/low 개수)
2. cycle 안에 fix 된 것 vs deferred vs ADR escalated 통계
3. *다음 cycle 의 trigger* 명시 (e.g., "다음 release v0.8.0 출시 후 1주 사용 후 cycle 3 진입")
4. cycle-handoff 한 줄 요약 (다른 세션이 이어받을 때 첫 5분 안에 파악할 수 있도록)

---

## 5. Release cadence × dogfood

이 프로젝트는 *frequent small patch* cadence:

| Trigger | Release type | 예시 |
|---------|--------------|------|
| blocker fix | patch (v0.7.X+1) | hot-fix |
| 다수 UX-fix 묶음 | patch (v0.7.X+1) | UX-batch |
| 1개 새 feature | minor (v0.X+1.0) | W3-2 follow-on bundle |
| 다수 feature + ADR | minor (v0.X+1.0) | W3-2/W3-4 bundle |
| Plugin v1.0.0 entry conditions 모두 만족 | major (v1.0.0) | trigger: B-2 production-proven |

dogfood 의 *각 finding* 은 적절한 release type 에 묶여서 publish. 5 release in 1 day 도 가능 (이번 session 의 v0.7.0~v0.7.4 패턴).

---

## 6. Reference materials

### 6.1 Per-surface 상세

- **plugin buddy 9-phase**: `docs/superpowers/specs/2026-05-06-lifecycle-orchestrator-architecture.md`
- **cli buddy spec**: `docs/cli-buddy-spec.md` (ADR-005 locked-in)
- **hook reliability**: `DOGFOOD.md` (v0.1 가이드, hook monitor 부분만 유효)

### 6.2 Decision history

- **ADR Index**: `docs/superpowers/decisions/README.md`
- 최신 ADR (ADR-006~008): B6 lint / router scope / feature-management-mcp scope

### 6.3 Track identity

- **charter**: `docs/two-tracks-charter.md` (plugin buddy vs cli buddy 책임 경계)
- **session handoff**: `docs/HANDOFF.md` (최신 release / Phase 표 / deferred items)

### 6.4 v1.0.0 entry status

- **B-1 cli buddy W3 cascade**: ✅ Done (v0.7.0~v0.7.2)
- **B-2 production dogfood**: ❌ *이 가이드 의 대상*
- **B-3 PROCEDURE B6 + --strict**: ✅ Done (ADR-006)
- **B-4 router smart-skip**: ✅ Done (ADR-007 supersede)

→ B-2 만 남음. dogfood 가 *진척이고 동시에 trigger*.

---

## 7. 부록 — Quick-start 매크로

### 7.1 dogfood DB 빠른 셋업

```bash
DB=/tmp/dogfood-$(date +%Y%m%d).db
rm -f "$DB"
buddy agent --db "$DB" create examples/webtoon-agent/spec.yaml 2>/dev/null
buddy agent --db "$DB" list
buddy tui --db "$DB"
```

### 7.2 hook monitor 일주일 모니터링

```bash
# 시작
buddy install
buddy daemon start

# 매일 한 번 (cron 또는 수동)
buddy stats --window 24h > ~/dogfood-$(date +%Y%m%d)-stats.txt

# 일주일 후
buddy stats --window 24h | head -20
buddy events --limit 1000 | tail -100   # 최근 1000 events 중 100 줄 샘플
```

### 7.3 finding 즉시 기록 alias

```bash
# ~/.zshrc 또는 ~/.bashrc 에 추가
alias dfnote='echo "### B-tbd — {category} — {title}\n\n**Surface**: ...\n**Severity**: ...\n**Repro**: ...\n**Expected vs Actual**: ...\n**Recommendation**: ..." >> docs/notes/$(date +%Y-%m-%d)-dogfood-result-cycle-N.md'
```
