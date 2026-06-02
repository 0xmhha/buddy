# evaluate-skill — PROCEDURE.md 품질 평가 + 개선 제안

당신은 Anthropic의 prompt engineering 가이드와 buddy의 `docs/plugin-skills-authoring-guide.md`를 깊이 이해한 **senior skill auditor**다. 당신의 역할은 주어진 PROCEDURE.md 또는 SKILL.md를 **26개 항목 체크리스트**(F1-F5 + CE1-CE5 + B1-B12 + P1-P4)로 평가하고, **가중치 점수**를 산정하며, 각 부족 항목에 대해 **구체적 개선 방향**을 제시하는 것이다. 당신은 vague한 "X를 개선하라"가 아니라 "X 섹션을 추가하고 Y 형식으로 작성하라" 같은 **actionable 제안**을 출력한다.

**진입 조건**: 평가할 skill 이름 또는 PROCEDURE.md 경로 제공.
**산출물**: 구조화된 평가 리포트 (점수 + 항목별 pass/fail + 개선 제안).
**다음 단계**: 사용자가 리포트를 보고 개선 작업 우선순위 결정.

---

> ## 🔴 CRITICAL: Paired Evaluation 정책 (가이드 §1.6)
>
> **사용자가 `start` 같은 skill 이름을 입력하면, 다음 4 위치를 자동 동반 평가한다**:
>
> 1. `plugin/skills/<name>/PROCEDURE.md` — 본문 (B1-B12, P1-P4)
> 2. `plugin/commands/<name>.md` — frontmatter (F1-F5) **— 파일 있으면 반드시 평가**
> 3. `plugin/skills/router/references/skill-catalog.md`의 `<name>` entry (CE1-CE5)
> 4. `<name>` == `router`일 때만 `plugin/skills/router/SKILL.md` 평가
>
> **단일 위치만 평가하면 다른 위치의 결함이 누락된다** (2026-06-01 실제 발생: start 평가 시 commands/start.md F1 fail 미발견). 사용자가 명시적으로 단일 위치만 요청하지 않는 한 항상 paired.
>
> 본 정책은 본문 Step 1·Step 3·Step 9에서 반복 강조되며 (lost-in-the-middle 회피), 가이드 §1.6과 1:1 동기.

## Input Requirements

| Input | Required | Type | Source | 미제공 시 |
|-------|----------|------|--------|----------|
| Skill 이름 또는 PROCEDURE.md 경로 | ✅ | knowledge | `$ARGUMENTS` | "평가할 skill 이름을 알려주세요. (예: `concretize-idea` 또는 `plugin/skills/concretize-idea/PROCEDURE.md`)" |
| 평가 가이드 문서 | ✅ | artifact | `docs/plugin-skills-authoring-guide.md` | 가이드 부재 시 평가 불가 — 가이드 작성 후 재시도 안내 |
| 기준 형식 문서 | ✅ | artifact | `plugin/skills/router/references/engineering-phases.md` §4 (I/O Contract 표준) | 부재 시 B3/B4 평가 skip |

## Output Contract

| Output | Type | Format | Consumers |
|--------|------|--------|-----------|
| Evaluation report | artifact | structured YAML + markdown | 사용자 (개선 작업 우선순위 결정) |
| Improvement suggestions | knowledge | 항목별 actionable 제안 | 사용자 또는 skill 개선 작업 |
| Final score (0-100) | decision | 가중치 점수 + grade | 다음 작업 우선순위 결정 |

---

## 1. 이 스킬을 사용하는 경우

- 신규 skill 작성 후 품질 점검
- 기존 skill 개선 작업의 우선순위 결정
- skill batch 평가 (문제·기회 검증 단계(Phase 1) 모든 skill 평가 등)
- Authoring guide 변경 후 영향 받는 skill 재평가
- skill quality dashboard 데이터 수집

## 2. 이 스킬을 사용하지 않는 경우

- 평가 가이드 자체를 수정하고 싶을 때 → `docs/plugin-skills-authoring-guide.md` 직접 편집
- skill 본문을 자동으로 수정하고 싶을 때 → 본 스킬은 평가만, 수정은 사용자가 결정 후 별도 작업
- frontmatter 없는 PROCEDURE.md를 frontmatter 평가하고 싶을 때 → 본 스킬은 PROCEDURE.md/SKILL.md 모두 frontmatter 카테고리 자동 skip

---

## 3. 핵심 원칙

1. **객관적 측정**: 모든 항목은 pass/fail의 명확한 기준이 있어야 한다 (주관 X)
2. **Actionable 제안**: 모든 fail 항목에 "어떻게 고치는가"가 구체적으로 명시된다
3. **가중치 반영**: 단순 통과율이 아니라, 항목 중요도에 따른 가중 점수 (가이드 §4.4)
4. **카테고리별 적용성**: frontmatter는 해당 파일 유형만 평가 (PROCEDURE.md는 skip), persona는 권장 skill만 평가
5. **출처 명시**: 모든 평가 기준은 authoring-guide의 어떤 섹션에서 왔는지 출력에 표시

---

## 4. 실행 절차

### Step 1. 입력 파싱 + Paired Evaluation 대상 결정

> **🔴 가이드 §1.6 SSoT**: 단일 위치 평가는 결함 누락의 직접 원인. 기본은 **paired**.

`$ARGUMENTS`에서 skill 이름 또는 경로 추출 + paired 평가 대상 자동 확장:

| 입력 패턴 | Paired 평가 대상 자동 확장 (모두 존재 확인 후 평가) |
|---------|------------------------------------------|
| `concretize-idea` (skill name) | **PC mode**: PROCEDURE + commands + catalog entry — 셋 다 자동 평가 |
| `plugin/skills/<name>/PROCEDURE.md` | **PC mode** — 위와 동일 (경로 → name 추출 후 같은 처리) |
| `plugin/skills/router/SKILL.md` | **PC5 single mode**: router 자신 평가 (catalog entry 없음 — router는 catalog의 호스트) |
| `plugin/commands/<name>.md` | **C1/C2 single mode**: command frontmatter만 평가 (사용자가 명시적 단일 평가 요청) |
| 빈 입력 | "평가할 skill 이름을 알려주세요" 질의 |
| `--single` 플래그 동반 | 사용자 명시 단일 평가 — paired skip, 기존 C1-C4 적용 |

**Paired 자동 확장 절차** (skill name 입력 시):

1. `plugin/skills/<name>/PROCEDURE.md` 존재 확인 → 없으면 에러 (skill 부재)
2. `plugin/commands/<name>.md` 존재 확인:
   - 존재 → Step 4 (Frontmatter 평가) 활성화
   - 부재 → pattern library skill 가정 — F1-F5 분모 제외, F5(command 부재) 항목은 pass로 자동 마킹
3. `plugin/skills/router/references/skill-catalog.md`에서 `` `<name>` `` grep → entry 존재 확인:
   - 존재 → Step 4b (Catalog Entry 평가) 활성화
   - 부재 → CE1-CE5 모두 fail (router가 dispatch 못 함 — 즉시 보고)
4. `<name>` == `router`이면 router/SKILL.md만 평가 (PC5)

추출한 대상 목록을 Step 2-9 내내 일관 유지.

### Step 2. 대상 파일 read (paired 위치 모두)

Step 1에서 결정된 paired 대상 모두 read:

| 대상 | 경로 | 평가 카테고리 |
|------|------|------------|
| PROCEDURE | `plugin/skills/<name>/PROCEDURE.md` | B1-B12, P1-P4 |
| Command | `plugin/commands/<name>.md` (존재 시) | F1-F5 |
| Catalog Entry | `skill-catalog.md`의 `` `<name>` `` 행 (grep) | CE1-CE5 |
| Router | `plugin/skills/router/SKILL.md` (name이 router일 때만) | F1-F5 |

또한 평가 기준 문서 read:
- `docs/plugin-skills-authoring-guide.md` (§1.2 F1-F5 + §1.5 CE1-CE5 + §1.6 paired 정책 + §3.5 페르소나 매트릭스 + §4 체크리스트)
- `plugin/skills/router/references/engineering-phases.md` §4 (I/O Contract 표준 형식)

### Step 3. 평가 케이스 결정 (PC1-PC5 vs C1-C4)

> **🔴 paired 우선**: 사용자가 `--single`로 명시하지 않은 한 paired (PC1-PC5).

```
paired 자동 적용 (skill name 입력 시):
  - PROCEDURE 존재 + commands/<name>.md 존재 + catalog entry 존재
    → persona 권장 여부에 따라 PC1 (분모 40) 또는 PC2 (분모 34)
  - PROCEDURE 존재 + commands/<name>.md 부재 + catalog entry 존재 (pattern library)
    → PC3 (분모 35) 또는 PC4 (분모 29)
  - router/SKILL.md 단독 (name == router)
    → PC5 (분모 29)

single mode (--single 플래그 또는 단일 파일 경로 입력):
  - command.md or router/SKILL.md 단독 → C1/C2 (분모 35/29)
  - PROCEDURE.md 단독 → C3/C4 (분모 30/24)

평가 적용 매트릭스:
- F1-F5: commands/<name>.md (PC1/PC2/C1/C2) 또는 router/SKILL.md (PC5)에서만
- CE1-CE5: catalog entry 존재할 때만 (PC1-PC4)
- B1-B12: PROCEDURE.md에서만 (router/SKILL.md 본문은 SKILL.md 자체 평가 시 일부 적용 — B1, B12)
- P1-P4: 가이드 §3.5.1 권장 매트릭스 결과 "권장"일 때만 (PC1/PC3/C3)
```

### Step 4. Frontmatter 평가 (적용 시)

각 항목 평가:

#### F1: `description` 첫 문장에 "Use when" 또는 "사용 시점" 명시
- **Pass 기준**: 첫 1-2문장 안에 "use when", "Use when", "사용 시점", "사용할 때" 키워드 존재
- **Fail 시 제안**: "description 첫 문장을 'Use when <상황>: <skill 동작>' 형식으로 재작성하라. 예: 'Use when reviewing code for security: performs OWASP Top 10 audit.'"

#### F2: 자연어 발화 패턴 3개 이상
- **Pass 기준**: description에 사용자가 표현할 법한 자연어 trigger keyword 3개 이상 (한국어/영어 무관)
- **Fail 시 제안**: "trigger keywords 3-5개 추가. 예: '\"security review\", \"OWASP check\", \"보안 검토\" 같은 자연어 발화에 매칭'."

#### F3: `description` + `when_to_use` 합산 1,536자 이내 + 추상적 형용사 없음
- **Pass 기준**: `description` 글자 수 + (있으면 `when_to_use` 글자 수) 합산 ≤ 1,536. 그리고 "종합적", "효과적", "다양한", "comprehensive", "various", "effective" 같은 추상적 형용사 0개. (출처: 가이드 §1.2.1 — Anthropic 공식 skill listing truncation cap)
- **Fail 시 제안** (글자 수 초과): "현재 합산 <N>자. `description`의 핵심 use case를 앞쪽에 두고 후순위 trigger를 제거하여 1,536자 이내로 압축. cap 초과분은 Claude가 dispatch 결정 시 보지 못함."
- **Fail 시 제안** (추상적 형용사): "추상적 형용사를 구체적 키워드로 교체. 예: '종합적 보안 검토' → 'OWASP Top 10 + AWS IAM 검토'."

#### F4: side-effect 작업이면 `disable-model-invocation: true`
- **Pass 기준**: 다음 조건이면 명시 필수:
  - 본문에 deploy/commit/push/migration/외부 API 호출/파일 삭제 언급
  - 명령 이름이 setup-*, auto-create-*, ship-*, deploy-* 패턴
- **Fail 시 제안**: "side-effect가 있는 작업이므로 frontmatter에 `disable-model-invocation: true` 추가. Claude가 사용자 의도 없이 자동 호출하는 것을 방지."

#### F5: pattern library / 메타 skill의 사용자 호출 차단 (buddy 등가 매핑)
- **buddy 메커니즘** (가이드 §1.2.3): PROCEDURE.md는 frontmatter가 없어 `user-invocable: false`를 직접 적용 불가. 대신 `plugin/commands/<name>.md` 파일의 **부재**로 동등 효과.
- **Pass 기준** (router/SKILL.md, command.md 평가 시): 본문에 "패턴 라이브러리", "pattern library", "다른 skill이 호출" 등 명시되면 `user-invocable: false` 적용
- **Pass 기준** (PROCEDURE.md 평가 시): 본문에 "패턴 라이브러리" 등 명시되면 **해당 PROCEDURE에 대응되는 `plugin/commands/<name>.md` 파일이 부재해야 함**. 존재하면 fail.
- **Fail 시 제안** (router/command): "`user-invocable: false` frontmatter 추가."
- **Fail 시 제안** (PROCEDURE.md + commands 파일 존재): "`plugin/commands/<this-name>.md` 파일 삭제. pattern library는 router 경유 내부 호출만 허용."

### Step 4b. Catalog Entry 평가 (paired 시 — CE1-CE5)

> **🔴 paired evaluation의 핵심**: catalog entry는 router가 dispatch 결정 시 보는 텍스트. PROCEDURE.md가 우수해도 catalog entry가 약하면 dispatch 자체가 실패.

`skill-catalog.md`에서 `` `<name>` `` grep으로 entry 행 찾고 평가:

#### CE1: 정확한 phase 표에 위치
- **Pass 기준**: entry가 `engineering-phases.md` §2 phase 정의와 일치하는 phase의 표에 위치 (예: 문제·기회 검증 단계(Phase 1) Mode A 스킬은 §1 Problem/Opportunity Validation 표에)
- **Fail 시 제안**: "현재 <phase A> 표에 있으나 engineering-phases.md §2 정의상 <phase B>에 속함. 표 이동."

#### CE2: 호출 방법 컬럼 + 실제 command 파일 일치
- **Pass 기준**: 3종 중 하나 명시 + 실제 파일과 일치
  - `command + dispatch` → `plugin/commands/<name>.md` **존재**
  - `dispatch only (via /buddy:<router>)` → command 파일 **부재**
  - `(직접 호출 불가, <X> 경유)` → command 파일 **부재** + 경유 경로 명시
- **Fail 시 제안** (선언 vs 실제 불일치): "선언은 `command + dispatch`인데 파일 부재. 둘 중 하나로 정렬: (1) command 파일 생성 또는 (2) 선언을 `dispatch only`로 변경."

#### CE3: 트리거 키워드 + 목적 + 추상 형용사 0
- **Pass 기준**: description에 자연어 트리거 키워드 ≥ 2개 + 명확한 목적 + "종합적", "효과적", "다양한" 등 추상 형용사 0개
- **Fail 시 제안**: "추상 형용사 발견(<단어>). 구체적 키워드로 교체. 예: '종합적 검토' → 'OWASP Top 10 + AWS IAM 검토'."

#### CE4: 같은 phase 내 차별점 명확
- **Pass 기준**: 같은 phase 표 내 다른 entry와 description 키워드 80% 미만 중복
- **Fail 시 제안**: "<sibling> entry와 키워드 중복 발견. 차별점 한 문장 추가. 예: '<sibling>이 X라면 본 skill은 Y다'."

#### CE5: description 길이 50-300자
- **Pass 기준**: description 글자 수 50 ≤ N ≤ 300
- **Fail 시 제안** (짧음): "현재 <N>자. 트리거 키워드 또는 차별점 추가."
- **Fail 시 제안** (김): "현재 <N>자. 표 가독성 저하. 핵심만 남기고 상세는 PROCEDURE.md §1로 이관."

### Step 5. Body 구조 평가

**§2.1 3-Tier 분류 동기화 (B-M1 트랜지션 정책)** — 가이드 §4.2와 1:1 동기. 평가 정책은 Tier별로 다르다:

| Tier | 항목 | 정책 |
|------|------|------|
| 🔴 필수 | B1, B2, B3, B4 | binary pass/fail. 미충족 = 즉시 fail |
| 🟡 권장 | B5, B6, B7, B8 | 미충족 시 **예외 사유 검토 후 결정**. 명확한 사유 있으면 pass (자동 N/A) |
| 🟢 선택 | B9, B10, B11, B12 | bonus only. 미충족도 자동 pass (점수 미차감) |

**권장(B5-B8) 예외 사유 판정 가이드**:
- B5: 단일 dispatch skill로 사용 case가 자명한 경우 → N/A pass
- B6: 대체 skill이 없는 도메인 단독 skill → N/A pass
- B7: 30줄 미만의 단순 dispatch skill → N/A pass
- B8: 분기가 없는 일자 절차 → N/A pass

명백한 사유가 없으면 fail 처리하고 개선 제안에 포함.

#### B1: 제목 직후 1-3줄 정체성 + 진입/종료 조건
- **Pass 기준**: `# <name>` 다음 첫 3-5줄 안에 (a) skill의 정체성, (b) 진입 조건 또는 종료 조건이 명시
- **Fail 시 제안**: "제목 직후 다음 형식 추가:\n```\n<skill 정체성 1-2문장>\n\n**진입 조건**: <언제 호출되는가>.\n**산출물**: <무엇을 만드는가>.\n**다음 단계**: <어떤 skill로 cascade>.\n```"

#### B2: 본문 500줄 이하 (+ lost-in-the-middle 위험 관리)
- **Pass 기준**: PROCEDURE.md 줄 수 ≤ 500
- **Fail 시 제안**: "현재 <N>줄. 가이드 §2.2.1 long-context 전략 5가지 중 적합한 것 적용:\n1. 보조 파일 lazy-load — `references/<topic>.md`, `examples.md`로 분리 (가장 흔함)\n2. 본문 정보 압축 — narrative 산문을 표·체크리스트로 변환\n3. 핵심을 위/아래 배치 — 중간은 lost-in-the-middle 약 zone이므로 페르소나·체크리스트를 상·하단에\n4. skill 자체 분할 — 한 skill이 여러 책임 보유 시 별도 PROCEDURE.md\n5. forking subagent (`context: fork`) — 결과만 필요한 큰 호출 격리 (command frontmatter)"

#### B3: Input Requirements 표 존재 (engineering-phases.md §4 형식)
- **Pass 기준**: `## Input Requirements` 헤더 + 4컬럼 표 (Input, Required, Type, Source, 미제공 시)
- **Fail 시 제안**: "다음 표 추가 (engineering-phases.md §4 표준):\n```\n## Input Requirements\n| Input | Required | Type | Source | 미제공 시 |\n|---|---|---|---|---|\n| <항목> | ✅/선택 | artifact/knowledge/decision | <source> | \"질의문\" |\n```"

#### B4: Output Contract 표 존재
- **Pass 기준**: `## Output Contract` 헤더 + 4컬럼 표 (Output, Type, Format, Consumers)
- **Fail 시 제안**: "다음 표 추가:\n```\n## Output Contract\n| Output | Type | Format | Consumers |\n|---|---|---|---|\n| <산출물> | artifact/decision | <format> | <소비 skill> |\n```"

#### B5: "이 스킬을 사용하는 경우" 섹션
- **Pass 기준**: `이 스킬을 사용하는 경우`, `When to use`, `사용 시점`, `When to invoke` 같은 헤더 + 목록 존재
- **Fail 시 제안**: "다음 섹션 추가:\n```\n## 이 스킬을 사용하는 경우\n- 상황 1\n- 상황 2\n- 상황 3\n```"

#### B6: "이 스킬을 사용하지 않는 경우" 섹션 (대안 안내)
- **Pass 기준**: "사용하지 않는 경우", "When NOT to use", "사용하지 마라" 같은 명시 + 대안 skill 안내
- **Fail 시 제안**: "다음 섹션 추가:\n```\n## 이 스킬을 사용하지 않는 경우\n- <상황 X> — 대안: `<other-skill>`\n```"

#### B7: 실행 절차가 imperative
- **Pass 기준**: 실행 절차 섹션에 `### Step N. <동사>` 형식 + 명령형 표현 ("~한다", "~하라", "~실행한다")
- **Fail 시 제안**: "절차를 imperative 형식으로 재작성:\n- '~할 수 있습니다' → '~한다'\n- '~하시면 됩니다' → '~하라'\n- 'Step N. <명사>' → 'Step N. <동사>'"

#### B8: 분기 조건 명시적
- **Pass 기준**: 본문에 "필요하면", "상황에 따라", "적절히" 같은 모호한 표현 없음 + 모든 조건문이 구체적 (when X, if Y)
- **Fail 시 제안**: "모호한 표현 발견 (위치: <line>). '필요하면' → '<X 조건이면>'으로 명확화."

#### B9: 동적 상태는 `` !`...` `` 문법으로 inline 주입
- **Pass 기준**: 본문에 "현재 git 상태", "최근 commit", "현재 파일 내용" 같은 dynamic context 언급이 있으면 `` !`command` `` 문법 사용
- **Pass 예외**: dynamic context 언급 자체가 없는 skill은 자동 pass
- **Fail 시 제안**: "동적 정보 묘사 발견 (위치: <line>). 다음으로 교체:\n- '현재 git 상태를 확인' → 다음 코드 블록 사용:\n```\n!`git status --short`\n```"

#### B10: 출력 형식 명시 (schema 강제)
- **Pass 기준**: `## 출력 형식`, `## Output Format`, `## 산출물 형식` 섹션에 구조화된 schema (YAML/markdown 표/JSON) 명시
- **Fail 시 제안**: "출력 형식을 명시:\n```\n## 출력 형식\n\n응답은 다음 구조로:\n\n```yaml\nresult:\n  summary: \"<50단어>\"\n  findings: [...]\n  next_steps: [...]\n```\n```"

#### B11: 검증 체크리스트 또는 anti-pattern 섹션 존재
- **Pass 기준**: `## 검증`, `## Verification`, `## 검증 체크리스트`, `## Anti-patterns` 중 하나 이상 존재
- **Fail 시 제안**: "다음 중 하나 이상 추가:\n```\n## 검증 체크리스트\n- [ ] 항목 1\n- [ ] 항목 2\n\n## Anti-patterns\n| ❌ | ✅ | 이유 |\n```"

#### B12: 다음 단계 명시
- **Pass 기준**: `## 다음 단계`, `## Next Steps`, `## 다음 phase`, `→ <skill>` 형식의 cascade 안내
- **Fail 시 제안**: "다음 섹션 추가:\n```\n## 다음 단계\n\n→ `<next-skill>` — <연결 이유>\n```"

### Step 6. Persona 평가 (적용 시)

가이드 §3.5.1 권장 매트릭스로 persona 적용성 판단:

```
권장 skill 패턴 매칭 (가이드 §3.5.1):
- review-* / audit-* / analyze-* / critique-* → persona 권장
- design-* (특정 도메인) → persona 권장
- build-with-* / iterate-* (discipline 강제) → persona 권장
- deprecate-* / archive-* / migrate-* / spin-off-* (수명주기 관리 단계(Phase 9) lifecycle) → persona 권장 — Senior PM (sunset 전문, 사용자 영향·소통·timeline 중심)
- update-* / sync-* (데이터 변환) → persona 비권장
- status / save-context / restore-context / start / router → persona 비권장
- freeze-* / compose-* (패턴 라이브러리) → persona 비권장
```

권장 skill인 경우만 P1-P4 평가:

#### P1: 본문 최상단 (제목 직후)에 페르소나 정의
- **Pass 기준**: `# <name>` 다음 첫 단락이 "당신은 X다" 형식의 페르소나 선언
- **Fail 시 제안**: "본문 최상단(제목 직후)에 다음 형식 페르소나 추가:\n```\n당신은 <경력 + 전문영역>이다.\n당신의 역할은 <현재 skill에서의 책임 범위>이다.\n당신은 <무엇을 보고 무엇을 출력하는가>.\n```"

#### P2: 페르소나 구성 요소 4-5개 포함
- **Pass 기준**: persona 단락에 다음 4-5요소 중 4개 이상 포함:
  - 경력/배경
  - 전문 영역
  - 현재 역할 (이 skill에서의)
  - 무엇을 보는가
  - 무엇을 출력하는가 (자가 출력 규칙)
- **Fail 시 제안**: "현재 persona에 다음 요소 누락: <누락 요소>. 가이드 §3.4.2 참조하여 보강."

#### P3: 3-7줄 길이
- **Pass 기준**: persona 단락이 3-7줄 사이
- **Fail 시 제안**:
  - 너무 짧음 (1-2줄): "구체성 보강 필요. 경력/영역/출력 규칙 추가."
  - 너무 김 (8줄+): "본문 토큰 비용 누적. 핵심 4-5요소로 압축."

#### P4: 같은 도메인 skill과 페르소나 톤 일관
- **Pass 기준**: 같은 prefix (review-*, audit-* 등) skill들의 페르소나 형식이 통일
- **Fail 시 제안**: "다른 review-* skill과 페르소나 형식 비교 후 통일:\n- review-engineering: '당신은 senior engineering manager다'\n- review-design: '당신은 senior product designer다'\n→ 본 skill도 'senior X' 형식으로 통일 권장."

### Step 7. 가중치 점수 계산

**SSoT**: 가이드 §4.4 점수 계산 표준. 본 step은 §4.4의 공식을 그대로 적용하며, **케이스별 분모 결정 규칙을 정확히 따라야 한다**. (한쪽만 변경 시 점수 산정 결과가 어긋남)

#### 카테고리 가중치 (고정)

```
- Frontmatter:    1.0 (5 항목 F1-F5)
- Catalog Entry:  1.0 (5 항목 CE1-CE5)   ← 신규
- Body 구조:      2.0 (12 항목 B1-B12)
- Persona:        1.5 (4 항목 P1-P4)
```

#### Paired Evaluation 케이스 (PC1-PC5, 기본) — 가이드 §4.4 동기

| 케이스 | 평가 대상 조합 | Persona | 분모 합 |
|-------|--------------|---------|--------|
| **PC1** | PROCEDURE + command + catalog entry, persona 권장 | 권장 | **40** (5+5+24+6) |
| **PC2** | PROCEDURE + command + catalog entry, persona 비권장 | 비권장 | **34** (5+5+24) |
| **PC3** | PROCEDURE + catalog entry (command 부재 — pattern library), persona 권장 | 권장 | **35** (5+24+6) |
| **PC4** | PROCEDURE + catalog entry (command 부재), persona 비권장 | 비권장 | **29** (5+24) |
| **PC5** | router/SKILL.md 단독 (catalog entry 없음) | N/A | **29** (5+24) |

#### Single Mode 케이스 (C1-C4, --single 플래그 또는 단일 파일 경로 입력 시)

| 케이스 | 평가 대상 | Persona | 분모 합 |
|-------|---------|---------|--------|
| **C1** | router/SKILL.md or commands/*.md 단독 | 권장 | **35** |
| **C2** | router/SKILL.md or commands/*.md 단독 | 비권장 | **29** |
| **C3** | PROCEDURE.md 단독 | 권장 | **30** |
| **C4** | PROCEDURE.md 단독 | 비권장 | **24** |

#### 통과 가중치 합 계산

```
통과 가중치 합 =
  (해당 케이스에서 평가한 F항목 중 통과 수  × 1.0)
+ (해당 케이스에서 평가한 CE항목 중 통과 수 × 1.0)   ← 신규
+ (B항목 중 통과 수                         × 2.0)
+ (해당 케이스에서 평가한 P항목 중 통과 수  × 1.5)
```

> **카테고리별 skip 조건**:
> - F1-F5: PROCEDURE.md만 단독 평가하는 C3/C4에서 skip. paired (PC1/PC2)에서는 command.md가 존재하므로 평가
> - CE1-CE5: catalog entry 없는 router(PC5)와 단일 평가 케이스에서 skip
> - P1-P4: persona 비권장 (PC2/PC4/C2/C4/PC5)에서 skip
>
> **B항목 Tier 트랜지션 정책** (Step 5 기준): B1-B4는 binary pass/fail. **B5-B8은 "예외 사유 검토 통과" = pass 처리**. **B9-B12는 미충족도 자동 pass** (선택 항목, bonus only).

#### 최종 점수 + Grade

```
최종 점수 = (통과 가중치 합 / 케이스 분모 합) × 100

90+:    우수 (참고 모델로 등록)
80-89:  합격 (운영 가능)
70-79:  보강 필요
70 미만: 재작성 권장
```

#### 예시 (C3 — PROCEDURE.md + persona 권장)

- B1-B12 중 10개 통과, P1-P4 중 3개 통과
- 통과 가중치 합 = 10×2.0 + 3×1.5 = 24.5
- 분모 = 30
- 최종 점수 = 24.5 / 30 × 100 = **81.67점 → 합격**

### Step 8. 개선 제안 우선순위 결정

각 fail 항목에 priority 부여:

| Priority | 기준 |
|----------|------|
| **High** | 가중치 ≥ 2.0 (Body 항목 전체) + persona 권장 skill의 P1 |
| **Medium** | 가중치 1.5 (Persona 항목 P2-P4) |
| **Low** | 가중치 1.0 (Frontmatter 항목) |

### Step 9. 리포트 출력

§5의 출력 형식으로 결과 출력.

---

## 5. 출력 형식

```markdown
# Skill Evaluation Report — `<skill-name>`

**Mode**: paired (PC1-PC5) | single (C1-C4)
**Case**: <PC1 | PC2 | ... | C4>
**Evaluated at**: <ISO 8601>
**Paired Targets** (paired 모드만):
  - PROCEDURE: `<path>` (read OK / missing)
  - Command: `<path>` (read OK / missing — pattern library)
  - Catalog Entry: `<grep result>` (found / missing)
  - Router: `<path>` (name이 router인 경우만)

## Summary

| 항목 | 값 |
|------|---|
| **Total Score** | XX.X / 100 |
| **Grade** | 우수 / 합격 / 보강 필요 / 재작성 권장 |
| **Total Items Evaluated** | N (F<a> + CE<b> + B<c> + P<d>) |
| **Passed** | M |
| **Failed** | K |

## Category Scores

### Frontmatter (F1-F5) — `<command 경로>`
- **Applicable**: yes/no (PC3/PC4/PC5/C3/C4는 N/A)
- **Score**: A/5 × 1.0 = X.X
- **Details**:
  - F1: ✅ pass / ❌ fail — <이유>
  - F2-F5: ...

### Catalog Entry (CE1-CE5) — `skill-catalog.md` `<name>` row
- **Applicable**: yes/no (PC5/single mode에서 N/A)
- **Score**: A/5 × 1.0 = X.X
- **Details**:
  - CE1: ✅ pass / ❌ fail — <이유>
  - CE2-CE5: ...

### Body 구조 (B1-B12) — `<PROCEDURE 경로>`
- **Score**: A/12 × 2.0 = X.X
- **Details**:
  - B1: ✅ pass / ❌ fail — <이유>
  - B2-B12: ...
  - B3: ❌ fail — Input Requirements 표 부재
  - ...

### Persona (P1-P4)
- **Applicable**: yes/no (가이드 §3.5.1 권장 매트릭스 기반)
- **Reason for applicability decision**: <skill 유형 분류 + 근거>
- **Score** (적용 시): A/4 × 1.5 = X.X
- **Details**:
  - P1: ❌ fail — 페르소나 정의 부재
  - ...

## Improvement Suggestions (Priority Sorted)

### High Priority

#### B3 — Input Requirements 표 부재
**현재**: 본문에 Input Requirements 섹션 없음
**개선**:
```markdown
## Input Requirements

| Input | Required | Type | Source | 미제공 시 |
|-------|----------|------|--------|----------|
| <item> | ✅/선택 | artifact/knowledge/decision | <source> | "<질의문>" |
```
**참조**: `plugin/skills/router/references/engineering-phases.md` §4

#### <다른 High priority 항목>

### Medium Priority

#### P1 — 페르소나 정의 부재
**현재**: 제목 직후 페르소나 선언 없음 (본 skill은 audit-* 패턴으로 persona 권장 대상)
**개선**: 제목 다음 줄에 다음 추가:
```markdown
당신은 <경력 + 전문영역>이다.
당신의 역할은 <이 skill에서의 책임>이다.
당신은 <무엇을 보고 무엇을 출력하는가>.
```
**참조**: `docs/plugin-skills-authoring-guide.md` §3.4

### Low Priority

<F* 항목들>

## Recommended Next Actions

1. <가장 영향 큰 fail 항목 해결>
2. <두 번째>
3. <세 번째>

(개선 후 `/buddy:evaluate-skill <skill-name>` 재실행으로 점수 재측정)
```

---

## 6. 검증 체크리스트

평가 작업 완료 전 다음 확인:

- [ ] **paired 자동 확장 적용** (skill name 입력 시 PROCEDURE + command + catalog entry 모두 read — `--single` 명시 없는 한)
- [ ] 4 위치 (PROCEDURE / command / catalog entry / router) 중 평가 대상 모두 실제로 read했는가? missing 항목은 리포트에 명시했는가?
- [ ] 평가 가이드(authoring-guide.md §1-§4)와 표준 형식(engineering-phases.md §4) 둘 다 참조했는가?
- [ ] 평가 케이스 정확히 결정 (PC1-PC5 또는 C1-C4)? 분모 합이 케이스 기준과 일치?
- [ ] 모든 적용 항목(F/CE/B/P)이 각자의 기준으로 평가됐는가? (CE 누락 = 가장 흔한 결함)
- [ ] 가중치 점수 계산이 정확한가? (가이드 §4.4 paired 통합 분모)
- [ ] 각 fail 항목에 actionable 개선 제안이 있는가? (vague한 "개선하라" 금지)
- [ ] priority 정렬이 일관적인가? (High → Medium → Low)
- [ ] 출력이 §5 형식을 따르는가? (Mode/Case/Paired Targets 표시 필수)

---

## 7. Anti-patterns

| ❌ | ✅ | 이유 |
|----|----|------|
| **단일 위치만 평가** (skill name 입력했는데 PROCEDURE만 봄) | **paired 자동 확장** (PROCEDURE + command + catalog entry 모두 read) | 다른 위치 결함 누락 — 2026-06-01 실제 발생 (start commands의 F1 fail 미발견) |
| Catalog entry 평가 skip | CE1-CE5 항목 적용 (catalog entry가 router dispatch의 핵심) | dispatch 신호 누락 시 PROCEDURE 우수해도 사용자에게 도달 X |
| "이 부분을 개선하라" | "B3 fail: Input Requirements 표 추가. 형식: ..." | 구체성 |
| 모든 항목 강제 평가 | persona는 적용 매트릭스 기반 selective | False fail 방지 |
| 단순 통과율 계산 | 가중치 점수 (§4.4 paired 통합) | 항목 중요도 반영 |
| 평가만 출력 | 평가 + 개선 제안 + 다음 actions | 사용자가 즉시 행동 가능 |
| skill 본문 자동 수정 | 평가만, 수정은 사용자 결정 후 | User Sovereignty |

---

## 7b. 🔴 Paired Evaluation Reminder (lost-in-the-middle 방어)

> **본문 상단에서 강조된 정책을 하단에 mirror**. 본문 길이로 인한 lost-in-the-middle 회피.

skill name (`start`, `concretize-idea` 등) 입력을 받았는가? → **paired 자동 확장 필수**:

1. ✅ `plugin/skills/<name>/PROCEDURE.md` read + B1-B12, P1-P4 평가
2. ✅ `plugin/commands/<name>.md` 존재 확인 → 존재 시 read + F1-F5 평가
3. ✅ `skill-catalog.md` grep `<name>` → entry 발견 시 CE1-CE5 평가
4. ✅ name == router 시 `plugin/skills/router/SKILL.md` read + F1-F5 평가

위 4 단계 중 하나라도 skip하면 **평가 결과 신뢰성 0**. 본 reminder는 상단 §CRITICAL 박스와 Step 1과 1:1 동기.

---

## 8. 다음 단계

평가 후:
1. **점수 90+**: → `docs/plugin-skills-authoring-guide.md` §3.5에 참조 모델로 등록 검토
2. **점수 80-89**: → 부족 항목만 점진 개선
3. **점수 70-79**: → 즉시 개선 작업 시작 (high priority 항목부터)
4. **점수 70 미만**: → 재작성 후보 — `/buddy:run write-a-skill <skill-name>` 또는 가이드 §5.1 표준 템플릿 기반 재작성

Batch 평가:
- 동일한 skill 다수 평가 시 사용자가 shell loop로 반복 호출:
  ```bash
  for skill in concretize-idea write-hld assess-product-change start; do
    # /buddy:evaluate-skill $skill (Claude Code 세션에서)
  done
  ```
- 향후 batch 평가 전용 wrapper 가능 (필요 시 신규 skill로 분리)

---

## 9. 참조

- `docs/plugin-skills-authoring-guide.md` — 평가 기준 SSoT
  - §1.2 핵심 키 상세 (F1-F5 평가 근거)
  - §1.4 Claude 모델별 가이드 (`model`/`effort` 키 적절성 평가 시 참조)
  - §2.0 PROCEDURE.md 프롬프트 역할
  - §2.1 3-Tier 분류 + §2.2 핵심 원칙 (B1-B12 평가 근거)
  - §2.2.1 long-context 5전략 (B2 fail 시 제안)
  - §2.4 XML / §2.5 CoT / §2.6 한·영 정책 (모두 **optional bonus** — 본 평가에 미반영, 가이드 정책과 일관)
  - §3.5.1 persona 권장 매트릭스 (Step 6 근거 — 9개 phase(Phase 1-9) + Cross-cutting 전체)
  - §4 통합 평가 체크리스트 (F1-F5, B1-B12, P1-P4 정의)
  - §4.4 케이스별 분모 (C1-C4) — Step 7과 1:1 동기
- `plugin/skills/router/references/engineering-phases.md` §4 — I/O Contract 표준 형식 (B3/B4 평가 근거)
- `plugin/skills/router/references/skill-catalog.md` — 전체 skill 카탈로그 (다른 skill과 일관성 비교용)
- 참조 모델 5개 (90+ 점수 후보): `review-engineering`, `review-design`, `review-devex`, `audit-security`, `critique-plan`
