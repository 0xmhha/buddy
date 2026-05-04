---
name: status
description: "This skill should be used when the user asks '지금 어디야?' / '다음에 뭐 해?' / '어떤 커맨드 써야 해?' / '진행 상황 알려줘' / '어느 phase야?' / 'what command next?' / 'where are we?' / 'show commands' / '커맨드 목록'. Detects the current lifecycle phase from available artifacts, shows what's done, what's blocked, and which /buddy command to run next."
type: skill
---

# Status — 현재 Phase + 다음 커맨드 안내

사용자에게 **"지금 어디 있는지"**와 **"다음에 어떤 커맨드를 쓸지"**를 알려준다.

---

## 실행 절차

### Step 1. Artifact 탐지로 현재 Phase 추론

아래 순서대로 artifact 존재 여부를 확인해 현재 phase를 판단:

| 탐지 artifact | 추론 phase |
|--------------|-----------|
| `docs/actor-track-plan.yaml` 존재 + 미완료 task | §5 build-feature |
| `docs/actor-track-plan.yaml` 존재 + 모든 task 완료 | §6 verify-quality |
| `docs/tech-spec.md` 또는 `docs/design/` 존재 | §4 plan-build |
| `docs/feature-spec/` 또는 `docs/features.yaml` 존재 | §3 design-system |
| `docs/prd.md` 또는 `docs/PRD.md` 존재 | §2 define-features |
| 위 없음 + idea/concept만 언급 | §1 concretize-idea |
| `dist/` 또는 `CHANGELOG.md` 존재 + release tag | §7 이후 |
| 다수 존재 + production traffic 언급 | §8 iterate-product |

탐지 불가 시: "현재 phase를 특정할 수 없어. 어느 단계에 있는지 알려줘."

### Step 2. 중단 원인 식별 (선택)

현재 phase에서 진행이 멈춘 이유가 있으면 같이 표시:
- 인간 의사결정 필요: `[결정 대기]` 표시
- 선행 artifact 부재: `[선행 작업 필요]` 표시
- 코드 품질 문제: `[Quality gate 미통과]` 표시

### Step 3. 출력

아래 형식으로 출력:

```
## 현재 위치

Phase: §N <phase-name>
근거: <탐지된 artifact>
상태: <진행 중 / 결정 대기 / 완료 대기>

## 지금 바로 실행할 커맨드

/buddy:<current-command>  ← 지금 이걸 실행해

## 완료 후 다음 단계

/buddy:<next-command>  ← <current> 완료 후 이걸 실행해

## 전체 커맨드 맵

┌─────────────────────────────────────────────────────────────────────┐
│  Phase  │ 커맨드                  │ 용도                            │
├─────────┼─────────────────────────┼─────────────────────────────────┤
│ §1      │ /buddy:concretize-idea  │ Idea → PRD + 사업성 검증        │
│ §2      │ /buddy:define-features  │ Feature backlog + actor 정의    │
│ §3      │ /buddy:design-system    │ 기술 설계 + API 계약             │
│ §4      │ /buddy:plan-build       │ 구현 계획 (actor-track plan)    │
│ §5      │ /buddy:build-feature    │ TDD 개발 + 병렬 agent dispatch  │
│ §6      │ /buddy:verify-quality   │ 품질 gate (test/lint/security)  │
│ §7      │ /buddy:ship-release     │ PR + 릴리즈 태깅 + changelog     │
│ §8      │ /buddy:iterate-product  │ A/B 분석 + 인시던트 + 개선 루프  │
│ §9      │ /buddy:manage-lifecycle │ Feature/product 수명주기 관리   │
├─────────┼─────────────────────────┼─────────────────────────────────┤
│ 크로스  │ /buddy:autoplan         │ 어느 phase의 plan이든 리뷰       │
│         │ /buddy:diagnose-bug     │ 버그 재현 → 원인 → fix           │
│         │ /buddy:audit-security   │ 보안 취약점 점검                 │
│         │ /buddy:auto-create-pr   │ feature → PR 자동 생성          │
│         │ /buddy:build-with-tdd   │ TDD 단독 실행                   │
└─────────┴─────────────────────────┴─────────────────────────────────┘

현재 phase: §N ← 여기
```

현재 phase 행에 `← 여기` 마커를 붙인다.

---

## 의사결정 대기 상황 안내

사용자가 의사결정 후 이어가야 할 때:

```
## 의사결정 필요

[무엇]: <결정 내용>
[선택지]:
  A) <선택지 A> → /buddy:<next-A> 실행
  B) <선택지 B> → /buddy:<next-B> 실행

결정 후 위 커맨드를 실행해줘.
```

---

## 구현 재개 안내

중단된 구현 작업을 이어갈 때:

```
## 구현 재개

중단 지점: <마지막 완료 task>
남은 작업: <미완료 task 목록>

이어가려면: /buddy:build-feature
```

---

## 주의

- 이 스킬은 **읽기 전용**. 파일을 수정하거나 커맨드를 실행하지 않는다.
- phase 추론이 불확실하면 단정하지 말고 사용자에게 묻는다.
- 의사결정 사항은 자동 결정하지 않는다.
