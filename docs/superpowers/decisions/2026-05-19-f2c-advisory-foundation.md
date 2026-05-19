# ADR-014 — F2.C Advisory foundation: knowledge retrieval (W7-3, v1.0 entry C-3 split into 3 phases)

**Status**: Accepted (2026-05-19)
**Authors**: mhha (cli buddy track)
**Supersedes**: —
**Related**: ADR-009 (cli buddy vision expansion — F2.C area), ADR-010 (v1.0.0 scope — C-3 condition), ADR-011 (release policy — milestone-driven), ADR-012 (F2.A Session Monitor — W7-1 substrate), ADR-013 (F2.B Usage Analysis — W7-2 substrate)

## Context

ADR-009 declared F2.C Advisory as "actionable Korean prose recommendations from usage analysis". Initial 4-Q design poll surfaced a fundamental gap in the rule-based proposal:

> **User feedback 2026-05-19 (F2.C Q1)**:
> "rule-based 가 가능한 영역이야?? 사람들마다 사용하는 방식이 모두 다를텐데, 그때마다 분석이 필요하지 않을까?? ... 토큰을 비효율적으로 사용한다면, prompt 사용과 skill 사용등에 있어서 추천 제안을 해줄수 있다는 것이야. 또한, 유저의 사용 패턴을 통해, 반복적인 작업에 대해서는 skill 을 generate 하여, 해당 스킬을 사용하도록 해줄수도 있어. 이와 같은 방식에 대해서, llm-driven 으로 처리하는것이 효율적인데, claude api 호출을 하지 않기 위해서는 buddy plugin 의 agent 기능으로 구현 하면 효율적일것 같아. ... session 별 사용기록들을 knowledge data 로 청킹과 임베딩을 통해 vector 화 하여 기록할수 있도록 기능을 제공하는 방법이 있을것 같아."

Implication: F2.C 의 진정한 가치는 단순 metric threshold alert 이 아니라 **개인 사용 패턴 분석 → prompt/skill 추천 + 반복 작업 skill autogen 제안**. 이를 deterministic rule 만으로 처리하기 어렵고, Claude API 호출 비용 / 외부 의존도 회피하려면 **local knowledge retrieval + 자체 embedding/BM25 인프라**가 필요.

Current fragments:

- W7-1 sessions table (transcript_path / token usage 등) — knowledge source.
- W7-2 usage.Service — metric primitive.
- `internal/agent/` (W3-3) — Python sub-process 호출 가능 (Executor 가 임의 shell 명령 지원).
- MCP server (`buddy-mcp`) — 외부 tool surface 준비됨.

Gap: knowledge data 의 chunking / embedding / BM25 / vector store 가 전무. Python 의존성 관리 정책 (venv) 미정. Local embedding model 선택 미정.

## Decision

**W7-3 = F2.C 의 3-phase split. v0.10.0 은 첫 phase (knowledge retrieval foundation) 만 ship. C-3 condition 은 v0.11.0 (advisor phase) 종료 시 closed.**

### Phase split

| Version | Phase | Scope | Closes |
|---------|-------|-------|--------|
| **v0.10.0** (W7-3a, this ADR) | **Foundation** | Knowledge pipeline: chunking + BM25 + local embeddings + vector store + MCP `knowledge_query` + CLI `buddy knowledge`. No advisory generation logic yet — retrieval primitive only. | — (C-3 still open) |
| **v0.11.0** (W7-3b) | **Advisor** | Retrieval + rule + (optionally LLM via local Python agent) → friend-tone Korean advisory. CLI `buddy advise` / TUI Usage pane advisory section / MCP `usage_advise`. | **C-3** |
| **v0.12.0** (W7-3c) | **Skill autogen** | 반복 패턴 detect → skill spec generate proposal. User confirm → plugin/skills/ 에 추가 PR. | — (post-v1.0 scope) |

→ ADR-014 = Phase 1 (foundation) 만 lock-in. v0.11 / v0.12 의 design ADR 은 trigger 발화 시 별도 작성.

### v0.10.0 (Phase 1) design choices

**Q1 Stack — Python agent (local ML 생태계)**

- buddy agent (W3-3) 의 Executor 가 Python sub-process 호출.
- Python deps: `sentence-transformers` (embedding) + `rank-bm25` (sparse) + `chromadb` 또는 `sqlite + numpy` (vector store).
- 사용자 측 venv 자동 설치는 `buddy install --with-knowledge` 옵션으로 (W7-3a follow-on). v0.10.0 ship 은 *manual venv* 가이드 + 의존성 미설치 시 friend-tone "Python venv 가 안 보여..." 메시지.

**Q2 Vector store — SQLite + JSON blobs (zero-CGo)**

- `chunks` 테이블 (new migration v6): `id / session_id / content / embedding (BLOB, JSON float32 array)`. modernc.org/sqlite pure Go.
- 작은 N (~10k chunks 가정) 에서는 in-memory cosine similarity 충분. v0.11+ N 증가 시 chromem-go 또는 sqlite-vec extension 도입 검토.

**Q3 BM25 — Go-only (별도 lib 없이 자체 구현, 작은 구현 size)**

- BM25 는 알고리즘 자체가 단순 (~50 LoC Go). bleve 도입 (전체 full-text engine) 은 v0.10.0 에 과잉.
- term posting list 는 `chunks` 테이블의 derived view 로 구성.

**Q4 Knowledge surface (v0.10.0 ship)**

- **CLI**:
  - `buddy knowledge ingest [--session <id>|--all] [--rebuild]` — sessions transcript → chunks + embedding (Python agent 호출).
  - `buddy knowledge query <text> [--k N]` — BM25 + vector hybrid retrieval. top-N chunks + score.
  - `buddy knowledge stats` — chunk count / index size / 마지막 ingest 시각.
- **MCP**: `knowledge_query(text, k)` → top-N chunks (W7-3b advisor 가 input 으로 사용).
- TUI: v0.10.0 에서는 추가 안 함. v0.11 advisor phase 에서 Usage pane 의 advisory section 과 함께.

## Alternatives considered

### Option A — End-to-end ship in v0.10.0 (rejected)

Knowledge layer + advisor + skill-gen 한 ship. 4-6 주 소요 예상. milestone-driven 원칙 위반 (단일 milestone 이 너무 큼). User 의 "3 phase 분해" 답으로 reject.

### Option B — LLM-driven first (rejected)

Claude API 호출 → advisory. Q1 사용자 답 명시: "claude api 호출을 하지 않기 위해서는 buddy plugin 의 agent 기능으로 구현". Local stack 선호.

### Option C — Go-only stack (rejected)

chromem-go (vector) + bleve (BM25) + Go embedding model (ONNX runtime via CGo). zero-CGo 정책 위반 + Go embedding model 생태계 빈약. Python agent 가 더 합리.

### Option D — Rule-based fixed threshold (rejected)

doctor 패턴 그대로. Q1 사용자 답 명시: "사람들마다 사용 방식이 다른데 일률 rule 한계". reject.

## Consequences

- **`internal/db/migrations.go` v6** — `chunks` 테이블 추가:
  ```sql
  CREATE TABLE chunks (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id  TEXT    NOT NULL,
    content     TEXT    NOT NULL,
    token_count INTEGER NOT NULL,
    embedding   BLOB,                            -- JSON float32 array; NULL until embedding ingest
    created_at  INTEGER NOT NULL,
    FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE
  );
  CREATE INDEX idx_chunks_session ON chunks(session_id);
  CREATE INDEX idx_chunks_created ON chunks(created_at);
  ```
- **`internal/knowledge/`** (new package):
  - `chunker.go` — transcript JSONL → chunks (user message + 응답 그룹 단위, max 500 토큰).
  - `bm25.go` — BM25 sparse retrieval over chunk content.
  - `vector.go` — cosine similarity over embedding BLOB.
  - `hybrid.go` — BM25 + vector 결합 score (reciprocal rank fusion).
  - `store.go` — chunks CRUD.
  - `embedder.go` — Python sub-process invoker. Executor 가 buddy/scripts/embed.py 호출.
- **`scripts/embed.py`** (new) — sentence-transformers 호출 stdin/stdout (JSON 라인 in / JSON BLOB out).
- **`cmd/buddy/knowledge_cmd.go`** — 3 subcommand.
- **`internal/mcp/knowledge_tool.go`** — `knowledge_query` MCP tool.
- **`internal/mcp/server.go`** — `addKnowledgeTools(s, opts)` + `Options.Knowledge`.
- **`docs/cli-buddy-spec.md` §9 W6** → "split into W6a foundation (v0.10.0) + W6b advisor (v0.11.0) + W6c skill-gen (v0.12.0)".
- **No code change to advisory generation** — v0.10.0 은 retrieval primitive 만. CLI `buddy advise` / MCP `usage_advise` 는 v0.11.0 ship.
- **Python venv 의존**: buddy 가 자동 설치 안 함. `~/.buddy/venv/bin/python3 -m pip install -r requirements.txt` 가 사용자 manual step. 미설치 시 friend-tone 메시지.
- **`CHANGELOG.md [0.10.0]`** — milestone-driven format. C-3 status 는 "여전히 open (Phase 1 만 ship)" 으로 명시.
- **BACKLOG / HANDOFF** — C-3 still 0 closed; v0.10.0 ship 은 retrieval foundation 으로 표기.
- **v1.0.0 entry update**: 5/9 closed 유지 (v0.10.0 가 C-3 닫지 않음). v0.11.0 ship 시 6/9.

## Verification

When W7-3a ships (v0.10.0):

```bash
# After buddy v0.10.0 install + Python venv 준비 + sessions 데이터:

buddy knowledge ingest --all                  # transcript → chunks + embedding
buddy knowledge stats                          # chunk count visible
buddy knowledge query "어떤 skill 이 자주 실패해" --k 5
                                              # top-5 retrieved chunks + score

# MCP 측 (Claude Code 안):
# mcp__buddy__knowledge_query                   # JSON top-N chunks

go test -race -count=1 ./internal/knowledge/... # 8+ race-clean tests
make verify-versions                            # 5 sources on 0.10.0
```

Python 의존 없는 환경에서 (`buddy knowledge ingest` 호출 시 venv 없음) → friend-tone error 메시지 출력, exit 1. CI 는 Python 의존 미설치 환경에서 `buddy knowledge stats` (read-only) 만 동작 확인.

## Trigger to revisit

- Python venv 자동 설치 UX 가 user 마찰 일으킴 → `buddy install --with-knowledge` 옵션 도입 (v0.10.x).
- Chunk count 가 10k 넘어가 in-memory cosine 느려짐 → chromem-go 또는 sqlite-vec 도입 (v0.11+).
- sentence-transformers 모델 size (~100MB) 가 disk footprint 문제 → 더 작은 모델 (e.g., bge-micro) 선택지 추가.
- 사용자가 advisor (v0.11.0) ship 후 "local 모델 품질이 부족, Claude API 호출 옵션 필요" 피드백 → v0.12+ hybrid LLM path 재검토 (사용자 명시 의향 변경 시).

## References

- ADR-009 (cli buddy vision expansion) — defines F2.C area.
- ADR-010 (whole-product v1.0.0 scope) — C-3 condition that v0.11.0 (Phase 2) closes.
- ADR-011 (release policy) — milestone-driven v0.10.0 framing.
- ADR-012 / ADR-013 — substrate (sessions table + usage metric primitive).
- User feedback 2026-05-19 (F2.C Q1 + scope split + stack) — full quotes in §Context.
- sentence-transformers — https://www.sbert.net/
- rank-bm25 — https://github.com/dorianbrown/rank_bm25
- chromem-go (future trigger) — https://github.com/philippgille/chromem-go
