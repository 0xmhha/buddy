# assess-product-change — 기존 프로덕트 변경 평가

기존 프로덕트에 대한 모든 변경 요청(버그 수정, 신규 기능, 성능 개선, 기술 부채 정리, 의존성 갱신 등)을 받아 **영향 평가 + scope 분류 + 다음 phase 결정**을 수행한다.

`concretize-idea`가 "없는 것을 만든다"면, 본 스킬은 "있는 것을 바꾼다".

**진입 조건**: 기존 프로덕트(코드베이스)가 존재하고, 변경이 필요한 상황.
**산출물**: validated work item + impact assessment + scope classification + routing decision.
**다음 phase**: scope에 따라 §2 / §3 / §5 중 하나로 routing.

---

## Input Requirements

| Input | Required | Type | Source | 미제공 시 |
|-------|----------|------|--------|----------|
| Change trigger | ✅ | knowledge | 사용자 발화 (버그 리포트, 기능 요청, 성능 불만, 기술 부채 인지 등) | "어떤 변경이 필요한가요? 버그, 새 기능, 성능 개선, 기술 부채 등 상황을 설명해 주세요." |
| Existing codebase | ✅ | artifact | 현재 작업 디렉토리의 코드 | "이 프로젝트의 코드베이스가 맞나요?" (자동 감지 — 현재 디렉토리) |

## Output Contract

| Output | Type | Format | Consumers |
|--------|------|--------|-----------|
| Validated work item | artifact | structured (문제/기회 설명 + 재현 방법 또는 근거 + 수용 기준) | §2 `define-features`, §3 `design-system`, §5 `build-feature` (scope에 따라) |
| Impact assessment | artifact | structured (영향 범위 + severity/priority + 기존 시스템 호환성) | routing decision 근거 |
| Scope classification | decision | small / medium / large | 다음 phase 결정 |
| Routing decision | decision | 다음 phase 번호 (§2 / §3 / §5 / defer) | 사용자 확인 후 해당 phase 진입 |

---

## 실행 절차

### Step 1. 변경 요청 접수

사용자의 변경 trigger를 접수한다. 유형을 묻지 않는다 — 유형 분류는 Step 2에서 자동으로 처리.

사용자에게 확인할 것:
1. "어떤 변경이 필요한가요? 상황을 설명해 주세요."
2. "이 변경이 필요한 이유(동기)가 있나요?" (비즈니스 요청, 사용자 불만, 기술적 필요 등)

### Step 2. 현재 상태 파악

코드베이스를 빠르게 탐색하여 변경 대상 영역을 파악한다:

1. **변경 대상 식별** — 어떤 파일/모듈/패키지가 영향을 받는지
2. **현재 테스트 상태** — 해당 영역의 테스트 커버리지
3. **의존성 확인** — 변경이 다른 모듈에 미치는 영향 범위
4. **최근 변경 이력** — 해당 영역의 최근 commit/PR (git log)

버그인 경우 추가:
- **재현 시도** — 사용자 설명 기반으로 문제 재현
- **에러 로그/스택트레이스 확인**

### Step 3. 영향 평가 (Impact Assessment)

| 평가 차원 | 질문 | 출력 |
|----------|------|------|
| **영향 범위** | 몇 개의 파일/모듈이 변경되는가? | 파일 수 + 모듈 수 |
| **사용자 영향** | 사용자가 체감하는 변화인가? | direct / indirect / none |
| **호환성 리스크** | 기존 API/데이터/동작이 깨지는가? | breaking / non-breaking |
| **테스트 영향** | 기존 테스트가 수정되어야 하는가? | yes / no / new tests only |
| **가역성** | 문제 발생 시 쉽게 롤백 가능한가? | easy / hard / irreversible |

### Step 4. Scope 분류

영향 평가 결과를 종합하여 scope를 분류한다:

| Scope | 기준 | 예시 |
|-------|------|------|
| **Small** | 1-3 파일 변경, non-breaking, 기존 테스트 유지, 설계 변경 없음 | 버그 수정, 설정 변경, 오타 수정, 작은 UI 수정 |
| **Medium** | 4-15 파일 변경, 또는 새 모듈/API 추가, 또는 스키마 변경, 설계 검토 필요 | 새 API endpoint, 컴포넌트 리팩토링, DB 스키마 변경, 미들웨어 추가 |
| **Large** | 15+ 파일 변경, 또는 아키텍처 변경, 또는 새 actor/use case 도입, feature 정의 필요 | 신규 기능, 대규모 재설계, 새 서비스 추가, 인증 체계 변경 |

### Step 5. 결과 출력 + Routing 제안

아래 형식으로 출력한다:

```
## 변경 평가 결과

### 작업 항목
- 설명: {문제/기회 설명}
- 동기: {비즈니스/기술적 근거}
- 수용 기준: {완료 조건}

### 영향 평가
- 영향 범위: {파일 수} 파일, {모듈 수} 모듈
- 사용자 영향: {direct / indirect / none}
- 호환성: {breaking / non-breaking}
- 가역성: {easy / hard / irreversible}

### Scope 분류: {Small / Medium / Large}

### 다음 단계 제안
→ {다음 phase 번호 + 스킬명 + 이유}
```

**Routing 결정:**

| Scope | 다음 경로 | 실행할 커맨드 |
|-------|----------|-------------|
| Small | → §5 Development | `/buddy:build-feature` 또는 `/buddy:diagnose-bug` |
| Medium | → §3 Technical Design | `/buddy:design-system` (설계 검토 후 구현) |
| Large | → §2 Feature Definition | `/buddy:define-features` (feature 정의부터) |
| Defer | → backlog 기록 | 현재 cycle 종료. 우선순위 재평가 시점 안내 |

사용자에게 routing 제안을 확인받은 후 해당 phase로 진입한다. 사용자가 다른 경로를 원하면 그 판단을 따른다 (User Sovereignty).

---

## 이 스킬을 사용하는 경우

- "이 버그 좀 봐줘" / "이거 고쳐야 해"
- "이 기능 추가하고 싶어" / "여기에 X를 넣으면 좋겠어"
- "성능이 느려졌어" / "이 부분 최적화해야 해"
- "의존성 업데이트해야 해" / "기술 부채 정리하자"
- "리팩토링 하고 싶어" / "이 코드 구조를 바꾸자"
- 기존 프로덕트에 대한 모든 변경 요청

## 이 스킬을 사용하지 않는 경우

- 아이디어만 있고 코드베이스가 없는 경우 → `concretize-idea`
- 이미 scope와 설계가 결정되어 바로 구현에 들어가는 경우 → `build-feature`
- Hotfix로 즉시 수정이 필요한 경우 → `diagnose-bug` 직접 호출
