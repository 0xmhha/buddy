# Design Accessibility Baseline — WCAG 2.2 AA + a11y annotation 컨벤션

## 1. 목적

**WCAG 2.2 Level AA** 기준의 *baseline accessibility* 설계. 4 원칙 (POUR — Perceivable / Operable / Understandable / Robust) 의 success criteria 를 디자인 / 개발 / QA 단계에 사전 통합.

`apply-design-system` + `audit-ui-quality` 와 cascade. `audit-accessibility` (§6, 구현됨) 의 입력.

## 2. 사용 시점

- §3 design-system 안에서 디자인 시스템 적용 직전
- 새 component / view 작성 전 — 디자인 단계 a11y 사전 점검
- `audit-accessibility` 가 *baseline 미달* 보고 시 — 본 skill 의 baseline 재검토
- 법률 / 규제 (ADA / EAA — 유럽 Accessibility Act / 한국 장애인차별금지법) 진출 시

## 3. 입력

### 필수
- `apply-design-system` 산출 — 사용 component library
- target user (`map-customer-segments`) — 장애 사용자 segment
- 사용 frontend 프레임워크 (`define-tech-stack`)

### 선택
- 기존 audit 결과 (있으면 — gap 식별 baseline)
- 법률 요구 (ADA / EAA / 한국 KS X 9211)

## 4. Stage 흐름

### Stage 1: WCAG 2.2 AA — POUR 4 원칙

| 원칙 | 핵심 success criteria (AA) |
|------|---------------------|
| **Perceivable** (지각 가능) | text alternative (1.1.1), 색 대비 ≥4.5:1 (1.4.3), 크기 조정 가능 (1.4.4) |
| **Operable** (운용 가능) | keyboard 전용 navigation (2.1.1), focus visible (2.4.7), seizure 회피 (2.3.1) |
| **Understandable** (이해 가능) | page language (3.1.1), input label (3.3.2), error identification (3.3.1) |
| **Robust** (견고) | parsing valid (4.1.1), name/role/value (4.1.2), status messages (4.1.3) |

→ AA 수준 success criteria 30+ 모두 적용. AAA 는 선택 (특정 영역 — 의료 / 금융 / 정부).

### Stage 2: a11y annotation 컨벤션

디자인 단계에 *annotation* 추가 (Figma / Sketch / Penpot):

| annotation | 의미 | 예시 |
|-----------|-----|------|
| `aria-label` | screen reader 음성 | "닫기 버튼" |
| `role` | semantic role | button / dialog / nav |
| `tabindex` | keyboard 순서 | 0 / -1 |
| `aria-live` | 동적 변화 알림 | polite / assertive |
| heading hierarchy | h1 → h6 순서 | 단일 h1, level skip 금지 |

→ 디자인 핸드오프 시 annotation 포함. 개발자가 *추측* 안 하게.

### Stage 3: Component-level baseline

각 component 에 *기본 a11y 속성* 표준화:

| component | 기본 a11y |
|-----------|---------|
| Button | role="button" + aria-label (icon-only 시) + keyboard activation (Enter/Space) |
| Modal/Dialog | role="dialog" + aria-modal="true" + focus trap + ESC close |
| Input | label associated (htmlFor or aria-labelledby) + aria-invalid + aria-describedby (error) |
| Tab | role="tablist" + aria-selected + arrow key navigation |
| Tooltip | role="tooltip" + aria-describedby + ESC dismiss |

→ component library 에 *기본 적용*. 개발자가 *추가* 노력 0.

### Stage 4: Color + 시각

- **대비**: WCAG AA = 4.5:1 (텍스트 일반) / 3:1 (large text 18pt+) / 3:1 (UI element)
- **색만으로 정보 전달 X**: error 는 빨간색 + icon + 텍스트 동시
- **focus indicator**: 2px outline + 색 대비 3:1+
- **motion**: prefers-reduced-motion 대응 (CSS media query)

### Stage 5: Keyboard / screen reader 검증

- **Keyboard**: Tab / Shift+Tab / Enter / Space / Arrow / ESC 모두 동작
- **Screen reader**: NVDA (Windows) / VoiceOver (macOS/iOS) / TalkBack (Android) 검증
- **Focus order**: 시각 순서 == 논리 순서

### Stage 6: CI 자동화 (audit-accessibility cascade)

본 skill 의 baseline 을 *코드 레벨 lint* 로:
- ESLint + jsx-a11y plugin
- axe-core (자동 a11y 검사) CI 통합
- Storybook a11y addon (component 단위)

→ `audit-accessibility` (§6) 에서 *runtime audit* 수행.

## 5. 산출물 형식

```markdown
## Accessibility Baseline — {제품}

### Target standard
- WCAG 2.2 Level AA + (선택) AAA 영역: ...

### POUR success criteria 적용
| 원칙 | criteria | 적용 도구 / 검증 |

### Component-level baseline
| component | 기본 a11y 속성 |

### Color + 시각
- 대비 비율: ...
- focus indicator: ...
- motion 대응: ...

### Keyboard + screen reader
- 검증 도구: ...

### CI 자동화
- lint: ...
- runtime audit: ...
```

## 6. 검증

- [ ] WCAG 2.2 AA criteria 30+ 모두 적용 plan?
- [ ] Component-level baseline 5+ component 정의?
- [ ] Color 대비 4.5:1 / focus indicator 3:1+?
- [ ] Keyboard 전용 navigation 가능 (mouse 없이도)?
- [ ] CI 자동화 (lint + axe-core + Storybook addon)?
- [ ] 법률 요구 (ADA / EAA / KS X 9211) 적용 시 추가 criteria 포함?

## 7. 다음 phase

- `apply-design-system` 의 component library a11y 보강
- `audit-accessibility` (§6) 의 runtime 검증 입력
- `audit-ui-quality` 의 visual + interaction 정합 검증

## 8. 참조

- WCAG 2.2 specification (W3C)
- ARIA Authoring Practices Guide (W3C APG)
- Inclusive Components (Heydon Pickering)
- designer-skills/interaction-design (외부 reference, MIT — a11y 패턴 보조)
- buddy `audit-accessibility` (§6, 구현됨) — runtime cascade
