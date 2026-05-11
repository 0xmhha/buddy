# {skill-name} — {1 줄 정체성}

> **Template (B6 — ADR-004 후속)**: 본 파일은 *신규 Stage skill 작성 시 reference*. 복사 후 placeholder 채움. Orchestrator (concretize-idea / define-features / design-system / plan-build / build-feature / verify-quality / ship-release / iterate-product / manage-lifecycle) 9개는 별도 양식 (Stage 흐름 / 실행 절차 / 다음 phase / 참조 의 simple 4-section). Special skill (router / status / validate-idea forcing-question / autoplan) 도 의도된 별도 양식. 본 8-section template 은 *그 외 stage skill* 강제.

---

## 1. 목적

{이 skill 의 책임을 1~3 문장으로. *언제 / 어떤 산출* 이 핵심.}

## 2. 사용 시점

| 진입 조건 | 의미 |
|---------|------|
| {조건 1} | {설명} |
| {조건 2} | {설명} |

{또는 산문 — 1~2 문장.}

## 3. 입력

| 입력 | source | 필수 |
|------|--------|------|
| {입력 1} | {위치 / 선행 stage} | yes / no |
| {입력 2} | {…} | … |

## 4. Stage 흐름

### Stage 1: {이름}
{1~2 문장 + bullet 절차}

### Stage 2: {이름}
{…}

### (Stage N 까지 반복)

## 5. 산출물 형식

```markdown
## {산출물 제목} — {제품 / 분기}

### {sub-section 1}
{format}

### {sub-section 2}
{format}

…
```

## 6. 검증 (self-check)

- [ ] {Acceptance 1 — 자동 검증 가능한 boolean}
- [ ] {Acceptance 2}
- [ ] {…}

산출물 commit 전 위 체크 모두 통과.

## 7. 다음 phase

- *Default*: → `{next-skill}` ({phase 매핑})
- *분기 조건*: {조건 1} → `{alt-skill-1}` / {조건 2} → `{alt-skill-2}`

## 8. 참조

- `{relative path to related skill or doc}` — {왜 참조}
- `{external attribution}` — {NOTICE entry, 외부 자산 차용 시}
