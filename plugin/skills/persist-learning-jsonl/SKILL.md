---
name: persist-learning-jsonl
description: "[패턴 라이브러리] JSONL append-only learning store data model + 누적/조회 패턴 (pattern/pitfall/preference taxonomy, latest-winner dedup, confidence ranking, cross-session retrieval). 직접 invoke보다 orchestrator가 import해 사용. 트리거: '학습 기록' / 'lesson learned 남겨' / 'JSONL 누적' / 'learning 저장'. 참조 위치: 모든 스킬의 prior learning 저장, autoplan/critique-plan/validate-idea 종료 시점, postmortem."
type: skill
---

# Snippet: JSONL 학습 저장소


> 이 snippet은 저장소 설계(storage design)에 집중. JSONL append-only 학습 저장소 데이터 모델 + 누적 패턴. CLI 서브커맨드(show/search/prune/export/stats/add)는 별도 스킬에서.

## 이 snippet을 사용하는 경우

- AI 어시스턴트용 프로젝트 메모리 레이어 구축
- 데이터베이스 없이 cross-session 학습 persistence
- "메모리"가 서비스가 아닌 flat 파일이어야 할 때 참조
- 에이전트 관찰용 append-only vs mutable 저장소 선택

## 왜 JSONL?

| Aspect | JSON array | JSONL | 왜 JSONL이 이김 |
|--------|------------|-------|----------------|
| Append | 전체 파일 re-write | `>>`로 한 줄 | Atomic, race 없음, 빠름 |
| Read | 전체 파일 parse | 라인별 스트림 | 메모리 제한 |
| 동시 쓰기 | Locking 필요 | `O_APPEND`로 lock-free | 멀티 프로세스 안전 |
| Diff/merge | 고통스러움 | 라인 추가형 | Git 친화적 |
| 손상 blast radius | 전체 파일 파싱 불가 | 나쁜 라인 하나 | 복구 가능 (`grep -v` skip) |

JSONL이 맞는 선택인 경우:

- Write가 append-only (과거 엔트리 수정 안 함)
- 동시 producer 존재 (여러 세션이 병렬 쓰기)
- Grep/awk/jq 가능한 평문 원함
- Dedup이 쓰기 시점이 아니라 읽기 시점

## 저장 위치

```
~/.claude/learnings/<project-slug>/
  learnings.jsonl       # 메인 append-only 저장소
```

Project slug: 작업 디렉토리 basename (또는 `git rev-parse --show-toplevel`)에서 유도.

## 타입 Taxonomy

모든 learning은 `type` 필드를 가진다. 원본 6가지 타입:

### `pattern` — 반복되는 문제에 대한 재사용 가능한 해결책

```json
{"skill":"review","type":"pattern","key":"jsonl-append-only","insight":"Use JSONL + latest-winner dedup for multi-session writes; lock-free.","confidence":8,"source":"observed","files":["bin/learnings-log"],"ts":"2026-04-23T..."}
```

### `pitfall` — 회피할 안티패턴, footgun, 실수

```json
{"skill":"investigate","type":"pitfall","key":"json-array-store","insight":"Storing learnings as a JSON array forces full-file rewrite on every append; races under concurrent writes.","confidence":9,"source":"observed","ts":"2026-04-23T..."}
```

### `preference` — 코드에서 유도 불가능한 사용자/팀 선호

```json
{"skill":"learn","type":"preference","key":"no-attribution","insight":"User does not want Co-Authored-By attribution in commit messages.","confidence":10,"source":"user-stated","ts":"2026-04-23T..."}
```

### `architecture` — 고수준 설계 결정과 근거

```json
{"skill":"review-plan-architecture","type":"architecture","key":"learnings-flat-file","insight":"Per-project flat JSONL store at ~/.claude/learnings/<slug>/. No DB. Dedup by key+type at read time.","confidence":9,"source":"observed","ts":"2026-04-23T..."}
```

### `tool` — 도구/바이너리/명령 동작에 대한 지식

```json
{"skill":"ship","type":"tool","key":"gh-pr-create-heredoc","insight":"gh pr create needs heredoc body to preserve newlines; --body 'string' collapses them.","confidence":8,"source":"observed","ts":"2026-04-23T..."}
```

### `operational` — 세션 경험 발견 (빌드 순서, env, 타이밍, auth)

```json
{"skill":"qa","type":"operational","key":"playwright-needs-display","insight":"Playwright headless on Linux requires Xvfb when running in CI containers without DISPLAY.","confidence":7,"source":"observed","ts":"2026-04-23T..."}
```

### 필드 스키마 (모든 타입)

| 필드 | 필수 | 메모 |
|-------|----------|-------|
| `skill` | yes | 이 learning을 캡처한 스킬 |
| `type` | yes | 위 6개 중 하나 |
| `key` | yes | 짧은 kebab-case ID, `^[a-zA-Z0-9_-]+$`에 매치 (injection surface 없음) |
| `insight` | yes | 한 문장 설명. injection-like 패턴(`ignore previous instructions`, `you are now`, `system:` 등) 포함 시 쓰기 시점에 strip |
| `confidence` | yes | 정수 1-10 |
| `source` | optional | `observed` / `user-stated` / `inferred` / `cross-model` |
| `files` | optional | 관련 파일 경로 배열 (staleness 체크용) |
| `ts` | auto | ISO8601, 누락 시 쓰기 시점에 주입 |
| `trusted` | auto | `source == "user-stated"`일 때만 `true` |

## Confidence 랭킹

Confidence는 정수 1-10. 출처별 calibration:

| 출처 | 초기 confidence | 이유 |
|--------|-------------------|-----|
| `user-stated` (사용자가 명시적으로 말함) | 9-10 | 사용자는 자기 선호에 권위 있음 |
| `cross-model` (여러 AI 동의) | 7-9 | Cross-model 동의는 강한 신호 |
| `observed` (이 세션에서 발생 관찰) | 5-8 | 실제 증거지만 제한된 샘플 |
| `inferred` (한 관찰에서 유도) | 3-5 | 합리적 추측, 아직 증명 안 됨 |

### 누적: (key, type)별 latest-winner

원본은 **읽기 시점 latest-winner dedup** 사용 (in-place 편집 없음). Corroboration 후 "confidence 높이기"는 같은 `key` + `type`에 더 높은 confidence 값으로 새 엔트리 append. 리더는 `(key, type)` 쌍별로 최신 `ts`의 엔트리만 유지.

```bash
# 첫 관찰
learnings-log '{"skill":"qa","type":"pitfall","key":"flaky-timeout","insight":"...","confidence":5,"source":"observed"}'

# 2주 뒤 재관찰 — 더 높은 confidence로 새 라인 append
learnings-log '{"skill":"qa","type":"pitfall","key":"flaky-timeout","insight":"...","confidence":8,"source":"observed"}'
```

Corroboration counter 대신 latest-winner인 이유:

- Append-only invariant 유지 — in-place mutation 없음
- 리더 로직이 단순 유지 (한 패스 + Map)
- 히스토리가 audit용으로 JSONL에 남음 (`grep '"key":"flaky-timeout"' learnings.jsonl`)
- 트레이드오프: 자동 decay나 가중 평균 없음 — confidence는 lifetime 증거 가중이 아니라 최신 평가 반영

Corroboration count 원하면 읽기 시점에 계산:

```bash
jq -s 'group_by(.key + "|" + .type) | map({key: .[0].key, type: .[0].type, count: length, latest_confidence: (max_by(.ts)).confidence})' learnings.jsonl
```

## 중복 감지 (읽기 시점)

원본은 쓰기 시점에 dedup 안 함. Append는 항상 성공. Dedup은 search/show 바이너리가 파일 읽을 때 발생:

```javascript
const seen = new Map();
for (const line of lines) {
  const e = JSON.parse(line);
  const dk = e.key + '|' + e.type;
  const existing = seen.get(dk);
  if (!existing || new Date(e.ts) > new Date(existing.ts)) seen.set(dk, e);
}
// seen.values() = deduped 현재 상태
```

이게 JSONL을 장기 실행 저장소로 만드는 트릭. 편집하지 않고, 계속 append하고, 리더가 현재 상태를 project.

### Stale 엔트리 플래그 시점 (수동 prune)

스킬의 `prune` 서브커맨드는 deduped 엔트리에 대해 다음 체크 실행:

1. **파일 존재**: `files` 필드가 repo에 더 이상 존재하지 않는 경로를 참조하면 `STALE: <key> references deleted file <path>` 플래그.
2. **모순**: 같은 `key`지만 다른 `type`의 두 엔트리가 반대 insight를 주면 `CONFLICT: <key> has contradicting entries` 플래그.

사용자가 플래그 엔트리별로 결정: 제거(JSONL에서 그 라인 빼고 재작성), 유지, 또는 업데이트(수정 엔트리 append).

## Cross-Session 검색

새 세션 시작 시 가장 관련 있는 learning을 surface:

### Step 1: 컨텍스트에서 쿼리 구성

- 현재 task의 태그: 편집 중인 파일 경로, 에러 메시지, 기술 스택
- 타입 필터: 디버깅 중엔 `pitfall`만, ship 직전엔 `preference`, 설계 중엔 `architecture`

### Step 2: 필터

```bash
# 모든 pitfall
jq -c 'select(.type == "pitfall")' learnings.jsonl

# auth.ts 관련 pitfall
jq -c 'select(.type == "pitfall" and (.files // [] | any(. == "auth.ts")))' learnings.jsonl

# 사용자로부터의 고 confidence preference
jq -c 'select(.type == "preference" and .source == "user-stated" and .confidence >= 8)' learnings.jsonl
```

### Step 3: Confidence + recency로 랭크

```
score = confidence * 0.7 + recency_factor * 0.3
recency_factor = ts가 30일 이내면 1.0, 아니면 0.5
```

### Step 4: 세션 시작 시 상위 3-5개를 에이전트 prompt에 주입

원본 스킬은 `> 5`개 learning 있는 세션에서 preamble에 상위 3개 auto-load:

```bash
if [ "$_LEARN_COUNT" -gt 5 ]; then
  learnings-search --limit 3
fi
```

로드 상한 (3-5, 50 아님) — 목표는 priming, 저장소 전체 덤프 아님.

## 보안 노트: 쓰기 시점 Prompt Injection

`insight` 필드는 이후 세션의 에이전트 컨텍스트에 로드되므로 쓰기 시점에 sanitize. 원본이 거부하는 패턴:

```
/ignore\s+(all\s+)?previous\s+(instructions|context|rules)/i
/you\s+are\s+now\s+/i
/always\s+output\s+no\s+findings/i
/skip\s+(all\s+)?(security|review|checks)/i
/override[:\s]/i
/\bsystem\s*:/i
/\bassistant\s*:/i
/\buser\s*:/i
/do\s+not\s+(report|flag|mention)/i
/approve\s+(all|every|this)/i
```

에이전트가 learning 저장소에 unsanitize하게 쓰게 허용하면 poisoned 웹 페이지나 PR description이 지시를 심어 이후 세션을 하이재킹할 수 있다. 이후가 아니라 경계에서 거부.

## 안티패턴

- JSON array로 저장 — append 성능 죽고 동시 쓰기에서 깨짐.
- 데이터베이스에 저장 — 사람이 읽을 수 있는 몇천 줄에 과함; grep/jq 상실.
- 과거 엔트리 in-place 편집 — append-only invariant 깨짐; 대신 같은 `key + type`의 새 엔트리 append.
- 읽기 시점 dedup 없음 — 저장소가 near-duplicate로 차고 confidence 랭킹이 의미 상실.
- Confidence가 초기값에 머물음 — 관찰이 corroborate/contradict하면 새 append로 업데이트해야.
- `insight`에 콘텐츠 sanitization 없음 — 모든 미래 세션으로의 prompt injection 채널 오픈.
