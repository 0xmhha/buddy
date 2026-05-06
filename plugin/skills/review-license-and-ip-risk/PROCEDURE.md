---
name: review-license-and-ip-risk
description: "프로젝트의 모든 의존성/asset/AI 생성 코드에 대한 라이선스 호환성, IP 출처, 상업 사용 가능성을 검토하고 risk register와 remediation list를 산출. 트리거: '라이선스 검토' / '상업 사용 가능해?' / 'copyleft 충돌 없어?' / 'AGPL 위험 있어?' / 'SBOM 만들어줘' / 'MIT랑 GPL 섞어 써도 돼?' / 'AI 생성 코드 라이선스 어떻게?'. 입력: 의존성 목록, asset, AI 생성 코드 인벤토리. 출력: SBOM + license risk register + remediation. 흐름: assess-business-viability/audit-security → review-license-and-ip-risk → define-product-spec/write-changelog."
type: skill
---

# Review License & IP Risk — 라이선스/IP 7단계 검증

## 1. 목적

audit-security가 보안 취약점을 본다면, 이 스킬은 **라이선스 호환성, copyleft 전염, IP 출처, 상업 사용 가능성, attribution 의무**를 본다. 보안과 라이선스는 서로 다른 도메인 — 보안 통과해도 라이선스 위반 시 상용 출시 불가.

이 스킬은 Stage 9 (Legal/Compliance) 게이트로 동작한다. 코드가 작동해도, 보안이 통과해도, 라이선스 충돌 1건이면 출시 막힌다. 특히 SaaS 모델에서 AGPL 의존성 1개는 전체 백엔드 source 공개 의무를 트리거할 수 있고, 인수합병 due diligence 단계에서 unknown license 1건은 거래 자체를 무산시킬 수 있다.

핵심 가치 제안: **법무팀 없이도 상업 출시 전 IP/라이선스 위험을 자체 진단**할 수 있는 게이트. 단, 이 스킬은 법률 자문이 아니다. 고위험 항목은 변호사 검토 권장.

## 2. 사용 시점

- assess-business-viability에서 commercial use 결정 후 (수익 모델 확정 시점)
- 새 의존성 추가 시 (npm install, go get, pip install 등) — 자동 호출
- AGPL/GPL 의존성 발견 시 SaaS 영향 검토
- AI 생성 코드(Claude/Copilot/Cursor)가 코드베이스에 누적된 시점 (분기별 권장)
- 상용 출시 전 (write-changelog 직전 게이트)
- 인수합병/투자 due diligence 응답 자료 작성 시
- third-party asset(폰트/아이콘/이미지/오디오) 추가 시
- distribution 모델 변경 시 (on-prem → SaaS 전환 등)
- 오픈소스 ↔ 상용 dual license 전환 검토 시

자동 트리거 키워드: "라이선스", "license", "GPL", "MIT", "Apache", "copyleft", "SBOM", "상업 사용", "commercial use", "attribution", "NOTICE", "patent", "trademark", "due diligence".

## 3. 입력

필수:
- package.json / pyproject.toml / go.mod / Cargo.toml / Gemfile / pom.xml 등 manifest 파일
- 프로젝트 자체 라이선스 (MIT / Apache-2.0 / proprietary / dual / unlicensed)
- third-party asset 인벤토리 (fonts, icons, images, audio, video, 3D 모델 등)
- AI 생성 코드 식별 표시 (있으면 — git history, 주석, author tag)

선택:
- 기존 SBOM (SPDX-2.3, CycloneDX-1.5 format)
- 라이선스 정책 문서 (organization-level approved/banned list)
- distribution 모델 (SaaS / on-prem / embedded / mobile-app / hybrid)
- 매출 모델 (free / freemium / subscription / one-time / open-core)
- 회사 IP 정책 (특허 출원 계획, trademark 등록 현황)

입력 부족 시 forcing question:
- "이 프로젝트 자체 라이선스가 뭐야? proprietary면 의존성 라이선스도 강하게 봐야 함."
- "SaaS로 호스팅해? on-prem 배포해? AGPL 영향 다름."
- "AI 도구로 생성한 코드 비율 알아? Copilot/Claude/Cursor — 학습 데이터 라이선스 분쟁 가능."
- "third-party 폰트/아이콘 라이선스 추적 중이야? Google Fonts? Font Awesome Pro?"
- "transitive dependency 검사 도구 돌리고 있어? `npm ls --all` 또는 `pip-licenses` 등."
- "회사 차원 banned license 리스트 있어? 보통 GPL-3, AGPL은 banned."
- "이 프로젝트 매각하거나 투자 받을 계획? 그러면 IP clean 필수."

## 4. 핵심 원칙 (Principles)

1. **MIT/Apache-2.0/BSD ≠ 자유 사용** — attribution 의무는 항상 있음. NOTICE 파일 또는 README 명시. 빠뜨리면 라이선스 위반.

2. **GPL은 copyleft 전염** — 정적 링크 시 파생작 전체 GPL. 동적 링크는 해석 분쟁 가능. LGPL은 더 약하지만 modification은 여전히 공개 의무.

3. **SaaS는 AGPL 트리거** — 네트워크 사용도 source 공개 의무. SaaS 모델에 AGPL 의존성 = 전체 백엔드 source 공개 위험. MongoDB가 AGPL → SSPL로 바꾼 이유.

4. **dual-license는 어느 라이선스 채택하는지 명시** — Sentry, MongoDB, Elastic 같은 OSS + commercial 모델. AGPL 옵션 vs Commercial 옵션 — 어느 것을 따르는지 명문화.

5. **코드 출처 모르면 IP 폭탄** — SBOM(SPDX/CycloneDX)으로 추적. unknown license 항목은 critical risk. 인수 due diligence에서 deal-breaker.

6. **AI 생성 코드 출처 별도 기록** — Copilot/Claude/Cursor — 학습 데이터 라이선스 분쟁 가능 (GitHub Copilot 집단소송 사례). 출처 표시 + 검토 흔적 보관.

7. **patent grant 없는 라이선스(BSD-2, MIT 등)는 특허 무방비** — Apache-2.0의 patent grant 조항 활용 권장. 특허 보유 회사 의존성은 특히 주의.

8. **third-party asset도 동일 강도** — 폰트(Google Fonts? OFL? 상용 폰트?), 아이콘(Heroicons MIT? Font Awesome Pro?), 이미지(Unsplash 라이선스? 스톡 사이트?), 오디오(CC-BY? Royalty-free?). 디자이너가 가져온 자료라고 면책되지 않음.

9. **transitive dependency가 진짜 위험** — direct는 100개여도 transitive는 2000개+. `npm ls --all` 또는 `pip-licenses --with-license-file`로 전수 추출.

10. **라이선스는 시점별로 달라질 수 있음** — 의존성이 v2.x에서 MIT였다가 v3.x에서 BSL로 바뀌는 사례 (Terraform, Redis, Sentry). 버전 고정 + 업그레이드 시 재검토.

## 5. 단계 (Phases)

### Phase 1. Inventory (전수 식별)

1. manifest 파일에서 모든 의존성 추출 (transitive 포함)
   - JS/TS: `npm ls --all --json` 또는 `pnpm licenses list`
   - Python: `pip-licenses --format=json --with-license-file`
   - Go: `go mod download && go list -m -json all`
   - Rust: `cargo license --json`
   - Java: `mvn license:add-third-party`
2. third-party asset 인벤토리
   - fonts: `find ./assets -name "*.woff2" -o -name "*.ttf"` + 라이선스 파일 동행
   - icons: 디자인 시스템 라이선스 명시 (Heroicons, Lucide, Font Awesome 등)
   - images: 스톡 사이트 출처 + 라이선스 영수증
   - audio/video: BGM, sound effect 출처
3. AI 생성 코드 영역 식별
   - git blame + author tag (`AI-generated: claude-3.5` 같은 주석)
   - PR description에 AI 도구 사용 명시
   - 비율 추정 (전체 LoC 대비 AI 비율)
4. 외부 코드 복붙 영역
   - Stack Overflow, GitHub gist, 블로그 코드 — 출처 주석 필수
   - 레퍼런스 구현(reference implementation) 차용 — 라이선스 확인

### Phase 2. License Classification

각 항목을 SPDX identifier로 분류:
- `permissive`: MIT, Apache-2.0, BSD-2-Clause, BSD-3-Clause, ISC, 0BSD
- `weak-copyleft`: LGPL-2.1, LGPL-3.0, MPL-2.0, CDDL-1.0, EPL-2.0
- `strong-copyleft`: GPL-2.0, GPL-3.0, AGPL-3.0
- `proprietary`: 상업 라이선스 (Sentry Business, MongoDB Atlas Commercial 등)
- `source-available`: BSL (Business Source License), SSPL, Elastic License — OSI 미인증
- `unknown`: 라이선스 명시 없음 (critical risk — 사용 자체 불법 가능)
- `public-domain`: CC0, Unlicense, WTFPL
- `creative-commons`: CC-BY, CC-BY-SA, CC-BY-NC (NC는 상업 사용 금지)

분류 시 자동 도구:
- license-checker (npm)
- ScanCode Toolkit
- FOSSA, Snyk, GitHub Dependency Graph
- Tern (container image scanning)

### Phase 3. Compatibility Matrix

프로젝트 라이선스와 의존성 라이선스 충돌 검사:
- proprietary 프로젝트 + GPL 의존성 = 충돌 (소스 공개 의무)
- MIT 프로젝트 + AGPL 의존성 = SaaS 시 모두 AGPL 영향
- Apache-2.0 + GPL-2-only = 비호환 (Apache-2.0의 patent retaliation 조항 vs GPL-2 추가 제약 금지)
- Apache-2.0 + GPL-3 = 호환 (FSF가 명시적으로 호환 인정)
- MPL-2.0 + 다른 라이선스 = 파일 단위 분리 가능 (file-level copyleft)
- CC-BY-NC + 상업 사용 = 충돌 (NonCommercial)

dual-license 케이스:
- Sentry: BSL → 4년 후 Apache-2.0 (시간 지연 OSS)
- MongoDB driver: Apache-2.0 (driver) + SSPL (server)
- Qt: LGPL-3 (OSS) + Commercial (proprietary 빌드용)
- 어느 옵션 선택하는지 명시 + 라이선스 영수증 보관

호환성 표 (자주 쓰이는 조합):

| 프로젝트 \ 의존성 | MIT | Apache-2.0 | LGPL-2.1 | GPL-3 | AGPL-3 | BSL |
|---|---|---|---|---|---|---|
| Proprietary | OK | OK | OK (dynamic link) | 충돌 | 충돌 (SaaS) | 조건부 |
| MIT | OK | OK | OK | 결과물 GPL-3 | 결과물 AGPL-3 | 충돌 |
| Apache-2.0 | OK | OK | OK | 결과물 GPL-3 | 결과물 AGPL-3 | 충돌 |
| GPL-3 | OK | OK | OK | OK | OK | 충돌 |

### Phase 4. Commercial-Use Gate

각 의존성에 대해:
- commercial use allowed? (라이선스 본문 명시 — "commercial" 키워드 확인)
- modification allowed?
- redistribution allowed?
- private use (proprietary fork) allowed?
- patent retaliation 조항 (Apache-2.0의 grant + termination)
- field-of-use restriction (예: 군사 사용 금지 — 일부 ethical license)

블록되는 항목 → 다음 중 하나:
- drop: 의존성 제거 (대안 없으면 기능 포기)
- replace: 다른 라이선스로 교체 (예: GPL → MIT 대안)
- commercial license 구매 (dual-license의 commercial 옵션)
- 사용 방식 변경 (정적 → 동적 링크, in-process → IPC 분리)
- distribution 모델 변경 (SaaS → on-prem만)

### Phase 5. Attribution Requirements

NOTICE 파일 / README / about 페이지 / 앱 설정 화면에 명시 필요한 항목:
- author (copyright holder)
- copyright year (range)
- license text (full or SPDX-License-Identifier)
- modifications (있으면 — Apache-2.0 4(b))
- disclaimer of warranty

자동화:
- license-checker, ScanCode 같은 도구로 NOTICE 자동 생성
- generate-license-file (npm)
- yarn licenses generate-disclaimer
- pip-licenses --format=html

배포물 형태별 attribution 위치:
- web app: about 페이지 + footer link
- desktop app: 메뉴 → About → Open Source Licenses
- mobile app: Settings → About → Acknowledgements
- npm package: NOTICE 파일 + README 링크
- docker image: /usr/share/doc/<package>/copyright 또는 /licenses/

### Phase 6. Patent / Trademark / Copyright 별도 검토

- **patent**:
  - Apache-2.0의 patent grant 활용 권장 (3조: contributor 특허 무상 라이선스)
  - BSD-2/MIT는 patent 무방비 — 특허 보유 회사가 나중에 소송 가능 (비록 가능성 낮음)
  - 회사 자체가 특허 보유 시: 의존성 사용이 자사 특허와 cross-license 영향 검토

- **trademark**:
  - brand 충돌 (예: 프로젝트명에 "Linux", "Apple", "Docker" 같은 trademark 단어 — 해당 회사 trademark policy 위반)
  - 의존성의 trademark 사용 제한 (예: Mozilla trademark는 Firefox 빌드 외 사용 금지)
  - 자사 trademark 등록 — 출시 전 USPTO/KIPO 검색

- **copyright**:
  - AI 생성 코드의 저작권 누구? (사용자? 모델 회사? 학습 데이터 원저자?)
  - US Copyright Office 입장: 인간 저작 부분만 저작권 보호. AI 자동 생성은 PD에 가까움.
  - Copilot 집단소송 (Doe v. GitHub) 등 판례 모니터링
  - 회사 내부 정책 명문화: "AI 도구로 생성한 코드는 인간 검토 후에만 commit"

### Phase 7. Risk Register & Remediation

각 위험 항목 다음 형식으로 기록:
- id (R1, R2, ...)
- severity:
  - critical: 출시 불가 (commercial use 금지된 항목 사용 등)
  - high: 수정 필요 (attribution 누락, copyleft 의존성)
  - medium: 모니터링 (license 변경 가능성, 향후 위험)
  - low: 기록만 (문서화 미흡)
- category: copyleft_contamination | unknown_license | patent_exposure | trademark | attribution_missing | ai_provenance | dual_license_choice | source_available_restriction
- description
- remediation 옵션:
  - drop
  - replace
  - dual-relicense
  - commercial-license-purchase
  - usage-change (정적→동적 링크 등)
  - distribution-change
  - workaround (NOTICE 추가, 사용 영역 격리)

verdict:
- `clear`: 모든 항목 합법 — 출시 게이트 통과
- `needs-remediation`: 수정 후 재검토 — 주로 high/medium
- `blocked`: critical 위험 1개+ — 출시 정지, 즉시 조치

## 6. 출력 템플릿

```yaml
project_license: MIT | Apache-2.0 | proprietary | dual-license | unlicensed
distribution_model: SaaS | on-prem | embedded | mobile-app | hybrid
revenue_model: free | freemium | subscription | one-time | open-core | dual-license

sbom:
  format: SPDX-2.3 | CycloneDX-1.5
  generator: syft | cdxgen | manual
  components:
    - name: "<dep>"
      version: "<ver>"
      license: "<SPDX id>"
      source: npm | pypi | github | manual | ai-generated | stack-overflow
      direct: yes | no
      transitive_depth: 0 | 1 | 2 | ...
      risk: low | medium | high | critical

license_classification:
  permissive: ["<dep>", ...]
  weak_copyleft: ["<dep>", ...]
  strong_copyleft: ["<dep>", ...]
  proprietary: ["<dep>", ...]
  source_available: ["<dep>", ...]
  unknown: ["<dep>", ...]
  creative_commons_nc: ["<dep>", ...]

compatibility_matrix:
  conflicts:
    - dep: "<name>"
      project_license: "<self>"
      dep_license: "<dep>"
      issue: "<설명>"
      remediation: drop | replace | dual-relicense | commercial-license

copyleft_exposure:
  agpl_in_saas:
    detected: yes | no
    components: [...]
    impact: "<설명>"
  gpl_static_link:
    detected: yes | no
    components: [...]
    impact: "<설명>"
  lgpl_modified:
    detected: yes | no
    components: [...]
    impact: "<설명>"

attribution_required:
  - dep: "<name>"
    notice_text: "<원문>"
    placement: NOTICE | README | about-page | settings-screen
    auto_generated: yes | no
    completed: yes | no

patent_grants:
  apache_2_present: ["<dep>"]
  missing_patent_grant: ["<dep>"]  # MIT/BSD-2 등
  patent_retaliation_risk: ["<dep>"]

trademark_concerns:
  - keyword: "<word>"
    risk: high | medium | low
    advice: "<remediation>"

ai_generated:
  - file: "<path>"
    tool: claude | copilot | cursor | manual-prompt
    coverage: "<% of file>"
    human_reviewed: yes | no
    risk_note: "<설명>"

third_party_assets:
  fonts:
    - name: "<font>"
      license: "<id>"
      source: "<URL>"
      receipt: yes | no
  icons: [...]
  images: [...]
  audio: [...]
  video: [...]
  with_licenses: yes | no

commercial_use:
  allowed: yes | conditional | no
  conditions: [...]
  blocked_components: [...]

risks:
  - id: R1
    severity: critical | high | medium | low
    category: copyleft_contamination | unknown_license | patent_exposure | trademark | attribution_missing | ai_provenance | dual_license_choice | source_available_restriction
    component: "<dep>"
    description: "<설명>"
    remediation: "<액션>"
    deadline: "<날짜>"
    owner: "<담당자>"

verdict: clear | needs-remediation | blocked
next_review: "<날짜>"
legal_counsel_required: yes | no
```

## 7. 자매 스킬

- 앞 단계: `assess-business-viability` — 수익 모델 확정 후 호출. `Skill` tool로 invoke.
- 페어: `audit-security` — 보안 별도 검토. 둘 다 통과해야 출시 게이트 OK.
- 다음 단계:
  - verdict가 `clear`면 → `define-product-spec` 또는 `write-changelog`
  - verdict가 `needs-remediation`면 → `triage-work-items` (remediation 백로그 생성)
  - verdict가 `blocked`면 → 즉시 의사결정 회의 + `critique-plan` 호출
- 관련: `review-architecture` (정적/동적 링크 결정 시), `sync-release-docs` (NOTICE 파일 동기화)

## 8. Anti-patterns

1. **package.json만 보고 transitive 의존성 무시** — npm install 시 들어오는 200+ transitive가 실제 위험. 직접 의존성 10개여도 전체 노드는 1500개 가능. SBOM 필수, `npm ls --all` 또는 `lockfile-lint` 활용.

2. **"MIT면 다 OK"** — attribution 의무 누락 시 라이선스 위반. NOTICE 파일 자동 생성 필수. 누락 시 GitHub에서 issue로 incident 발생 사례 다수.

3. **AGPL 의존성을 npm 패키지명으로만 보고 OK 판정** — README/LICENSE 직접 확인. dual-license일 수 있음. 예: MongoDB driver는 Apache-2.0이지만 server는 SSPL.

4. **AI 생성 코드를 자기 저작권으로 가정** — 학습 데이터 라이선스 분쟁 가능. 출처 표시 + 인간 검토 흔적 보관. Copilot 집단소송 등 판례 모니터링 필수.

5. **third-party 폰트/아이콘을 코드와 분리해서 검토** — "디자인은 제가 모르는 영역" — 동일 강도로. 디자이너가 가져온 폰트 1개 라이선스 위반으로 출시 후 takedown 사례 존재.

6. **dual-license 의존성을 OSS 옵션으로만 가정** — commercial 옵션 선택 시 비용 발생. 배포 모델 따라 결정. Sentry, Qt, Elastic 등 명확히 선택 + 영수증 보관.

7. **fundraise 직전에 monkeypatch** — due diligence에서 발각. 처음부터 SBOM 운영 권장. Series A 단계에서 IP audit 통과 못 해 deal 무산 사례 다수.

8. **라이선스 변경 무시** — Terraform이 BSL, Redis가 SSPL+RSAL로 바뀐 사건. 버전 업그레이드 시 라이선스 재검토 필수. Renovate/Dependabot에 license check 통합.

9. **on-prem만 한다고 AGPL OK 판정 후 SaaS 전환** — 비즈니스 모델 변경 시 AGPL 의존성 즉시 critical로 승격. 미리 distribution 모델 변경 시나리오 검토.

10. **OSI 미인증 라이선스(BSL, SSPL, Elastic)를 OSS로 가정** — 이들은 source-available이지 OSS 아님. 상업 경쟁 제한 조항 등 별도 검토 필요.

11. **Stack Overflow 코드 복붙 후 라이선스 무시** — SO 답변은 CC-BY-SA 4.0 (2018년 이후). attribution + ShareAlike 의무. 단순 복붙도 라이선스 트리거.

12. **Creative Commons NC를 상업 프로젝트에 사용** — CC-BY-NC는 NonCommercial. 상업 사용 명백히 금지. 무료 SaaS도 광고 모델이면 NC 위반 가능.
