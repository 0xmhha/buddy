# Design Secret Management — secret store + rotation + audit + leak detection

## 1. 목적

API key / DB credential / OAuth token / 인증서 등 *secret* 의 **저장 (where) + 회전 (how often) + 접근 audit (who) + 누수 탐지 (leak detection)** 4 영역을 통합 설계.

`define-tech-stack` (hosting) + `design-observability` (audit log) + `audit-security` 와 cascade.

## 2. 사용 시점

- §3 design-system 안에서 tech stack 결정 후
- 외부 API 통합 시점 (3rd-party key 추가)
- 인증 / 인가 model 결정 (`design-auth-model` 구현됨) 직전 또는 안에서
- secret leak incident 사후
- 정기 audit (분기 / 반기) — rotation 주기 검증

## 3. 입력

### 필수
- `define-tech-stack` — hosting (cloud / on-prem) + secret manager 후보
- 사용 secret 목록 (API key / DB / OAuth / TLS cert)
- 접근 actor (dev / CI / runtime / admin)

### 선택
- 컴플라이언스 요구 (SOC2 / ISO / PCI-DSS) — rotation 주기 강제

## 4. Stage 흐름

### Stage 1: Secret store 결정

> **AI 편향 차단 게이트** — 본 Stage 시작 전 *3개 이상의 orthogonal secret store 후보*(예: cloud KMS / HashiCorp Vault / k8s Secret + sealed-secrets / managed identity 기반 keyless / 자체 운영 PKI) 발산. 직접 또는 `verify-best-alternative` 호출. *첫 답 commit 금지*. 산출물에 3+ 후보의 *(rotation 비용·access audit·multi-cloud 지원·운영 복잡도·compliance)* rubric 비교 존재해야 §6 검증 통과. 적용 근거: [`docs/superpowers/specs/2026-05-21-engineering-decision-gate-mapping.md`](../../../../docs/superpowers/specs/2026-05-21-engineering-decision-gate-mapping.md) §3.

| 옵션 | 적합 시점 | 단점 |
|------|--------|------|
| Cloud KMS (AWS Secrets Manager / GCP Secret Manager / Azure Key Vault) | cloud-native, 단일 cloud | vendor lock-in |
| HashiCorp Vault | multi-cloud / on-prem | 운영 복잡 |
| .env file + sealed-secrets (k8s) | 작은 규모 | rotation 자동화 없음 |
| 1Password / Bitwarden teams | dev secret | runtime 비호환 |

→ runtime secret + dev secret 분리. *코드 / repo 에 plaintext secret 절대 금지*.

### Stage 2: Rotation 정책

| secret 종류 | 권장 주기 |
|----------|--------|
| OAuth client secret | 90일 |
| DB password | 30일 (자동) / 90일 (수동) |
| API key (외부 SaaS) | 180일 (compromise 의심 시 즉시) |
| TLS certificate | 90일 (Let's Encrypt 자동) |

각 secret 에 *expiry alert* 설정 (rotation 30일 전 경고).

### Stage 3: Access audit

| 차원 | 기록 |
|------|------|
| who | actor (user / service / CI) |
| when | timestamp |
| what | secret name (값은 X) |
| why | request context (PR # / ticket #) |

→ `design-observability` 의 audit log pillar 와 integrate.

### Stage 4: Leak detection

#### 4.1 Pre-commit
- `gitleaks` / `trufflehog` git pre-commit hook
- `.gitignore` + secret pattern allowlist

#### 4.2 Repo scan
- 매주 historical commit scan
- public repo 시 *주기 X 시간* 으로 scan

#### 4.3 Runtime
- log redaction (`design-observability` Stage 5 정합)
- error message 에 secret 노출 차단

#### 4.4 Public exposure
- pastebin / GitHub gist / S3 bucket public 검사 (외부 도구)

### Stage 5: Incident response

leak 발견 시 *5 분 내* 절차:
1. 해당 secret *즉시 invalidate*
2. 새 secret 발급 + 배포
3. 사용 history audit (어디서 사용됐나)
4. 외부 공지 필요 여부 판단 (privacy / legal)

## 5. 산출물 형식

```markdown
## Secret Management Design — {제품}

### Secret store
- Runtime: {tool}
- Dev: {tool}

### Rotation 표
| secret | 주기 | 자동 |

### Access audit
- log 위치: ...
- 보존: ...

### Leak detection 4 layer
- pre-commit / repo scan / runtime / public

### Incident response 5 step
1~5
```

## 6. 검증

- [ ] runtime + dev secret store *분리*?
- [ ] 모든 secret 에 rotation 주기?
- [ ] Access audit *who/when/what/why* 4 차원 모두 기록?
- [ ] Leak detection 4 layer 모두 활성화?
- [ ] Incident response 5 step 명시?
- [ ] *코드 / repo 에 plaintext secret 0건* (gitleaks 검증)?
- [ ] **`verify-best-alternative` 1회 이상 호출 완료** — AI 편향 방지 의무. secret store / rotation 주기 / detection 전략이 *첫 답*이 아니라 다관점 검토 후 최선임을 확인
- [ ] **3+ orthogonal secret store 후보의 rubric 비교 표가 산출물에 존재** — 체크박스만 체크하는 *anti-rationalization 회피* 금지. Stage 1의 store 옵션 표로 증명

## 7. 다음 phase

- `design-auth-model` (구현됨) 와 정합
- `audit-security` 의 secret 영역 입력
- `design-observability` 의 audit log integrate

## 8. 참조

- OWASP Cryptographic Storage Cheat Sheet
- 12-Factor App (config / secret 분리)
- gitleaks / trufflehog (leak detection 도구)
