# SE Lifecycle Naming — buddy 명사 표준

> **목적**: buddy 라이프사이클을 **artifact 의존성 그래프(DAG)** 로 정의하고, 각 노드(단계)에 대해 **(a) 내부 작업 약어** + **(b) 한국어 일반 명사** + **(c) 영문 표준 용어(SE/Agile)** + **(d) Input(DoR)/Output(DoD) 계약**을 표준화한다. 모든 buddy 문서·출력은 본 SSoT를 따른다.
>
> **핵심 원리**: 라이프사이클은 *시간순 단계*가 아니라 *산출물 의존성 그래프*다. "Phase 1~9" 번호는 그래프를 사람이 읽기 쉽게 묶은 **클러스터 라벨**이며 강제 실행 순서가 아니다. 각 노드는 **자신이 생산하는 산출물(output)로 명명**하므로 단어만 보고 산출물이 유추되고, input은 그래프의 직전 노드(엣지)에서 읽힌다. 진입·완료 계약은 Agile의 **Definition of Ready(DoR = input gate)** / **Definition of Done(DoD = output gate)** 로 표기한다.
>
> **배경**: 기존 문서들이 같은 노드를 미세하게 다른 명사로 부르고, 선형 SDLC 어휘(Phase 1→9)로 순환·다진입 구조를 설명해 일관성·정합성이 결여됨. 표준 SE/Agile 용어로 통일하여 단어 자체가 input/output을 함의하게 한다. "Phase 1", "CE1-CE5" 같은 약어를 외부 출력에 그대로 노출하면 사용자 이해가 어려워지고 불필요한 확인 토큰이 소비됨.
>
> **사용 시점**:
> - 모든 사용자 대면 문서·답변·출력 작성 시
> - 신규 skill 작성 시 (`write-a-skill`)
> - 평가 리포트 작성 시 (`evaluate-skill`)
> - ADR, 가이드, README 등 모든 영속 문서

---

## §1. 라이프사이클 DAG — 노드 명사 매핑

라이프사이클은 **노드(산출물을 생산하는 단계) + 엣지("왼쪽 노드의 Output = 오른쪽 노드의 Input")** 로 구성된 의존성 그래프다. 각 노드는 4가지 표현을 가진다: 내부 약어 · 한국어 명사 · 영문 표준 용어 · Input(DoR)/Output(DoD) 계약. **노드는 자신이 생산하는 산출물(output)로 명명**하므로 단어만 보고 산출물이 유추되고, input은 그래프의 직전 노드에서 읽힌다.

| 내부 약어 | 한국어 명사 (사용자 대면) | 영문 표준 용어 (SE/Agile) | Input — DoR | Output — DoD (이름에 박힌 산출물) | 표준 출처 |
|---|---|---|---|---|---|
| Phase 1 · Mode A | **문제·기회 검증 단계** | Discovery / Inception | idea·hypothesis | 검증된 PRD + HLD | Lean·Agile / RUP |
| Phase 1 · Mode B | **변경 영향 분석 단계** | Impact Analysis | change request + 기존 codebase | Impact Assessment + Scope 분류 | ISO/IEC 14764 |
| Phase 2 | **기능 정의 단계** | Requirements Specification | PRD / validated work item | SRS = feature backlog + actor·UC map | SWEBOK · IEEE 830/29148 |
| Phase 3 | **기술 설계 단계** | Software Design | requirements spec (SRS) | SDD = ADR·API contract·data model | SWEBOK · IEEE 1016 |
| Phase 4 | **구현 계획 단계** | Iteration Planning (WBS) | design (SDD) + specs | Iteration Plan = task DAG·actor-track | Scrum · PMBOK |
| Phase 5 | **개발 단계** | Construction (Implementation) | iteration plan | working code + developer tests | SWEBOK (Software Construction) |
| Phase 6 | **검증·확인 단계** | Verification & Validation (V&V) | code + developer tests | V&V evidence = QA·security·compliance | IEEE 1012 |
| Phase 7 | **출시·배포 단계** | Release & Deployment | V&V passed code | release package + deployed artifact | DevOps · ITIL |
| Phase 8 | **운영·개선 단계** | Operation & Maintenance | deployed artifact + production traffic | ops metrics + improvement backlog | ISO 12207 · 14764 |
| Phase 9 | **폐기·종료 단계** | Retirement / Decommissioning | usage data + business decision | deprecation·migration·EOL plan | ISO/IEC/IEEE 12207 (Disposal) |

> **번호 = 클러스터 라벨, 순서 아님**: "Phase 1~9"는 그래프를 사람이 읽기 쉽게 묶은 라벨이며 강제 실행 순서가 아니다. 실제 진입점은 "지금 존재하는 artifact의 frontier"로 결정된다(§1.1).
>
> **2개 정합 교정**: (a) Phase 6 `Verification & Quality` → **V&V** — 빠져 있던 *Validation*("올바른 제품인가") 복원. (b) Phase 9 `Lifecycle Management`(상위 lifecycle 개념과 충돌하던 오기) → **Retirement / Decommissioning**(ISO 12207 Disposal). 추가로 Phase 1 Mode B를 ISO 14764 정식 유지보수 용어 **Impact Analysis**로 명명.

### §1.1 DAG 구조 (엣지 = DoD→DoR 계약)

```
trigger: idea | change request | incident | metric signal
   │  (어떤 source + 어떤 artifact가 이미 존재하나 → 진입 노드 결정)
   ▼
Discovery ─(PRD)─► Requirements Spec ─(SRS)─► Software Design ─(SDD)─► Iteration Planning
   ─(task DAG)─► Construction ─(code+tests)─► V&V ─(evidence)─► Release & Deployment
   ─(deployed)─► Operation & Maintenance ─(improvement = 새 source)─► [Requirements Spec로 역류]
                                          └─(EOL 결정)─► Retirement

· ─(…)─ = "왼쪽 노드의 DoD = 오른쪽 노드의 DoR"
· 진입 = 현재 존재하는 artifact의 frontier (status 스킬이 자동 탐지)
    아무것도 없음        → Discovery부터 (전체 빌드)
    코드만 + 작은 변경   → Construction부터 (Mode B small)
    SRS 있음             → Software Design부터 (Mode B medium)
    production + metrics → Operation부터, 결과가 Requirements Spec로 역류
· prerequisite(DoR) 미충족 → 그 input을 생산하는 upstream 노드로 자동 선행 (구 "backtrack")
· 사이클 = trigger가 새 source를 주입하면 downstream만 증분 재평가 (Make/Bazel stale-rebuild 의미론)
```

### §1.2 각 노드 핵심 활동

| 노드 | 핵심 활동 |
|------|---------|
| 문제·기회 검증 (Discovery) | 아이디어 검증 + 사업성 평가 + PRD/HLD 작성 |
| 변경 영향 분석 (Impact Analysis) | 변경 영향 평가 + scope 분류 + routing (기존 제품) |
| 기능 정의 (Requirements Spec) | actor·use case·system boundary 분해 → feature 합성 → backlog |
| 기술 설계 (Software Design) | tech stack ADR + infra + API contract + data model + 설계 검토 |
| 구현 계획 (Iteration Planning) | actor별 task 분해 + 의존성 DAG + 병렬 실행 계획 |
| 개발 (Construction) | TDD 루프 + 병렬 worker agent + 코드 + 자체 테스트 작성 |
| 검증·확인 (V&V) | 상용 quality bar 검증 (coverage / security / a11y / compliance / code health) |
| 출시·배포 (Release & Deployment) | 패키징 + 태깅 + 배포 + canary + UAT + launch readiness |
| 운영·개선 (Operation & Maintenance) | A/B 실험 + funnel 분석 + 인시던트 대응 + improvement backlog |
| 폐기·종료 (Retirement) | feature/product deprecation + 마이그레이션 + EOL |

### §1.3 고도(altitude) — 같은 그래프, 다른 재평가 빈도

| 고도 | 노드 | 재평가 빈도 |
|------|------|------------|
| 제품 수명 (macro) | Discovery, Retirement | 드물게 — 제품 1회 탄생 / 폐기 |
| 반복 (micro) | Requirements Spec ~ Operation & Maintenance | 매 iteration · trigger마다 |

---

## §2. Phase 1 내부 — Mode A / Mode B

문제·기회 검증 단계는 프로덕트 존재 여부에 따라 2가지 모드로 동작 (그래프 진입 노드가 갈림).

| 내부 약어 | 한국어 일반 명사 (사용자 대면) | 영문 표준 용어 |
|---------|--------------------------|-------------|
| Mode A | **신규 제품 기획 모드** | Discovery / Inception (Greenfield) |
| Mode B | **기존 제품 변경 모드** | Impact Analysis (Change Assessment) |

- **신규 제품 기획 모드** (Mode A): 코드베이스가 없는 상태에서 아이디어로부터 시작 → PRD + HLD + 사업성 검증
- **기존 제품 변경 모드** (Mode B): 기존 코드베이스가 있는 상태에서 변경 요청 접수 → 영향 평가 + scope 분류 → 다음 단계로 라우팅

---

## §3. 평가 항목 약어 (`evaluate-skill` 내부)

| 내부 약어 | 한국어 일반 명사 (사용자 대면) | 영문 일반 명사 |
|---------|--------------------------|-------------|
| F1-F5 | **frontmatter 평가 항목** | Frontmatter Checklist |
| CE1-CE5 | **카탈로그 항목 평가 기준** | Catalog Entry Criteria |
| B1-B12 | **본문 구조 평가 항목** | Body Structure Checklist |
| P1-P4 | **페르소나 평가 항목** | Persona Checklist |
| PC1-PC5 | **paired 평가 케이스** | Paired Evaluation Cases |
| C1-C4 | **단일 평가 케이스** | Single Evaluation Cases |

---

## §4. 용어 사용 규칙 (적용 정책)

| 컨텍스트 | 사용 형식 | 예시 |
|--------|---------|------|
| **내부 작업** (Claude 추론, 코드, task name, 식별자) | 내부 약어 — 간결 | "Phase 1", "F3", "PC2" |
| **사용자 대면 답변·output** | 일반 명사 우선 + 첫 등장 시 약어 부기 | "문제·기회 검증 단계(내부 약어: Phase 1)" |
| **문서 본문** (가이드, README, ADR, notes) | 일반 명사 우선, 약어는 첫 등장 시 정의 | "본문 구조 평가 항목(이하 B1-B12)" |
| **표 헤더 / 식별자 컬럼** | 약어 OK (공간 제약) | `| ID | 평가 결과 | ...| F1 | ✅ pass |` |
| **코드 주석 / 변수명** | 약어 OK (코드 가독성) | `// PC2 case: command exists + persona not recommended` |

### §4.1 첫 등장 시 약어 부기 방법

처음 약어를 사용할 때는 일반 명사 + 약어를 함께 표기. 같은 문서에서 두 번째 등장부터는 약어만 사용해도 됨.

```
✅ 권장: "문제·기회 검증 단계(이하 Phase 1)에서..."
✅ 권장: "Frontmatter 평가 항목(F1-F5)은 5개로 구성된다."
❌ 비권장: "Phase 1에서 CE1-CE5를 평가한다." (어떤 단계인지·어떤 평가인지 불명)
```

### §4.2 외부 의존 용어 (Anthropic 정의 등)

Anthropic 공식 docs의 용어는 그대로 사용 + 출처 명시:

| 외부 의존 용어 | 사용 형식 |
|------------|---------|
| `skillListingBudgetFraction` | 그대로 사용 + "(Claude Code 공식 설정 키)" 부기 |
| `disable-model-invocation`, `when_to_use` 등 frontmatter 필드 | 그대로 사용 + 첫 등장 시 의미 설명 |
| `claude-opus-4-7`, `claude-sonnet-4-6` 등 모델 ID | 그대로 사용 |

---

## §5. 기존 문서 정렬 작업 가이드

본 SSoT 신설 후, 기존 buddy 문서들은 다음 정렬 작업이 필요:

| 문서 | 정렬 필요 영역 |
|------|----------|
| `plugin/skills/router/SKILL.md` "Skill index" 표 | Phase 명사를 §1 표 영문 명사로 통일 + 한국어 부기 |
| `plugin/skills/router/references/engineering-phases.md` 각 Phase 정체성 표 | 영문 명사를 §1 표와 일치시킴 |
| `plugin/skills/router/references/routing-rules.md` §5 약칭 | §1 표 한국어 명사로 통일 |
| `plugin/skills/router/references/skill-catalog.md` §2 phase 헤더 | §1 표 형식으로 통일 |
| `docs/plugin-skills-classification-matrix.md` | §1 표 영문 명사로 통일 |
| `docs/plugin-skills-authoring-guide.md` | §3.5.1 등 phase 언급 모두 일반 명사 부기 |
| `plugin/skills/evaluate-skill/PROCEDURE.md` | 평가 리포트 출력 형식에 일반 명사 적용 |
| `plugin/skills/write-a-skill/PROCEDURE.md` | 사용자 대면 yaml 산출물 영역에 일반 명사 적용 |
| `plugin/skills/start/PROCEDURE.md` | dispatch context schema는 약어 그대로(내부), 사용자 질문은 일반 명사(이미 적용됨) |

정렬은 점진 — 신규 문서·수정 시점에 §1 SSoT 따름.

---

## §6. 변경 trigger

| trigger | 갱신 대상 |
|---------|---------|
| 새 phase 추가/분리/병합 | §1 phase 표 |
| 새 평가 항목 카테고리 추가 | §3 평가 항목 약어 표 |
| 외부 의존 용어 추가 (Anthropic 신규 발표) | §4.2 |
| 한국어/영문 명사 표현 변경 결정 | §1 + §2 표 동기 갱신 |

---

## §7. 참조

- `plugin/skills/router/references/engineering-phases.md` — 9 phase 각 정체성·산출물·전이 규칙 (artifact-based 정의)
- `plugin/skills/router/references/skill-catalog.md` — phase별 skill 등재 카탈로그
- `docs/plugin-skills-authoring-guide.md` — skill 작성·평가 기준 SSoT
- `plugin/skills/write-a-skill/references/naming-convention.md` — 스킬 작성 명명 규칙 (skill 디렉토리/파일/식별자)

본 문서는 **명사 정의의 SSoT**, 위 문서들은 본 SSoT를 따르는 정합 문서.
