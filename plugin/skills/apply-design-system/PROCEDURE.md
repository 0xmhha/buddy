# Apply Design System — 기존 design system 채택 / 적용

## 1. 목적

*신규 design 작성* (= `prototype-from-spec`) 이 아닌 **기존 design system 채택** 의 절차. token / component / pattern / a11y baseline 의 일관 적용.

`decide-form-factor-app-vs-web` 산출 후 form factor 별 design system 후보 선정 + 채택 + 운영 정합.

## 2. 사용 시점

- §3 design-system 의 form factor 결정 후 design system 채택 직전
- 기존 design system 변경 / migration 결정 시
- 새 component / view 작성 전 — *기존 token / pattern* 활용 검증
- 신규 designer / developer 온보딩 — 채택된 system 학습

## 3. 입력

### 필수
- `decide-form-factor-app-vs-web` 산출 — form factor 별 system 후보 narrow
- `design-accessibility-baseline` — a11y 정합 검증
- 브랜드 아이덴티티 (있으면 — color / typography / voice)

### 선택
- 기존 design 자산 (있으면 — migration plan)
- 라이선스 / 비용 budget (proprietary 시)

## 4. Stage 흐름

### Stage 1: Design system 후보 매트릭스

form factor 별:

| form factor | 후보 | 라이선스 | a11y baseline |
|-----------|-----|--------|------------|
| Web | shadcn/ui / Radix UI / MUI / Ant Design / Chakra | MIT (대부분) | 양호 (radix / shadcn) |
| iOS | Apple HIG + SwiftUI | proprietary (Apple) | 강제 |
| Android | Material 3 + Jetpack Compose | Apache (Google) | 강제 |
| React Native | NativeBase / Tamagui / RN Paper | MIT | 양호 |
| Desktop (Electron) | shadcn/ui / Fluent UI / Carbon | MIT | 다양 |

### Stage 2: 4 차원 평가 + 채택

| 차원 | 측정 |
|------|------|
| Component coverage | 필요 component 가 system 안에 있나 (빠진 것은 직접 작성 비용) |
| Customization 자유도 | token 변경 / 새 variant 추가 용이성 |
| a11y baseline | WCAG AA 준수 (`design-accessibility-baseline` 정합) |
| 활성도 / 커뮤니티 | 마지막 commit / issue response / 사용자 수 |

→ 1~2 후보 narrow 후 *prototype 빠르게* (1~2일) 만들고 결정.

### Stage 3: Token 정합

design system 채택 시 *token* (color / spacing / typography / radius / shadow) 을 *brand* 와 정합:

| token 종류 | 결정 |
|---------|------|
| Color | brand primary + system color (semantic — success / error / warning / info) |
| Typography | font family + scale (h1 ~ caption) |
| Spacing | 4px / 8px grid |
| Radius | 0 / 2 / 4 / 8 / full |
| Shadow | 0 / 1 / 2 / 3 elevation |

→ token 변경은 system 의 update 와 정합 (override file).

### Stage 4: Pattern library

자주 쓰는 *조합 pattern* 정의 (system 의 component 보다 상위):

| pattern | component 조합 |
|---------|-------------|
| Form layout | Label + Input + Helper + Error |
| Empty state | Icon + Title + Description + Action |
| Confirmation dialog | Modal + Title + Body + Cancel/Confirm |
| Data table | Table + Filter + Pagination + Empty |

### Stage 5: 운영 — adoption tracking

채택 후 *실제 사용* 추적:
- 새 view 작성 시 *system 사용 비율* (vs 직접 CSS)
- *system 외 component* 발생 시 PR review 에서 *system 흡수 검토*

→ adoption 비율 80%+ 권장. 50% 이하면 *system mismatch* 신호.

## 5. 산출물 형식

```markdown
## Design System Adoption — {제품}

### 채택 system
- form factor: {web / iOS / ...}
- system: {shadcn / MUI / ...}
- 라이선스: {MIT / ...}
- 버전 lock: {v X.Y}

### Token 정합
| token | 값 | brand 정합 |

### Pattern library
| pattern | component 조합 |

### Adoption tracking
- 측정 방법: ...
- 목표 비율: 80%+
```

## 6. 검증

- [ ] form factor 별 후보 매트릭스 ≥3?
- [ ] 4 차원 평가 + 결정?
- [ ] Token 5 종 (color / typography / spacing / radius / shadow) 정합?
- [ ] Pattern library 4+ pattern?
- [ ] a11y baseline 정합 (`design-accessibility-baseline`) ?
- [ ] Adoption tracking 방법 명시?

## 7. 다음 phase

- `audit-ui-quality` — system 적용의 *디테일 품질* 검증
- `prototype-from-spec` — 새 view / feature 작성 시 system 사용
- `design-interaction-pattern` — interaction layer 의 system 정합

## 8. 참조

- shadcn/ui (Anthony Castelo) — token-driven web component library
- Apple HIG / Material 3 — native form factor 표준
- designer-skills/design-systems plugin (외부 reference, MIT) — adoption pattern
