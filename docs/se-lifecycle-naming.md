# SE Lifecycle Naming — buddy 명사 표준

> **목적**: buddy 9-phase 라이프사이클의 각 단계에 대해 **(a) 내부 작업 약어** + **(b) 한국어 일반 명사** + **(c) 영문 일반 명사** 3가지 표현을 표준화한다. 모든 buddy 문서·출력은 본 SSoT를 따른다.
>
> **배경**: 기존 문서들이 같은 phase를 미세하게 다른 명사로 부르고 있어 일관성 결여. "Phase 1", "CE1-CE5" 같은 약어를 외부 출력에 그대로 노출하면 사용자 이해가 어려워지고 불필요한 확인 토큰이 소비됨.
>
> **사용 시점**:
> - 모든 사용자 대면 문서·답변·출력 작성 시
> - 신규 skill 작성 시 (`write-a-skill`)
> - 평가 리포트 작성 시 (`evaluate-skill`)
> - ADR, 가이드, README 등 모든 영속 문서

---

## §1. 9-Phase 라이프사이클 명사 매핑

각 단계는 다음 3가지 표현을 가진다.

| # | 내부 약어 (작업·식별) | 한국어 일반 명사 (사용자 대면) | 영문 일반 명사 (공식 문서) |
|---|---------------------|--------------------------|------------------------|
| 1 | Phase 1 | **문제·기회 검증 단계** | Problem/Opportunity Validation |
| 2 | Phase 2 | **기능 정의 단계** | Feature Definition |
| 3 | Phase 3 | **기술 설계 단계** | Technical Design |
| 4 | Phase 4 | **구현 계획 단계** | Implementation Planning |
| 5 | Phase 5 | **개발 단계** | Development |
| 6 | Phase 6 | **품질 검증 단계** | Verification & Quality |
| 7 | Phase 7 | **출시 단계** | Release |
| 8 | Phase 8 | **운영·개선 단계** | Operations & Iteration |
| 9 | Phase 9 | **수명주기 관리 단계** | Lifecycle Management |

### §1.1 각 phase별 1줄 설명

| 단계 | 핵심 활동 | 대표 산출물 |
|------|---------|---------|
| 문제·기회 검증 단계 | 아이디어 검증 + 사업성 평가 + PRD/HLD 작성 (신규) 또는 변경 영향 평가 + scope 분류 (기존 제품) | PRD, HLD, 영향 평가 리포트 |
| 기능 정의 단계 | actor·use case·system boundary 분해 → feature 합성 → backlog | Feature backlog (priority + estimate) |
| 기술 설계 단계 | tech stack ADR + infra + API contract + data model + 설계 검토 | ADR, API spec, data model, design docs |
| 구현 계획 단계 | actor별 task 분해 + 의존성 DAG + 병렬 실행 계획 | Task DAG, parallel execution plan |
| 개발 단계 | TDD 루프 + 병렬 worker agent + 코드 + 자체 테스트 작성 | Working code + developer-authored tests |
| 품질 검증 단계 | 상용 quality bar 검증 (coverage / security / a11y / compliance / code health) | QA report, security audit, compliance sign-off |
| 출시 단계 | 패키징 + 태깅 + 배포 + canary + UAT + launch readiness | Tagged release, deployed artifact, launch checklist pass |
| 운영·개선 단계 | A/B 실험 + funnel 분석 + 인시던트 대응 + improvement backlog | Experiment results, postmortem, improvement tasks |
| 수명주기 관리 단계 | feature/product deprecation + 마이그레이션 + EOL | Deprecation plan, migration plan, EOL documentation |

---

## §2. Phase 1 내부 — Mode A / Mode B

문제·기회 검증 단계는 프로덕트 존재 여부에 따라 2가지 모드로 동작.

| 내부 약어 | 한국어 일반 명사 (사용자 대면) | 영문 일반 명사 |
|---------|--------------------------|-------------|
| Mode A | **신규 제품 기획 모드** | Greenfield (New Product) |
| Mode B | **기존 제품 변경 모드** | Existing Product (Change Assessment) |

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
