# Skill Naming Convention — 스킬 작성 명명 규칙

> **목적**: buddy 스킬 작성 시 따르는 명명 규칙 SSoT. 스킬 자체 구성 요소(디렉토리·파일·식별자·섹션·표·약어)의 명명에 한정.
>
> **적용 범위 외**: 코드 언어별 명명 규칙 (JavaScript camelCase, Python snake_case 등) 은 별도 언어 정책 문서에서 정의. 본 문서는 스킬 마크다운 작성에 한정.
>
> **사용 시점**: 신규 스킬 작성 시 (`write-a-skill` Step 1) + 기존 스킬 검토 시 + `evaluate-skill` 평가 시 명명 일관성 검증 항목으로 활용.
>
> **상위 참조**:
> - `plugin/skills/router/references/se-lifecycle-naming.md` — 9 phase 명사 SSoT (본 문서 N8 참조)
> - `docs/plugin-skills-authoring-guide.md` — skill 작성·평가 기준 SSoT (본 문서가 그 안의 명명 규칙 부분을 상세화)

---

## N1. 스킬 디렉토리 명명

| 규칙 | 내용 |
|------|------|
| 형식 | `kebab-case` (소문자 + 하이픈) |
| 시작 단어 | **영어 동사 시작 권장** (예: `validate-idea`, `design-system`, `build-feature`) |
| 한국어 화자가 명사로 인식하는 단어 | 영어 사전상 동사 형태이면 허용 (예: `brainstorm`, `audit`) |
| 명사형 허용 케이스 | **패턴 라이브러리 항목만** (`router`, `catalog` 등) |
| 위치 | `plugin/skills/<skill-name>/` |
| 길이 | 2-4 단어 권장 (너무 짧으면 모호, 너무 길면 가독성 저하) |

**예시**:
- ✅ `validate-idea`, `design-data-model`, `audit-security`, `concretize-idea`
- ❌ `Validate_Idea` (snake_case + PascalCase 혼용), `idea-validation` (명사 시작), `vi` (너무 짧음)

---

## N2. 핵심 파일 명명

| 파일 | 명명 규칙 | 비고 |
|------|---------|------|
| `PROCEDURE.md` | **대문자 고정** | 모든 스킬 본문. buddy 표준 (frontmatter 미사용, router 경유 lazy-load) |
| `SKILL.md` | **대문자 고정** | router 1개만 사용 (Anthropic 공식 표준 호환). 일반 스킬에는 사용 안 함 |
| `references/<topic>.md` | kebab-case, 토픽 명사형 | 보조 자료. 예: `yaml-output-spec.md`, `attribution-classification.md` |
| `examples/case-<n>.md` | kebab-case + 번호 | few-shot 예시 케이스. 예: `case-1.md`, `case-2.md` |

**디렉토리 구조 예시**:
```
plugin/skills/<skill-name>/
├── PROCEDURE.md                          # 필수 (본문)
├── references/                           # 선택 (긴 보조 자료)
│   ├── <topic-1>.md
│   └── <topic-2>.md
└── examples/                             # 선택 (긴 예시)
    ├── case-1.md
    └── case-2.md
```

---

## N3. Frontmatter 필드 명명

PROCEDURE.md에 frontmatter 도입 시 (향후 마이그레이션), Anthropic 공식 규격 그대로 사용:

| 필드 | 형식 | 출처 |
|------|------|------|
| `name` | kebab-case (디렉토리명과 일치) | Anthropic 공식 |
| `description` | 영문 + 한국어 trigger 키워드 혼용 | Anthropic 공식 |
| `when_to_use` | 영문/한국어 (description과 합산 1,536자 cap) | Anthropic 공식 |
| `disable-model-invocation` | boolean (kebab-case) | Anthropic 공식 |
| `argument-hint` | 사용자 가시 hint 문자열 | Anthropic 공식 |
| `model` | model ID 형식 (예: `claude-opus-4-7`) | Anthropic 공식 |
| `effort` | `low`/`medium`/`high`/`xhigh`/`max` | Anthropic 공식 |
| `allowed-tools` | YAML 리스트 (tool 이름 정확) | Anthropic 공식 |

**규칙**: Anthropic 공식 docs(https://code.claude.com/docs/en/skills.md)의 필드명·형식을 그대로 따른다. 임의 변형 금지.

---

## N4. 본문 섹션 헤더 명명

| 규칙 | 내용 |
|------|------|
| 우선순위 | **한국어 일반 명사 우선** + 영문 보조 표기 (필요 시) |
| 약어 사용 | 헤더에는 약어 금지 (사용자 가독성) |
| 형식 | `## <한국어 명사> (<영문 보조>)` 또는 `## <한국어 명사>` |

**예시**:

| ✅ 권장 | ❌ 비권장 |
|--------|--------|
| `## 입력 요구사항 (Input Requirements)` | `## I/O Contract` |
| `## 실행 절차 (Steps)` | `## §3` |
| `## 출력 형식 (Output Format)` | `## Schema` |
| `## 검증 체크리스트` | `## §V` |
| `## 다음 단계 (Next Steps)` | `## ▸` |

---

## N5. 표 컬럼 헤더 명명

| 규칙 | 내용 |
|------|------|
| 한국어 우선 | 일반 명사 |
| 외부 의존 키는 영문 고정 | Anthropic 정의 키 (`Type`, `Required`, `Source`) 등 |

**Input Requirements 표 예시**:

```markdown
| Input | Required | Type | Source | 미제공 시 |
|-------|----------|------|--------|----------|
| <항목> | ✅/선택 | artifact/knowledge/decision | <출처> | "<질의문>" |
```

`Input`, `Required`, `Type`, `Source`는 `engineering-phases.md` §4 표준이므로 영문 고정. `미제공 시`는 buddy 자체 컬럼이므로 한국어.

**Output Contract 표 예시**:

```markdown
| Output | Type | Format | Consumers |
|--------|------|--------|-----------|
| <산출물> | artifact/decision | <형식> | <소비 skill> |
```

---

## N6. 식별자 명명 (Step, 검증 항목 등)

### N6.1 Step 식별자

| 규칙 | 형식 | 예시 |
|------|------|------|
| 형식 | `Step N. <imperative 동사>` | `Step 1. 입력 파싱`, `Step 2. 대상 파일 read` |
| 번호 | 정수, 1부터 시작 | Step 1, Step 2, ... |
| 분할 step | `Step N.M` 형식 (sub-step) | `Step 2.1`, `Step 2.2` |
| 명사형 금지 | 동사 시작 (절차 명확) | ✅ "Step 1. 분류" ❌ "Step 1. 분류 작업" |

### N6.2 검증 항목 식별자

| 카테고리 | 약어 형식 | 예시 |
|--------|---------|------|
| Frontmatter | F1, F2, ..., F5 | `F3: 합산 1,536자 cap` |
| Catalog Entry | CE1, CE2, ..., CE5 | `CE4: 차별점 명확` |
| Body 구조 | B1, B2, ..., B12 | `B7: imperative 절차` |
| Persona | P1, P2, ..., P4 | `P1: 최상단 페르소나 정의` |
| 평가 케이스 | C1-C4 (single) / PC1-PC5 (paired) | `C3 케이스: PROCEDURE.md 단독, persona 권장` |

**약어 정의는 `plugin/skills/router/references/se-lifecycle-naming.md` §3에 SSoT.**

---

## N7. 외부 의존 용어 처리

Anthropic 공식 정의 용어는 그대로 사용 + 첫 등장 시 출처/의미 명시.

| 외부 의존 용어 | 처리 |
|------------|------|
| `skillListingBudgetFraction` | 그대로 사용 + "(Claude Code 공식 설정 키)" 부기 |
| `disable-model-invocation`, `when_to_use` 등 frontmatter 필드 | 그대로 사용 + 첫 등장 시 의미 설명 |
| `claude-opus-4-7`, `claude-sonnet-4-6` 등 모델 ID | 그대로 사용 |
| `${CLAUDE_PLUGIN_ROOT}` 등 환경 변수 | 그대로 사용 |
| `PreToolUse`, `Stop` 등 hook event | 그대로 사용 |

---

## N8. SE Lifecycle 단계 명명

9 phase 명명은 `plugin/skills/router/references/se-lifecycle-naming.md` §1 SSoT를 따른다. 본 문서는 그 매핑을 다시 정의하지 않고 참조만.

요약:
- 내부 작업: `Phase 1` ~ `Phase 9`
- 사용자 대면: 한국어 일반 명사 (예: "문제·기회 검증 단계")
- 영문 문서: 영문 일반 명사 (예: "Problem/Opportunity Validation")
- 첫 등장 시 약어 부기 ("문제·기회 검증 단계(이하 Phase 1)")

상세는 `plugin/skills/router/references/se-lifecycle-naming.md` §1 + §2 (Mode A/B) + §3 (평가 약어) + §4 (용어 사용 규칙) 참조.

---

## N9. 사용자 대면 출력 vs 내부 작업 용어 분리 원칙

스킬의 출력에는 두 영역이 공존:
- **사용자 대면 영역**: 답변·리포트·사용자 질문 등 사람이 읽는 부분
- **내부 작업 영역**: 코드 식별자·task name·디버그 로그·내부 yaml schema

| 영역 | 약어 사용 | 일반 명사 |
|------|---------|---------|
| 사용자 대면 답변 본문 | ❌ 단독 사용 금지 | ✅ 우선 + 약어 부기 |
| 사용자 대면 리포트 표 | 🟡 표 헤더에만 OK (공간 제약) | ✅ 본문 설명에 사용 |
| 내부 작업 식별자 (task, code, yaml key) | ✅ 약어 OK | (선택) |
| 디버그·로그 메시지 | ✅ 약어 OK | (선택) |
| ADR / 가이드 / 영속 문서 | 🟡 첫 등장 시 정의 후 약어 OK | ✅ 첫 등장 + 정의 |

**Default**: 모호한 경우 일반 명사 선택 (사용자 가독성 우선).

---

## N10. 명명 규칙 자동 검증

`evaluate-skill`은 paired evaluation 시 본 명명 규칙 준수 여부를 다음 항목으로 검증:

| 검증 영역 | 검사 항목 |
|---------|---------|
| 스킬 디렉토리명 (N1) | kebab-case + 동사 시작 (패턴 라이브러리 예외) |
| 핵심 파일명 (N2) | `PROCEDURE.md` 대문자 고정 |
| 본문 섹션 헤더 (N4) | 약어 단독 사용 금지 |
| 표 컬럼 헤더 (N5) | Input Requirements / Output Contract 표 표준 형식 준수 |
| Step 식별자 (N6.1) | `Step N. <동사>` 형식 |
| 검증 항목 약어 (N6.2) | 카테고리 + 번호 형식 |

위반 시 `evaluate-skill` 리포트에 actionable 개선 제안 출력.

---

## N11. 변경 trigger

| trigger | 갱신 대상 |
|---------|---------|
| Anthropic 공식 frontmatter 필드 추가/변경 | N3 |
| 새 검증 항목 카테고리 추가 | N6.2 + `se-lifecycle-naming.md` §3 동기 |
| buddy 표 표준 형식 변경 (engineering-phases.md §4) | N5 |
| 스킬 디렉토리/파일 구조 변경 | N1, N2 |

---

## 참조

- `plugin/skills/router/references/se-lifecycle-naming.md` — SE Lifecycle 명사 SSoT
- `docs/plugin-skills-authoring-guide.md` — 스킬 작성·평가 기준 SSoT
- `plugin/skills/router/references/engineering-phases.md` §4 — I/O Contract 표 표준
- `plugin/skills/write-a-skill/PROCEDURE.md` — 본 references를 참조하는 메타 스킬
- Anthropic 공식 docs — https://code.claude.com/docs/en/skills.md
