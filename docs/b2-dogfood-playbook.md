# B-2 Production Dogfood — Execution Playbook

> **이 문서의 목적**: plugin v1.0.0 entry condition 9 개 중 마지막 1 개 unclosed = **B-2 production dogfood** 의 *사용자 직접 실행* playbook. *Day-by-Day* 진행 단위 + *자가 점검 매트릭스* + *common failure 대응*.
>
> **General guide 와 분담**:
> - [`dogfood-guide.md`](./dogfood-guide.md) — *어느 cycle 에나 적용* 되는 general 절차 (3 surface 정의 / finding 양식 / triage rules / release cadence). **본 playbook 의 모든 §는 거기 §X 를 전제**.
> - **본 playbook** — B-2 *충족* 에 특화된 *사용자 실행 가능* 절차. 매일 무엇을 / 얼마나 / 어떻게 측정하는지.
>
> **현 cycle**: [`notes/2026-05-20-dogfood-result-cycle-3.md`](./notes/2026-05-20-dogfood-result-cycle-3.md) — finding 은 모두 거기 §C 로. 본 playbook 은 *어떻게 실행해서 거기에 채울지*.

---

## §1. B-2 인정 기준 (정확한 조건)

[`BACKLOG.md`](./BACKLOG.md) W1-1 의 4 조건:

| # | 조건 | 측정 방법 | 인정 임계 |
|---|------|----------|----------|
| **C1** | 외부 SaaS 1건 이상 end-to-end | dogfood 안에서 *실제 외부 SaaS 프로젝트* 의 plugin+cli 통합 사용 | 1 건 |
| **C2** | 9-phase 중 ≥3 phase 사용 | `/buddy:*` slash 또는 PROCEDURE 호출 로그 기준 — concretize / define-features / design-system / plan-build / build-feature / verify-quality / ship-release / iterate-product / manage-lifecycle 중 사용된 phase | 3 phase 이상 |
| **C3** | agent ≥1 회 schedule 실행 완주 | `buddy agent scheduler` 가 한 cron tick 을 *완전히* 실행 + result JSON 산출 | 1 회 |
| **C4** | hook stats 누락 0 | `buddy stats` 의 `expected_count - actual_count = 0` (outbox 누락 없음) | 0 누락 |

→ 4 조건 모두 만족 시 **B-2 closed**. cycle-3 note 의 §E "Cycle close" 표에 각 카운터 채움.

---

## §2. 사전 준비 (Day 0, ~30 분)

### §2.1 환경 점검 (이미 끝났으면 skip)

```bash
buddy --version            # 0.13.0+ (cycle-3 baseline)
buddy doctor               # daemon 가동 / outbox / hook 설치 상태 한눈에
make verify-versions       # 5 SSoT 일치
```

### §2.2 daemon 가동 (C4 의 전제)

```bash
buddy daemon status        # not running 이면 →
buddy daemon start         # background
# 다시 확인
buddy doctor               # daemon ✅ 표시
```

**중요**: C4 (hook stats 누락 0) 는 *daemon 이 가동된 기간만* 측정. 며칠 멈춰 있으면 그 기간은 측정 불가. cycle-3 note 의 baseline-as-of 마커처럼 *daemon-up-as-of* 시점을 본 playbook 진행 시작 시점으로 둠.

### §2.3 진행 카운터 초기화

cycle-3 note 의 §E 표에 본인의 *현재 카운터* 를 한 번 기록:

```markdown
| C1 외부 SaaS | 0 건 |
| C2 9-phase | 0 phase |
| C3 agent schedule | 0 회 |
| C4 hook 누락 | (측정 시작 ___-__-__) |
```

매 Day 끝에 이 4 카운터를 *갱신* (덮어쓰기). diff 가 *그 Day 의 진척*.

---

## §3. Day-by-Day 실행 매트릭스

> **현실 인정**: 1주일 통째 못 비울 가능성 100%. 하루 *최소 30 분* 만 확보하면 5 ~ 10 일 안에 4 조건 충족 가능. 매일 *반드시* 해야 하는 것 + *기회 닿을 때만* 하는 것을 분리.

### §3.1 매일 (Daily — 5 분, 무조건)

| 작업 | 명령 | 의미 |
|------|------|------|
| 1. daemon 살아있는지 | `buddy doctor` | C4 측정 연속성 유지 |
| 2. 24h 통계 capture | `buddy stats --window 24h > ~/dogfood-$(date +%Y%m%d)-stats.txt` | C4 evidence 누적 |
| 3. cycle-3 note 갱신 | §E 의 4 카운터만 덮어쓰기 | self-tracking |

→ 매일 5 분. 빠지면 그 날은 *측정 단절* 로 표기 (C4 인정 불가 day).

### §3.2 Day 1-2 — C2 진척 (plugin path)

**목표**: 9-phase 중 ≥3 phase 사용. cycle-3 note 의 §B.1 path.

**진행**:

```bash
# Claude Code 세션 안에서 (~30-60 분)
/buddy:concretize-idea      # phase 1
  ↳ 본인의 진짜 외부 SaaS 아이디어 1 개 (C1 의 후보) — forcing question 6 개 답변
/buddy:define-features      # phase 2
  ↳ concretize 결과로 actor / use-case / feature backlog
/buddy:autoplan             # phase 4 (plan-build) 의 sub-tool
```

**자가 점검** (Day 끝):

- [ ] `/buddy:concretize-idea` 가 *demand reality* 분리해 줬는가? (Q1 "내일 사라지면 진짜로 화내는 사람" 가 작동했는가)
- [ ] AskUserQuestion 폭주 없었는가? (한 turn 에 2~4 개로 묶였는가)
- [ ] 한국어 friend-tone 깨짐 없었는가?
- [ ] *마찰 finding* 발견했으면 즉시 cycle-3 note §C 에 `B1-<n>` 으로 추가 (4-field format)

**카운터 갱신**:
- C2 += 3 phase (concretize / define-features / plan-build 라면)
- C1 = 1 (이 아이디어가 외부 SaaS 면 — 이미 만족)

### §3.3 Day 3 — C3 진척 (cli buddy agent path)

**목표**: agent schedule 실행 1 회 완주. cycle-3 note §B.2.

**진행**:

```bash
DB=/tmp/dogfood.db
# (a) on-demand 1 회 — spawn 동작 확인 먼저
buddy agent --db "$DB" run hello-world
# 실 claude CLI subprocess spawn → result JSON 출력 확인

# (b) scheduler 모드 — 별 shell 에서
buddy agent --db "$DB" scheduler &
# cron tick (default 60s) 안에 hourly-cleanup 또는 daily-report 가 fire 하길 대기
# 한 cron tick 완주 = result JSON 산출 + buddy agent log 에 chain steps 기록
```

**자가 점검**:

- [ ] subprocess spawn 의 stdout/stderr 가 `buddy agent log <run-id>` 에 line-by-line 보이는가?
- [ ] auto_cascade 가 depth 제한 (max_depth: 3) 준수했는가?
- [ ] scheduler 가 *한 cron tick 안에서* run 완주했는가?

**카운터 갱신**:
- C3 += 1 (scheduler 가 한 tick 완주 시)

### §3.4 Day 4-7 — C4 누적 측정 (hook monitor path)

**목표**: hook stats 누락 0. cycle-3 note §B.3. *passive*.

**진행**: 매일 §3.1 의 daily 5 분만 반복. 별도 작업 X — 평소 Claude Code 사용 자체가 evidence.

**자가 점검** (Day 7 시점):

- [ ] `buddy stats --window 7d` 의 `expected_count == actual_count` 인가?
- [ ] outbox 적체 < 100 인가? (`buddy doctor` 의 outbox row)
- [ ] daemon crash 흔적 없는가? (`buddy daemon status` 의 restart count)

**카운터 갱신**:
- C4 = 0 누락 (또는 측정 break 시 그 day 제외하고 7 일 연속 확보 시점까지 연장)

### §3.5 Day 8+ (optional, W7 surface 검증)

§3.2 ~ §3.4 로 4 조건 모두 충족하면 **B-2 closed 진입 가능**. 다만 cycle-3 note 의 §B.4 (W7 surface 5 add-on) 가 *primary 끝나고 시간 남으면* 적극 활용 추천 — production-proven 의 폭이 넓어짐.

**add-on dependency DAG**: `session → usage / knowledge → advise → notify`. 순서 따라 진행 (각 5-10 분).

이미 cycle-3 note §A.1 에서 *5/5 surface dependency DAG 완주* 가 mechanical 검증된 상태 — 실 production use 에서 friction 만 확인하면 충분.

---

## §4. 자가 점검 매트릭스 (cycle 중 언제든)

```bash
# 한 줄로 4 카운터 확인
echo "C1 (외부 SaaS): __ 건"
echo "C2 (9-phase): $(grep -oE '/buddy:[a-z-]+' ~/.claude/projects/*/*.jsonl 2>/dev/null | sort -u | wc -l) phase"
echo "C3 (agent schedule): $(buddy agent --db /tmp/dogfood.db list 2>/dev/null | grep -c 'completed')"
echo "C4 (hook 누락): $(buddy doctor 2>&1 | grep -i 'outbox' | head -1)"
```

> *C2 의 grep 은 휴리스틱* — `/buddy:` slash 호출 흔적을 jsonl 에서 찾음. 정확하지 않으면 본인이 *실제로 호출한 unique slash* 를 손으로 세는 게 더 빠름.

### §4.1 진척 상태 4 가지

| 상태 | 의미 | 다음 액션 |
|------|------|----------|
| **0/4** | playbook 시작 전 | §2 사전 준비부터 |
| **1-2/4** | 진행 중 | 아직 안 한 condition 의 Day 진입 |
| **3/4** | 거의 다 | 남은 1 개 condition 만 집중 |
| **4/4** | B-2 충족 | §5 close-out 진입 |

---

## §5. Cycle close-out

4 조건 모두 충족 시:

1. cycle-3 note 의 **§E "Cycle close"** 표를 채움:
   - Cycle 종료 = 충족 시점 (YYYY-MM-DD)
   - Surface 별 사용 시간 = plugin __ h / cli-buddy __ h / hook-monitor __ days
   - Findings 분류 = §C 의 B*-* finding 총합 (blocker / high / medium / low 카운트)
   - Cycle 안에 fix 된 것 = within-cycle close 한 finding 수
   - Deferred = §D 로 보낸 finding 수
   - ADR escalated = §C 안에서 ADR 로 승격된 finding 수
2. **BACKLOG.md** 갱신:
   - §1 의 "8/9 closed" → "9/9 closed"
   - §4 Wave 1 의 W1-1 / W1-2 / W1-3 → ✅ Done
   - §2.A 의 B-2 entry condition → ✅ Done
3. **v1.0.0 release** 진입 (별 세션 권장):
   - `make set-version VERSION=1.0.0`
   - CHANGELOG 작성 (8/9 → 9/9 close-out 강조)
   - release tag + binary publish
4. **W4-5b unblocked**: post-dogfood trigger 가 비로소 발화. scrollback design ADR 진입 가능.

---

## §6. Common failure 대응

### §6.1 daemon 이 자꾸 죽는다

```bash
buddy daemon status               # restart count 확인
tail -100 ~/.buddy/daemon.log     # crash 직전 메시지
```

→ crash 직전 메시지를 cycle-3 note §C 에 `B3-<n>` finding 으로 추가. severity = high (C4 측정 단절).

### §6.2 claude CLI 가 PATH 에 없다

`buddy agent run` 이 즉시 "claude not found" 류 에러 → cycle-3 note §C 에 `B2-<n>` finding. 우회: `BUDDY_CLAUDE_BIN=/full/path/to/claude buddy agent run hello-world` 시도 후 결과 기록.

### §6.3 외부 SaaS 1건 정의가 모호하다

C1 의 "외부 SaaS" = *본인이 평소 진짜 작업* 하는 프로젝트. dogfood 용 toy project 는 인정 X. *이미 진행 중* 인 프로젝트의 한 feature 도 가능. concretize-idea 의 출력물이 실제로 *그 프로젝트의 backlog* 에 들어가야 인정.

→ 모호하면: cycle-3 note §C 에 `BA-<n>` 으로 "C1 정의 모호 — 본인 기준 ___ 으로 해석함" 명시 후 진행. 사후 ADR 로 정의 명확화 가능.

### §6.4 9-phase 중 3 phase 가 안 채워진다

대부분 build-feature / verify-quality 가 *실제로 작성된 코드를 reflect* 해야 의미 — concretize / define-features 만으로 끝나면 phase 1-2 만. *최소 3 phase 충족* 위해서는 concretize → define-features → design-system (또는 plan-build) 까지 진행 필요. *실 코드 작성* 까지 갈 필요는 없음 (design / plan 산출물로 충분).

### §6.5 hook 누락이 0 이 안 된다

```bash
buddy events --limit 1000 | grep -i 'failed\|error' | head
```

→ 누락 패턴 발견 시 cycle-3 note §C 에 `B3-<n>` finding. severity 는 누락 빈도로:
- 1 ~ 5 누락 / 7 days = medium
- 6+ 누락 / 7 days = high
- daemon 자체가 빈 경우 = blocker

---

## §7. Reference

- **General guide**: [`dogfood-guide.md`](./dogfood-guide.md) — 어느 cycle 에나 적용되는 절차.
- **Feedback template**: [`dogfood-feedback-template.md`](./dogfood-feedback-template.md) — finding 4-field format.
- **Active cycle note**: [`notes/2026-05-20-dogfood-result-cycle-3.md`](./notes/2026-05-20-dogfood-result-cycle-3.md) — 모든 finding 의 sink.
- **BACKLOG W1-1**: [`BACKLOG.md`](./BACKLOG.md#wave-1) — B-2 조건 원본 정의.
- **v1.0.0 framework**: ADR-010 — 9-조건 framework (B-1~B-4 + C-1~C-5).
- **이전 cycle**: [`notes/2026-05-19-dogfood-result-cycle-2.md`](./notes/2026-05-19-dogfood-result-cycle-2.md) — superseded, baseline staled lesson.
- **첫 cycle**: [`notes/2026-05-10-dogfood-result-cycle-1.md`](./notes/2026-05-10-dogfood-result-cycle-1.md) — B1~B5 fix cycle 의 reference.
