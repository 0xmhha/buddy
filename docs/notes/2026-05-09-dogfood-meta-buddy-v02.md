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
| §2 define-features | ⏸ | (다음 세션) | — |
| §3 design-system | 🟡 진입 | plan §3.1 (cascade chain 식별, SaaS pattern 미적용 판단) | (다음 세션 진행) |
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

---

## 3. 다음 세션 entry point

1. 본 ledger §1 표에서 ⏳ 항목 중 첫 번째 (§1 concretize-idea) PROCEDURE.md를 읽어 v0.2 PRD 섹션 작성.
2. 진행하면서 마찰 발견 시 §2 ledger 에 F-N 으로 추가.
3. §3 design-system 끝에서 1차 plan commit.
4. Cycle 완주 후 A-4.3 진행 — A-2 잔여 29 skill 재정렬.

---

## 4. 검증 evidence (A-4.1)

- `claude plugin list` (N-1 closure §4.2): `buddy 1.0.8 enabled`
- `~/.claude/plugins/marketplaces/buddy/plugin/commands/*.md` = 57 files all with `disable-model-invocation: true`
- 본 cycle 진행 자체가 plugin install 후 router skill discovery 동작의 indirect evidence (system-reminder 의 available-skills 에 `buddy:router` 표시됨)
