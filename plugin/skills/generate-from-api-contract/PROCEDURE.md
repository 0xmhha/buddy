# Generate From API Contract — OpenAPI / GraphQL → 클라이언트 / 서버 stub 자동 생성

## 1. 목적

`design-api-contract` 산출 (OpenAPI 3.x / GraphQL SDL / gRPC proto) 을 입력으로 *클라이언트 SDK + 서버 stub + type 정의* 자동 생성. 수동 코딩의 *type drift* 차단.

contract = single source of truth. 변경 시 generator 재실행 = 코드 + 클라이언트 동시 동기화.


## Input Requirements

| Input | Required | Type | Source | 미제공 시 |
|-------|----------|------|--------|----------|
| API contract (OpenAPI / GraphQL / gRPC) | ✅ | artifact | `design-api-contract` 산출물 | 먼저 `/buddy:design-api-contract` 를 실행하세요 |

## Output Contract

| Output | Type | Format | Consumers |
|--------|------|--------|-----------|
| Generated SDK / stub code (AUTO-GENERATED 헤더) | artifact | source files | `build-with-tdd` |

## 2. 사용 시점

- §5 build-feature 안 — API 계약 정의 후 첫 코드 생성 시
- API 변경 시 (endpoint 추가 / 필드 추가 / version 증가) — generator 재실행
- 새 클라이언트 언어 추가 (예: 기존 TS + 신규 Swift) — 같은 contract 에서 SDK 추가
- type drift 의심 시 — generator 결과와 hand-written 코드 비교 audit

## 3. 입력

### 필수
- `design-api-contract` 산출 — OpenAPI / GraphQL / proto file
- target language (frontend + backend + 외부 클라이언트)
- 사용 generator 도구 (또는 결정 입력)

### 선택
- 기존 hand-written client 코드 (있으면 — migration plan)
- contract version 정책 (semver / breaking change rule)

## 4. Stage 흐름

### Stage 1: Generator 도구 결정

| contract format | 권장 generator |
|---------------|----------|
| OpenAPI 3.x | openapi-generator-cli / oazapfts (TS) / openapi-typescript |
| GraphQL SDL | GraphQL Code Generator (TS / Swift / Kotlin) |
| gRPC proto | protoc + grpc plugin (TS / Go / Swift / Java) |
| AsyncAPI (event) | asyncapi-generator |

→ *언어별 maturity* 차이 있음. 결정 시 *generator 출력 코드 quality* + *active maintenance* 검증.

### Stage 2: 출력 분리

| 출력 | 위치 |
|------|------|
| 클라이언트 SDK (TS) | `packages/client-ts/src/generated/` |
| 클라이언트 SDK (Swift / Kotlin) | 각 native repo 의 `Generated/` 디렉토리 |
| 서버 type / handler stub | `internal/api/generated/` |
| 공통 type 정의 | `packages/types/` |

→ *generated 디렉토리는 git track* (재현성). 단 `// AUTO-GENERATED — DO NOT EDIT` 헤더 강제.

### Stage 3: CI 자동화

contract 변경 → generator 재실행 → diff PR:

1. `pre-commit` hook 으로 contract 변경 감지
2. CI 에서 generator 재실행 + 결과 commit
3. PR 검토자가 *contract 변경 + generated diff* 동시 review
4. merge 시 generated 코드 동기화

→ *generated 코드를 hand-edit* 하면 다음 generator run 에서 덮어씌움. *재현성* 우선.

### Stage 4: Hand-written wrapper

generated 코드는 *raw API* 만 cover. 사용자 친화적 wrapper 는 hand-written:

| layer | 역할 |
|------|------|
| Generated (raw) | API call + type | (auto) |
| Wrapper (hand) | retry / cache / rate limit / error normalization | hand-write |
| Application | 비즈니스 로직 | hand-write |

→ wrapper layer 가 *generated 와 application 분리* — generator 변경에 application 영향 minimize.

### Stage 5: Type drift 검증

분기별:
- generated type vs hand-written type 비교
- 부적절한 hand-edit 발견 시 *re-generate*
- contract test (`design-api-contract` 산출) 정합 검증

## 5. 산출물 형식

```markdown
## API Contract Generation — {제품}

### Generator 결정
| contract format | generator | language |

### 출력 위치
| output | path |

### CI 자동화
- contract 변경 trigger: ...
- generated 동기화 PR: ...

### Wrapper layer
- raw → wrapper → app 3 layer 분리

### Type drift 검증 주기
- 분기별 audit
```

## 6. 검증

- [ ] Contract format 별 generator 결정?
- [ ] Generated 디렉토리 git track + AUTO-GENERATED 헤더?
- [ ] CI 자동화 (contract 변경 → re-generate)?
- [ ] Wrapper layer 분리 (generated 직접 호출 X)?
- [ ] Type drift 분기 audit?

## 7. 다음 phase

- §5 build-feature 의 *비즈니스 로직 작성* 단계
- `generate-tests-from-spec` 와 cascade — contract test 의 baseline
- `monitor-regressions` 의 contract drift 감지

## 8. 참조

- OpenAPI Generator (openapi-generator.tech)
- GraphQL Code Generator (the-guild.dev/graphql/codegen)
- gRPC documentation (grpc.io)
- buddy `design-api-contract` (구현됨) — cascade in
