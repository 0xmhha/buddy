# Audit UI Quality — visual / interaction / detail 디자인 품질 검토

## 1. 목적

`apply-design-system` 채택 후 *적용된 결과* 의 품질 검토. **visual consistency / interaction polish / micro-detail / a11y / performance** 5 차원 audit.

`audit-accessibility` (§6, 구현됨) 와 책임 분리 — 본 skill 은 *디자인 품질* (인지된 quality), audit-accessibility 는 *기술적 a11y* 준수.


## Input Requirements

| Input | Required | Type | Source | 미제공 시 |
|-------|----------|------|--------|----------|
| 기존 UI | ✅ | artifact | 배포된 UI 또는 로컬 dev 서버 | "감사할 UI의 URL 또는 경로를 알려주세요." |

## Output Contract

| Output | Type | Format | Consumers |
|--------|------|--------|-----------|
| UI quality audit report | artifact | structured report | `iterate-fix-verify` |

## 2. 사용 시점

- feature 구현 후 release 직전 — visual gate
- design review (디자이너 + dev 합동) 시점
- 사용자 피드백 *어색하다 / 클릭하기 어렵다* 같은 모호한 신호 시
- competitor 비교 *우리가 떨어진다* 인식 시
- 분기 polish sprint (UI debt 청산)

## 3. 입력

### 필수
- 검토 대상 view / feature URL 또는 prototype
- `apply-design-system` 산출 — 적용 system + token
- `design-accessibility-baseline` — a11y 기준선

### 선택
- 사용자 피드백 (있으면 — *어디가 어색한지*)
- competitor screenshot (비교 baseline)

## 4. Stage 흐름

### Stage 1: Visual consistency

| 항목 | 점검 |
|------|------|
| Spacing | 4 / 8 px grid 정합? 임의 값 있나 |
| Color | token 사용? hardcoded hex 있나 |
| Typography | scale 정합? font family 일관 |
| Border radius | system 값만? |
| Shadow | system elevation 만? |

→ *임의 값 ≥3건* 발견 시 *system gap* — `apply-design-system` 의 token 추가 검토.

### Stage 2: Interaction polish

| 항목 | 점검 |
|------|------|
| Hover state | 모든 clickable 에 명시 |
| Active / pressed state | mouse down + touch 반응 |
| Focus state | keyboard tab 시 visible (WCAG 2.4.7) |
| Loading state | 1초+ 작업에 명시 (skeleton / spinner) |
| Empty state | 데이터 0 시 helpful 안내 (action 포함) |
| Error state | 명시적 (red + icon + 안내 + recover action) |

### Stage 3: Micro-detail

`make-interfaces-feel-better` 패턴 (외부 reference):

| detail | 효과 |
|------|------|
| Transition timing | 200~300ms ease-out (느림 / 빠름 X) |
| Input 즉시 feedback | 타이핑 → instant validation (debounce 300ms) |
| Optimistic UI | submit 직후 즉시 반영 (실패 시 rollback) |
| Skeleton 매칭 | 실제 content shape 와 같은 skeleton (jarring 회피) |
| Number formatting | locale 정합 (1,234 vs 1.234 vs 1 234) |
| Truncation | ellipsis + tooltip on hover |
| Undo 가능 | destructive action 에 5초 undo |

→ 위 7 detail 중 *적용된 것 vs 미적용* 표시.

### Stage 4: a11y 정합 (audit-accessibility cascade)

기술적 a11y 검증은 `audit-accessibility` 호출. 본 skill 은 *시각 인지된 a11y* :
- focus indicator *시각적으로* 명확 (대비 + 두께)
- error message *시각적으로* 명시 (색 only X)
- icon 의미 명확 (text label 없이도 추측 가능)

### Stage 5: Performance perception

| 항목 | 측정 |
|------|------|
| First Contentful Paint | < 1.5s (mobile 3G) |
| Time to Interactive | < 3s |
| 입력 → response | < 100ms |
| Animation FPS | ≥ 60 (jank 없음) |
| Layout shift (CLS) | < 0.1 |

→ Lighthouse / Web Vitals / `audit-performance` 도구 활용.

### Stage 6: Issue severity 분류

발견 issue 를:

| severity | 정의 | 대응 |
|---------|-----|------|
| Critical | 사용 불가 / 데이터 손실 | 즉시 fix (release block) |
| Major | 인지된 *어색* | sprint 안 fix |
| Minor | polish 영역 | backlog |
| Trivial | nit-pick | (선택) |

## 5. 산출물 형식

```markdown
## UI Quality Audit — {feature / view}

### Visual consistency
| 항목 | pass / gap |

### Interaction polish 6 state
| state | 상태 |

### Micro-detail 7 항목
| detail | 적용 |

### a11y 시각 인지
| 항목 | pass / gap |

### Performance perception
| 측정 | 값 | 목표 |

### Issue 분류 (severity)
| severity | count | items |
```

## 6. 검증

- [ ] 5 차원 모두 audit?
- [ ] Visual / interaction / micro-detail / a11y / performance 각 ≥3 항목 점검?
- [ ] Issue severity 분류?
- [ ] *임의 값 ≥3* 발견 시 system gap 명시?
- [ ] `audit-accessibility` cascade 호출 (기술적 a11y 검증)?

## 7. 다음 phase

- `apply-design-system` 의 token 추가 (system gap 발견 시)
- `audit-accessibility` (§6) — 기술적 a11y
- `monitor-regressions` — visual regression 자동 검증

## 8. 참조

- make-interfaces-feel-better (외부 reference, MIT — micro-detail 패턴)
- designer-skills/designer-toolkit (외부 reference, MIT — toolkit 패턴)
- Refactoring UI (Adam Wathan + Steve Schoger) — visual quality 원칙
- buddy `audit-accessibility` (§6, 구현됨) — 기술적 a11y cascade
