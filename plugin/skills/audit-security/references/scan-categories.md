# Audit Security — 8 스캔 카테고리 상세

> 본 파일은 `audit-security` 스킬의 detailed reference입니다. 메인 SKILL.md에서 참조하며, on-demand 로드됩니다.

## 1. Secrets Archaeology

실제 세계 #1 vector. Leak된 자격증명에 대해 git 히스토리 (현재 상태만 아님) 스캔. 대부분 leak이 첫 commit에 발생하고 영원히 남음.

**스캔 패턴** (ripgrep / git log / `git rev-list` 사용):

| Provider | 패턴 | Source 위치 |
|----------|---------|----------------|
| AWS | `AKIA[0-9A-Z]{16}` | `*.env`, `*.yml`, `*.json`, `*.tf` |
| OpenAI / Anthropic | `sk-[A-Za-z0-9]{20,}`, `sk-ant-[A-Za-z0-9-]{20,}` | `*.env`, `*.ts`, `*.js`, `*.py` |
| GitHub | `ghp_`, `gho_`, `github_pat_`, `ghs_` | 모든 파일 |
| Slack | `xoxb-`, `xoxp-`, `xapp-` | 모든 파일 |
| Stripe | `sk_live_`, `rk_live_` | 모든 파일 |
| Generic | `password|secret|token|api_key` | Config 파일만 |

**추적된 .env 파일:** `git ls-files`의 `*.env`나 `.env.*` (`.example`, `.sample`, `.template` 제외)가 finding.

**인라인 시크릿 있는 CI config:** `password:`나 `token:`가 `${{ secrets.X }}`나 동등 시크릿 스토어 interpolation 안에 있지 않은 GitHub workflow / GitLab CI / CircleCI config.

**심각도:**
- CRITICAL: git 히스토리에 발견된 활성 키 포맷 (AWS, Stripe live, OpenAI sk-)
- HIGH: git이 추적하는 `.env`, 인라인 자격증명 있는 CI config
- MEDIUM: `.env.example`의 의심 값, 부분 매치

**필터할 false positive:**
- 플레이스홀더: `your_key_here`, `changeme`, `TODO`, `example`
- 테스트 fixture, 같은 값이 비테스트 코드에 나타나지 않는 한
- 이미 rotate된 시크릿도 여전히 flag (한 번 노출됨)

**Diff 모드:** 현재 브랜치의 commit만 스캔 (`git log <base>..HEAD`).

## 2. 의존성 Supply Chain

`npm audit` 넘어섬. 실제 supply-chain 리스크는 install 스크립트, unpinned 버전, 누락 lockfile.

**패키지 매니저 감지:** `package.json` / `Gemfile` / `requirements.txt` / `pyproject.toml` / `Cargo.toml` / `go.mod` / `composer.json`.

**표준 취약점 스캔:** 매니저와 함께 ship되는 audit 도구 실행 (`npm audit`, `pip-audit`, `cargo audit`, `bundle audit`, `govulncheck`). 설치 안 됐으면 "SKIPPED — tool not installed" 주목 + 설치 instruction. 이는 정보 제공, finding 아님.

**Install-script audit (Node/npm 특정):** 프로덕션 의존성의 `preinstall`, `postinstall`, `install` 스크립트 체크. 패키지 설치 시 임의 코드 실행. 악의적 업데이트가 backdoor 드롭 가능.

**Lockfile 무결성:**
- Lockfile 존재 필수
- Lockfile git 추적 필수
- Lockfile이 manifest와 매치 필수 (`npm ci --dry-run` 또는 동등)

**심각도:**
- CRITICAL: 직접 프로덕션 의존성의 known CVE (CVSS ≥ 7.0)
- HIGH: prod dep의 install 스크립트 / 누락 lockfile / lockfile out of sync
- MEDIUM: 버려진 패키지 / 중간 CVE / lockfile 추적 안 됨

**필터할 false positive:**
- devDependency CVE는 MEDIUM에 cap
- `node-gyp`, `cmake`, 네이티브 빌드 install 스크립트는 MEDIUM에 cap
- known active exploit 없는 no-fix-available advisory → 제외
- Lockfile 없는 라이브러리 repo (앱 아님) → 제외 (ecosystem 규범)

## 3. CI/CD 파이프라인 보안

누가 파이프라인 수정 가능? 어떤 시크릿 접근? Fork PR이 exfiltrate 가능?

**GitHub Actions 체크:**
- **Unpinned 서드파티 action:** `@<sha>` 없는 `uses:` (tag-pinning은 불충분 — tag는 가변). First-party `actions/*` unpinned = MEDIUM.
- **`pull_request_target` + PR ref checkout:** CRITICAL. Fork PR이 쓰기 접근 얻고 시크릿 exfiltrate 가능.
- **스크립트 주입:** `run:` 블록 안의 `${{ github.event.*.body }}`, `${{ github.event.issue.title }}`. CRITICAL.
- **마스킹 없이 env var로 시크릿:** `env: TOKEN: ${{ secrets.X }}` 그다음 `run: echo $TOKEN` — 로그에 leak.
- **CODEOWNERS 보호:** `.github/workflows/`가 수정에 codeowner 승인 요구해야.

**GitLab CI / CircleCI:** 인라인 자격증명, 신뢰 안 된 runner, fork-PR run에 동등 체크.

**심각도:**
- CRITICAL: `pull_request_target` + PR-ref checkout, 스크립트 주입
- HIGH: unpinned 서드파티 action, 마스킹 없이 env로 시크릿
- MEDIUM: workflow 파일의 누락 CODEOWNERS

**필터할 false positive:**
- First-party `actions/*` unpinned → HIGH 아니라 MEDIUM
- PR ref checkout **없는** `pull_request_target` → 안전
- `with:`로 전달된 시크릿 (런타임 처리) → 안전

## 4. LLM / AI 보안

새 공격 클래스. 대부분 팀이 아직 internalize 안 함.

**Prompt injection vector:**
- 문자열 interpolation으로 **시스템 prompt**나 **tool schema**로 흐르는 사용자 입력. (노트: user-message 위치의 사용자 입력은 prompt injection 아님 — 의도된 채널.)
- Sanitization 없이 RAG에 공급된 외부 문서 (RAG 포이즈닝)
- 출력 validation 없는 Tool/function-calling

**HTML로 렌더링되는 sanitize 안 된 LLM 출력:**
- React의 unsafe-innerHTML setter ("dangerously..." 계열), Vue의 `v-html`, jQuery `.html()`, 어떤 프레임워크든 raw HTML 주입 helper — sanitization 없이 LLM 출력 공급하면 모두 위험
- Evaluator-style 함수, shell 명령, 또는 function-from-string 생성자로 흐르는 LLM 출력

**API 키 leak:**
- 코드의 `sk-`, `sk-ant-`, 하드코딩 API 키 할당 (env var 아님)
- 클라이언트 번들의 API 키 (브라우저 노출)

**비용 / 리소스 공격:**
- 사용자당 unbounded LLM 호출 (chat endpoint에 rate limit 없음)
- max_tokens cap 없음, 세션당 spend cap 없음
- 깊이 제한 없는 recursive agent 루프

**Tool-calling 권한:**
- 인자 validation 없이 실행되는 LLM tool 호출
- 임의 파일 경로 읽거나 임의 명령 실행하는 tool

**심각도:**
- CRITICAL: 시스템 prompt의 사용자 입력, HTML/코드로서 sanitize 안 된 LLM 출력, LLM 출력의 evaluator-style 실행
- HIGH: 누락 tool-call validation, 노출 AI API 키, unbounded 비용 증폭
- MEDIUM: 입력 validation 없는 RAG, 사용자별 rate limit 누락

**필터할 false positive:**
- User-message 위치의 사용자 content → 주입 아님
- 프레임워크 기본 text 렌더링(React, Vue text binding)으로 렌더링된 LLM 출력 → 안전
- 하드코딩 API 키 있는 테스트 fixture → finding 아님

**"DoS finding은 보안 아님"의 중요 예외:** unbounded LLM 비용은 denial of service 아님, **재정 리스크**. 비용 finding auto-discard 금지.

## 5. OWASP Top 10

각 카테고리에 targeted 분석 실행. 감지된 스택에 파일 확장자 scope.

**A01 — Broken Access Control:**
- 라우트에 누락 auth (`skip_before_action`, 데코레이터 없는 `@app.route`, auth 미들웨어 없는 bare `app.get`)
- 직접 object 참조: 사용자 A가 ID 변경으로 사용자 B의 리소스 읽기 가능?
- 수평/수직 권한 상승 경로

**A02 — 암호화 실패:**
- 약한 암호화: MD5, SHA1, DES, ECB, RC4
- 하드코딩 키, 하드코딩 IV, 예측 가능 nonce
- 저장 또는 전송 중 암호화 안 된 민감 데이터

**A03 — Injection:**
- SQL: raw 쿼리, SQL의 문자열 interpolation, unparameterized ORM escape hatch
- Command: 사용자 입력 있는 shell-out helper (`system`, `exec`, `spawn`, `popen`)
- Template: 사용자 제어 template 렌더링, evaluator helper, `html_safe`, `raw()`
- LLM prompt injection: 카테고리 4에서 커버

**A04 — Insecure Design:**
- Auth endpoint(login, 비밀번호 리셋, signup)에 rate limit?
- N 실패 시도 후 계정 lockout?
- 서버에 비즈니스 로직 validation (클라이언트만 아님)?

**A05 — 보안 Misconfiguration:**
- CORS: 프로덕션의 wildcard origin?
- CSP 헤더 존재하고 restrictive?
- 프로덕션 응답의 debug 모드 / verbose 스택 트레이스?

**A06 — 취약 컴포넌트:** 카테고리 2 참조.

**A07 — 식별과 인증 실패:**
- 세션 생성, 저장 (HttpOnly, Secure, SameSite), 무효화
- 비밀번호 정책: 복잡성, HIBP로 breach 체크
- MFA: 사용 가능? admin에 강제?
- JWT: 만료 ≤ 24h, 서명 알고리즘 고정 (`alg: none` 없음)

**A08 — 소프트웨어와 데이터 무결성 실패:**
- 신뢰 안 된 입력의 deserialization (Python binary deserializer, Java ObjectInputStream, unsafe YAML loader) — 이들에 네트워크 입력 절대 공급 금지
- 외부 데이터에 무결성 체크 (서명, 체크섬)
- 파이프라인 무결성은 카테고리 3 참조

**A09 — 보안 로깅과 모니터링 실패:**
- 인증 실패 로그?
- 인가 거부 로그?
- Audit trail의 admin action?
- 변조에서 보호된 로그 (append-only, 별도 저장소)?

**A10 — SSRF:**
- 사용자 입력에서 URL 구성
- 사용자 제어 호스트로 outbound 요청
- Outbound HTTP allowlist
- 사설 IP 범위 block (`169.254.*`, `10.*`, `127.*`, `localhost`)

## 6. STRIDE 위협 모델

아키텍처 pass 중 식별된 각 주요 컴포넌트에 대해 STRIDE matrix walk:

```
COMPONENT: [Name]
  Spoofing:               Can an attacker impersonate a user / service?
  Tampering:              Can data be modified in transit / at rest?
  Repudiation:            Can actions be denied? Is there an audit trail?
  Information Disclosure: Can sensitive data leak?
  Denial of Service:      Can the component be overwhelmed?
  Elevation of Privilege: Can a user gain unauthorized access?
```

체크리스트 아니라 reasoning 연습. 출력은 "yes/no" box 테이블이 아니라 컴포넌트당 간단한 위협 narrative.

STRIDE에서 finding은 구체적이고 악용 가능할 때만 주 리포트에 공급. 경로 없는 이론적 "spoof 가능"은 finding 아님.

## 7. 웹훅 & 통합 Audit

웹훅 receiver는 종종 뭐든 수락하는 inbound endpoint. 흔한 miss: 서명 검증 없는 receiver.

**웹훅 route:** 웹훅 / hook / callback route 패턴 매치 파일 찾기. 각각, 같은 파일 (또는 그 미들웨어 체인)에 서명 검증 로직: `signature`, `hmac`, `verify`, `digest`, `x-hub-signature`, `stripe-signature`, `svix` 포함 여부 체크.

웹훅 route 있지만 미들웨어 체인 어디에도 서명 검증 없는 파일 → finding.

**TLS 검증 비활성:** `verify=false`, `VERIFY_NONE`, `InsecureSkipVerify`, `NODE_TLS_REJECT_UNAUTHORIZED=0`, `rejectUnauthorized: false` 같은 패턴.

**OAuth scope 분석:** 과도하게 광범위 scope 요청하는 OAuth config (예: `public_repo`만 필요할 때 GitHub `repo`, `drive.file`만 필요할 때 Google `drive`).

**심각도:**
- CRITICAL: 서명 검증 전혀 없는 웹훅 receiver
- HIGH: 비테스트 코드의 TLS 검증 비활성, 과도 광범위 OAuth scope
- MEDIUM: 서드파티로의 문서화 안 된 outbound 데이터 흐름

**코드 트레이싱만:** 미들웨어 체인 읽어 검증. 웹훅 endpoint에 live HTTP 요청 금지.

## 8. 인프라 Shadow Surface

과도 접근 있는 shadow 인프라 찾기. 아무도 소유 안 하는 것이 pwned되는 것.

**Dockerfile:**
- 누락 `USER` 지시 → root로 실행
- `ARG`로 전달된 시크릿 (이미지 레이어에 baked)
- 이미지에 복사된 `.env` 파일 (`COPY .env`, `ADD .env`)
- 문서화된 목적 없는 노출 포트

**Prod 자격증명 있는 config 파일:**
- Committed config의 데이터베이스 연결 문자열 (`postgres://`, `mysql://`, `mongodb://`, `redis://`), localhost / 127.0.0.1 / example.com 제외
- Prod 참조하는 staging/dev config (예: prod DB host 있는 `staging.config.js`)

**IaC 보안:**
- Terraform: IAM `actions`나 `resources`의 `"*"`, `.tf` / `.tfvars`의 하드코딩 시크릿
- Kubernetes: 프로덕션 manifest의 `privileged: true`, `hostNetwork: true`, `hostPID: true`

**심각도:**
- CRITICAL: committed config의 자격증명 있는 prod DB URL, 민감 리소스의 `"*"` IAM, Docker 이미지 레이어에 baked 시크릿
- HIGH: prod의 root 컨테이너, prod DB용 구성된 staging, privileged K8s
- MEDIUM: 누락 `USER` 지시, 문서화 안 된 노출 포트

**필터할 false positive:**
- localhost 있는 로컬 dev용 `docker-compose.yml` → finding 아님
- Terraform `data` 블록의 `"*"` (read-only) → 안전
- `test/` / `dev/` / `local/`의 K8s manifest → finding 아님
