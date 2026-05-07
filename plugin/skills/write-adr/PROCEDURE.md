# write-adr — Architecture Decision Record 작성

ADR (Architecture Decision Record) 는 기술 의사결정을 표준 양식으로 보존해 미래 의 maintainer / 신규 팀원 / 본인 자신이 "왜 이 결정을 했는지" 를 추적 가능하게 만드는 도구다. `define-tech-stack`, `design-data-model`, `design-api-contract` 같은 의사결정 skill 의 산출물을 받아 `docs/adr/NNNN-<slug>.md` 형식으로 영속화한다. 본 skill 은 ADR 의 **표준 7 섹션 강제 + 결정 supersede 체인 관리 + Index 갱신 + Status 라이프사이클** 을 보장한다.

## 0. STOP — 시작 전 읽기

이 skill 은 다음 anti-pattern 들을 방지한다. 산출물에 다음이 발견되면 §5 로 돌아가 보강한다:

- **Status 누락 / 모호** — `proposed` / `accepted` / `deprecated` / `superseded` 4 가지 중 명시 안 함. ADR 의 lifecycle 추적 불가.
- **Context 가 "필요해서 만들었다"** — 해결하려는 문제 / 제약 / 가정이 명시 안 됨. 미래의 reader 가 결정 적용 가능한지 판단 불가.
- **Decision 만 적고 Alternatives 누락** — "Postgres 쓴다" 만 있고 검토했던 후보 + 거부 이유 없음. evidence 영속화 실패.
- **Consequences 가 positive 만** — 부정 / 중립 영향 누락. 결정 비용을 미래 reader 가 이해 못 함.
- **Supersede 체인 단방향** — 새 ADR 가 기존 ADR 대체할 때 양방향 link 누락. 추적 끊김.
- **Index 미갱신** — 새 ADR 만들고 `docs/adr/README.md` 또는 INDEX 안 건드림. discovery 불가.
- **ADR id 충돌** — 기존 max id 확인 안 하고 임의 부여.

§5 의 모든 phase 누락 없이 수행하라.

## 1. 목적

기술 의사결정의 영속화를 강제. "왜 이 선택을 했는가" 의 답을 미래 reader 가 ADR 1 파일 만으로 재구성할 수 있게 표준화. 의사결정 skill (define-tech-stack 등) 의 chain 종착점으로 호출되어 그 결정을 ADR 양식으로 전환한다.

## 2. 사용 시점 (When to invoke)

- `define-tech-stack` / `design-data-model` / `design-api-contract` 종료 직후 (chain 권장)
- 기존 ADR 의 결정 변경 / supersede 시
- 외부 system / 3rd-party 결정 inheritance 영속화 (예: company-wide 표준 채택)
- post-incident review 결과 결정 변경 (예: "X 를 안 쓰기로 결정") 의 영속화
- ad-hoc 의사결정 (chat 으로 끝낸 결정) 의 retrospective ADR 작성

## 3. 입력 (Inputs)

### 필수
- 결정 context (어떤 문제 / 제약 / 가정)
- 검토했던 alternatives + rejection reason
- 선택지 (decision)
- 결정자 / 영향 범위
- supersede 대상 ADR (있다면)

### 선택
- 기존 ADR repository 위치 (default `docs/adr/`)
- ADR id 형식 규칙 (default `NNNN`, 4 digit zero-pad)
- Status 초기값 (default `accepted`, draft 인 경우 `proposed`)

### 입력이 부족할 때 forcing question
- "이 결정으로 영향 받는 사람 / 시스템은? 명시 안 하면 ADR 의 reader 가 적용 가능성을 판단 못 함."
- "결정을 뒤집을 수 있는 조건은? '계속 동작' 만으론 부족 — '어떤 measurable signal' 명시 필요."
- "검토했던 alternatives 가 단 하나? Real alternatives 검토했나? '검토 안 함' 도 정직하게 기록 (그게 자체로 ADR 의 caveat)."
- "이 ADR 가 기존 ADR 와 충돌하면? supersede 체인 정리 필요."

## 4. 핵심 원칙 (Principles + Posture)

이 skill 의 운영 posture:

- **입장 취함, hedge 금지** — Decision 섹션은 "X 를 선택" 으로 단호. "X 또는 Y 가능" 은 ADR 가 아님. 명확한 선택 + 변경 조건.
- **사용자 입력을 challenge** — 사용자가 Context 를 1줄로만 적으려 하면 push back. 미래 reader 가 적용 가능성 판단할 만큼 specific 한지 검증.
- **Specificity 강제** — "performance 좋음" 거부. 숫자 / 측정 가능 조건 / vendor 이름 / version 명시.
- **Honest about unknowns** — 검토 안 한 alternatives, 측정 못한 trade-off 는 솔직하게 명시. 가짜 정확성 회피.

도메인 원칙:

1. **표준 7 섹션 강제** — Title / Status / Context / Decision / Consequences / Alternatives / References. 누락 시 ADR 가치 무너짐.
2. **Status 는 lifecycle** — proposed → accepted → (deprecated 또는 superseded). 단방향 transitions 만 허용.
3. **Supersede 양방향 link** — 새 ADR 가 기존 대체 시 양쪽 모두 link 명시.
4. **ADR id 는 immutable** — 한번 부여하면 변경 안 함. supersede 도 새 id 발급, 기존 id 유지.
5. **Index 는 source of truth** — `docs/adr/README.md` 또는 INDEX.md 가 모든 ADR 의 entry. ADR 만 있고 index 누락 시 discovery 불가.

## 5. 단계 (Phases)

### Phase 1. ADR id 결정 + slug 생성

```bash
# 기존 max id 확인
ls docs/adr/[0-9]*.md 2>/dev/null | sort -V | tail -1
# 결과의 NNNN + 1 = 새 ADR id

# slug: kebab-case, 5단어 이내, 결정 핵심
```

### Phase 2. 7 섹션 작성

표준 ADR 양식:

```markdown
# ADR-NNNN: <Title>

- **Status**: proposed | accepted | deprecated | superseded by [ADR-MMMM](./MMMM-...)
- **Date**: YYYY-MM-DD
- **Decider**: <name / role>
- **Stakeholders**: <list>

## Context

<문제 / 제약 / 가정 / 결정해야 하는 이유 — 미래 reader 가 적용 가능성 판단할 만큼 specific. 1~3 단락.>

## Decision

<선택지를 단호하게. "X 를 선택한다." 한 문장 + 1~2 단락 detail.>

## Consequences

### Positive
- <긍정 결과>

### Negative
- <부정 결과 / 비용 / 책임>

### Neutral
- <중립 영향 — 다른 의사결정에 영향 / 새 제약 추가 등>

## Alternatives Considered

### Alternative A: <name>
- Pros: ...
- Cons: ...
- Why rejected: <1 줄 명시>

### Alternative B: <name>
- ...

(검토 안 한 후보가 있다면 솔직하게: "검토 안 함 — <왜 시간 / 정보 부족>")

## References

- 관련 ADR: [ADR-XXXX](./XXXX-...)
- 외부 reference: <link>
- 산출물: <define-tech-stack 등 결정 skill 의 출력 link>
```

### Phase 3. Supersede 체인 처리

기존 ADR 를 대체하면:

1. 새 ADR 의 Status: `accepted, supersedes [ADR-MMMM](./MMMM-...)`
2. 기존 ADR 의 Status 갱신: `superseded by [ADR-NNNN](./NNNN-...)` (수정 commit)
3. 양방향 link 무결성 확인

### Phase 4. Index 갱신

`docs/adr/README.md` (또는 INDEX.md) 에 row 추가:

```markdown
| ID | Title | Status | Date |
|----|-------|--------|------|
| 0001 | ... | accepted | YYYY-MM-DD |
| 0002 | ... | superseded by 0005 | YYYY-MM-DD |
| ... | ... | ... | ... |
```

또는 list 형식이면 chronological 순서로 entry 추가.

### Phase 5. Validation

다음 자동 검증:

```bash
# 모든 ADR 가 7 섹션 (Title + 4 metadata + Context + Decision + Consequences + Alternatives + References) 을 가지는지
for f in docs/adr/[0-9]*.md; do
  for sec in "Status" "Context" "Decision" "Consequences" "Alternatives Considered"; do
    grep -q "^## $sec\|^- \*\*$sec" "$f" || echo "MISSING $sec in $f"
  done
done
```

`MISSING` 출력이 있으면 §2 로 돌아가 해당 섹션 보강.

## 6. 산출물 형식 (Output format)

> **Note**: 본 skill 의 핵심 산출물은 ADR 마크다운 파일 그 자체다. 출력은 다음 두 파트:

### Part A: ADR 파일 내용

위 §5 Phase 2 의 표준 양식 그대로 작성한 markdown.

### Part B: Skill 실행 보고

```markdown
## write-adr Output

### Created
- **File**: `docs/adr/NNNN-<slug>.md`
- **ADR id**: NNNN
- **Title**: <title>
- **Status**: accepted | proposed | ...
- **Length**: <line count>

### Validation
| Section | Present |
|---------|---------|
| Status | ✓ |
| Context | ✓ |
| Decision | ✓ |
| Consequences (positive + negative + neutral) | ✓ |
| Alternatives Considered | ✓ |
| References | ✓ |

### Supersede Chain (해당 시)
| ADR | Direction | Status update |
|-----|-----------|---------------|
| ADR-MMMM | superseded by NNNN | needs commit |

### Index Update
- **File**: `docs/adr/README.md`
- **Entry added**: `| NNNN | <title> | accepted | YYYY-MM-DD |`

### Next Step
<구체 action — 1줄: 예 "supersede 한 기존 ADR 의 Status 라인 갱신 후 commit", 또는 "관련 design skill 호출">
```

## 7. Cross-phase cascade

ADR 는 **모든 phase** 의 의사결정 영속화 도구로 cross-cutting:

- §1 idea / business 결정도 ADR 화 가능 (예: pricing model 결정)
- §3 design 결정 (tech stack / data model / API) — primary use case
- §4 implementation 결정 (예: actor track 분배 전략)
- §5 build 결정 (예: testing framework 선택)
- §6 quality 결정 (예: SLO 정의)
- §7 release 결정 (예: canary 전략)
- §8 operate 결정 (예: incident response runbook 채택)
- §9 lifecycle 결정 (예: deprecation 일정)

## 8. 다음 skill (next in stage flow)

본 skill 은 chain 의 종착점이라 "다음 skill" 보단 "이전 skill 의 산출물 받기":

- 직전: `define-tech-stack` / `design-data-model` / `design-api-contract` / 기타 결정 skill
- 후속: ADR 가 생성되면 사용자 / 팀 review → 필요 시 status 변경 (proposed → accepted)

## 9. 다른 skill 과의 경계 (충돌 방지)

- **vs `define-tech-stack` / `design-data-model` / `design-api-contract`** — 그들은 의사결정 process. 본 skill 은 그 결정의 영속화 format. chain 으로 함께 사용. process 와 format 분리.
- **vs `summarize-retro`** — summarize-retro 는 git history 기반 회고. 본 skill 은 명시적 architecture 결정 영속화. 회고에서 결정 발견 시 본 skill 로 영속화.
- **vs `save-context`** — save-context 는 작업 진행 상태 checkpoint (휘발성). 본 skill 은 architectural 결정 영속화 (불변). 다른 layer.

## 10. 중요 규칙

- **ADR id 는 immutable** — 부여 후 변경 금지.
- **Supersede 시 양방향 link 의무** — 단방향만 두면 추적 끊김.
- **Status 4 가지 외 사용 금지** — proposed / accepted / deprecated / superseded.
- **Index 갱신 누락 = ADR 작성 미완료** — 둘 다 단일 commit 으로 묶기.
- **ADR 본문 수정 시 caveat** — accepted ADR 의 본문은 일반적으로 수정 안 함. 결정 변경 시 새 ADR + supersede.

## 11. Verification gate — 완료 선언 전 self-check

- [ ] §5 의 5 phase 누락 없이 실행
- [ ] ADR 파일이 표준 7 섹션 (Title + Status + Context + Decision + Consequences + Alternatives + References) 모두 포함
- [ ] Status 가 4 valid 값 (proposed / accepted / deprecated / superseded by) 중 하나
- [ ] Consequences 가 positive + negative + neutral 3 sub-section 모두 포함
- [ ] Alternatives Considered 가 ≥ 2 alternatives + 각 rejection reason
- [ ] References 가 직전 결정 skill 의 산출물 link 포함 (define-tech-stack output 등)
- [ ] Supersede 한 ADR 가 있다면 양방향 link 작성됨
- [ ] `docs/adr/README.md` (또는 INDEX) 에 entry 추가됨
- [ ] §4 posture 적용 — Decision 섹션 hedge 없음 (단호)
- [ ] §0 anti-pattern 들이 산출물에 등장하지 않음
- [ ] §6 의 Skill 실행 보고 (Part B) 가 사용자에게 출력됨

하나라도 no 면 해당 phase 로 돌아가 보강 후 재검증.
