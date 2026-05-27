# Design Interaction Pattern — interaction / motion / micro-feedback 설계

## 1. 목적

UI 의 *동적 layer* — 사용자 action 에 *어떻게 반응* 하는지 설계. **gesture / motion / feedback / state transition** 4 영역 패턴 lock-in.

`apply-design-system` 의 token + `design-accessibility-baseline` 의 motion 대응 + `audit-ui-quality` 의 polish 검증과 cascade.


## Input Requirements

| Input | Required | Type | Source | 미제공 시 |
|-------|----------|------|--------|----------|
| UX 요구사항 | ✅ | knowledge | 사용자 도메인 지식 | "어떤 사용자 인터랙션을 설계하나요? (gesture, motion, feedback 등)" |

## Output Contract

| Output | Type | Format | Consumers |
|--------|------|--------|-----------|
| Interaction pattern 설계 (gesture + motion + feedback 규칙) | artifact | structured YAML | `build-feature`, `apply-design-system` |

## 2. 사용 시점

- §3 design-system 의 design 적용 stage
- 새 component / view 작성 전 — interaction 패턴 사전 결정
- 사용자 피드백 *반응이 어색하다* 시
- competitor 의 *부드러운 interaction* 비교 후

## 3. 입력

### 필수
- `apply-design-system` 산출 — system 의 default interaction
- `design-accessibility-baseline` — `prefers-reduced-motion` 대응
- target user device 패턴 (mobile / desktop — touch vs cursor)

### 선택
- 사용자 피드백
- competitor pattern (좋다고 인식된 reference)

## 4. Stage 흐름

### Stage 1: Gesture / input 패턴

| device | primary gesture | 권장 mapping |
|--------|---------------|---------|
| Desktop (mouse) | hover / click / right-click / scroll / drag | hover = intent / click = action |
| Mobile (touch) | tap / long-press / swipe / pinch | tap = action (hover X) |
| Hybrid (touch + mouse) | 둘 다 — 우선 touch | tap action 명시 |

→ *desktop-only hover* 는 mobile 에서 *동일 정보 접근 경로* 필요.

### Stage 2: Motion / animation 패턴

| 영역 | 패턴 |
|------|------|
| Page transition | fade 200ms / slide 300ms |
| Modal / sheet | slide-up 250ms ease-out |
| Tooltip | fade 150ms (hover delay 500ms) |
| Drawer | slide 300ms ease-in-out |
| Skeleton | shimmer 1.5s linear infinite |
| Loading | spinner / progress (정해진 시간 모를 때) / determinate (시간 알 때) |
| Confirmation | checkmark 400ms scale + fade |

**원칙**:
- 200~300ms 가 *기본* (느림 X / 빠름 X)
- ease-out (감속) — 사용자 시선 정착
- *prefers-reduced-motion* 활성 시 transition X 또는 fade only

### Stage 3: Feedback 패턴 (immediate response)

| action | feedback |
|------|---------|
| Button press | 100ms 안 visual 변화 (color / scale 0.98) |
| Form submit | spinner + button disable + 결과 (success / error) |
| Drag | ghost element + drop target highlight |
| Long task | progressive (예상 시간 + 진행률) |
| Optimistic UI | 즉시 반영 + 실패 시 rollback + 사용자 알림 |

**원칙**: 사용자 action 후 *100ms 안* 무엇이라도 반응. 그 이상 침묵 = "고장났나?".

### Stage 4: State transition 일관성

같은 component 의 state transition 양식 통일:

```
default → hover → active → loading → success / error → default
                                         ↓
                                      (toast / inline)
```

각 state 의 *시간 / motion* 동일.

### Stage 5: Mobile 특화 — gesture vocabulary

| gesture | 의미 (system 표준) |
|---------|----------------|
| Pull-to-refresh | 데이터 새로고침 |
| Swipe-left / -right | next / previous (또는 delete) |
| Long-press | context menu |
| Pinch | zoom |
| Edge swipe | back navigation |

→ system 표준 따름. 임의 gesture 도입 시 *명시적 onboarding* 필수.

### Stage 6: a11y motion 대응

`prefers-reduced-motion: reduce` 시:
- transition 시간 0 또는 ≤80ms
- spinner 회전 X (정적 indicator)
- parallax / autoplay X
- shake / bounce X (vestibular disorder 사용자)

→ CSS:
```css
@media (prefers-reduced-motion: reduce) {
  *, *::before, *::after {
    animation-duration: 0.01ms !important;
    transition-duration: 0.01ms !important;
  }
}
```

## 5. 산출물 형식

```markdown
## Interaction Pattern — {제품}

### Gesture mapping
| device | gesture | mapping |

### Motion 양식
| 영역 | 패턴 (시간 + ease) |

### Feedback
| action | feedback (100ms 안) |

### State transition
- default → hover → active → loading → success/error → default
- 일관 시간 / motion

### Mobile gesture vocabulary
| gesture | 의미 |

### a11y motion 대응
- prefers-reduced-motion: ...
```

## 6. 검증

- [ ] Gesture mapping device 별 명시?
- [ ] Motion 시간 200~300ms 기본 (느림 / 빠름 회피)?
- [ ] Feedback *100ms 안* 반응?
- [ ] State transition 일관 양식?
- [ ] Mobile gesture vocabulary system 표준 (임의 도입 X)?
- [ ] `prefers-reduced-motion` 대응 CSS?

## 7. 다음 phase

- `apply-design-system` 의 interaction layer 보강
- `audit-ui-quality` 의 polish 검증 입력
- `design-accessibility-baseline` 의 motion 정합

## 8. 참조

- Material Motion (Google Material 3) — motion guidelines
- Apple HIG Animation — iOS motion 표준
- designer-skills/interaction-design plugin (외부 reference, MIT — interaction 패턴)
- About Face (Alan Cooper) — interaction principle
