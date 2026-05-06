---
name: consult-codex
description: "독립 컨텍스트의 외부 LLM CLI(codex 등)를 호출해 review/challenge/consult 3 modes로 second opinion을 얻음. dual voice 합의 패스. 트리거: '외부 voice 받아줘' / '다른 LLM 의견' / '두 번째 의견' / 'dual voice로 가자' / 'codex와 논의' / '외부 컨텍스트 의견' / 'second opinion'. 입력: plan, critique, 의사결정 후보. 출력: 외부 LLM 응답 (single-voice degrade 가능). 흐름: autoplan dual voice / critique-plan에서 호출."
type: skill
---

# Codex — Multi-Mode 외부 Second-Opinion Wrapper


이 스킬은 **OpenAI Codex CLI**를 wrap해 다른 AI 시스템에서 독립적, brutally honest second opinion을 얻는다. IP는 CLI invocation이 아니라 — **mode-switching 패턴**: 한 스킬이 사용자 prompt의 flag로 선택된 세 personality (reviewer, adversary, peer) 노출. 각 모드는 purpose-built prompt 템플릿과 (review mode엔) deterministic pass/fail 게이트 보유.

**중요 이유:** 코드를 몇 시간 본 뒤에는 약점이 안 보인다. 공유 context 없는 다른 모델이 당신이 합리화한 것을 catch. 아래 패턴은 generic — 어떤 second-opinion CLI (Gemini, 로컬 llama, 다른 wrapper)로 교체해도 세 모드는 적용.

---

## 사전 요구사항

- `codex` CLI 설치: `npm install -g @openai/codex` 또는 https://github.com/openai/codex
- 인증: `codex login` (브라우저 기반 ChatGPT auth) **또는** `$CODEX_API_KEY` / `$OPENAI_API_KEY` 설정

이 스킬은 `codex review`와 `codex exec` 호출. CLI는 **read-only 샌드박스 모드**로 실행 — 파일 수정 불가. Output은 verbatim 텍스트로 스트림백되어 action 가능.

---

## Mode-Switching 패턴 (IP)

세 모드가 한 wrapper 공유. 모드는 스킬 이름 뒤 첫 토큰으로 결정:

| Subcommand | 모드 | Posture | 사용 시점 |
|------------|------|---------|-------------|
| `review` | Review | 독립 reviewer | 머지 전 — PASS/FAIL 게이트 있는 fresh-eyes diff critique |
| `challenge` | Challenge | 적대자 | 보안 민감 코드, 결제 flow, 데이터 migration — 깨려 시도 |
| 기타 | Consult | 동료 전문가 | 아키텍처 결정, 라이브러리 선택, 디버깅 경로 |

사용자 prompt로 전환:
- `/codex review` → Review 모드
- `/codex review focus on security` → focus 있는 Review 모드
- `/codex challenge` → Challenge 모드
- `/codex challenge security` → focus 있는 Challenge 모드
- `/codex how should I handle race conditions in X?` → Consult 모드

사용자가 인자 없이 `/codex` 타이핑하고 베이스 브랜치 대비 diff 존재하면 어떤 모드 원하는지 질문 (A/B/C).

---

## Step 0: Codex binary 체크

```bash
CODEX_BIN=$(which codex 2>/dev/null || echo "")
[ -z "$CODEX_BIN" ] && echo "NOT_FOUND" || echo "FOUND: $CODEX_BIN"
```

`NOT_FOUND`이면 중단하고 사용자에게:
"Codex CLI not found. Install it: `npm install -g @openai/codex` or see https://github.com/openai/codex"

## Step 0.5: Auth probe

```bash
# Multi-signal: $CODEX_API_KEY OR $OPENAI_API_KEY OR ~/.codex/auth.json
if [ -z "$CODEX_API_KEY" ] && [ -z "$OPENAI_API_KEY" ] && [ ! -f "${CODEX_HOME:-$HOME/.codex}/auth.json" ]; then
  echo "AUTH_FAILED"
fi
```

`AUTH_FAILED`면 중단하고 사용자에게:
"No Codex authentication found. Run `codex login` or set `$CODEX_API_KEY` / `$OPENAI_API_KEY`, then re-run."

## Step 1: 모드 감지

사용자 입력 parse:

1. `/codex review [instructions]` → **Review 모드** (Step 2A)
2. `/codex challenge [focus]` → **Challenge 모드** (Step 2B)
3. 인자 없는 `/codex` → **Auto-detect:**
   - Diff 체크: `git diff origin/<base> --stat 2>/dev/null | tail -1`
   - Diff 존재하면 질문:
     ```
     Codex detected changes against the base branch. What should it do?
     A) Review the diff (code review with pass/fail gate)
     B) Challenge the diff (adversarial — try to break it)
     C) Something else — provide a prompt
     ```
   - 아니면 질문: "What would you like to ask Codex?"
4. `/codex <기타>` → **Consult 모드** (Step 2C), 나머지 텍스트가 prompt

**Reasoning effort 기본값** (모드별, `--xhigh` flag로 override):
- Review: `high` — bounded diff 입력, 철저함 필요
- Challenge: `high` — adversarial이지만 diff에 bounded
- Consult: `medium` — 큰 context, 대화형, 속도 필요

---

## Step 2A: Review 모드 (독립 Diff Critique)

**목적:** 머지 전 자기 diff에 second opinion. Codex는 공유 context 없는 외부 reviewer — 코드를 차갑게 봄.

**Prompt 템플릿** (`codex review`에 prompt 인자로 전달):

> Review the diff between this branch and the base branch. Look for: logic bugs,
> missed edge cases, security holes, error-handling gaps, race conditions,
> performance issues, and incorrect assumptions. Mark critical findings with `[P1]`
> and non-critical with `[P2]`. Be terse. No compliments — just the findings.

사용자가 커스텀 instruction 추가하면 (예: `/codex review focus on security`), newline으로 분리해 base prompt 뒤에 append.

**Invocation:**

```bash
TMPERR=$(mktemp /tmp/codex-err-XXXXXX.txt)
_REPO_ROOT=$(git rev-parse --show-toplevel) || { echo "ERROR: not in a git repo" >&2; exit 1; }
cd "$_REPO_ROOT"

timeout 330 codex review "<prompt>" \
  --base <base> \
  -c 'model_reasoning_effort="high"' \
  --enable web_search_cached \
  < /dev/null 2>"$TMPERR"
_CODEX_EXIT=$?

if [ "$_CODEX_EXIT" = "124" ]; then
  echo "Codex stalled past 5.5 minutes. Common causes: model API stall, long prompt, network issue."
fi
```

Bash 호출에 `timeout: 300000` (5분) 사용.

**Stderr에서 비용 parse:**
```bash
grep "tokens used" "$TMPERR" 2>/dev/null || echo "tokens: unknown"
```

### Pass/Fail 게이트 로직

Review 완료 후 output에서 severity marker 스캔:

- Output에 `[P1]` 포함 → **GATE: FAIL** (critical finding — 머지 전 처리 필수)
- Output에 `[P2]` marker만 또는 finding 없음 → **GATE: PASS** (머지 승인)

게이트는 deterministic하고 grep 가능 — LLM 판단 루프 없음. 이게 review 모드를 머지 전제조건으로 안전하게 쓸 수 있게 만든다.

### Output 제시

```
CODEX SAYS (code review):
════════════════════════════════════════════════════════════
<full codex output, verbatim — do not truncate or summarize>
════════════════════════════════════════════════════════════
GATE: PASS                    Tokens: 14,331
```

또는

```
GATE: FAIL (N critical findings)
```

**Cleanup:**
```bash
rm -f "$TMPERR"
```

---

## 공유 JSONL parser

Challenge와 Consult 모드 둘 다 `codex exec --json` output을 이 Python parser로 pipe해 reasoning, tool call, 최종 응답 surface:

```python
import sys, json
for line in sys.stdin:
    line = line.strip()
    if not line: continue
    try:
        obj = json.loads(line)
        t = obj.get('type','')
        if t == 'item.completed' and 'item' in obj:
            item = obj['item']
            itype, text = item.get('type',''), item.get('text','')
            if itype == 'reasoning' and text:
                print(f'[codex thinking] {text}', flush=True)
            elif itype == 'agent_message' and text:
                print(text, flush=True)
            elif itype == 'command_execution':
                cmd = item.get('command','')
                if cmd: print(f'[codex ran] {cmd}', flush=True)
        elif t == 'turn.completed':
            usage = obj.get('usage', {})
            tokens = usage.get('input_tokens',0) + usage.get('output_tokens',0)
            if tokens: print(f'\ntokens used: {tokens}', flush=True)
    except: pass
```

아래 단계의 `<jsonl-parser>`는 `python3 -u -c "<위 스크립트>"`로 pipe.

---

## Step 2B: Challenge 모드 (Adversarial)

**목적:** 코드를 깨려 시도. 일반 리뷰가 놓칠 edge case, race condition, 보안 hole, 리소스 leak, silent 데이터 손상 경로 발견. Codex는 적대적 — 코드가 틀렸다 가정하고 증명.

**Prompt 템플릿** (기본, focus 없음):

> Review the changes on this branch against the base branch. Run
> `git diff origin/<base>` to see the diff. Your job is to find ways this code
> will fail in production. Think like an attacker and a chaos engineer. Find
> edge cases, race conditions, security holes, resource leaks, failure modes,
> and silent data corruption paths. Be adversarial. Be thorough. No compliments —
> just the problems.

**focus 있는 Prompt 템플릿** (예: `/codex challenge security`):

> Review the changes on this branch against the base branch. Run
> `git diff origin/<base>` to see the diff. Focus specifically on SECURITY.
> Your job is to find every way an attacker could exploit this code. Think
> about injection vectors, auth bypasses, privilege escalation, data exposure,
> and timing attacks. Be adversarial.

"focus" 단어 (security / performance / correctness / concurrency)가 "SECURITY"와 공격 vector body 교체. Adversarial framing 유지 — 이게 모드 작동시키는 prompt engineering 트릭.

**Invocation** (`codex exec`를 JSONL 스트리밍과 함께 사용 — 아래 "공유 JSONL parser" 참조):

```bash
TMPERR=$(mktemp /tmp/codex-err-XXXXXX.txt)
_REPO_ROOT=$(git rev-parse --show-toplevel) || { echo "ERROR: not in a git repo" >&2; exit 1; }

timeout 600 codex exec "<prompt>" \
  -C "$_REPO_ROOT" -s read-only \
  -c 'model_reasoning_effort="high"' \
  --enable web_search_cached --json \
  < /dev/null 2>"$TMPERR" | <jsonl-parser>
```

`--json` flag가 reasoning trace, tool call, 최종 응답을 JSONL 이벤트로 stream. Parser가 `[codex thinking]` 라인 surface해 모델의 reasoning을 결론만이 아니라 볼 수 있다. Challenge 모드에서 가치 있음 — reasoning 경로가 종종 실제 공격 vector 드러냄.

### Output 제시

```
CODEX SAYS (adversarial challenge):
════════════════════════════════════════════════════════════
<full output, verbatim>
════════════════════════════════════════════════════════════
Tokens: N
```

Challenge 모드는 **게이트 없음** — 모든 finding이 잠재 이슈지만, action 여부는 사람 (또는 Claude) 결정. Output은 informational, merge 전제조건 아님.

---

## Step 2C: Consult 모드 (Q&A)

**목적:** 전체 repo context로 Codex에게 무엇이든 질문. 세션 continuity 지원 — follow-up이 이전 턴 기억.

**Prompt 템플릿** (일반 consult):

> You are a brutally honest technical reviewer. Answer the following question.
> Be direct. Be terse. No compliments.
>
> <user's question>

**Prompt 템플릿** (plan review variant — plan inline embed):

> You are a brutally honest technical reviewer. Review this plan for: logical
> gaps and unstated assumptions, missing error handling or edge cases,
> overcomplexity (is there a simpler approach?), feasibility risks (what could
> go wrong?), and missing dependencies or sequencing issues. Be direct. Be
> terse. No compliments. Just the problems.
>
> THE PLAN:
> <full plan content, embedded verbatim>

**중요 — content embed, 경로 참조 금지.** Codex는 repo root (`-C`)에 sandbox. Repo 밖 파일 read 불가. Codex가 access 못 하는 경로 참조하면 tool call 낭비하고 실패. Content를 직접 read해 inline.

### 세션 continuity

기존 세션 체크:
```bash
cat .context/codex-session-id 2>/dev/null || echo "NO_SESSION"
```

세션 존재하면 질문:
```
You have an active Codex conversation from earlier. Continue or start fresh?
A) Continue (Codex remembers prior context)
B) Start a new conversation
```

**새 세션:**
```bash
timeout 600 codex exec "<prompt>" \
  -C "$_REPO_ROOT" -s read-only \
  -c 'model_reasoning_effort="medium"' \
  --enable web_search_cached --json \
  < /dev/null 2>"$TMPERR" | <jsonl-parser-with-session-id>
```

Consult 모드엔 JSONL parser 확장해 `thread.started` 이벤트에서 세션 ID 캡처: `if t == 'thread.started': print(f'SESSION_ID:{obj.get("thread_id","")}', flush=True)`. 출력된 `SESSION_ID:<id>` 라인을 follow-up 위해 `.context/codex-session-id`에 저장.

**재개 세션:** 새와 동일, `codex exec resume <session-id> "<prompt>" ...` invoke.

### Output 제시

```
CODEX SAYS (consult):
════════════════════════════════════════════════════════════
<full output, verbatim — includes [codex thinking] traces>
════════════════════════════════════════════════════════════
Tokens: N
Session saved — run /codex again to continue this conversation.
```

Codex 분석이 자기 이해와 다르면 output 뒤에 flag: "Note: I disagree on X because Y."

---

## Generic Wrapper 패턴 (다른 도구에 Adapt)

3-모드 패턴은 **모든 외부 second-opinion CLI**에 작동 — Codex는 교체 가능, 구조는 아님. 예시:

```bash
# Gemini CLI
gemini review --diff origin/<base> --prompt "<review prompt>"
gemini exec --prompt "<adversarial prompt>" --read-only
gemini exec --prompt "<consult prompt>" --read-only --session-id <id>

# 로컬 LLM (llama.cpp) — 내장 diff 지원 없음, stdin으로 diff 공급
git diff <base> | llama-cli -m model.gguf -p "$(cat review-prompt.txt)"
git diff <base> | llama-cli -m model.gguf -p "$(cat challenge-prompt.txt)"
echo "<question>" | llama-cli -m model.gguf -p "$(cat consult-prompt.txt)"
```

**substrate 간 portable:** 3-모드 taxonomy, `[P1]/[P2]` severity 컨벤션, challenge 모드의 adversarial framing, sandboxed CLI를 위한 embed-don't-reference 규칙, verbatim-output 규칙 (second opinion 절대 요약 금지 — 차갑게 제시).

**CLI 특정:** auth probe, sandbox flag, 세션 continuity API, 스트리밍 포맷 (JSONL vs SSE vs plain text).

---

## 비용 & 성능 노트

- **Codex CLI 사용은 OpenAI API 비용 유발.** Reasoning effort `high`는 typical PR review에 ~10-20K 토큰; `xhigh`는 ~23x 더 많은 토큰 사용, 큰 context에 50+ 분 hang 가능. Latency 허용되는 고위험 코드에 `xhigh` 예약.
- **Routine 체크엔 review 모드 선호.** 저렴, 빠름, 추가 reading 없이 action 가능한 binary 게이트 생성.
- **Challenge 모드는 고위험 코드 예약** — 보안 경계, 결제 flow, 데이터 migration, silent 실패가 실제 돈이나 신뢰 비용 유발하는 모든 것.
- **Consult 모드는 대화형** — 사용자가 reply 기다리므로 `medium` reasoning이 올바른 기본값.

---

## 중요 규칙

- **파일 수정 금지.** 이 스킬은 read-only. Codex는 read-only 샌드박스 모드 (`-s read-only`)에서 실행.
- **Output verbatim 제시.** 보이기 전 Codex output truncate, summarize, editorialize 금지. `CODEX SAYS` 블록 안에 full 표시.
- **Synthesis는 after, replace 아님.** 어떤 commentary든 full Codex output 뒤에, 대체 아님.
- **Review 모드엔 5분 timeout** (`timeout: 300000`), challenge / consult엔 10분.
- **이중 review 없음.** 사용자가 이미 주 review 실행했으면 Codex가 second 독립 opinion 제공. 주 review 재실행 금지.

---

## 에러 처리

| 에러 | 감지 | 응답 |
|-------|-----------|----------|
| Binary 없음 | Step 0 `which codex` empty 반환 | 중단. 설치 명령 출력. |
| Auth 실패 | Step 0.5 probe 실패, 또는 stderr가 `auth\|login\|unauthorized` 매치 | 중단. 사용자에게 `codex login` 실행 알림. |
| 외부 timeout (Bash) | Bash가 5/10분에 호출 kill | 사용자에게 재시도 또는 scope 축소 알림. |
| 내부 timeout (`timeout 124`) | Wrapper가 Bash 게이트 전에 fire | Actionable hang 메시지 출력; 사용자가 prompt 분할 또는 로그 체크 가능. |
| Empty 응답 | Output 파일 empty | 사용자에게 stderr 체크 알림. |
| 세션 재개 실패 | `codex exec resume` 에러 | 세션 파일 삭제, fresh 시작. |
