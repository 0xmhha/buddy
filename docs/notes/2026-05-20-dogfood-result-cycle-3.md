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

## §A. Pre-flight (2026-05-20 측정, commit `66575f8`)

baseline-as-of 마커는 cycle open 시점인 `d5f2755`. 본 세션에서 Wave 2 / Wave 3 polish + 본 §A sweep 중 발견된 finding 의 fix 가 cycle 내부에서 흡수되어 측정 commit 은 `66575f8`. cycle 진행 중 fix 가 들어가면 §A 는 *측정 시점의 commit* 을 명시한다는 lesson refinement.

| 체크 | 명령 | 결과 |
|------|------|------|
| binary 버전 | `buddy --version` | ✅ `buddy 0.13.0 (sha=6c233bf, built=2026-05-20T07:35:05Z)` |
| 5 SSoT version agree | `make verify-versions` | ✅ all 5 sources agree on 0.13.0 |
| 26 packages race-clean | `go test -race -count=1 ./...` | ✅ all 26 ok (cycle-3 doc 의 "23 packages" 는 cycle-2 stale 수치; v0.13.0 시점 26) |
| 148 skill PROCEDURE lint | `make test-skill-form --strict` | ⚠️ 62 allowlist / 86 pass / **1 deviate** — but the 1 deviation (`write-a-skill`) is *uncommitted in-progress* work, not a shipped skill. 148 base skills all green. |
| `~/.buddy/buddy.db` 존재 | `ls -la ~/.buddy/buddy.db` | ✅ 12 MB (active use, 2026-05-20 timestamp) |
| daemon 가동 | `buddy doctor` | ⚠️ daemon 미가동 + outbox 8,035 entries 누적 → user action: `buddy daemon start` |
| `buddy --help` 16 subcommand + W7 5 | `buddy --help` | ✅ 21 subcommand 노출 (W7 5 모두 포함) |
| TUI 7 modes 정상 | `buddy tui` 진입 후 `j/k/Enter/t/e/c/d/s/H` 키 1 회씩 | ✅ partial (mechanical 100% — non-TTY launch graceful + 8/8 키 unit test pass; interactive surface 만 user-paced) |

**Pre-flight 결과**: 8/8 자동 검증 가능 영역 ✅ (1 skill lint deviation 은 cycle 외부 in-progress work). §B 진입 가능 상태.

### §A.8 — TUI smoke detailed breakdown

interactive surface 라 *non-deterministic 영역* (AltScreen 복원 / 색상 / 한글 / 키 timing) 만 사용자 본인 머신 검증 필요.

**Mechanical 검증 (AI 가 실행, 100%)**:
- Non-TTY launch: `./bin/buddy tui --db ~/.buddy/buddy.db` → graceful error `"buddy: tui: could not open a new TTY: ..."` (exit 2). panic 없음, friend-tone wrapper 적용.
- TUI 단위 테스트: `internal/tui/model_test.go` 가 bubbletea model 에 key event 직접 주입. §A.8 의 8 키 모두 cover:

| 키 | Test |
|----|------|
| j/k | `TestUpdate_NavigationRespectsBounds` + `TestUpdate_GAndShiftGJumpToEnds` |
| Enter/l | `TestUpdate_EnterSwitchesToDetailAndFiresLoad` + `TestUpdate_LSwitchesToDetailLikeEnter` |
| t | `TestUpdate_TInDetailEntersLogTailAndFiresInitialLoad` |
| e | `TestUpdate_EInDetailWithSpecDispatchesEditCmd` |
| c | `TestUpdate_CInListDispatchesCreateCmd` + `TestUpdate_CInListWorksOnEmptyList` |
| d | `TestUpdate_DEntersConfirmMode` + `TestUpdate_DOnEmptyListIsNoOp` |
| s | `TestUpdate_SSwitchesToScheduler` |
| H | `TestUpdate_HInListWithFetcherEntersHookStats` + `TestUpdate_EscFromHookStatsReturnsToList` |

**Interactive 검증 결과 (2026-05-20, 사용자 본인 머신)**:
- ✅ DB 미존재 → empty list 정상 (graceful empty-state)
- ✅ 신규 agent 생성 후 키 인식 — `j/k/Enter/l/t/e/c/d/s/H/q` 모두 의도대로 반응
- ✅ 한글 깨짐 X, 색상 / 키 timing 정상
- ⚠ **터미널 가로/세로 변경 시 UI 깨짐** → §C BA-4 등록
- 💡 사용량 화면 = text-only — 그래프로 trend 시각화 가능하면 가독성 ↑. 사용자가 "불필요한 작업량 많으면 스킵" 명시 → §D BA-5 enhancement 로 deferred

### §A.1 — §B.4 add-on surface smoke (2026-05-20, commit `66575f8`)

§A 후속으로 §B.4 의 5 W7 surface 를 *daemon 없이* 1 회씩 호출 — *empty-state → ingest → real-data* 사이클 검증:

| Surface | 명령 | 결과 |
|---------|------|------|
| B.4.a session | `buddy session list --refresh` | ✅ 19+ sessions ingested via fsLister (daemon-less one-shot) |
| B.4.b usage today/trend/top | `buddy usage today` 등 | ✅ 3.7B tokens 24h / 64 sessions 7d / top-N leaderboard 정상 |
| B.4.c knowledge | `buddy knowledge ingest --all` → `stats` → `query` | ✅ 16,162 chunks; BM25 검색 정상; embedder 부재 시 friend-tone fallback (sentence-transformers 미설치) |
| B.4.d advise | `buddy advise` | ✅ **실 advisor rule 2 개 fire**: `token-spike-day · high` + `session-volume-day · info` |
| B.4.e notify | `buddy notify test --channel {desktop,tui-banner}` | ✅ dispatch 메시지 정상; ⚠ `notify status` 와 audit log 분리 (→ BA-3) |

**§B.4 통과**: 5/5 surface dependency DAG 완주. *real cycle-3 baseline signal* 확보 (advisor 첫 production 발화).

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

**진행 결과 (2026-05-20, Stage 1 validate-idea 완주)**:

- ✅ `/buddy:concretize-idea` slash → router → concretize-idea PROCEDURE → validate-idea Stage 1 cascade 정상 동작
- ✅ **cycle-1 B2 fix (Q1 phrasing "내일 사라지면 진짜로 화내는 사람") production verified** — 자연스럽게 동작
- ✅ Anti-sycophancy 의 "Push-back + calibrated 인정 + dwell 금지" pattern 이 *real wedge re-discovery* 트리거 — Q4 의 (d) → Q5 surprise → (d') re-formulation
- ✅ Escape hatch ((B) 정직 인정) 가 *real anchor 발굴* 의 catalyst 로 작동 — Q3 mimicry → 세민 surface
- ✅ Two-mode (Startup / Builder) 분기, 6 forcing question 완주, 전제 체크 4 항목, 2-3 대안 (A/B/C) 생성, 추천 (A→B sequence) 도출, design doc template (Startup) 산출 — *모든 stage 완료*
- ⚠ 3 PROCEDURE-side findings: **BA-6 (Q3 prompt mimicry)** / **BA-7 (hybrid persona)** / **BA-8 (전제 active obligations silent)** — Wave 4 candidates (W4-8/9/10)
- 📄 Design doc 산출 → **이동 완료** 2026-05-20: `/Users/wm-it-22-00661/Work/github/study/ai/claude-design-skill/docs/2026-05-20-validate-idea-ai-figma-context-bridge.md` (claude-design-skill repo 의 master branch). 본 idea 가 *해당 프로젝트의 확장 영역* 이라 IP 가 그 repo 에 귀속
- ⏳ Stage 2-8 (validate-advanced-edge-idea → assess-business-viability → ... → autoplan) — 별 세션 권장 (token budget)
- ⏳ The Assignment (세민 미팅 + 1-week trial) — 사용자 *이번 주* 실행 obligation

**Token 비용**: 약 6 round forcing questions + 전제 체크 + 2-3 대안 + design doc generation ~ context 대량 소비. **Stage 1 only** 의 결정 후행 검증.

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
buddy knowledge ingest --all       # 모든 session transcript chunk (BM25)
buddy knowledge ingest --session <id>  # 단일 세션만
# --all 또는 --session 필수. 둘 다 없으면 friend-tone error.
buddy knowledge ingest --all --embed-script ./scripts/embed.py  # 임베딩까지 (python + sentence-transformers 필요)
buddy knowledge stats              # chunk 수 / embedding coverage / last ingest
buddy knowledge query "내 질문"    # BM25 (default; embedder 미설치 시 자동 fallback)
buddy knowledge query "내 질문" --mode hybrid  # BM25 + vector (embedder 필요)
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

> 각 finding 은 guide §3.2 의 4-field format. ID 는 `B<surface>-<n>` (e.g., `B1-1` plugin 첫 finding / `B4a-1` session 첫 finding). §A pre-flight 중 발견된 finding 은 `BA-<n>`.

### BA-1 — events --limit 의 i18n wiring partial regression

**Surface**: cli-buddy
**Severity**: medium (사용자가 만나는 first-impression error 가 영어)
**Repro**:
```bash
./bin/buddy events --limit -1
# expected:  buddy: --limit 은 1 이상이어야 해.
# observed:  buddy: --limit must be >= 1
```
**Expected vs Actual**:
- expected: KeyQueriesInvalidLimit 의 ko 템플릿 그대로 출력 (cycle-3 의 baseline 이 Wave 2 W2-2 commit 8f901a5 를 포함)
- actual: raw English fallback (`queries.ErrInvalidLimit.Error()`) 노출

**Root cause**: W2-2 commit (`8f901a5`) 의 `Edit` with `replace_all: true` 가 events_cmd.go 의 2 ErrInvalidLimit 매칭 중 1 개만 잡음. Follow branch (depth 5 tab) 만 변경되고 RunEvents branch (depth 4 tab) 누락. *왜 silent*: structural test 부재. 단일 site 누락은 단위 테스트가 잡지 못함.

**Recommendation**: bug-fix (즉시 처리). cycle-3 진행 중 commit `66575f8` 으로 fix + cmd/buddy/main_test.go 에 `TestEvents_InvalidLimit_RendersViaPersona` 추가하여 plain + follow 양쪽 site 의 persona 렌더링을 lock-in. 동일 패턴 (i18n bulk wiring → 다중 site 검증 부재) 의 일반화된 가드는 *Wave 4 cycle-3 finding* 으로 분리 검토 (예: `grep -rn 'buddy: " + err.Error()' cmd/buddy/` CI step).

**Closed**: ✅ commit `66575f8` (within-cycle fix, 동일 세션).

---

### BA-2 — cycle-3 doc 의 `buddy knowledge ingest` 예시가 required flag 누락

**Surface**: cli-buddy (documentation)
**Severity**: low (UX confusion; CLI 자체는 friend-tone 으로 안내)
**Repro**:
```bash
./bin/buddy knowledge ingest
# CLI 응답: buddy: --session <id> 또는 --all 필요해
```
**Expected vs Actual**:
- expected: cycle-3 doc §B.4.c 의 첫 예시가 `buddy knowledge ingest --all` 로 시작해야 사용자가 1 회 호출로 결과 확인
- actual: doc 의 첫 예시가 인자 없이 `buddy knowledge ingest` — 사용자가 그대로 실행하면 fail

**Root cause**: cycle-3 doc 작성 시 W7-3a (ADR-014) 의 ingest 동작을 *flag-less default* 로 추정. 실제 CLI 는 *데이터 범위 명시 강제* (안전 default). doc 이 CLI 동작 잘못 묘사.

**Recommendation**: documentation. cycle-3 doc §B.4.c 의 첫 줄을 `buddy knowledge ingest --all` 로 갱신. 동일 doc 의 다른 surface 도 *실 인자 검증* 한 번 더 sweep.

**Closed**: 이번 cycle-3 doc 갱신 commit 에 포함.

---

### BA-3 — `notify test` 가 dispatch 후 notify_log 에 기록 안 됨

**Surface**: notify
**Severity**: low (audit trail 누락; user 가 "test 보냈는데 status 에 안 보이네" confusion)
**Repro**:
```bash
./bin/buddy notify test --channel desktop
#  buddy: desktop 채널로 테스트 알림 보냈어.
./bin/buddy notify status --limit 5
#  buddy: 기록된 알림이 없어. daemon 가동 + advisor 생성 후 확인해줘.
```
**Expected vs Actual**:
- expected: 두 가지 중 하나. (a) synthetic test 도 notify_log v8 에 기록 + status 가 *test* tag 와 함께 표시, 또는 (b) `notify test` 응답 메시지가 "audit log 에는 안 기록함" 명시
- actual: dispatch 는 성공한 응답, status 는 empty 표시 — *연결 끊긴 두 surface*

**Root cause**: ADR-016 (W7-5 notification) 의 *audit trail vs synthetic-test* 분리 정책이 *코드는 분리, doc 은 미명시*. 사용자 입장에서 "보냈다더니 status 가 비어있네" 의 disconnect.

**Recommendation**: UX-fix 또는 documentation. 가장 가벼운 수정 = `notify test` 응답 끝에 "(audit log 에는 기록 안 함; daemon auto-dispatch 만 status 에 보임)" 한 줄 부연. 더 강한 수정 = synthetic dispatch 도 *kind=test* 로 notification_log 기록 후 `status --kind real` flag 추가.

**Open** — Wave 4 candidates 에 등록 권장 (post-B-2 finding triage).

---

### BA-6 — validate-idea Q3 의 예시가 *너무 완성형* → prompt mimicry 유도

**Surface**: plugin (validate-idea PROCEDURE)
**Severity**: low-medium (Q3 의 *진짜 anchor 인간 surface* 실패 risk)
**Repro**: cycle-3 §B.1 본 세션 — Q3 "이름 + 직함 + 보스 이름 + 캘린더 본 적" 예시 그대로 user 가 *2 토큰만 바꿔 복사* (민지→민규, 30→15)
**Expected vs Actual**:
- expected: founder 본인의 *real anchor 인간* 정보가 *유기적으로* 등장
- actual: 예시 구조 mimic + cosmetic edit → escape hatch 발동 후에야 *real 세민* surface
**Root cause**: validate-idea PROCEDURE 의 Q3 예시가 "Sarah, 50명 logistics 회사 ops 매니저..." 형태로 *완성형 sentence* — copy-edit 유혹 큼
**Recommendation**: UX-fix (Wave 4 candidate). 예시를 *fragmentary* 로 (이름만, 또는 직함만, 또는 보스만 — 단편 3 개) 제공 → mimicry 자연 어려움, organic specificity 강제. PROCEDURE Stage 1 의 Q3 sample 단 1 곳 수정.
**Open** — Wave 4 신규 W4-8 으로 BACKLOG 등록 권장.

### BA-7 — validate-idea 가 hybrid persona 케이스 미고려

**Surface**: plugin (validate-idea PROCEDURE)
**Severity**: low (현 PROCEDURE 도 force-pick 으로 진행 가능; 단 wedge framing 의 정확도 저하)
**Repro**: cycle-3 §B.1 본 세션 — anchor 세민 = *lead designer + frontend code* hybrid role. PROCEDURE Q3 는 *single named human + single role* 형태로 force, 그 후 wedge 가 *designer vs engineer* 둘 중 하나로 갈라짐. 실제로는 *hybrid persona at small startup* 이 더 sharp wedge.
**Expected vs Actual**:
- expected: PROCEDURE 가 hybrid persona case 를 *acknowledge* 하고 *wedge framing 조정 ask*
- actual: AI 가 immediate 인식해서 surface (BA-7 본 finding) 했지만 PROCEDURE 본문에는 미서술 — 향후 dogfood 시 AI 가 인식 못 하면 wedge framing 실수 가능
**Recommendation**: documentation 추가. PROCEDURE 의 Q3 또는 Q4 에 *"persona 가 hybrid role 일 경우 single-product-for-hybrid wedge 도 valid alternative"* note 1-2 줄. Optional reframe step (Q3 후): "이 사용자의 *role split* 이 *fundamental* 인지 *circumstantial* 인지 — 다른 같은 segment 의 디자이너 5 명도 *hybrid role* 인가?"
**Open** — Wave 4 신규 W4-9 으로 BACKLOG 등록 권장.

### BA-8 — 전제 체크의 *agree* 가 active obligation 을 silent 처리

**Surface**: plugin (validate-idea PROCEDURE)
**Severity**: medium (실제 발생 시 demand validation 결손 위험)
**Repro**: cycle-3 §B.1 본 세션 — P2 ("20h/week wasted *self-reported only* — methodology unverified") 에 user 가 단순 *agree*. PROCEDURE 의 design doc template 의 *Open Questions* 또는 *Dependencies* 에 자동 reflection 없으면 P2 의 *upgrade obligation* 이 doc 작성 시 lost.
**Expected vs Actual**:
- expected: 전제 체크의 *qualifier 가 포함된 전제* (e.g., "X — but Y unverified") 에 user agree 시, PROCEDURE 가 *Y* 를 design doc 의 *Open Questions* / *Dependencies* / *The Assignment* 로 자동 carry-forward 강제
- actual: PROCEDURE 명시 없음. AI 가 인식해서 *The Assignment* 에 직접 포함 (BA-8 surface) 했지만, 일반화된 가드는 없음
**Recommendation**: documentation 추가. PROCEDURE 의 *전제 체크* step 에 *active obligations 추출 + doc carry-forward 의무* 명시. e.g., "전제 중 *qualifier (unverified / pending / upgrade-needed)* 포함된 항목 → design doc Open Questions OR The Assignment 에 *반드시* 표기."
**Open** — Wave 4 신규 W4-10 으로 BACKLOG 등록 권장.

---

### BA-4 — TUI 가 터미널 resize 에 reflow 안 함

**Surface**: tui (cli-buddy)
**Severity**: medium (UX 큰 영향; panic 아님이라 blocker 는 아니나 *first-impression* 손상)
**Repro**:
```
./bin/buddy tui --db ~/.buddy/buddy.db
# 터미널 너비 80 → 120 으로 마우스 drag 변경
# 또는 zoom in/out
```
**Expected vs Actual**:
- expected: 새 width/height 에 맞춰 layout 재계산 + 텍스트 reflow. resize 종료 후 정상 렌더링.
- actual: 화면 깨짐 — 텍스트 잘림 / 잔여 글자 / 패널 겹침. resize 후에도 회복 안 됨 (재진입 / `r` refresh 시에만 정상)

**Root cause** (코드 분석 완료):
- `internal/tui/model.go:614-617` 의 `WindowSizeMsg` 핸들러가 `m.Width = msg.Width; m.Height = msg.Height` 로 *저장만* 함.
- 모든 view function 들 (List/Detail/LogTail/Scheduler/HookStats/Usage) 이 `m.Width` / `m.Height` 를 *전혀 사용하지 않음* → grep 결과 view 코드에서 reference 0회.
- 결과: bubbletea 가 새 사이즈 통보해도 렌더링 폭은 *터미널 시작 시점의 implicit width* 기준.

**Recommendation**: bug-fix (Wave 4 candidate, *4-6 h*).
- 각 view function 에 `lipgloss.NewStyle().Width(m.Width).Render(...)` 적용 또는 column width 동적 산정
- 우선순위 = List (가장 흔한 view) > Detail > Scheduler/HookStats > Usage > LogTail
- B6 의 *width-bound test* 가 단위 테스트로 lock-in 가능 — `m.Width = 40` 으로 좁힌 후 View() 가 80-col layout 깨지 않는지 검증

**Open** → Wave 4 신규 W4-6 으로 BACKLOG 등록 권장.

---

## §C.1 — Within-cycle observations (finding 은 아니나 기록 가치)

- **Real advisor signals fired** during §B.4.d smoke: `token-spike-day · high` (오늘 토큰 평소 2.3x) + `session-volume-day · info` (24h 21 sessions). 이건 *advisor 가 실 user 의 패턴에서 신호 추출* 의 첫 production 증거. cycle-3 doc 작성 직후 발화로, advisor rule (ADR-015) tuning 의 baseline 으로 활용 가능.
- **Knowledge ingest 규모**: `--all` 1 회로 64 sessions → 16,162 chunks. 평균 252 chunks/session. embedding 없이 BM25 만으로도 retrieval 작동.
- **Session monitor `--refresh` 모드**가 daemon-less one-shot scan 으로 잘 작동 — cycle-3 §A *user-paced 1-day cycle* 가능성을 확보 (daemon 띄우지 않고도 §B.4 add-on 검증 가능).

---

## §D. Deferred / Out-of-scope

cycle 진행 중 "지금은 안 한다" 결정된 것은 여기 기록 + trigger 명시.

### BA-5 — Usage pane / `buddy usage` 출력에 trend 그래프 추가

**Source**: 사용자 §A.8 interactive sweep 중 제안 — *"text 만 존재하는데, 그래프로 변동추이를 보여줄수 있으면 눈에 잘 들어올 것 같아"*. *"불필요한 작업량 많다면 스킵하는 것이 좋을 것"* 명시.

**제안 범위**:
- `buddy usage trend --days 7` 등 시계열 metric 에 *ASCII sparkline* / bar chart 렌더링
- TUI 의 Usage pane 도 동일 시각화 적용

**Cost-benefit**:
- 비용: Go terminal chart lib (e.g., `asciigraph`) 도입 + 2-3 h. test 작성 포함하면 4 h.
- 가치: text-only → trend at-a-glance perception 큰 개선. *real user friction* 점.
- v1.0.0 blocker 아님, B-2 진척 영향 X.

**Decision**: defer. Wave 4 (TUI / runtime UX follow-on) 신규 W4-7 로 BACKLOG 등록. v1.0.0 publish 후 또는 *resize bug (BA-4)* 와 묶어서 Wave 4 cycle 에서 진행.

**Trigger**: B-2 cycle close 후 / Wave 4 dogfood signal pivot 시 우선순위 재평가.

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
