# Decision 1: Schema 필드 — 상세 분석

> v0.1 spec §6.1을 보강. 각 필드의 *출처·비용·활용·위험*을 평가하여
> 사용자가 의미 있는 결정을 할 수 있도록 정보를 제공.

---

## 0. 배경: Claude Code hook이 실제로 주는 정보

Claude Code의 hook은 **stdin으로 JSON을 받는다**. 예를 들어 PreToolUse hook이
받는 입력 형태는 대략 이렇다 (실제 스키마는 Claude Code 버전에 따라 변동):

```json
{
  "session_id": "550e8400-e29b-41d4-a716-446655440000",
  "transcript_path": "/Users/.../.claude/sessions/550e8400.../transcript.jsonl",
  "cwd": "/Users/wm/Work/buddy",
  "hook_event_name": "PreToolUse",
  "tool_name": "Bash",
  "tool_input": {
    "command": "git status",
    "description": "check git state"
  }
}
```

**핵심 인사이트:**
- `tool_name`, `tool_input`은 **이미 stdin에 있다** → 별도 비용 없음
- `transcript_path`는 JSONL 파일 경로 → 파싱하면 model/token 정보 얻을 수 있음
- `git`/`terminal` 정보는 stdin에 없음 → 별도 subprocess 호출 필요

이 사실이 필드별 비용 평가의 근거가 된다.

---

## 1. 필드별 상세 분석

### 1-A. `toolName` — 호출된 tool 이름

| 항목 | 값 |
|------|---|
| 출처 | stdin JSON의 `tool_name` |
| 추가 비용 | **0** (이미 받은 데이터) |
| Storage | 평균 10 bytes (Bash, Read, Edit, Glob, Grep, Task...) |
| Privacy 위험 | 없음 |
| Cardinality | 낮음 (Claude Code 내장 tool ~15종 + MCP tools) |

**무엇이 가능해지는가:**
- "어떤 tool이 가장 느린가?" → Bash p95 vs Read p95 비교
- "어떤 tool이 자주 실패하는가?" → tool별 실패율
- v0.2 dashboard의 *주축 분석 차원*이 됨 (없으면 hook 단위 통계만 가능)

**없으면 잃는 것:**
- 같은 PostToolUse hook이라도 Bash 호출 vs Read 호출의 성격이 매우 달라
  hook 단위 통계만으로는 노이즈가 큼

**평가 (Opinion, High):** v0.1 **필수 포함**. 비용 0, 활용 최고.

---

### 1-B. `toolArgs` — tool 인자 (Bash command, file path 등)

| 항목 | 값 |
|------|---|
| 출처 | stdin JSON의 `tool_input` (object) |
| 추가 비용 | **0** (이미 받은 데이터) |
| Storage | 평균 200 bytes ~ 5KB (Bash 긴 명령어 시 큼) |
| Privacy 위험 | **높음** — 명령어/경로에 시크릿·PII 노출 가능 |
| Cardinality | 매우 높음 (free text) |

**Privacy 위험 구체 예시:**
```bash
# 이런 명령어가 그대로 DB에 들어감
aws s3 cp ./backup s3://my-bucket --aws-access-key-id=AKIA...
psql postgres://user:password@host/db
gh auth login --with-token < ./secret.txt
```

**무엇이 가능해지는가:**
- 디버그: "어떤 Bash 명령어가 timeout 났는지 정확히"
- 패턴 발견: 자주 쓰는 명령어의 평균 시간

**완화책 (구현 비용 있음):**
1. Sanitize 정규식 (`AKIA[A-Z0-9]{16}`, `password=...`, `Bearer ...` 마스킹)
2. opt-in 정책 (`config.recordToolArgs: 'never' | 'sanitized' | 'raw'`)
3. 자동 truncate (`tool_input` 길이 > 1KB이면 hash만 저장)

**평가 (Opinion, High):**
- v0.1에서는 **opt-in으로 가능 + default off**가 안전
- 즉, 컬럼은 만들되 default config가 `'never'`. sanitize 모듈은 v0.2

---

### 1-C. `modelName` — 사용된 모델

| 항목 | 값 |
|------|---|
| 출처 | stdin JSON에 **없음**. transcript JSONL의 assistant 메시지 `model` 필드 파싱 필요 |
| 추가 비용 | JSONL tail 파싱 (파일 IO + 파싱). 캐싱 시 평균 <5ms |
| Storage | 평균 30 bytes (`claude-opus-4-7` 등) |
| Privacy 위험 | 없음 |
| Cardinality | 낮음 (~10종) |

**무엇이 가능해지는가:**
- Opus가 hook을 더 자주/느리게 호출하는지 비교
- 모델 변경 시점과 hook 패턴 변화 상관관계

**구현 부담:**
- transcript JSONL 파싱 로직 추가 (recon이 이미 한 패턴 — 참고 가능)
- `transcript_path`가 hook stdin에 없는 버전 대응 필요할 수 있음
- session_id별 model 매핑 캐시 (동일 세션은 모델 변경 빈도 낮음)

**평가 (Opinion, Mid):**
- v0.1 포함하면 transcript 파싱 인프라가 함께 들어옴 → v0.2 dashboard 가속
- 다만 v0.1 scope가 늘어남 (M3에 0.5일 추가)
- **권장: v0.1 포함 (transcript 파싱은 어차피 v0.2에 필요한 기반)**

---

### 1-D. `tokenUsage` — input/output/cache 토큰

| 항목 | 값 |
|------|---|
| 출처 | transcript JSONL의 assistant 메시지 `usage` 객체 |
| 추가 비용 | modelName과 동일 파싱 경로에 무료 추가 |
| Storage | 평균 40 bytes (4개 정수 + JSON) |
| Privacy 위험 | 낮음 (수치만) |
| Cardinality | continuous |

**무엇이 가능해지는가:**
- 토큰 burn rate (시간당 토큰)
- 비용 추정 (model price × tokens)
- cache hit rate 모니터링

**중요한 트레이드오프 — token-monitor와 중복:**
- `harness/token-monitor` 프로젝트가 이미 같은 일을 함 (Claude JSONL 스캔)
- Buddy가 자체 구현하면: 일관된 단일 dashboard 가능
- Buddy가 token-monitor 통합하면: 중복 제거, 그러나 의존성 추가
- **권장 (Opinion, High):** v0.1은 자체 구현 (transcript 파싱 인프라가 어차피
  필요), v1.0에서 token-monitor를 plugin으로 흡수 검토

**평가 (Opinion, High):** v0.1 포함 — modelName과 같은 파싱 경로라 한계 비용 거의 0

---

### 1-E. `gitBranch` — 현재 브랜치명

| 항목 | 값 |
|------|---|
| 출처 | hook wrapper에서 `git -C $cwd branch --show-current` 실행 |
| 추가 비용 | subprocess fork (~5-15ms per hook) |
| Storage | 평균 25 bytes |
| Privacy 위험 | 중간 (브랜치명에 jira ticket·고객명·기능명 노출) |
| Cardinality | 중간 (사용자별 수십~수백) |

**무엇이 가능해지는가:**
- 브랜치별 작업 패턴 (어떤 기능 작업 시 hook이 더 느린지)
- "main에서는 hook이 안 도는데 feature/X에서만 timeout" 류 발견

**비용의 진짜 의미:**
- hook이 분당 100회 호출되면 매번 5ms = 0.5초/분 추가 부하
- buddy 자체가 hook을 늦추는 원인이 됨 (자기 모순)
- 완화: cwd별 브랜치 캐시 (5초 TTL) → 비용 거의 0

**평가 (Opinion, Mid):**
- 비용 완화 가능하나 캐시 로직이 v0.1 복잡도 증가
- **권장: v0.1 보류, v0.2에서 캐시 인프라와 함께 추가**

---

### 1-F. `gitDirty` — uncommitted 변경 여부

| 항목 | 값 |
|------|---|
| 출처 | `git -C $cwd status --porcelain | head -1` |
| 추가 비용 | gitBranch와 비슷 (5-15ms), 큰 repo에서 더 느림 |
| Storage | 1 byte (boolean) |
| Privacy 위험 | 없음 (boolean) |
| Cardinality | 2 |

**무엇이 가능해지는가:**
- "dirty 상태에서 hook이 더 느린가?" — 큰 repo에서 git stash/diff 영향
- 사용자의 commit 빈도 패턴

**평가 (Opinion, Low):**
- 활용도 낮음, 비용 있음
- **권장: v0.1 보류, v1.0에서도 우선순위 낮음**

---

### 1-G. `terminalCols` — 터미널 가로 폭

| 항목 | 값 |
|------|---|
| 출처 | 환경변수 `COLUMNS` 또는 `tput cols` |
| 추가 비용 | 환경변수면 0, tput 호출이면 ~3ms |
| Storage | 2 bytes |
| Privacy 위험 | 없음 |

**무엇이 가능해지는가:**
- "좁은 터미널에서 hook이 출력을 깨뜨리는가" 분석
- TUI 환경 vs SSH 일반 vs CI 분류

**평가 (Opinion, Low):**
- v0.2 dashboard에서 시각화할 때 의미 있음
- v0.1에서는 분석 동기 부족
- **권장: v0.1 보류**

---

### 1-H. `customTags` — 사용자 정의 태그

| 항목 | 값 |
|------|---|
| 출처 | hook script가 명시적으로 환경변수/CLI 인자로 buddy에 전달 |
| 추가 비용 | 0 (수동 전달) |
| Storage | 평균 50 bytes (태그 1-3개) |
| Privacy 위험 | 사용자 책임 |
| Cardinality | 사용자 정의 |

**활용 예시:**
```bash
# 사용자의 hook script 안에서:
BUDDY_TAGS="experiment:foo,branch:wip-graphql" \
  /usr/local/bin/buddy hook-wrap "$@"
```

→ 이후 `buddy stats --tag experiment:foo` 같은 쿼리 가능

**무엇이 가능해지는가:**
- 사용자가 *자기 워크플로우의 의미 차원*을 직접 정의
- 실험 비교 ("이번 주 experiment:foo가 평소보다 느려진 시점은?")

**평가 (Opinion, High):**
- 비용 0, schemaless 확장 슬롯
- **권장: v0.1 포함 (단순한 환경변수 파싱만 추가)**

---

## 2. 종합 평가 매트릭스

| 필드 | 비용 | 활용도 | Privacy | v0.1 권장 |
|------|------|--------|---------|----------|
| toolName | 0 | ★★★★★ | ✅ | **포함** |
| toolArgs | 0 (raw) | ★★★ | ❌ 위험 | **컬럼만 + default off** |
| modelName | 중 (파싱) | ★★★★ | ✅ | **포함** |
| tokenUsage | 0 (modelName과 함께) | ★★★★★ | ✅ | **포함** |
| gitBranch | 중 (subprocess) | ★★★ | △ | v0.2 |
| gitDirty | 중 (subprocess) | ★★ | ✅ | v1.0+ |
| terminalCols | 0~저 | ★★ | ✅ | v0.2 |
| customTags | 0 | ★★★★ | 사용자 책임 | **포함** |

---

## 3. 추천 v0.1 schema (Opinion, High)

```typescript
export const HookEventPayload = z.object({
  // ── baseline (모든 hook event 공통) ──
  ts:          z.number().int().positive(),
  event:       z.enum(['SessionStart','PreToolUse','PostToolUse',
                       'Stop','PreCompact','UserPromptSubmit']),
  hookName:    z.string().min(1).max(100),
  durationMs:  z.number().int().min(0),
  exitCode:    z.number().int(),
  sessionId:   z.string().optional(),
  pid:         z.number().int().positive().optional(),
  cwd:         z.string().optional(),

  // ── v0.1 포함 결정 (제안) ──
  toolName:    z.string().optional(),                    // 1-A
  toolArgs:    z.unknown().optional(),                   // 1-B (default off, opt-in)
  modelName:   z.string().optional(),                    // 1-C
  tokenUsage:  z.object({                                // 1-D
    inputTokens:        z.number().int().nonnegative(),
    outputTokens:       z.number().int().nonnegative(),
    cacheReadTokens:    z.number().int().nonnegative(),
    cacheCreateTokens:  z.number().int().nonnegative(),
  }).optional(),
  customTags:  z.record(z.string()).optional(),          // 1-H

  // ── 확장 슬롯 ──
  meta:        z.record(z.unknown()).optional(),         // schemaless
})
```

```typescript
// config.json default
{
  "recordToolArgs": "never",   // 'never' | 'sanitized' | 'raw'
  "transcriptParse": true,     // modelName/tokenUsage 활성화
}
```

**왜 이 조합인가 (Opinion, High):**
1. 비용 거의 없는 필드만 포함 (toolName, customTags)
2. transcript 파싱 인프라는 v0.2에 어차피 필요 → v0.1에서 한 번 만들면 재활용
3. toolArgs는 *컬럼만* 만들어 두고 default off → 미래 sanitizer 추가 시 schema migration 불필요
4. git/terminal 필드는 hook 자체 지연 위험 vs 활용도 trade-off에서 보류 우세

---

## 4. 사용자 결정 (이 결정 후 spec lock)

다음 중 하나를 선택해 주세요:

| 옵션 | 내용 |
|------|------|
| **A) 추천대로** | toolName + toolArgs(off) + modelName + tokenUsage + customTags |
| B) 최소 | toolName + customTags 만 (transcript 파싱 v0.2로 미룸) |
| C) 최대 | A + gitBranch (v0.2 작업을 일부 당김) |
| D) 자유 | 직접 조합 명시 |

A를 추천합니다 (Opinion, High). 이유는 §3 참조.
