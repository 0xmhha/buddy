---
name: summarize-retro
description: "git history를 evidence-based weekly retrospective로 변환 — work types, hotspots, focus score, AI collaboration, person별 praise. 트리거: '주간 retro 정리' / '이번 주 회고' / 'commit 분석 retro' / '팀 회고' / 'weekly retro 작성' / '회고 만들어줘' / '이번 sprint 회고'. 입력: git log, commit history, 이슈 트래커. 출력: 회고 보고서 (성과/문제/배움/액션). 흐름: monitor-regressions/persist-learning-jsonl → summarize-retro → triage-work-items."
type: skill
---

# Retro — 증거 기반 주간 회고


Raw git history를 측정 가능 신호(commit work type, 작업 세션, focus score, 파일 hotspot, AI 협업 비율, 사람별 기여)로 변환해 종합 엔지니어링 회고 생성.

차별화는 출력의 모든 feedback 줄에 적용되는 하나의 규칙: **칭찬과 성장 노트는 특정 commit이나 PR 인용 필수. Generic 언어 없음. Ever.** "이번 주 훌륭한 작업"이라 말하는 retro는 아무것도 안 가르침. "`auth/middleware.ts`의 3 nested callback을 composable 미들웨어로 추출 (commit a3f8b2)"이라 말하는 retro는 엔지니어에게 정확히 반복할 것과 다음 리뷰 사이클에 가리킬 구체 것을 줌.

## 이 스킬을 사용하는 경우
- 주간 cadence (월요일 아침 planning, 금요일 오후 wrap)
- Sprint 완료 후
- 마일스톤 또는 릴리스 완료 후
- "이번 주 뭐 ship?" 또는 "주간 retro" 요청
- 인상이 아니라 패턴 보고 싶은 모든 윈도우 후

---

## 핵심 규칙: 증거 기반 Feedback

**모든 칭찬과 모든 성장 노트는 commit, PR, 파일 명명 필수. Generic 언어 없음. Ever.**

이게 이 스킬의 단일 가장 중요한 규칙. 칭찬, 성장 영역, 패턴, 추천 모두에 적용. Commit 명명 불가면 주장 불가.

### Bad vs. Good — 칭찬

| Bad (generic, 안 가르침) | Good (구체, actionable) |
|---|---|
| "이번 주 훌륭한 refactoring 작업!" | "`auth/middleware.ts` (commit `a3f8b2`)의 훌륭한 refactor — 3 nested callback을 composable 미들웨어로 추출. 새 shape가 auth chain을 처음으로 isolation에서 테스트 가능하게." |
| "좋은 테스트 coverage." | "PR #482에서 `payment/processor.ts`에 14 edge-case 테스트 추가, currency 반올림, retry idempotency, partial-refund flow 커버. Retry 테스트가 중복 청구 유발할 off-by-one catch." |
| "강한 feature shipping 주." | "사용자 대상 feature 3개 end-to-end ship: 프로필 avatar (PR #471), CSV export (PR #475), 다크 모드 토글 (PR #479). 세 개 모두 같은 PR의 테스트와 문서와 함께 landing." |

### Bad vs. Good — 성장 영역

| Bad (모호, action 불가) | Good (구체, level up 가능) |
|---|---|
| "테스트 coverage 개선 가능." | "이번 주 `payment/processor.ts`에 동작 추가한 commit 8, 테스트 추가는 1만 (commit `f9c1d3`가 happy 경로 커버; edge case 테스트 없음). 다음 주: 모든 동작 commit을 같은 PR의 최소 한 테스트 commit과 페어." |
| "더 작은 PR 시도." | "PR #468이 31 파일에 걸쳐 2,400 LOC (refactor + feature + test infra 혼합). Reviewer가 한 번에 로드 너무 커서 4일 뒤 코멘트. 미래 작업을 <500 LOC PR로 분할 시도." |
| "Fix-chain 패턴 주시." | "6시간에 3 fix commit이 `notifications/sender.ts` 건드림 (commit `b1`, `c4`, `d7`) — commit `a0`의 원래 변경이 SMS 경로 exercise 없이 ship한 것으로 보임. 거기 다음 변경에 SMS 테스트 함께 추가." |

### 중요 이유
Generic feedback은 unfalsifiable. 수령자가 동의, 반대, 반복, 개선 불가. 구체 feedback이 엔지니어가 자기 retro, 1:1, 승진 packet에 가리킬 구체 artifact 생성. 구체 feedback은 또한 retro 작성자가 인상을 투사하는 게 아니라 실제 데이터 보도록 강제.

**Generic 라인 쓰는 자신을 발견하면 멈추고 commit 가져오기. Anchor할 commit 없으면 라인 삭제.**

---

## 워크플로우

### Step 0: 시간 윈도우 정의

기본: 지난 7일. 다른 흔한 윈도우:
- `24h` — 어제 작업
- `14d` — 2주 sprint
- `30d` — 월
- `since-last-retro` — 이전 retro 파일 읽고, 그 종료 날짜를 새 시작으로 사용
- `compare` — 두 윈도우 나란히 실행 (현재와 이전 같은 길이)

기본으로 UTC 사용 (`--since="N days ago"`는 git의 로컬 해석; 크로스팀 작업엔 `TZ=UTC`로 normalize 원할 수도). Retro가 한 timezone의 단일 팀용이면 output에 명시 전달 ("시간은 `America/Los_Angeles`로 표시").

Day-aligned 윈도우에 시작을 정확히 N×24 시간 전이 아니라 선택된 timezone의 자정에 anchor — 세션이 자연스럽게 달력 day와 정렬.

### Step 1: Reader 식별

```bash
git config user.name
git config user.email
```

반환된 이름이 retro의 **"you"** — 읽는 사람. 다른 모든 author는 동료. 이 레이블링이 Step 9에서 "you"에게 깊은 개인 처리 주는 동시 동료를 표준 포맷 유지.

### Step 2: Raw Git 데이터 pull

이 명령들 실행. 독립적 — runtime 지원하면 병렬 fire.

```bash
# timestamp, subject, hash, author, 변경 파일 있는 모든 commit
git log --since="<window>" --format="%H|%aN|%ae|%ai|%s" --shortstat

# Commit별 numstat (테스트 vs 총 LOC, author와 함께)
git log --since="<window>" --format="COMMIT:%H|%aN" --numstat

# 세션 감지와 시간별 분포용 commit timestamp
git log --since="<window>" --format="%at|%aN|%ai|%s" | sort -n

# 파일 hotspot (가장 많이 건드려진 파일)
git log --since="<window>" --format="" --name-only \
  | grep -v '^$' | sort | uniq -c | sort -rn

# Commit subject의 PR 번호 (GitHub `#123`와 GitLab `!123` 작동)
git log --since="<window>" --format="%s" \
  | grep -oE '[#!][0-9]+' | sort -u

# Author별 파일 hotspot
git log --since="<window>" --format="AUTHOR:%aN" --name-only

# Author별 commit count (merge 제외)
git shortlog --since="<window>" -sn --no-merges

# 윈도우에 변경된 테스트 파일
git log --since="<window>" --format="" --name-only \
  | grep -E '\.(test|spec)\.|_test\.|_spec\.' | sort -u | wc -l

# AI 귀속 commit (Co-Authored-By trailer, "Generated with" 등)
git log --since="<window>" --format="%H|%aN|%B" \
  | grep -iE 'co-authored-by:.*(claude|gpt|copilot|cursor|ai)|generated with|🤖'
```

특정 브랜치 타겟팅하면 브랜치 prefix (예: `git log origin/main --since=...`) — 로컬 WIP가 데이터 오염 안 하도록.

### Step 3: 각 commit을 work type으로 분류

conventional-commit prefix 있으면 사용, 없으면 diff에서 추론:

| Type | 신호 |
|---|---|
| **feat** | `feat:` prefix, 또는 src dir의 net-new 파일 |
| **fix** | `fix:` prefix, issue/bug 참조, 작은 diff로 기존 로직 수정 |
| **refactor** | `refactor:` prefix, 동작 변경 없는 큰 diff (test diff 없음, balanced ±) |
| **docs** | doc 파일만 (`.md`, `docs/`, `README`) 건드림 |
| **test** | 테스트 파일만 건드림 |
| **chore** | dep, config, CI, 빌드 (lockfile, `.github/`, `package.json`만) |

Edge case:
- 혼합 commit (feature + test + docs 한 번에) → LOC로 지배 type 분류
- Merge commit → work-type tally 제외
- Revert → `fix`로 bucket

### Step 4: 작업 세션 감지 (45분 gap 규칙)

Author별 commit 시간순 정렬. 이전 commit과 gap이 **45분** 초과하면 새 세션 시작. 세션 내, 첫 commit에서 마지막까지 경과 시간 합계.

각 세션 분류:
- **Deep** — 50+ 분 경과
- **Medium** — 20–50 분
- **Micro** — <20 분 (종종 단일 commit)

Author별 계산:
- 총 세션 count
- 총 활성 코딩 분
- 평균 세션 길이
- 활성 시간당 LOC

45분 threshold는 heuristic: 커피 휴식이나 코드 리뷰 흡수 충분히 길고, 관련 없는 두 작업 블록 merge 안 되게 충분히 짧음. 한 방향으로 증거 있으면 팀에 맞게 tune.

### Step 5: Focus Score 계산 (0–1)

```
focus_score = (commits_in_top_directory) / (total_commits)
```

"Top 디렉토리"는 윈도우에서 가장 많이 건드려진 최상위 디렉토리.

| Score | 해석 |
|---|---|
| 0.7–1.0 | 깊은 focus — 한 영역의 지속 작업, 단일 큰 것 ship하는 중일 가능성 |
| 0.4–0.7 | 혼합 — feature 작업 + side fix, 정상 주 |
| 0.0–0.4 | 분산 — 여러 영역 걸쳐 context switching, 가능성 firefighting |

낮은 score가 자동 나쁨은 아님 (릴리스 주는 자연스럽게 분산); 유지보수 윈도우의 높은 score는 한 영역에 over-investment 나타낼 수 있음. 항상 context로 해석.

### Step 6: Hotspot 식별

윈도우에서 가장 많이 건드려진 파일 top 5. 우려 신호로 플래그:

- 윈도우에 **>X commit** 있는 단일 파일 (X = 주간 기본 5)
  → 불안정성 또는 불명확 소유권
- 같은 파일이 **3주 연속** hotspot으로 나타남 → 아키텍처 smell
- Top 5의 테스트 파일 → 훌륭한 규율 (코드와 함께 진화하는 테스트) 또는 반복 patch되는 flaky 테스트 (diff 보기)
- Top 5의 `VERSION`, `CHANGELOG`, lockfile → 릴리스 heavy 주, 정상

### Step 7: AI 협업 감지

메시지 body나 trailer의 AI 귀속 marker 있는 commit count:
- AI 이름 명명하는 `Co-Authored-By:` 라인 (Claude, GPT, Copilot, Cursor 등)
- `Generated with` / `🤖 Generated with` 패턴
- 커스텀 org 컨벤션 (예: `[ai-pair]` 태그)

리포트:
- AI-assisted commit의 절대 count
- 총 commit 대비 AI-assisted 비율
- Per-author breakdown (일부 동료는 AI 많이, 다른 사람은 전혀 안 쓸 수도)
- 가능하면 이전 주 대비 트렌드

이건 중립 metric — 추적, 도덕적 판단 금지. 작업 스타일 shift 주목에 유용.

### Step 8: 14-섹션 대시보드 생성

(아래 섹션 정의.)

---

## 14 대시보드 섹션

모든 retro output은 이 순서로 14 섹션 모두 포함. 불가능할 때만 섹션 skip (예: 솔로 repo가 Person별 Breakdown skip), skip 시 이유 주목.

### 1. Headline
한 줄. 포함: 윈도우, commit count, 기여자 count, net LOC, 지배적 work type, 하나의 notable 신호.
> Week of Mar 11–17 · 47 commits · 3 contributors · +3,200/-1,100 LOC · feature-heavy · browse refactor에 가장 긴 세션 4h

### 2. By the Numbers
Compact metric 블록:
- main에 commit: N
- 기여자: N
- 머지된 PR: N
- Insertions / Deletions / Net LOC
- Test LOC / Test 비율 (test LOC ÷ 총 LOC)
- 활성 days: N of N possible
- 감지된 세션: N
- 세션-시간당 평균 LOC: N

### 3. Work Type Breakdown
Step 3의 percentage bar chart. `fix` 비율 50% 초과 시 플래그 (ship-fast/fix-fast 패턴, 가능 리뷰나 테스트 gap).

```
feat:      20  (40%)  ████████████████████
fix:       18  (36%)  ██████████████████
refactor:   6  (12%)  ██████
docs:       4  ( 8%)  ████
test:       2  ( 4%)  ██
```

### 4. Sessions
- 총 세션 count와 총 활성 코딩 분
- 가장 긴 단일 세션 (author + commit subject에서 topic 포함)
- 평균 세션 길이
- Deep / Medium / Micro 분포

### 5. Focus Score
숫자 + 실제 top 디렉토리에 ground된 한 문장 해석.
> Focus score: 0.62 — `browse/`에 대부분 작업 집중 (47 commit 중 29), `extension/`과 `docs/`에 side 투자.

### 6. Hotspot
Commit count 있는 top 5 파일. 우려 패턴 명시 플래그:
> ⚠ `browse/src/server.ts` 8번 건드려짐 — 분할 필요 여부 조사.

### 7. Person별 Breakdown (team-aware)
각 기여자에 대해 이 순서:
1. **You** (reader) — 가장 충만한 처리, 1인칭 framing
2. 다른 기여자, commit count 내림차순 정렬

Person별 커버:
- Commit / LOC / PR
- 상위 3 focus 영역 (디렉토리 또는 모듈)
- 개인 work-type 믹스 (그들의 feat/fix/refactor 분할)
- 세션 패턴 (peak 시간, 세션 count, deep vs micro)
- 개인 test 비율
- 가장 큰 단일 ship (가장 높은 LOC PR 또는 commit)

솔로 repo면: 이 섹션 skip, "솔로 repo — 위 개인 stat 참조"로 주목.

### 8. AI 협업 Stats
- N AI-assisted commit (총의 X%)
- Person별 breakdown
- 이전 주 대비 트렌드
- AI 사용이 작업 shape 어떻게 shift했는지 하나 관찰 (예: "AI-assisted commit이 테스트 작성으로 skew — 8 중 6이 테스트 추가")

### 9. 칭찬 (증거 기반, person별)
각 사람에 대해 **commit, PR, 파일 명명하는 1–2 구체 칭찬**. 모든 라인 anchor. 작성 전 위 Bad vs Good 테이블 재독.

### 10. 성장 영역 (증거 기반, person별)
각 사람에 대해 **commit, PR, 파일 명명하는 1 구체 성장 영역**. 비판이 아니라 leveling up으로 frame. 같은 anchoring 규칙. Anchor 불가면 라인 쓰지 말 것.

### 11. 주목된 패턴
Person별 섹션이 캡처 안 하는 cross-cutting 관찰:
- Fix-chain 패턴 (<24h에 같은 파일 건드리는 3+ fix)
- 늦은 밤 클러스터 (로컬 밤 10시 이후 commit)
- Hand-off 패턴 (한 사람 commit 다음 다른 사람의 같은 파일 commit)
- 문서/테스트 부채 (문서나 테스트 없이 merge된 feature)

각 패턴이 관찰 트리거한 특정 commit 명명.

### 12. 이전 주 대비 트렌드
이전 retro 존재하면 delta 표시:

```
                지난주       이번주       Δ
Commits             32           47       +15
Test 비율           22%          41%      +19pp
Sessions            10           14       +4
Fix 비율            54%          30%      −24pp (개선)
Focus score         0.41         0.62     +0.21
AI-assisted        12%          28%      +16pp
```

이전 retro 없으면: "첫 retro — 트렌드 보려 다음 주 재실행."

### 13. 미해결 질문 / 조사 Hook
데이터가 raise했지만 답 안 한 것. 다음 팀 sync의 agenda가 됨.
- "왜 이번 주 `notifications/sender.ts`가 4 fix commit 얻음? 원래 변경이 under-tested였나?"
- "왜 아침 세션 블록이 지난주보다 얇은가?"

### 14. 다음 주 추천
3 작고, 실용, 현실적 습관 (각 <5 min/day 채택). 각 추천을 motivate하는 데이터에 anchor.
> `payment/`의 모든 동작 commit을 같은 PR의 테스트 commit과 페어 — 이번 주 거기 8 동작 commit이 총 1 테스트 commit.

---

## 출력 포맷

위 대시보드가 canonical form. 팀이 사용하는 어떤 surface든 렌더링.

### 기본: Markdown 대시보드
위 14 섹션 정확히 렌더링, heading (`##`), 테이블, code 블록 intact. Best for: GitHub Discussions, Notion, 내부 wiki, 이메일.

### 대안: Telegram / chat 친화 (compact)
테이블 제거 (대부분 chat 클라이언트가 나쁘게 렌더링). Bullet과 bold 강조로 교체. 가장 높은 신호 항목만 trim. 섹션당 한 viewport 목표.

```
**Week of Mar 11–17**
47 commits · 3 contributors · +3,200/-1,100 LOC

**You shipped:**
• Browse refactor (PR #482) — 서버 bootstrap을 composable layer로 추출
• payment/processor.ts의 14 edge-case 테스트 (retry 경로의 off-by-one catch)

**팀 칭찬:**
• @alice — extension/sidebar.ts의 깔끔한 state-machine 추출 (commit a3f8b2)
• @bob — PR #479의 flaky CI race condition catch

**다음 주 습관:**
• 모든 payment/ 동작 commit을 같은 PR의 테스트 commit과 페어
```

### 대안: Slack Block Kit
스켈레톤 (섹션별 채우기):

```json
{
  "blocks": [
    {"type": "header", "text": {"type": "plain_text", "text": "Weekly Retro · Mar 11–17"}},
    {"type": "section", "text": {"type": "mrkdwn", "text": "*47 commits · 3 contributors · +3,200/-1,100 LOC*"}},
    {"type": "divider"},
    {"type": "section", "text": {"type": "mrkdwn", "text": "*Praise*\n• @alice — ..."}},
    {"type": "section", "text": {"type": "mrkdwn", "text": "*Growth*\n• @bob — ..."}},
    {"type": "section", "text": {"type": "mrkdwn", "text": "*Habit for next week*\n• ..."}}
  ]
}
```

### 대안: 이메일 요약
Subject 라인 = Headline (섹션 1). Body = 섹션 2, 9, 10, 14만, plain text. 전체 대시보드는 첨부 또는 링크로 저장.

---

## 저장

각 retro를 `~/.claude/retros/<project>/<YYYY-WW>.md` (ISO week)에 저장. 파일명 예시: `~/.claude/retros/myapp/2026-W17.md`.

Raw metric (work-type count, 세션 list, hotspot list, person별 숫자) 포함 JSON snapshot도 함께 저장: `2026-W17.json`. 이게 다음 run에서 Step 12의 트렌드 추적이 읽는 것.

```json
{
  "window": {"start": "2026-04-21", "end": "2026-04-27", "tz": "UTC"},
  "totals": {"commits": 47, "contributors": 3, "insertions": 3200, "deletions": 1100, "test_ratio": 0.41},
  "work_types": {"feat": 20, "fix": 18, "refactor": 6, "docs": 4, "test": 2},
  "sessions": {"count": 14, "deep": 5, "medium": 6, "micro": 3, "longest_minutes": 240},
  "focus_score": 0.62,
  "hotspots": [{"file": "browse/src/server.ts", "commits": 8}],
  "ai_assisted": {"count": 13, "ratio": 0.28},
  "per_person": [
    {"name": "you", "commits": 32, "loc_net": 1900, "top_dir": "browse/", "test_ratio": 0.45}
  ]
}
```

---

## 트렌드 추적

매 run마다 Step 8 후 가장 최근 이전 JSON snapshot 로드하고 Section 12를 delta로 populate. 주시:

- **Focus score 트렌드** — 지속 drop (3+ 주 감소 focus)은 성장 surface 영역 또는 축적 distraction 제안
- **Work-type 비율 shift** — 릴리스 후 `fix` 비율 갑작스 jump는 정상; 지속 높은 `fix` 비율은 품질 신호
- **AI 협업 트렌드** — 방향성보다 인식 자체가 더 중요
- **Hotspot 지속성** — 3+ 연속 주 top 5의 파일은 Section 13 surface 가치 있는 아키텍처 신호
- **세션 패턴 drift** — 시간에 걸친 더 적은 deep 세션은 종종 증가 미팅 부담 또는 context switching과 상관

---

## Compare 모드

"compare" 요청 시, 같은 길이 두 윈도우 back-to-back 분석 실행 (현재 vs 바로 이전)하고 Section 2–8의 모든 metric에 대해 두 value 컬럼과 delta 컬럼 있는 단일 대시보드 제시. Section 9–10도 적용되지만 **현재** 윈도우만 draw — 이번 주 rating, 지난주는 context.

---

## 완료 상태

다음 중 하나 리포트:
- **DONE** — retro 생성, 히스토리 저장, 14 섹션 모두 렌더링
- **DONE_WITH_CAVEATS** — 생성됐지만 데이터 누락 (예: 트렌드용 이전 retro 없음, 또는 솔로 repo가 person별 섹션 skip). Caveat 명명.
- **BLOCKED** — git repo 아님, 윈도우에 commit 없음, 또는 git 히스토리 접근 불가. Blocker 진술.

---

## 주시할 안티패턴

Retro 생성 시 자기 초안을 이들에 대해 audit:

1. **Generic 칭찬** — "훌륭한 작업", "강한 주", "많이 shipping". Commit 가져오거나 라인 삭제.
2. **Unanchored 성장 노트** — "더 신중하게", "품질 개선". 노트 motivate한 commit 명명 또는 삭제.
3. **Commit count만 재진술하는 단일 문장 Per-Person 섹션.** Per-person breakdown은 각 엔지니어에게 가리킬 것 제공해야.
4. **AI commit을 도덕 점수로 count.** Metric이지 verdict 아님.
5. **Focus score를 target으로 취급.** 주가 어떻게 shape됐는지 describe, optimize할 목표 아님.
6. **트렌드에 사과와 오렌지 비교** — 릴리스 주 vs planning 주는 wildly diverge. Context 주목, delta만 표시 금지.
