# Prototype From Spec — 명세 → prototype 작성 (디자인 단계)

## 1. 목적

`define-feature-spec` / PRD 의 명세를 *시각화 prototype* 으로 변환. **low-fidelity → high-fidelity** 단계 적용. 사용자 검증 / dev 핸드오프 입력.

`apply-design-system` 채택된 system 의 component / token 활용 — 처음부터 *system 정합* prototype.


## Input Requirements

| Input | Required | Type | Source | 미제공 시 |
|-------|----------|------|--------|----------|
| Feature spec 또는 설명 | ✅ | artifact / knowledge | `define-feature-spec` 산출물 또는 사용자 설명 | "어떤 기능의 프로토타입을 만드나요?" |

## Output Contract

| Output | Type | Format | Consumers |
|--------|------|--------|-----------|
| 프로토타입 (탐색용 코드 또는 디자인) | artifact | 코드 / HTML / 디자인 파일 | `review-design` |

## 2. 사용 시점

- §3 design-system 의 design 적용 stage
- 새 feature 명세 후 *시각 검증* 단계
- 사용자 user testing 전 — testable prototype
- dev 핸드오프 직전 — spec → 시각 → 코드 cascade
- pivot 시 *빠른 시각 검증* (low-fi 우선)

## 3. 입력

### 필수
- `define-feature-spec` 산출 (acceptance criteria + interaction)
- `apply-design-system` 산출 (system / token)
- target user (`map-customer-segments`)

### 선택
- 기존 prototype 또는 production view (있으면 — 일관성 유지)
- 사용자 피드백 (이전 user testing)

## 4. Stage 흐름

### Stage 1: Low-fi prototype (wireframe)

목적: *layout / 정보 hierarchy / flow* 검증. 디테일 X.

| 도구 후보 | 용도 |
|--------|------|
| Figma / FigJam | 협업 |
| Excalidraw / tldraw | 빠른 sketch |
| Whimsical | flow + wireframe |
| 종이 + 사진 | 가장 빠른 |

산출:
- 주요 view 5~10 개
- 각 view 의 *주요 element* (header / content / action)
- view 간 flow 화살표

→ *내부 review* (1~2일) 로 충분. 사용자 testing 은 high-fi 단계.

### Stage 2: User flow + state diagram

각 critical user journey 의 *state* 명시:

| state | trigger | next |
|-------|--------|------|
| empty | 첫 접속 | onboarding 또는 sample data |
| loading | data fetch | content / error |
| content | data 도착 | (사용자 action) |
| error | network / server | retry / fallback |
| success | action 완료 | confirmation / next view |

→ 모든 state 가 prototype 에 *명시* 됐는지 검증 (state 누락 = production bug 자주).

### Stage 3: High-fi prototype (visual + interaction)

low-fi 통과 후 design system 적용:

| 영역 | 적용 |
|------|------|
| Component | apply-design-system 의 system component 만 사용 |
| Token | color / spacing / typography 모두 token |
| Interaction | hover / active / focus / loading / empty / error 6 state |
| Animation | transition timing 200~300ms ease-out |

→ 사용자 testing 가능 수준 (clickable prototype).

### Stage 4: User testing (선택)

5 user (Jakob Nielsen — usability 80% 발견율) 대상:
- task 기반 (e.g. "X 를 해보세요")
- think-aloud 기록
- 발견된 issue 분류 (critical / major / minor)

→ `conduct-customer-interview` skill 의 *prototype 검증 변형*.

### Stage 5: Dev handoff

high-fi prototype → 코드 핸드오프:

| 자료 | 도구 |
|------|------|
| Inspect (CSS) | Figma Dev mode / Zeplin |
| Component link | Storybook integration |
| a11y annotation | `design-accessibility-baseline` 정합 |
| Asset export | 1× / 2× / 3× / vector |

→ dev 가 *추측 없이* 구현 가능한 수준.

## 5. 산출물 형식

```markdown
## Prototype — {feature}

### Low-fi (wireframe)
- 도구: ...
- view list: ...

### State diagram
| state | trigger | next |

### High-fi
- design system: ...
- token 적용: ...
- 6 interaction state 모두 명시

### User testing (선택)
- N=5 결과: critical {n} / major {n} / minor {n}

### Dev handoff
- inspect link: ...
- a11y annotation: ...
- asset export: ...
```

## 6. 검증

- [ ] Low-fi 우선 (high-fi 직진 X)?
- [ ] State 5+ (empty / loading / content / error / success) 모두 prototype?
- [ ] High-fi 가 design system 정합?
- [ ] 6 interaction state 모두 명시?
- [ ] Dev handoff 자료 (inspect + a11y + asset) 완비?
- [ ] (선택) User testing 5 명 진행?

## 7. 다음 phase

- Dev handoff → §5 build-feature
- User testing 결과 → `define-feature-spec` 의 acceptance criteria 정정
- `audit-ui-quality` — 구현 후 prototype vs 실제 정합 검증

## 8. 참조

- About Face (Alan Cooper) — interaction design
- Don't Make Me Think (Steve Krug) — usability testing
- Jakob Nielsen — 5-user testing 80% 발견율
- designer-skills/prototyping-testing (외부 reference, MIT) — prototype + testing 패턴
