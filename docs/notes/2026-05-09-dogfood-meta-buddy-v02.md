# Meta-dogfood — buddy v0.2 Control Plane via 9-phase orchestrator

> **Date:** 2026-05-09
> **Plugin version:** v1.0.8 (`claude plugin list` 검증 완료, N-1 closure handoff §4.2)
> **Target:** buddy 자체 v0.2 Control Plane (multi-session dashboard) plan 도출
> **Method:** 9-phase orchestrator를 manual walkthrough — 각 phase orchestrator의 PROCEDURE.md 를 읽고 v0.2 입력에 적용. router 자동 dispatch 대신 사용자 cycle 외부 검증.
> **Output target plan:** [`../superpowers/plans/2026-05-09-v02-control-plane-plan.md`](../superpowers/plans/2026-05-09-v02-control-plane-plan.md)
> **A-4 항목 매핑:**
> - A-4.1 ✅ Plugin install (이미 v1.0.8 enabled — N-1 closure §4.2 evidence)
> - A-4.2 🟡 진행 중 — 단일 cycle walkthrough
> - A-4.3 ⏳ cycle 종료 후 마찰 분석 → A-2 잔여 29 skill 재정렬
> - A-4.4 ⏳ cycle 중 발견된 마찰 fix (별도 commit)

---

## 1. Cycle 진행 로그

| Phase | 상태 | 산출물 | 마찰 기록 |
|-------|------|--------|----------|
| status | ✅ | "현재 phase = §1 진입 직전" 판단 | F-1 |
| §1 concretize-idea | ✅ | plan §1.1~§1.7 (condensed, autoplan skip) | F-2, F-3, F-4, F-5 |
| §2 define-features | ✅ | plan §2.1~§2.4 (5 actor / 5 use case / 6 feature / DAG depth 매핑) | F-6 |
| §3 design-system | ✅ | plan §3.1 (cascade 5 stage 적용) + §3.2 (D-1 = TUI 채택, ADR-002 proposed) + §3.3 (ai-m 패턴 차용 분류) + §3.4 (trigger-driven backlog 4건) + §3.5 (ADR draft) | F-7, F-8 |
| §4 plan-build | ✅ | plan §4.1~§4.7 (3 implementation tracks / 24 atomic tasks / DAG critical path 7.75h / 8 batch schedule / acceptance test plan / timeline p50=5d p90=8d / autoplan skip) | F-9 |
| §5 build-feature | ⏸ | (다음 라운드 — batch B1 진입) | — |
| §6 verify-quality | ⏸ | (§4.5 acceptance test plan 이 입력, SaaS audit 부분 적용) | — |
| §7 ship-release | ⏸ | (v0.1 release.yml 재사용 + sessions migration + binary size check) | — |
| §4 plan-build | ⏸ | (다음 세션) | — |
| §5 build-feature | ⏸ | (다음 세션 이후) | — |
| §6 verify-quality | ⏸ | — | — |
| §7 ship-release | ⏸ | — | — |
| §8 iterate-product | n/a | (production traffic 의존) | — |
| §9 manage-lifecycle | n/a | (deprecation 시점) | — |

---

## 2. 마찰 ledger (dogfood findings)

### F-1: status orchestrator의 artifact 탐지가 informal PRD를 미인식

**관찰 (status PROCEDURE Step 1 적용 시):**
- `docs/prd.md` / `docs/PRD.md` / `docs/feature-spec/` / `docs/features.yaml` 모두 부재.
- 하지만 `docs/roadmap.md §4` 가 v0.2 informal PRD + task 6분해 (T1~T6) 를 보유.
- artifact 탐지표 기준으로는 "위 없음 + idea/concept만 언급 → §1 concretize-idea" 추론.
- 실제로는 informal PRD가 이미 있어 §3 design 직전 상태가 더 정확.

**확신도:** Mid (artifact 표가 entry path 가정에 묶임 — informal/distributed PRD 케이스 미반영).

**제안:**
- (a) PROCEDURE.md 의 탐지표를 "informal PRD = roadmap §X / decision doc / spec md" 까지 확장.
- (b) status에 사용자에게 추가 질문 hook ("PRD가 별도 파일이 아니라 roadmap 섹션이면 알려줘") 추가.
- (c) 이 cycle처럼 사용자가 manual override 가능하게 — 현재 PROCEDURE 본문에 "탐지 불가 시 사용자에게 묻는다" 라인은 있음, 하지만 *exists-but-non-canonical* 케이스는 묵시적으로 §1 추론.

**fix 우선순위:** Low — manual override로 우회 가능. 다만 다른 dogfood 사용자도 같은 마찰을 겪을 가능성이 있어 README/docs 차원의 "informal PRD 케이스" 가이드 한 줄 추가 권장.

### F-2: §1 PROCEDURE 의 Stage 1 / Stage 3 gate 가 *기존 PRD 재정리* 케이스에 과함

**관찰:**
- Stage 1 gate = idea 명확성 부족 시 사용자 confirm. Stage 3 gate = 사업성 치명결함 시 피벗 confirm.
- v0.2 같은 *이미 validated된 idea를 9-phase로 재정리*하는 케이스에는 두 gate 모두 자동 통과해야 자연스럽다.
- PROCEDURE 본문은 "Gate 없이 자동 진행하지 않는다" 라고 명시 — 신규 idea 가정.

**확신도:** Mid (대부분의 신규 idea 케이스에는 적절. 단, meta-dogfood / re-organization 케이스 미커버).

**제안:** PROCEDURE.md 에 "skip 조건" 명시 — informal PRD가 보유 (roadmap §X / decision doc 등) 라면 §1 을 condensed 모드로 적용 + gate 자동 통과.

**fix 우선순위:** Low.

### F-3: stage 4 (`analyze-competition-and-substitutes`), stage 6 (`map-customer-segments`) skill 미존재 — orchestrator fallback 부담

**관찰:**
- §1 PROCEDURE 의 stage 4, stage 6 은 skill 미작성 상태로 *orchestrator 가 직접 수행*하라고 본문에 명시.
- 실제 walkthrough 시 "직접 수행" 가이드는 본문 한 줄 (직군/매트릭스 양식) 정도 — competition / segmentation 이 깊은 분석을 요구하는 phase 임에도 fallback 가이드가 가벼움.
- 현재는 v0.2 가 OSS 단일 머신 도구라 competition/segmentation 가벼워서 큰 문제 아님. 하지만 *상용 SaaS 시나리오* 에서는 fallback 만으로 부족할 가능성 있음.

**확신도:** Mid.

**제안:** A-2.1 Cluster E (5 skill — `analyze-competition-and-substitutes`, `map-customer-segments` 포함) 우선순위는 dogfood 신호 따라 변동. 본 meta-dogfood 케이스는 fallback 으로 충분 — Cluster E 우선순위 낮음 유지 (Wave 6 위치 적절).

**fix 우선순위:** Low (현 우선순위 유지).

### F-4: 산출물 template 이 *condensed* 케이스 미반영

**관찰:**
- §1 PROCEDURE "산출물 형식" = `## 1단계 산출물 — {idea 이름}` 표준 template (Idea Validation Summary / Business Viability / PRD Draft / autoplan Review Summary 4 섹션).
- 본 cycle 에서는 *기존 informal PRD를 재정리* 하는 케이스라 v0.2 plan 의 §1 섹션을 1.1~1.8 로 풀어쓰되 PRD draft 는 별도 파일 미생성, autoplan review 도 skip.
- PROCEDURE 의 *required template* 와 *condensed 적용* 사이 가이드 부재.

**확신도:** Mid.

**제안:** PROCEDURE.md 산출물 형식에 "informal PRD 보유 시 condensed 적용 — autoplan review 는 §3 cascade 산출 후로 deferred 가능" 한 줄 추가. F-2 과 함께 묶어 single fix.

**fix 우선순위:** Low.

### F-5: §1 → §2 transition 의 *PRD draft → features.yaml* 변환 가이드 부재

**관찰:**
- §1 끝 = PRD draft. §2 시작 = features 목록. 둘 사이 변환을 누가 책임지는지 PROCEDURE 본문 미명시.
- §1 Stage 7 `define-product-spec` 산출물이 PRD 이고, §2 entry artifact 는 `docs/feature-spec/` or `docs/features.yaml` — 형식이 다름.
- Meta-dogfood 케이스에서는 informal PRD (roadmap.md §4) → §2 정식 features.yaml 변환을 어떻게 해야 할지 불명확. 다음 세션의 첫 마찰 후보.

**확신도:** Low (다음 세션 §2 진입 시 확인).

**제안:** §2 PROCEDURE.md 의 진입 조건 / preflight check 섹션 보강 — "PRD 가 별도 파일이 아닌 informal artifact 면 features 추출 절차" 가이드.

**fix 우선순위:** Mid (다음 세션 §2 entry 에서 결과적 신호 확인 후 결정).

### F-6: §2 PROCEDURE 가 *single-actor* 케이스에 9 stage 모두 강제

**관찰:**
- §2 PROCEDURE 의 stage 6~10 (query-feature-registry / score-feature-priority RICE/ICE/MoSCoW / map-feature-dependencies / split-work-into-features / triage-work-items) 은 다중 stakeholder + 백로그 운영 가정.
- v0.2 같은 single-user OSS 도구 backlog 에서는 RICE 같은 priority framework 가 over-engineering. MoSCoW 만으로 충분.
- query-feature-registry 는 buddy 자체에 registry 미존재라 자동 skip — PROCEDURE 가 "registry 부재 시 skip" 명시 안 함.

**확신도:** Mid.

**제안:** §2 PROCEDURE.md 에 단순 케이스 가이드 — single-actor / single-user / no-registry 시 stage 1-5 + stage 7 (MoSCoW only) 만 적용. F-2, F-4 와 묶어 "PROCEDURE 의 condensed/full 모드 분기" 단일 fix 후보.

**fix 우선순위:** Low (현 우회로 충분).

### F-7: 외부 reference (ai-m 같은 precedent) 차용 절차가 PROCEDURE 어디에도 명시 안 됨

**관찰:**
- §3 design-system 진행 시 사용자가 `0xmhha/ai-m` 을 reference 로 지정 — TUI vs web 결정의 직접 입력.
- §3 PROCEDURE 본문은 *internal* design 만 가정 (define-tech-stack, map-use-cases-to-infra 등 5 stage 모두 *결정* 자체에 집중, *외부 precedent 차용* 절차 없음).
- 결과: orchestrator (이번엔 manual) 가 reference repo 를 ad-hoc 로 탐색해 패턴 추출. 절차 무재현성.

**확신도:** High (이번 cycle 에서 직접 관찰 — ai-m 탐색이 §3.3 의 "차용 가능 ai-m 패턴" 표로 산출되기까지 정형 절차 없이 ad-hoc).

**제안 (택 1):**
- (a) §3 PROCEDURE 에 stage 0 신설: `survey-precedents` — 사용자에게 reference repo / paper / 제품 입력받아 design 입력으로 반영.
- (b) cross-cutting skill `survey-design-precedents` 신설 (Cluster B residual 4 와 별개) — 모든 design phase 에서 호출 가능.
- (c) 본 마찰만 기록하고 deferred — 차용은 사용자가 reference 를 명시할 때만 발생. orchestrator 가 자동 precedent 검색하면 노이즈 + IP 위험.

**fix 우선순위:** Mid — meta-dogfood 가 외부 reference 사용 케이스를 노출. 다른 dogfood 사용자도 자기 도메인의 precedent 를 자연스럽게 사용할 가능성 높음.

### F-8: §3 SaaS pattern 3 stage (event/auth/tenant) 가 *적용 안 됨* 명시 없이 silent skip

**관찰:**
- v1.0.8 에서 추가된 design-event-schema / design-auth-model / design-tenant-model 은 SaaS pattern 가정 (multi-tenant async event-driven backend).
- v0.2 같은 OSS 단일 머신 도구는 *세 stage 모두 N/A*.
- §3 design-system PROCEDURE 가 이 3 stage 를 cascade 의 *기본 chain* 으로 포함 — N/A 판단 절차 없음. orchestrator 가 manual 로 "OSS 단일 머신 → SaaS pattern 미적용" 결론.

**확신도:** Mid.

**제안:** §3 PROCEDURE 에 *applicability check* 한 줄 — "SaaS pattern stage 는 multi-tenant 또는 async event-driven 일 때만 적용. Decision 1 의 §1.4 business viability 의 segmentation 결과를 입력으로." F-7 과 묶어 single fix.

**fix 우선순위:** Low.

### F-9: §2 *actor* (user/system/3rd-party) 와 §4 *actor track* (frontend/backend/3rd-party = implementation domain) 의 단어 충돌

**관찰:**
- §2 PROCEDURE 의 actor 분류: user / system / 3rd-party / external-tool — *use case 합성*의 차원.
- §4 PROCEDURE 의 actor track: frontend / backend / 3rd-party — *implementation domain*의 차원.
- 같은 *actor* 단어가 두 phase 에서 다른 의미. cascade 진행 시 user 가 어떤 차원의 actor 인지 매번 추론 필요.
- v0.2 같은 single-actor (CLI user 1명) 도구 케이스에서는 §4 의 implementation domain track 으로 자연스럽게 분해 가능 (core / ui / i18n) — *§2 의 actor 차원과 무관*.

**확신도:** High (이번 cycle 진행 중 직접 마찰 — §4 진입 시 "v0.2 가 single-actor 인데 actor track 분해를 어떻게 하지?" 1차 질문 필요).

**제안:**
- (a) §4 PROCEDURE 의 *actor track* 을 *implementation track* 으로 rename. 단어 충돌 해소.
- (b) 또는 §2 PROCEDURE 의 actor 를 *role* 로 rename, §4 의 actor track 유지. 단, §2 stage 1 skill 명 `identify-actors` 도 영향.
- (c) 단어 그대로 유지하되 §4 PROCEDURE 첫 문단에 "여기서 actor 는 §2 의 actor (user/system) 와 다른 차원의 *implementation domain*" 명시.

**fix 우선순위:** Mid (변경 비용 낮음 — (c) 로는 PROCEDURE 한 문단 추가만, (a)/(b) 는 큰 명명 작업이라 trigger-driven 으로 deferred).

---

## 3. ai-m 차용 평가 (Wave 2 추가 산출물)

| 패턴 | 차용 결정 | 위치 (v0.2 plan) |
|------|---------|------------------|
| `bubbletea + lipgloss` TUI stack | ✅ 채택 (D-1 결정) | §3.1.1 + §3.2 |
| `internal/llm/backend/` interface + 2 impls | ✅ 부분 (`sessions.Lister` 1 impl, 2번째는 trigger-driven) | §3.3 |
| `internal/server/server.go` websocket 서버 | ❌ deferred (v1.0+ trigger-driven `BUDDY-D-WEB`) | §3.3, §3.4 |
| `apps/desktop/` Electron app | ❌ N/A (single-binary 정책 + non-goal) | §3.3, §3.4 |
| `STATUS.md` "Resolved + Trigger-driven backlog" 단일 문서 | 🟡 차용 가치 (A-1 SSoT polish 후보, deferred) | §3.3 |
| `--restart` policy / `--persistent` reattach | ❌ N/A v0.2 (v0.3+ task DAG 영역) | §3.3 |
| `fsnotify` file watcher | ✅ 채택 (UC-4 입력) | §3.1.1 |

**Net 결과**: ai-m 의 **stack 선택** + **interface abstraction 패턴** + **trigger-driven backlog 철학** 3개 차용. **세션 자체 관리** 영역은 책임 경계 (buddy = 세션 관찰, ai-m = 세션 운영) 따라 분리 유지.

---

## 4. 다음 세션 entry point

1. 본 ledger §1 표에서 ⏳ 항목 중 첫 번째 (§1 concretize-idea) PROCEDURE.md를 읽어 v0.2 PRD 섹션 작성.
2. 진행하면서 마찰 발견 시 §2 ledger 에 F-N 으로 추가.
3. §3 design-system 끝에서 1차 plan commit.
4. Cycle 완주 후 A-4.3 진행 — A-2 잔여 29 skill 재정렬.

---

## 5. 검증 evidence (A-4.1)

- `claude plugin list` (N-1 closure §4.2): `buddy 1.0.8 enabled`
- `~/.claude/plugins/marketplaces/buddy/plugin/commands/*.md` = 57 files all with `disable-model-invocation: true`
- 본 cycle 진행 자체가 plugin install 후 router skill discovery 동작의 indirect evidence (system-reminder 의 available-skills 에 `buddy:router` 표시됨)
