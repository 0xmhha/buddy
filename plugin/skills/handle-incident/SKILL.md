---
name: handle-incident
description: This skill should be used when the user says "we have an incident", "production is down", "handle this outage", "something is broken in production", "users are reporting errors", or needs to coordinate an incident response in a structured way.
---

# handle-incident

프로덕션 인시던트를 구조적으로 대응한다. 영향 범위 확인 → 즉각 완화 → 근본 원인 조사 → Fix 배포 → 고객 커뮤니케이션 순서를 따른다.

**§8 iterate-product stage skill.** 단독 호출도 가능 (dual-mode).

---

## 인시던트 대응 런북

### Step 1: 심각도 분류 (5분 내)

| 심각도 | 기준 | 대응 시간 |
|--------|------|---------|
| P0 (Critical) | 전체 서비스 중단 / 데이터 손실 | 즉시 |
| P1 (Major) | 핵심 기능 중단 (결제, 로그인 등) | 15분 내 |
| P2 (Minor) | 기능 일부 저하, 우회 가능 | 1시간 내 |
| P3 (Low) | 성능 저하, 사용자 영향 최소 | 4시간 내 |

### Step 2: 영향 범위 확인 (10분 내)

§2 use case 분해 기반으로 어느 actor가 영향받는지 확인:

```
영향 범위 체크:
- user-actor: 몇 % 사용자가 영향받는가? (geography / segment)
- backend-actor: 어느 service / endpoint가 failing인가?
- 3rd-party-actor: 외부 SaaS 장애인가? (status page 확인)

데이터:
- Error rate: {현재} vs {baseline}
- Affected users: {N}명 (전체의 {%})
- Started: {timestamp}
```

### Step 3: 즉각 완화 (15분 내)

**옵션 A: Feature Flag Off**
```bash
# Feature flag로 문제 기능 비활성화
feature_flag disable {feature_name} --env production
```

**옵션 B: Rollback**
```bash
# 이전 정상 버전으로 롤백
git revert {broken_commit}
# 또는
deployment rollback --to {previous_version}
```

**옵션 C: Rate Limiting**
```bash
# 부하 원인이면 rate limit 강화
nginx rate_limit_zone update --limit 10r/s
```

`guard-destructive-commands` skill을 invoke해 완화 작업 전 위험 명령을 가드한다.

### Step 4: 근본 원인 조사

`diagnose-bug` skill을 invoke해 재현 가능한 원인을 찾는다:
1. Minimized repro 생성
2. Hypothesis 목록 (actor별로 분류)
3. Targeted instrumentation으로 hypothesis 구분
4. Root cause 확정

### Step 5: Fix 배포

`build-feature` (§5) → `verify-quality` (§6 minimal) → `ship-release` (§7.3 fast path)

P0/P1의 경우 fast path:
- unit test만 실행 (E2E skip 가능)
- 1-click deploy (rollback 준비 상태)
- fix 후 즉시 full test suite 실행

### Step 6: 고객 커뮤니케이션

```markdown
## 인시던트 커뮤니케이션 템플릿

### 발생 시 (즉시)
"[서비스명]에서 [기능]에 영향을 미치는 이슈를 인지하고 있습니다. 
현재 조사 중이며 업데이트를 제공하겠습니다."

### 진행 중 (30분마다)
"업데이트: [현재 상태]. 원인 조사 중입니다. 다음 업데이트: [시간]."

### 해결 후
"[기능]이 복구되었습니다. [영향 기간]. 자세한 내용은 포스트모템에서 공유하겠습니다."
```

---

## 인시던트 기록 포맷

```yaml
incident:
  id: INC-{YYYYMMDD}-{N}
  severity: {P0/P1/P2/P3}
  started_at: {ISO 8601}
  resolved_at: {ISO 8601}
  duration: {분}
  
  affected_actors:
    - actor: user-actor
      impact: "{N}% 사용자 영향"
    - actor: backend-actor
      impact: "{service} {error_rate}% error rate"
  
  mitigation: {완화 조치}
  root_cause: {근본 원인 — conduct-postmortem 이후 채움}
  fix: {수정 내용}
  
  timeline:
    - time: {T+0}
      event: "이슈 감지"
    - time: {T+N분}
      event: "완화 적용"
    - time: {T+N분}
      event: "해결"
```

---

## 다음 단계

- `conduct-postmortem` — 인시던트 종료 후 48시간 내
- `generate-improvement-tasks` — postmortem action items → backlog
