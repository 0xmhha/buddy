---
name: audit-live-devex
description: "[패턴 라이브러리] 빌드/배포된 live developer product를 실제로 따라 하며 TTHW timing, evidence, literal doc-following으로 DX audit. review-devex의 live companion. 트리거(orchestrator 본문에서 호출 시): 'build 후 라이브 감사' / 'post-build DX check' / '실제 사용해보면서 점검' / 'live devex 감사' / 'deployed 상태 감사'. 참조 위치: review-devex post-build 단계 / 배포 후 검증."
type: skill
---

# Snippet: Live DX Audit Checklist


> 라이브 감사 방법론 — 도구별 브라우저 plumbing 제외.
> 짝: `review-devex` 스킬은 planning 커버; 이것은 post-build 라이브 감사 커버.

## 이 snippet을 사용하는 경우
- 자기 문서를 따라 하며 자기 제품의 온보딩 감사
- 릴리스 후보에서 TTHW 측정
- 출시 전 DX sanity 체크

당신은 **라이브 개발자 제품을 dogfood**하는 DX 엔지니어다. 계획을 리뷰하거나
경험에 대해 읽는 게 아니라 — 테스트하는 것. 문서 탐색, 실제 명령 실행,
개발자가 실제로 보는 것을 스크린샷. 추측 말고 측정.

---

## 패턴 1: TTHW 타이밍 (Time to Hello World)

**시계 시작**: "이 제품을 써보고 싶다" (문서/홈페이지 랜딩).
**시계 종료**: "첫 성공 출력" (CLI가 기대 결과 반환, API 호출이 실제 데이터로 200 반환, UI가 사용자의 첫 artifact 렌더링).

### TTHW Tier 벤치마크
| Tier | 시간 | 채택 임팩트 |
|------|------|-----------------|
| Champion | < 2분 | 3-4배 높은 채택 |
| Competitive | 2-5분 | Baseline |
| Needs Work | 5-10분 | 상당한 drop-off |
| Red Flag | > 10분 | 50-70% 포기 |

### 모든 단계, 모든 blocker를 항목화
```
GETTING STARTED AUDIT
=====================
Step 1: [what dev does]    Time: [actual]  Friction: [low/med/high]  Evidence: [screenshot/log]
Step 2: [what dev does]    Time: [actual]  Friction: [low/med/high]  Evidence: [screenshot/log]
...
TOTAL: [N steps, M minutes wall-clock]
BLOCKERS: [list every place you got stuck, what unblocked you]
```

Wall-clock은 실패를 포함. 명령을 오타치거나 prereq가 빠졌다면 — 그 시간도 포함. 실제 사용자는 re-roll 못 한다.

---

## 패턴 2: 증거 수집

증거 없는 점수는 단지 의견. 모든 주장은 artifact 필요.

### 스크린샷/로그 체크리스트
최소한 캡처:
- Getting-started flow의 **각 단계** (랜딩 페이지, 설치, 첫 명령, 첫 성공)
- 마주친 **각 에러** (전체 메시지 + 뭘 하려 했는지 + 뭐로 해결했는지)
- **각 모호한 순간** ("어떤 옵션을 고르지?", "맞는 경로인가?")
- **각 "매직 모먼트"** 그냥 동작한 곳 (좋은 게 뭔지 인식할 수 있도록)
- **각 dead-end** (문서 404, 깨진 예시, --help에 누락된 flag)

### 모든 artifact에 annotate
각 스크린샷/로그에 기록:
1. **뭘 하려 했는지** (명령이 아니라 사용자 목표)
2. **뭘 봤는지** (문자 그대로 출력 / UI 상태)
3. **뭘 기대했는지** (gap = friction signal)
4. **심각도**: blocker / friction / cosmetic
5. **Source**: TESTED (실행함) / INFERRED (파일에서 읽음) / ASSUMED (이거 하지 말 것)

모든 점수를 TESTED / PARTIAL / INFERRED로 표시. 추측 금지. 모든 dimension에 증거 출처 명시.

---

## 패턴 3: 문자 그대로 문서 따라하기

> **최악의 gap을 드러내는 규칙**: 문서를 문자 그대로 읽기. 추론 금지.
> 자기 지식으로 gap 채우지 말 것. 문서가 언급 못한 빠진 prereq를
> "당연히" 설치하지 말 것.

문서가 "패키지 설치"라고 말하면, 명명된 패키지만 설치. 빠진 peer dep 때문에 실패하면 — 그게 gap. 기록.

예시가 `client.send(message)`를 보여주지만 `client`를 어떻게 구성하는지 안 보여주면 — "당연히" import할 것을 import하지 말 것. 멈춤. 그게 gap. 기록.

Curl 예시가 키 출처 없이 `$API_KEY`를 쓰면 — 우연히 아는 대시보드로 가지 말 것. 멈춤. 그게 gap. 기록.

### 문자 그대로 따라하기가 중요한 이유
기존 지식은 DX 테스팅의 적. 오늘 도착하는 새 개발자는 당신의 관례를 모른다. "당연히 이것도 필요..."로 조용히 채운 모든 gap은 실제 첫 사용자가 전력으로 부딪히는 gap. 문자 그대로 따라하기가 이를 노출하는 유일한 방법.

---

## 에러 메시지 감사

흔한 에러 시나리오를 일부러 trigger:
- 무효 auth (키 누락, 만료 키, 잘못된 region)
- 필수 flag/arg 누락
- 잘못된 입력 (malformed JSON, 잘못된 타입, 범위 밖 값)
- 네트워크 실패 (timeout, DNS, 도달 불가 호스트)
- 권한 거부 (read-only 경로, 누락된 scope)

각 에러에 대해 3-tier 모델로 점수:
- **Tier 1 (Best)**: 문제 식별, 원인 설명, fix 표시, 문서 링크. (Elm, Rust, Stripe.)
- **Tier 2 (OK)**: 문제와 {원인, fix} 중 하나 식별.
- **Tier 3 (Pain)**: 스택 트레이스, generic "internal error", 또는 silent failure.

감사 중 마주친 모든 에러 카탈로그화. 심각도 분류:
| 심각도 | 정의 |
|----------|------------|
| Blocker | 외부 도움(Slack, GH 이슈, source dive) 없이 개발자가 진행 불가 |
| Friction | 개발자가 해결하지만 2-10분 손실 |
| Cosmetic | 잘못된 단어, 포맷팅, 대문자 — 시간 손실 없음 |

---

## 스코프 선언

감사가 무엇을 커버했고 안 했는지 명시적. 과장 금지.

라이브 테스트 보통 가능한 웹 접근 surface: 문서 페이지, API 플레이그라운드, 웹 대시보드, 가입 flow, 대화형 튜토리얼, 에러 페이지.

별도 감사가 흔히 필요한 것: 깨끗한 머신의 CLI 설치 friction, 터미널 출력 품질, 로컬 환경 설정, 이메일 검증 flow, 실제 credential 필요 auth, 오프라인 동작, 빌드 시간, IDE 통합.

테스트 불가 dimension: artifact (README, CHANGELOG, --help 출력) 읽고 점수를 INFERRED로 표시. 추측 금지. 모든 점수에 증거 출처 명시.

---

## 라이브 감사 스코어카드

```
+====================================================================+
|              DX LIVE AUDIT — SCORECARD                              |
+====================================================================+
| Dimension            | Score  | Evidence       | Method    |
|----------------------|--------|----------------|-----------|
| Getting Started      | __/10  | [screenshots]  | TESTED    |
| API/CLI/SDK          | __/10  | [screenshots]  | PARTIAL   |
| Error Messages       | __/10  | [screenshots]  | PARTIAL   |
| Documentation        | __/10  | [screenshots]  | TESTED    |
| Upgrade Path         | __/10  | [file refs]    | INFERRED  |
| Dev Environment      | __/10  | [file refs]    | INFERRED  |
| Community            | __/10  | [screenshots]  | TESTED    |
| DX Measurement       | __/10  | [file refs]    | INFERRED  |
+--------------------------------------------------------------------+
| TTHW (measured)      | __ min | [step count]   | TESTED    |
| Overall DX           | __/10  |                |           |
+====================================================================+
```

### Plan-vs-Reality (plan-time DX 리뷰 존재 시)
**live score < plan score - 2**인 모든 dimension 플래그. 그게 plan에서 현실이 못 미친 곳 — 설계와 ship 사이 가정이 깨진 곳. 다음 릴리스 전에 조사.
