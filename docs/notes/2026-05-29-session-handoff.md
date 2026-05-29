# Session Handoff — 2026-05-29 (engineering 스킬 구조 재정비 + UX 라우터 도입)

> **다음 세션의 진입 지점**. 본 doc + `git log -16 --oneline` + 본 doc §10의 핵심 4 문서만 읽으면 *cross-machine 컨텍스트 0* 으로 이어받을 수 있다.
>
> **표기 약속**: `<REPO_ROOT>` = buddy 레포 working tree 루트.

---

## 0. 다른 머신에서 시작 시 (전제 조건)

```bash
# 1. 최신 동기화
cd <REPO_ROOT>
git fetch --all
git pull --ff-only origin main         # HEAD bbe5f09 이상 확인

# 2. 핵심 문서 읽기 (5-10분)
cat docs/notes/2026-05-29-session-handoff.md          # 본 doc
cat plugin/skills/router/references/engineering-phases.md   # 9-phase 정의 + Phase별 스킬 테이블
cat docs/plugin-skills-classification-matrix.md       # 152 스킬 분류
cat docs/plugin-skills-flow-graph.md                  # 전이 그래프 (Mermaid)

# 3. 무결성 검증
make test-routing 2>&1 | tail -3                      # 10/10 통과 필수
go test -race -count=1 ./internal/permissions/        # permissions 테스트
make build                                            # bin/buddy + bin/buddy-mcp 빌드

# 4. 최신 작업 trail
git log --oneline -16                                 # 본 세션 14 commits
```

---

## 1. 한 줄 요약

본 세션은 **engineering 스킬의 구조적 정합성을 정비하고 사용자 진입 UX를 개선**한 작업. 14 commits, plugin version `1.0.0 → 1.0.2`로 dogfood 검증 가능 상태.

핵심 산출물:
- **engineering-phases.md** — 9-phase artifact-based 정의 (cross-machine SSoT)
- **classification-matrix.md** — 152 스킬 phase별 분류 + standalone 등급
- **flow-graph.md** — 149 I/O Contract connections 시각화
- **신규 스킬 3개**: `assess-product-change`, `write-hld`, `start`
- **신규 Go 기능**: `buddy permissions check/inject` (+ `buddy doctor` 통합)
- **131 스킬에 Input/Output Contract 섹션 추가**

---

## 2. 현재 working tree 상태 (세션 종결 시점)

| 항목 | 값 |
|------|---|
| 작업 디렉토리 | `<REPO_ROOT>` (본 머신 예시: `/Users/wm-it-22-00661/Work/github/study/ai/buddy`) |
| Branch | `main` |
| Last commit | `bbe5f09` — `refactor(skill): remove run-uat/run-load-test/run-beta-program direct commands` |
| Remote sync | (사용자가 push 필요한 경우 직접 진행) |
| Uncommitted modified files | 0 |
| Untracked files | 0 |
| Plugin version | `1.0.2` |
| PROCEDURE.md count | 155 (router SKILL.md 제외) |
| Command count | 102 |

---

## 3. 본 세션 commits 14개 (chronological)

```
34a978d docs(skill): add artifact-based engineering phases and reclassify cross-cutting skills
5b77193 feat(skill): add Input/Output Contracts to 131 skills and skill tables to engineering-phases
addf307 feat(permissions): add buddy permissions check/inject for subagent tool access
064dfdc docs(skill): register assess-product-change in catalog, command, and wireup test
ac58cb6 fix(skill): relocate 11 misplaced I/O Contract sections to document header
231a537 feat(doctor): integrate subagent permissions check into buddy doctor
aa1edcc docs(skill): add flow graph visualization and update status skill for Phase 1 Mode B
9c1baae feat(skill): add write-hld and restructure concretize-idea stage order
38cf8ee fix(skill): correct concretize-idea Mode A entry condition and outputs
db6672a feat(skill): add /buddy:start intent-based entry router
6444886 refactor(skill): rewrite start router with user-facing natural questions
878e118 refactor(skill): remove direct Mode A/B commands, route only via /buddy:start
9bd88b0 chore(plugin): bump version 1.0.0 → 1.0.1 for iterative testing
bbe5f09 refactor(skill): remove run-uat/run-load-test/run-beta-program direct commands
```

### 그룹별 정리

**A. Engineering 구조 정비 (5 commits)** — engineering-phases.md를 SSoT로 확립:
- `34a978d` — Artifact-based 9-phase 정의 + 분류 매트릭스 + 4 cross-cutting 재분류
- `5b77193` — 131 PROCEDURE.md에 Input/Output Contract 추가 (스크립트 일괄)
- `ac58cb6` — 11 스킬에서 I/O Contract 삽입 위치 잘못된 것 정정 (스크립트 부작용)
- `064dfdc` — `assess-product-change` catalog/command/wireup 등재
- `aa1edcc` — flow-graph.md (Mermaid) + status 스킬 Phase 1 Mode B 반영

**B. Subagent 권한 자동화 (2 commits)** — subagent의 Edit/Bash 권한 부족 문제 해결:
- `addf307` — `internal/permissions/` 패키지 + `cmd/buddy/permissions_cmd.go` 신규 (Edit 14 + Bash 15 패턴)
- `231a537` — `buddy doctor`에 permissions check 통합 (DB 없어도 검진 가능)

**C. Phase 1 재구조화 (2 commits)** — `concretize-idea` 흐름과 산출물 확장:
- `9c1baae` — `write-hld` 신규 스킬 (HLD 9 섹션) + concretize-idea stage 6→3 customer 이동 + stage 8 write-hld 추가 + define-product-spec에 Actors/Use Cases 섹션 + map-actor-use-cases에 데이터 흐름 필드
- `38cf8ee` — concretize-idea line 5 진입 조건 정정 (Mode A/B는 코드 유무만 기준, 상업성은 별개 차원)

**D. 진입 라우터 + 명령 정리 (5 commits)** — UX 단순화:
- `db6672a` — `/buddy:start` 신규 (의도 분류 + auto-dispatch)
- `6444886` — start 본문 재작성 (Mode A/B 용어 제거, 사용자 자연 질문 기반)
- `878e118` — `/buddy:concretize-idea` + `/buddy:assess-product-change` 직접 호출 명령 삭제 (start 경유만)
- `9bd88b0` — plugin version 1.0.0 → 1.0.1 (dogfood용 patch bump)
- `bbe5f09` — `/buddy:run-uat`, `/buddy:run-load-test`, `/buddy:run-beta-program` 직접 명령 삭제 + version 1.0.2

---

## 4. 본 세션의 lock-in 결정 사항 (다시 결정하지 말 것)

| # | 결정 | 근거 |
|---|------|------|
| L1 | **Phase 정의는 Artifact-based** (Activity/Decision-based 아님) | I/O Contract와 자연 정합 + ADR-007 stateless 호환 |
| L2 | **Phase 1 = Problem/Opportunity Identification & Validation** (이전 "Idea & Business Validation") | greenfield + existing product 변경 둘 다 포괄 |
| L3 | **Mode A = 코드베이스 없음 / Mode B = 있음** | 작업 시작점 기준. 상업/비상업과는 직교 |
| L4 | **상업/비상업은 Mode와 독립 차원** | 각 stage 내부 분기 (concretize-idea stage 5/6 skip 가능) |
| L5 | **write-spec(PRD)과 write-hld(HLD) 분리** | 청중 분리 (stakeholder vs 엔지니어), anti-pattern §8과 일관 |
| L6 | **Use case는 PRD에서 LOGICAL, HLD §5에서 PHYSICAL** | 데이터 흐름은 PRD, product 매핑은 HLD |
| L7 | **`identify-actors` + `map-actor-use-cases`는 write-spec 내부 호출** | actors가 PRD의 핵심 정보. Phase 2에서 분리하지 않음 |
| L8 | **concretize-idea stage 순서: 1-2 idea → 3 customer → 4 competition → 5 viability → 6 pricing → 7 PRD → 8 HLD → 9 review** | 후속 단계가 customer 정의를 입력으로 필요로 함 |
| L9 | **`/buddy:start`가 유일한 사용자 진입 라우터** | Mode A/B 직접 호출 명령 폐기 |
| L10 | **start는 Mode A/B 용어 노출 X, 자연 질문(프로젝트 경로/작업 유형/상업성)** | 사용자 mental model과 일치 |
| L11 | **분류 결과는 항상 사용자 confirm 후 dispatch** | auto-dispatch는 사용자 통제권 약화 |
| L12 | **subagent 권한 주입은 Go 바이너리** (`buddy permissions inject`) | 스크립트보다 tamper-resistant |
| L13 | **상업 vs 비상업 = stage 내부 분기** (Mode 결정 표 안에 X) | concretize-idea stage 5 entry gate에서 결정 |
| L14 | **autoplan의 4 review 검증 대상**: review-scope=PRD, review-design=HLD, review-devex=HLD(SDK/CLI 있을 때), review-engineering=HLD per-product tech stack + Phase 4 plan(있을 때) | 타이밍 미스매치 해소 |
| L15 | **patch version bump** (1.0.1, 1.0.2) for iterative testing | minor bump은 milestone 완료 시 (ADR-011) |

---

## 5. 다음 세션의 후보 작업

### 즉시 가능 (검토 단계)

| 작업 | 비용 | 우선순위 |
|------|------|---------|
| **`/buddy:start` 동작 검증** (dogfood — 의도 분류 정확도, 메뉴 흐름 자연스러움) | 30분~1시간 | HIGH |
| **`write-hld` 동작 검증** (실제 호출해서 9 섹션 결과 품질) | 30분 | HIGH |
| **`buddy permissions check/inject` dogfood** (다른 머신에서 신선한 ~/.claude 환경 시뮬레이션) | 15분 | HIGH |
| **Phase 1 Mode A의 나머지 stage skill 본문 검토** (validate-idea, validate-advanced-edge-idea, assess-business-viability 등) | 2-3시간 | MID |

### 분리 검토 필요 (사용자 결정 필요)

| 작업 | 결정 포인트 |
|------|---------|
| **6-cluster orchestrator 분리** | concretize-idea를 cluster로 쪼개기 (validate-concept / validate-target-users / validate-business / validate-compliance / write-spec / validate-spec). 본 세션에서 논의만, 미실행 |
| **Phase 6/7의 다른 직접 명령 정리** | run-* 3개와 동일 논리로 prepare-launch-checklist, setup-canary-deploy 등 검토 |
| **UI/UX 디자인 스킬 통합** | 사용자가 다른 프로젝트에서 작업 중 — 통합 시점 `write-design` 스킬과 `review-design` cascade 완성 |
| **Phase 2-9 스킬 본문 상세 검토** | Phase 2(11개)부터 시작. Phase 3(33개)이 가장 큼 |

### 보류 (별도 트랙)

| 작업 | 비고 |
|------|------|
| **ADR-019 implementation** (plugin MCP exposure layer) | 이전 세션 결정. 1-3일 작업, 본 세션 무관 |
| **v1.0.0 release engineering** | B-2 production dogfood 완료 시 |

---

## 6. 핵심 아키텍처 변경 (사용자 mental model 갱신 필요)

### 6.1 진입 명령 단순화

```
이전:                             현재:
/buddy:concretize-idea            (제거)
/buddy:assess-product-change      (제거)
/buddy:run-uat                    (제거)
/buddy:run-load-test              (제거)
/buddy:run-beta-program           (제거)
                                  /buddy:start  ← 통합 진입
                                  /buddy:status (기존)
                                  /buddy:run    (advanced, 단일 스킬 직접 호출)
```

### 6.2 Phase 1 흐름 (재구조화)

```
이전 8 stages:                    현재 9 stages:
1. validate-idea                  1. validate-idea
2. validate-advanced-edge-idea    2. validate-advanced-edge-idea
3. assess-business-viability      3. map-customer-segments         ← 위치 변경 (먼저)
4. analyze-competition            4. analyze-competition (customer 기준)
5. review-pricing-and-gtm         5. assess-business-viability     ← 위치 변경 (customer 후)
6. map-customer-segments          6. review-pricing-and-gtm
7. define-product-spec            7. define-product-spec (actors+use cases 포함)
8. autoplan                       8. write-hld                      ← 신규
                                  9. autoplan (PRD+HLD review)
```

### 6.3 산출물 변경

```
이전:                             현재:
PRD                               PRD + HLD + review report
  Problem                           PRD: Problem, Actors, Use Cases (logical), User Stories,
  Target user                            Success Criteria, Functional/Non-functional Req, ...
  Features                          HLD: Product Decomposition, Per-product Tech Stack,
  ...                                    Inter-product Communication, Use Case→Product Mapping,
                                         External Integration, Deployment, Licensing, Distribution
                                    Review: review-scope (PRD) + review-design (HLD) +
                                            review-devex (HLD SDK/CLI) + review-engineering (HLD)
```

### 6.4 Mode B (기존 프로덕트 변경)

```
/buddy:start → "기존 프로젝트가 있나요? 경로 알려주세요"
            → "어떤 작업? (신규 기능 / 버그 수정 / 개선 / 리팩토링 / 의존성 갱신 / 기타)"
            → "상업용인가요? (yes / no / 모름)"
            → 정보 확인 → assess-product-change dispatch
              → small scope: → Phase 5 직접
              → medium scope: → Phase 3 → 5 → ...
              → large scope: → Phase 2 → 3 → 4 → 5 → ...
```

---

## 7. 알려진 위험 / 함정

### 7.1 subagent 권한 fragility

subagent가 백그라운드 실행이라 대화형 승인 불가 → `~/.claude/settings.local.json`에 미리 등록된 도구만 사용 가능. 본 세션에서 Edit 14 + Bash 15 패턴 추가했지만, **새 머신에서는 다시 `buddy permissions inject` 실행 필요**.

**대응**: 다음 세션 시작 시 `buddy permissions check` 먼저 실행 → 누락 있으면 `buddy permissions inject`.

### 7.2 스크립트 일괄 처리의 부작용

`5b77193` commit의 스크립트가 I/O Contract를 잘못된 위치(문서 중간)에 삽입한 케이스 11건 발생 → `ac58cb6`에서 정정. **다음 일괄 작업 시 삽입 위치 검증 로직 강화 필요**.

### 7.3 review-design의 검증 대상 부재

`review-design`은 UI/UX 디자인 산출물을 검증하지만, **buddy에 UI/UX 디자인 산출 스킬이 없음**. 사용자가 다른 프로젝트에서 작업 중 — 통합 시점 cascade 완성. 현재는 외부 디자인 산출물 입력이나 condition-based skip으로 동작.

### 7.4 index.lock race (이전 세션 §9.1과 동일)

본 세션 commit 도중 1회 발생. `gpgconf --kill gpg-agent` 또는 `rm .git/index.lock` 후 retry. 다음 세션도 동일 가능성.

### 7.5 user 발화 정확도 의존 (start의 의도 분류)

start의 Mode 1(자연어 분류)는 LLM 판단에 의존. 모호한 발화 시 잘못된 dispatch 가능. 본 세션의 mitigation: **모든 분류 결과를 사용자 confirm 후 dispatch**.

---

## 8. 사용자 글로벌 룰 (본 세션 적용 빈도 높은 5건)

1. **언어**: 사용자가 한국어로 쓰면 한국어로 응답. 영어 technical identifier (commit hash / skill name / path) 영어 유지.
2. **Git commit attribution 금지**: `Co-Authored-By` / "Generated with Claude" 절대 금지. commit message 영어 + Conventional Commits.
3. **Push 패턴**: 사용자 명시 승인 후만. 본 세션은 commit만, push는 사용자가 직접.
4. **Uncommitted 변경 종료 시**: 사용자에게 commit 여부 *먼저* 확인. 자율 commit / 폐기 금지.
5. **3회 연속 동일 오류 시**: 사용자 보고 후 지시 대기.

본 세션 적용 사례:
- 사용자가 "patch version" 강조 → minor bump(1.1.0)를 patch bump(1.0.1)로 정정
- 사용자가 "권장 vs 호출 불가" 차이 짚어줌 → "직접 호출 불가, /buddy:start 경유"로 정정
- 사용자가 "Mode와 상업성 분리" 짚어줌 → Mode 결정 표에서 상업성 차원 제거

---

## 9. 다음 세션 진입 순서 (권장)

### Step 1 — 정합성 검증 (5분)

```bash
cd <REPO_ROOT>
git log -16 --oneline                  # 본 세션 14 commits + 직전 확인
make test-routing 2>&1 | tail -3       # 10/10 통과
make build                             # bin/buddy + bin/buddy-mcp 빌드
bin/buddy permissions check            # subagent 권한 상태 — 누락 시 inject
go test ./internal/permissions/ -count=1
```

### Step 2 — 본 doc + 핵심 4 문서 read (10-15분)

```bash
cat docs/notes/2026-05-29-session-handoff.md             # 본 doc
cat plugin/skills/router/references/engineering-phases.md  # 9-phase SSoT
cat docs/plugin-skills-classification-matrix.md          # 152 스킬 분류
cat docs/plugin-skills-flow-graph.md                     # I/O Contract 시각화
```

### Step 3 — 다음 작업 결정

사용자가 §5에서 후보 작업 선택. 권장 우선순위:
1. dogfood 검증 (start / write-hld / permissions inject)
2. Phase 1 Mode A 나머지 stage skill 본문 검토 (validate-idea 등)
3. 6-cluster orchestrator 분리 (대규모 작업 — 사용자 결정 필요)

---

## 10. 핵심 파일 인덱스

| 문서 | 역할 |
|------|------|
| **`docs/notes/2026-05-29-session-handoff.md`** | **본 doc** — 다른 세션 진입점 |
| `plugin/skills/router/references/engineering-phases.md` | 9-phase artifact-based 정의 + Phase별 소속 스킬 테이블 (cross-machine SSoT) |
| `plugin/skills/router/references/skill-catalog.md` | 155 스킬 catalog — phase별 정렬 |
| `plugin/skills/router/references/routing-rules.md` | 스킬 간 라우팅 충돌 결정 |
| `docs/plugin-skills-classification-matrix.md` | 152 스킬 영역별(engineering/product/marketing/meta) + standalone 등급 분류 |
| `docs/plugin-skills-flow-graph.md` | 149 I/O Contract connections 시각화 (Mermaid) |
| `docs/plugin-skills-engineering-flow.md` | SE 이론 10-phase × 76 step 매핑 (이전 세션, 여전히 유효) |
| `plugin/skills/concretize-idea/PROCEDURE.md` | Phase 1 Mode A orchestrator (9 stages, 재구조화 완료) |
| `plugin/skills/assess-product-change/PROCEDURE.md` | Phase 1 Mode B orchestrator (신규) |
| `plugin/skills/write-hld/PROCEDURE.md` | High Level Design 작성 (신규, 9 섹션) |
| `plugin/skills/start/PROCEDURE.md` | 사용자 의도 기반 진입 라우터 (신규) |
| `internal/permissions/permissions.go` | subagent 권한 자동 주입 (신규 Go 패키지) |
| `cmd/buddy/permissions_cmd.go` | `buddy permissions check/inject` CLI (신규) |
| `internal/diagnose/doctor.go` | `buddy doctor` — permissions check 통합 |

---

## 11. 자주 쓰는 명령

```bash
# 빌드 + 테스트
make build                            # bin/buddy + bin/buddy-mcp
make test-routing                     # 10/10 통과 필수
go test -race -count=1 ./...
go vet ./...

# 본 세션 작업 trail spot check
git log -16 --oneline                 # 본 세션 14 commits
git show bbe5f09 --stat               # 마지막 commit (run-* commands 삭제)
git show 9c1baae --stat               # 가장 큰 commit (write-hld + concretize-idea)

# subagent 권한 (다른 머신 첫 사용 시 필수)
bin/buddy permissions check
bin/buddy permissions inject          # 누락 시

# Plugin install (dogfood용)
# Plugin version 1.0.2가 Claude Code marketplace에 반영되었는지 확인 후 재설치

# /buddy:start 동작 검증
# Claude Code 세션에서:
#   /buddy:start                      ← 메뉴 흐름
#   /buddy:start "내가 만든 OSS에 새 기능 추가하고 싶어"   ← 자연어 분류
#   /buddy:status                     ← 현재 phase 안내
```

---

## 12. End-of-handoff sanity check

이 doc만 보고도 다음 세션이 다음 정보 회복 가능:

- [x] 현재 branch + last commit (`bbe5f09`)
- [x] 본 세션이 무엇을 했고 어디서 멈췄는지 (14 commits + 다음 작업 후보)
- [x] working tree의 변경 파일 리스트 (0 — clean)
- [x] 다음에 무엇을 할지 + 어느 순서로 (§5 + §9)
- [x] 사용자 글로벌 룰 + 본 세션의 결정 트레이스 (§8 + §4)
- [x] 알려진 위험 (5 개 명시)
- [x] cross-machine 진입 전제 (§0)

본 doc의 정확도가 stale 해지면 **본 doc부터 갱신**. 다른 세션이 의지하는 SSoT.

---

**문서 끝**. 다음 세션 시작 시 §0 → §9 sequence 진입.
