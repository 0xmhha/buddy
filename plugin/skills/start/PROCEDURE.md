# start — 사용자 의도 기반 진입 라우터

사용자가 "무엇을 하려는지" 표현하면 적절한 Phase 1 orchestrator로 라우팅한다. 사용자가 command 이름을 기억할 필요 없이, 자기 의도를 자연어로 표현하거나 메뉴에서 선택만 하면 됨.

**진입 조건**: 어떤 command를 써야 할지 모를 때 / 새 작업의 시작점 / 처음 buddy 사용 시.
**산출물**: 적절한 orchestrator로의 dispatch + 그 orchestrator의 산출물.
**다음 phase**: 라우팅된 orchestrator의 흐름을 따름.

## Input Requirements

| Input | Required | Type | Source | 미제공 시 |
|-------|----------|------|--------|----------|
| 사용자 의도 (자연어 발화) | 선택 | knowledge | `$ARGUMENTS` | 메뉴 제시 후 선택받음 (Mode 2) |

## Output Contract

| Output | Type | Format | Consumers |
|--------|------|--------|-----------|
| Dispatch decision + target orchestrator | decision | inline + Skill tool 호출 | `concretize-idea`, `assess-product-change`, `status` 등 |

---

## 동작 모드

### Mode 1: 입력이 있을 때 — 의도 분석 후 직접 라우팅

`$ARGUMENTS`에 사용자 발화가 있으면 아래 분류 규칙으로 의도를 판단하고 즉시 해당 orchestrator로 dispatch.

### Mode 2: 입력이 없을 때 — 메뉴 제시 후 라우팅

`$ARGUMENTS`가 비어있으면 메뉴를 보여주고 사용자 선택 + 상세 입력을 받음.

---

## Mode 1: 의도 분류 규칙

사용자 발화에서 다음 signal을 찾아 분류한다.

### → `concretize-idea` (Phase 1 Mode A: 신규 아이디어)

**Signal**:
- "아이디어가 있어", "구체화하고 싶어", "새로 만들고 싶어"
- "어떻게 시작할지", "PRD 만들어줘", "spec 작성하자"
- "스타트업", "신규 프로젝트", "처음 만드는"
- "프로토타입을 빌드하고 싶어"
- 기존 코드베이스 언급 없음

**라우팅 행동**: `Skill` tool로 `concretize-idea` 호출, `$ARGUMENTS` 전달.

### → `assess-product-change` (Phase 1 Mode B: 기존 프로덕트 변경)

**Signal**:
- "버그", "에러", "안 돌아가", "고쳐", "수정"
- "기능 추가", "feature 넣어", "새 기능"
- "리팩토링", "개선", "최적화"
- "의존성 갱신", "업데이트"
- "이 프로젝트", "내 코드" 같은 기존 프로덕트 지시

**라우팅 행동**: `Skill` tool로 `assess-product-change` 호출, `$ARGUMENTS` 전달.

### → `status` (현재 위치 확인)

**Signal**:
- "어디까지 했지", "현재 상태"
- "다음에 뭐 해야", "어디부터 시작"
- "지금 어느 단계"

**라우팅 행동**: `Skill` tool로 `status` 호출.

### → 분류 불가 (모호한 경우)

**Signal**:
- 짧고 모호한 발화 ("도와줘", "뭐 할까?")
- 두 가지 의도가 섞인 경우
- buddy scope 외 ("날씨 알려줘" 등)

**라우팅 행동**: 사용자에게 추가 질문:
- buddy scope 발화면: 메뉴 제시 (Mode 2 흐름으로 fallback)
- buddy scope 외면: "이 도구는 소프트웨어 빌딩 작업을 지원해요. 어떤 작업을 하시겠어요?" + 메뉴

---

## Mode 2: 메뉴 제시 흐름

`$ARGUMENTS`가 비어있을 때 아래 형식으로 출력:

```
무엇을 하려고 하세요?

1. 새 아이디어를 구체화하고 싶어요 (PRD + HLD 작성)
   → 코드베이스가 아직 없는 신규 프로젝트

2. 기존 프로젝트에 변경을 가하고 싶어요 (버그/기능/개선)
   → 코드베이스가 이미 있고, 무언가 수정/추가

3. 지금 어디 있는지 확인하고 싶어요
   → 현재 phase 추론 + 다음 권장 command

번호를 선택하거나 자유롭게 의도를 설명해 주세요:
```

### 사용자 응답 처리

| 사용자 응답 | 다음 동작 |
|-----------|---------|
| `1` 또는 "1번" | "어떤 아이디어를 구체화하나요? 한 문장으로 설명해 주세요." 질문 → 응답 받아 `concretize-idea` 호출 |
| `2` 또는 "2번" | "어떤 변경이 필요한가요? (버그/기능/개선 등) 상황을 설명해 주세요." 질문 → 응답 받아 `assess-product-change` 호출 |
| `3` 또는 "3번" | 즉시 `status` 호출 (별도 입력 불필요) |
| 자유 발화 | Mode 1의 의도 분류 규칙 적용 |

---

## 실행 절차

### Step 1. 입력 확인

`$ARGUMENTS` 내용 검사:
- 비어있음 → Mode 2로 진행
- 내용 있음 → Mode 1로 진행

### Step 2 (Mode 1). 의도 분류

위 4가지 분류 중 가장 강한 signal에 매칭되는 카테고리 선택. 분류 결과를 사용자에게 한 줄로 확인:

```
"새 아이디어 구체화"로 이해했어요. `concretize-idea`로 진행합니다.
(아니면 다시 알려주세요)
```

명시적 confirm 없이 즉시 dispatch (사용자가 다시 알려주지 않으면 진행).

### Step 3. Dispatch

`Skill` tool로 target orchestrator 호출. 사용자 발화 전체를 인자로 전달.

---

## 이 스킬을 사용하는 경우

- 처음 buddy를 사용해서 어떤 command가 있는지 모름
- 자기 의도를 자연어로 표현하고 싶음
- "어디서부터 시작할지 모르겠음"
- 시연 / 데모 / onboarding 시점

## 이 스킬을 사용하지 않는 경우

- 이미 어떤 작업인지 명확함 → 직접 해당 command 호출 (예: `/buddy:concretize-idea`)
- 진행 중인 작업 이어가기 → `restore-context` 또는 직접 작업 command

---

## 다음 단계

라우팅된 orchestrator의 흐름을 따른다:
- `concretize-idea` → Phase 1 Mode A 9 stages
- `assess-product-change` → Mode B scope 평가 + Phase 2/3/5 routing
- `status` → 현재 phase 안내
