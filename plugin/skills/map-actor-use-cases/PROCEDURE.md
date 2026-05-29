# map-actor-use-cases

actor별 use case를 식별한다. 2단계 `define-features`의 두 번째 stage.

UML use case diagram 등가 작업. actor 시점에서 시스템과의 상호작용을 동사+목적어 형태로 나열한다.

## Input Requirements

| Input | Required | Type | Source | 미제공 시 |
|-------|----------|------|--------|----------|
| Actor list | ✅ | artifact | `identify-actors` 산출물 | "시스템에 참여하는 actor(사용자/시스템/외부 서비스)를 알려주세요." |
| 제품/시스템 맥락 | ✅ | knowledge | PRD 또는 사용자 설명 | "이 시스템이 어떤 문제를 해결하나요?" |

## Output Contract

| Output | Type | Format | Consumers |
|--------|------|--------|-----------|
| Actor-use case map (actor별 use case 목록) | artifact | structured YAML | `map-use-case-to-system-boundary`, `compose-feature-from-use-cases`, `decompose-feature-to-actor-tracks` |

---

## Use Case 작성 원칙

1. **actor 시점**: "{actor}가 {동사} + {목적어}" 형태로 작성
2. **독립성**: 각 use case는 actor 단독으로 식별 가능한 상호작용
3. **측정 가능**: use case는 6단계 QA에서 테스트 가능해야 함
4. **현재 scope**: 8단계의 iteration에서 발견된 use case는 2단계 재진입 시 추가
5. **LOGICAL 수준**: 어느 product가 처리하는지는 명시 X — 그건 HLD(`write-hld`) §5에서 매핑. 본 스킬은 "누가/무엇/데이터"만 담당
6. **데이터 흐름 명시**: 각 use case는 actors 간 데이터 in/out을 포함 — service 제공 내용과 함께

---

## 매핑 절차

### 1. `identify-actors` 결과 입력

`identify-actors`의 출력(actor 목록)을 입력으로 받는다.

### 2. Actor별 Use Case 열거 (LOGICAL — 데이터 흐름 포함)

각 actor에 대해 시스템과의 상호작용 + 데이터 흐름을 나열한다.

```yaml
actor_use_cases:
  - actor_id: anonymous-visitor
    use_cases:
      - id: uv-001
        title: "회원가입 페이지에 접근한다"
        trigger: "landing page CTA 클릭"
        actors_involved: [anonymous-visitor, system]
        interactions:
          - from: anonymous-visitor
            to: system
            data: "URL navigation"
          - from: system
            to: anonymous-visitor
            data: "회원가입 폼 HTML/UI"
        service_provided: "회원가입 진입점 제공"
        outcome: "회원가입 폼 표시"

  - actor_id: authenticated-user
    use_cases:
      - id: au-001
        title: "이메일/비밀번호로 회원가입한다"
        trigger: "회원가입 폼 제출"
        actors_involved: [authenticated-user, system, sendgrid]
        interactions:
          - from: authenticated-user
            to: system
            data: "email, password, (optional) profile fields"
          - from: system
            to: sendgrid
            data: "verification email request (recipient, link)"
          - from: sendgrid
            to: authenticated-user
            data: "verification email"
          - from: authenticated-user
            to: system
            data: "verification link click (token)"
          - from: system
            to: authenticated-user
            data: "계정 활성화 확인 + JWT 발급"
        service_provided: "신규 계정 생성 + 이메일 검증 + 첫 인증 토큰 발급"
        outcome: "계정 생성 + verification email 수신 + 활성화 완료"
        
      - id: au-002
        title: "이메일/비밀번호로 로그인한다"
        trigger: "로그인 폼 제출"
        outcome: "JWT 발급 + dashboard redirect"
        
      - id: au-003
        title: "비밀번호를 재설정한다"
        trigger: "forgot password 링크 클릭"
        outcome: "reset email 수신 + 새 비밀번호 설정"
        
  - actor_id: auth-service
    use_cases:
      - id: as-001
        title: "이메일 형식과 비밀번호 요구사항을 검증한다"
        trigger: "POST /auth/signup"
        outcome: "validation result + 400/200"
        
      - id: as-002
        title: "비밀번호를 해시 저장한다"
        trigger: "validation 통과 후"
        outcome: "bcrypt hash → DB users table"
        
      - id: as-003
        title: "JWT를 발급한다"
        trigger: "user record 생성 후"
        outcome: "signed JWT (exp: 24h)"
        
  - actor_id: sendgrid
    use_cases:
      - id: sg-001
        title: "이메일 인증 메시지를 발송한다"
        trigger: "auth-service → sendgrid API call"
        outcome: "verification email delivered"
        
      - id: sg-002
        title: "인증 링크 클릭을 webhook으로 처리한다"
        trigger: "사용자 링크 클릭"
        outcome: "POST /webhooks/sendgrid → account verified"
```

### 3. Use Case 검증 체크

- [ ] 각 use case가 단일 actor의 단일 상호작용인가?
- [ ] Trigger와 outcome이 명확한가?
- [ ] 6단계 QA에서 테스트 가능한가?
- [ ] actor 간 중복이 없는가? (use case는 한 actor 소속)

---

## 출력 형식 (LOGICAL — physical product 매핑은 HLD §5에서)

```yaml
actor_use_cases:
  - actor_id: {id}
    actor_type: {user/system/3rd-party/external-tool}
    use_cases:
      - id: {actor_prefix}-{N}
        title: "{동사 + 목적어}"
        trigger: "{무엇이 이 use case를 시작하는가}"
        actors_involved: [{actor_ids}]              # 이 use case에 등장하는 모든 actor
        interactions:                                # actor 간 데이터 흐름 (logical)
          - from: {actor_id}
            to: {actor_id}
            data: "{전달 데이터의 의미적 설명}"
        service_provided: "{이 use case가 제공하는 서비스 한 문장}"
        outcome: "{성공 시 결과}"

total_use_cases: {N}
```

**참고**: `interactions`의 data는 의미적 설명(예: "credentials") 수준. 실제 형식(JSON/protobuf/SQL)과 product 매핑은 `write-hld` §5 (Use Case → Product Mapping)에서 명시.

---

## 다음 단계

→ `map-use-case-to-system-boundary` — use case별 시스템 경계 매핑
