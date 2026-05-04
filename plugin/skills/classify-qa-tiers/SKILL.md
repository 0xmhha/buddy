---
name: classify-qa-tiers
description: "[패턴 라이브러리] QA intensity를 Quick/Standard/Exhaustive 3 tiers로 분류 + fix→commit→re-verify loop + before/after health score. 직접 invoke보다 orchestrator가 import해 사용. 트리거: 'QA 강도 분류' / '테스트 깊이 결정' / 'QA tier' / 'Quick/Standard/Exhaustive'. 참조 위치: critique-plan 후 테스트 단계, autoplan QA 결정, run-browser-qa 깊이 선택."
type: skill
---

# Snippet: QA Tiers + Fix Loop


> 강도 분류 + fix 루프 + 스코어링 패턴 — 브라우저 자동화 바이너리는 도구별로 분리.

## 이 snippet을 사용하는 경우
- 선호 브라우저 도구(Playwright, Puppeteer, Selenium 등) 위에 프로젝트별 QA 스킬 구축
- 팀 간 QA 깊이 표준화
- Fix-and-verify atomic commit 규율 참조

## 3가지 강도 Tier

### Quick Tier (~5-10분)
- **목적**: commit 전 smoke test
- **커버리지**: 크리티컬 경로만 (login, 주요 CTA, 이커머스면 checkout)
- **surfaced되는 버그 심각도**: Critical만
- **사용 시점**: Pre-commit, 빠른 피드백 필요 시

### Standard Tier (~30분)
- **목적**: PR 전 confidence
- **커버리지**: 모든 주요 flow, 3 breakpoint 반응형, 기본 접근성
- **surfaced되는 버그 심각도**: Critical, High, Medium
- **사용 시점**: PR 오픈 전

### Exhaustive Tier (~2-4시간)
- **목적**: 릴리스 전 / 대규모 refactor 후
- **커버리지**: 모든 flow, edge case, 에러 상태, 모든 breakpoint, 접근성 audit
- **surfaced되는 버그 심각도**: 모두 (Critical, High, Medium, Cosmetic)
- **사용 시점**: 프로덕션 배포 전, 아키텍처 변경 후

## Tier 선택 결정 트리

```
Pre-commit 빠른 체크? → Quick
PR 오픈? → Standard
배포 전 또는 refactor 후? → Exhaustive
확실치 않음? → Standard (기본값)
```

## 심각도 분류

| 심각도 | 정의 | 예시 |
|----------|------------|----------|
| Critical | 코어 기능 차단 | Login 깨짐, checkout 실패, 데이터 손실 |
| High | 주요 저하 | Form validation 깨짐, >5s 느림, 크리티컬 페이지 콘솔 에러 |
| Medium | 심각하지만 우회책 존재 | Edge breakpoint에서 레이아웃 깨짐, 보조 기능 깨짐 |
| Cosmetic | 시각/polish만 | 간격 어긋남, hover 상태 누락, 오타 |

## Fix → Commit → Re-Verify 루프

발견된 각 버그마다:

### Step 1: 재현
- 정확한 재현 단계 캡처
- 스크린샷 증거 (어떤 브라우저 자동화 도구든)
- 콘솔/네트워크 에러 메시지

### Step 2: Fix
- 근본 원인을 다루는 가능한 가장 작은 변경
- 한 번에 한 버그 — 번들 금지

### Step 3: Atomic commit
- Commit 메시지: `fix: <한 줄 버그 설명>`
- 본문: 재현 단계, 근본 원인, fix 접근
- 한 commit에 여러 fix 번들 금지

### Step 4: Re-verify
- 실패했던 재현을 재실행
- Fix 확인
- 인접 기능 smoke test (회귀 체크)

### Step 5: 다음 버그로 이동
- Tier 내 모든 버그가 해결되거나 escalate될 때까지 반복

### Atomic commit이 필요한 이유
- fix가 회귀를 도입하면 쉬운 revert
- 향후 디버깅용 bisectability
- 무엇을 시도했는지 명확한 히스토리

## 헬스 스코어 (Before/After)

### 계산

```
Health Score (0-10) =
  (10 - critical_bug_count * 3)        # Critical: -3 each
  - (high_bug_count * 1.5)             # High: -1.5 each
  - (medium_bug_count * 0.5)           # Medium: -0.5 each
  - (cosmetic_bug_count * 0.1)         # Cosmetic: -0.1 each
  clamped to [0, 10]
```

### 리포팅

항상 before/after delta 표시:

```
QA Health Score
  Before fixes: 4.5/10 (3 Critical, 2 High, 5 Medium)
  After fixes:  9.2/10 (0 Critical, 0 High, 1 Medium remains)
  Delta: +4.7 over 4 fix commits
```

## 출력 템플릿

```
## QA Run: <date>, <tier>

### Bugs Found
| # | Severity | Description | Status | Fix Commit |
|---|----------|-------------|--------|------------|
| 1 | Critical | Login fails on mobile Safari | Fixed | abc1234 |
| 2 | High | ...                           | Fixed | def5678 |

### Health Score
- Before: 4.5/10
- After: 9.2/10
- Delta: +4.7

### Notes
- ...
```
