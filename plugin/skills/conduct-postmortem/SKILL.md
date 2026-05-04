---
name: conduct-postmortem
description: This skill should be used when the user wants to "conduct a postmortem", "write an incident postmortem", "do a blameless postmortem", "analyze what went wrong", or has resolved an incident and needs to do a structured retrospective to prevent recurrence.
---

# conduct-postmortem

인시던트 종료 후 비난 없는(blameless) 포스트모템을 진행한다. 타임라인 재구성 → 5 Whys → 재발 방지 action items 순서로 진행한다.

**§8 iterate-product stage skill.** 단독 호출도 가능 (dual-mode).
**진행 시점**: 인시던트 해결 후 24-72시간 내. 기억이 생생할 때.

---

## 포스트모템 원칙

1. **비난 없음**: 사람이 아닌 시스템 / 프로세스를 분석한다.
2. **사실 기반**: 추측이 아닌 로그 / 타임라인 데이터 기반.
3. **학습 중심**: 처벌보다 재발 방지 action items.
4. **공개 공유**: 팀 전체가 배울 수 있도록 공유.

---

## 포스트모템 진행 절차

### 1. 타임라인 재구성

로그, 모니터링 데이터, Slack 메시지를 기반으로 타임라인을 재구성한다:

```markdown
| 시각 | 이벤트 | 담당 |
|------|--------|------|
| T+0  | 이슈 최초 감지 (alert) | monitoring |
| T+5  | on-call 인지 | {이름} |
| T+15 | 완화 조치 적용 | {이름} |
| T+60 | 근본 원인 확인 | {이름} |
| T+90 | Fix 배포 | {이름} |
| T+95 | 서비스 정상화 확인 | {이름} |
```

### 2. 영향 범위 정량화

```yaml
impact:
  duration: {분}
  affected_users: {N} (전체의 {%})
  error_requests: {N}건
  sla_violation: {yes/no}
  estimated_revenue_impact: ${N}
  actor_breakdown:
    user_actor: "{N}% 사용자 — {기능} 불가"
    backend_actor: "{service} — {error_rate}% failure"
```

### 3. 5 Whys 분석

표면적 원인에서 근본 원인까지 파고든다:

```
증상: {관찰된 이슈}

Why 1: 왜 증상이 발생했는가?
→ {원인 1}

Why 2: 왜 원인 1이 발생했는가?
→ {원인 2}

Why 3: 왜 원인 2가 발생했는가?
→ {원인 3}

Why 4: 왜 원인 3이 발생했는가?
→ {원인 4}

Why 5: 왜 원인 4가 발생했는가?
→ {근본 원인}
```

예시:
```
증상: 회원가입 API 500 에러 급증

Why 1: 왜 500 에러인가?
→ DB connection pool 고갈

Why 2: 왜 connection pool이 고갈됐는가?
→ 특정 쿼리가 long-running transaction 점유

Why 3: 왜 long-running transaction인가?
→ index 없는 컬럼에 full table scan

Why 4: 왜 index가 없었는가?
→ migration 스크립트에서 index 추가가 누락됨

Why 5: 왜 누락됐는가?
→ 코드 리뷰에서 migration 파일을 별도 체크하지 않음

근본 원인: migration 파일 review 프로세스 부재
```

### 4. Contributing Factors

근본 원인 외 기여 요인을 나열한다 (비난 없이):
- 탐지 지연: 어떤 모니터링이 있었다면 더 빨리 잡았을까?
- 완화 지연: 더 빠른 rollback이 가능했다면?
- 범위 확대: 격리 실패 이유는?

### 5. Action Items 생성

```yaml
action_items:
  - id: PA-001
    title: "migration 파일에 index 체크리스트 추가"
    owner: {이름}
    due: {날짜}
    type: process  # process / tooling / monitoring / documentation
    prevents: {재발 방지 대상}
    
  - id: PA-002
    title: "DB connection pool exhaustion alert 추가"
    owner: {이름}
    due: {날짜}
    type: monitoring
    prevents: "pool 고갈 탐지 지연"
    
  - id: PA-003
    title: "slow query threshold alert 설정"
    owner: {이름}
    due: {날짜}
    type: monitoring
    prevents: "long-running query 조기 탐지"
```

---

## 포스트모템 문서 포맷

```markdown
# 포스트모템 — {incident_id}

**날짜**: {작성일}
**심각도**: {P0/P1/P2/P3}
**소요 시간**: {분}
**작성자**: {이름}

## 요약
{1-2문장 요약}

## 영향
{영향 범위 정량화}

## 타임라인
{재구성된 타임라인}

## 근본 원인
{5 Whys 분석 결과}

## Contributing Factors
{기여 요인 목록}

## 해결 방법
{실제 적용한 fix}

## Action Items
{action_items 목록}

## 교훈
{팀 전체가 배울 수 있는 핵심 인사이트 2-3개}
```

---

## 다음 단계

- `generate-improvement-tasks` — action items → improvement backlog 생성
- `persist-learning-jsonl` — 포스트모템 인사이트를 learning store에 저장
