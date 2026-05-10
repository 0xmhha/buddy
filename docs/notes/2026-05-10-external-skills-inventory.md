# External Reference Inventory — `aidax-dag/ai-cli/` 4 경로

> **목적**: [`docs/notes/2026-05-10-missing-skills-inventory.md`](./2026-05-10-missing-skills-inventory.md) 의 *Step 2 외부 4 경로 탐색* 산출물. 각 경로 하위 프로젝트의 skill / agent / harness pattern / MCP 자산 인벤토리.
>
> **다음 Step (Step 3)**: 본 인벤토리 × missing-skills inventory 매트릭스 작성. 호환성 / 궁합 점검.
>
> **사용자 발화 (2026-05-10)**:
> > "여기에 존재하는 스킬들 중 우리 프로젝트로 가져와서 사용함으로써, 스킬 능력이 오히려 떨어질 수 있는 부분도 존재하니 무조건 가져올 것이 아니라, 우리의 'buddy' 프로젝트 스킬들과 호환 및 궁합이 좋은지 점검이 필요하다."
>
> 본 인벤토리는 *식별만*. 호환성 점검은 Step 3 이후.

---

## 1. 4 경로 top-level 인벤토리

| 경로 | README 제외 프로젝트 수 | 1줄 성격 |
|------|---------------------|--------|
| `/skill/` | 23 | skill 자산 — buddy plugin/skills/ 와 직접 경쟁 영역 |
| `/agent/` | 26 | agent 자산 — plugin/agents/ + cli buddy 의 agent runtime 입력 |
| `/harness/` | 29 | harness pattern — Claude Code 환경 / orchestrator / cli 자산 |
| `/mcp/` | 14 | MCP server — plugin/mcp/ + 그룹 3 (analytics-mcp / feature-management-mcp) 입력 |
| **합계** | **92** | |

---

## 2. `/skill/` — 23 프로젝트 (plugin buddy 직접 매칭 후보)

명명 / 경로명 기반 *영역 추정*. 실제 내용 검증은 Step 3 매트릭스 작성 시.

### 2.1 영역 분류

| 프로젝트 | 추정 영역 | missing-skills 매칭 후보 |
|---------|---------|----------------------|
| `ai-professional-replacement-legal-exploration_skill` | **법률 / 규제** | 그룹 2 신규 `review-legal-regulatory` |
| `korean-legal-guide_skill` | **법률 / 한국 규제** | 그룹 2 신규 `review-legal-regulatory` (지역 특화) |
| `patent-application-drafting_skill` | **법률 / 특허** | 그룹 2 신규 `review-legal-regulatory` (특허 영역) |
| `marketingskills` | **마케팅** | **그룹 4 stage 11 마케팅 지원** ⭐ |
| `designer-skills` | **디자인** | **그룹 4 stage 4 디자인 적용** ⭐ |
| `ui-design-brain` | **UI 디자인** | 그룹 4 stage 4 |
| `ui-ux-pro-max-skill` | **UI/UX** | 그룹 4 stage 4 |
| `Star-Office-UI` | **UI 라이브러리?** | 그룹 4 stage 4 (불확실) |
| `make-interfaces-feel-better` | **UI 보강** | 그룹 4 stage 4 |
| `better-icons` | **icon 자산** | 그룹 4 stage 4 (지원 영역) |
| `humanizer` | **텍스트 변환 / 자연어** | (charter 외) — agent 활용 가능 |
| `autoresearch-skill` | **연구 / 리서치** | 그룹 1 §1 conduct-customer-interview / analyze-market-size 후보 |
| `notebooklm-skill` | **NotebookLM 연계** | (charter 외) — domain context 영역 |
| `obsidian-skills` | **Obsidian 연계** | (charter 외) — domain knowledge management |
| `scholar-translator` | **학술 번역** | (charter 외) |
| `awesome-claude-skills` | **메타 컬렉션** | 인벤토리 — 다른 skill 의 source pool |
| `superpowers` | **메타 / 패턴** | buddy 의 docs/superpowers/ 와 명명 동일 — 패턴 source |
| `skills` | **generic skill 컬렉션** | source pool |
| `strands-agentskills` | **agent skill** | cli buddy 트랙 |
| `supabase-agent-skills` | **Supabase 통합** | (charter 외) — DB 영역 |
| `vercel-agent-skills` | **Vercel 통합** | (charter 외) — deploy / hosting 영역 |
| `KESE-KIT` | **KESE (한국 표준?)** | (불확실) |
| `rustunnel` | **도구 — tunneling?** | (불확실) |

### 2.2 missing-skills 매칭 quick-win 후보 ⭐

| 그룹 / 영역 | 외부 후보 |
|-----------|--------|
| 그룹 2 신규 `review-legal-regulatory` | 3 후보 (`ai-professional-replacement-legal-exploration_skill` / `korean-legal-guide_skill` / `patent-application-drafting_skill`) — **다양한 법률 영역, scope 결정 시 단서 풍부** |
| 그룹 4 stage 11 마케팅 | 1 후보 (`marketingskills`) — *직접 매칭* |
| 그룹 4 stage 4 디자인 적용 | 6 후보 (designer-skills / ui-design-brain / ui-ux-pro-max-skill / Star-Office-UI / make-interfaces-feel-better / better-icons) — **풍부**, scope / 책임 분리 검토 필요 |

### 2.3 charter scope 외 (외부 자산이 *기존 plugin buddy* 영역에 있는 경우)

| 프로젝트 | 기존 plugin buddy 영역 |
|---------|-------------------|
| `awesome-claude-skills` / `skills` | source pool — 다양한 skill 의 retrieval |
| `superpowers` | docs/superpowers/ 패턴 — buddy 자산과 명명 동일성 검증 필요 |

---

## 3. `/agent/` — 26 프로젝트 (cli buddy 트랙 + plugin/agents/ 영역)

| 프로젝트 | 추정 영역 | 트랙 |
|---------|---------|------|
| `agent-manager` | **agent 관리 패턴** | **cli buddy** ⭐ — 진짜 목적과 직접 정합 |
| `agent-browser` | agent + browser 통합 | cli buddy |
| `agent-evaluation` | agent 평가 | cli buddy 의 quality gate |
| `agentation` | agent + automation 패턴 | cli buddy |
| `agentic-ai-prompt-research` | agent prompt 패턴 | plugin buddy + cli buddy |
| `agentkit-samples` | agent SDK 예시 | cli buddy |
| `atomic-agents` | atomic agent 패턴 | plugin/agents/ |
| `auto-researchtrading` | 자동 거래 — 도메인 특화 | (charter 외 — 사용 예시 reference 가능) |
| `autonomous-coding-agents` | 자동 coding agent | plugin/agents/ + cli buddy |
| `awesome-agents` | meta 컬렉션 | source pool |
| `deepagents` | deep agent | plugin/agents/ |
| `deer-flow` | flow 기반 agent | cli buddy 의 orchestration |
| `eko` | (불확실) | (불확실) |
| `gpt-researcher` | research agent | plugin/agents/ + 그룹 1 §1 conduct-customer-interview |
| `graphify` | graph + agent | (불확실) |
| `heartbeat-operator` | operator 패턴 | cli buddy 의 supervision |
| `langgraph` | LangGraph framework | cli buddy 의 orchestration |
| `multi-agent-shogun` | multi-agent | cli buddy ⭐ — 사용자 발화 "여러 자동화된 agent 관리" 직접 정합 |
| `openfang` | (불확실) | (불확실) |
| `owl` | (불확실) | (불확실) |
| `robot-agentic-ai-dev` | robotics agent | (charter 외) |
| `sage` | (불확실) | (불확실) |
| `trae-agent` | (불확실) | (불확실) |
| `UI-TARS-desktop` | UI 자동화 desktop agent | cli buddy ⭐ — 웹툰 agent 같은 사용 예시 reference |
| `A2A` | agent-to-agent | cli buddy 의 communication |
| `AP2` | (불확실 — agent protocol?) | (불확실) |

### 3.1 cli buddy 트랙 직접 정합 ⭐

| 후보 | 이유 |
|------|------|
| `agent-manager` | "agent 관리" 명명 — cli buddy 의 진짜 목적 (agent 생성/실행/종료/설정) 직접 매칭 |
| `multi-agent-shogun` | "multi-agent" — 사용자 발화 "여러 자동화된 agent 설정 / 관리" 정합 |
| `UI-TARS-desktop` | UI 자동화 desktop agent — 웹툰 agent 같은 사용 예시의 reference |

### 3.2 plugin/agents/ 영역 후보

`atomic-agents`, `deepagents`, `gpt-researcher` 등 — buddy plugin 의 sub-agent 정의 input.

---

## 4. `/harness/` — 29 프로젝트 (Claude Code 환경 / cli 자산)

charter §6.3 의 *호환성 우선 원칙* — *외부 자산 무조건 차용 X*. 본 경로는 *Claude Code 환경 / orchestrator / cli 패턴* 이라 plugin buddy 의 *PROCEDURE 양식* / cli buddy 의 *TUI / runtime* 패턴 reference.

### 4.1 Claude Code 직접 관련 (15)

| 프로젝트 | 영역 |
|---------|------|
| `claude-code` | Claude Code 자체 |
| `claude-code-organizer` | Claude Code 설정 / 세션 관리 — HANDOFF.md 의 "5단계 비전" 에 직접 언급된 reference |
| `claude-code-templates` | template pattern |
| `claude-config-editor` | config 편집 |
| `claude-for-android` | mobile |
| `claude-task-master` | task DAG (roadmap.md §5 v0.3 reference) |
| `claude-squad` | multi-instance |
| `oh-my-claude-desktop` | desktop wrapper |
| `oh-my-claudecode` | enhanced Claude Code |
| `oh-my-opencode` | OpenCode wrapper |
| `everything-claude-code` | meta 컬렉션 |
| `all-in-one-claude-code` | meta 컬렉션 |
| `clawflows` | flow + Claude Code (HANDOFF "메타-패턴" reference) |
| `openclaw` | (Claude 변형 / wrapper) |
| `kamar-taj` | (불확실) |

### 4.2 alternative cli (5)

| 프로젝트 | 영역 |
|---------|------|
| `codex` | OpenAI Codex CLI |
| `gemini-cli` | Gemini CLI |
| `opencode` | OpenCode |
| `dyad` | (불확실) |
| `rikkahub` | (불확실) |

### 4.3 orchestration / supervision pattern (5)

| 프로젝트 | 영역 |
|---------|------|
| `cli-wrapper` | cli buddy 의 v0.1 의존 (`0xmhha/cli-wrapper`) — 이미 buddy 가 차용 |
| `automaton` | 자동화 패턴 |
| `ghost-os` | (불확실 — OS 추상?) |
| `vibe-sunsang` | (불확실) |
| `shuri` | (불확실) |

### 4.4 기타 (4)

| 프로젝트 | 영역 |
|---------|------|
| `context-mode` | context 관리 (charter scope 외) |
| `gstack` | buddy README 에 명시된 차용 source (Garry Tan) |
| `get-shit-done` | productivity skill (charter 외) |
| `memory-bank` | memory pattern |

### 4.5 cli buddy 트랙 reference 후보 ⭐

| 후보 | 이유 |
|------|------|
| `claude-code-organizer` | HANDOFF "5단계 비전" 에 명시된 reference — TUI 설정 / 세션 관리 패턴 |
| `claude-task-master` | task DAG executor (cli buddy 의 task 단위 spec 작성 시 reference) |
| `claude-squad` | multi-instance (cli buddy 의 *여러 agent 동시 관리* 패턴) |

---

## 5. `/mcp/` — 14 프로젝트 (그룹 3 + plugin/mcp/ 입력)

| 프로젝트 | 추정 영역 | 매칭 |
|---------|---------|------|
| `agent-forge` | agent 생성 MCP | cli buddy ⭐ |
| `arxiv-mcp-server` | 학술 검색 | (charter 외 — 도메인) |
| `codebase-memory-mcp` | 코드 memory | plugin/mcp/ — context |
| `Connectome` | (불확실) | (불확실) |
| `context7` | context 7 (이미 buddy 가 사용) | 기존 의존 |
| `go-sdk` | MCP Go SDK | **cli buddy 의 MCP server 구현 의존** ⭐ |
| `korean-law-mcp` | 한국 법률 MCP | 그룹 2 review-legal-regulatory 자산 후보 |
| `mcp-go` | MCP Go (vs go-sdk 차이?) | (불확실) |
| `mcp-shield` | MCP 보안 | (charter 외 — 보안) |
| `MCP-Zero` | MCP 기초 | (불확실) |
| `pdf-translator-mcp` | PDF 변환 | (charter 외) |
| `ros-mcp-server` | ROS robotics | (charter 외) |
| `tavily-mcp` | Tavily 검색 | 그룹 1 §1 conduct-customer-interview / analyze-market-size 자산 후보 |
| `Unity-MCP` | Unity 통합 | (charter 외) |

### 5.1 그룹 3 (analytics-mcp / feature-management-mcp) 매칭

| 그룹 3 후보 | 외부 reference |
|-----------|--------------|
| `analytics-mcp` | (직접 매칭 없음 — 신규 작성) |
| `feature-management-mcp` (cli buddy 트랙) | (직접 매칭 없음 — 신규 작성) |

→ 두 MCP 모두 *외부에 직접 매칭 없음*. 신규 작성 시 `go-sdk` / `mcp-go` 가 *구현 의존* 으로 활용.

---

## 6. 종합 — 92 프로젝트 활용 정도

| 활용 정도 | 카운트 | 영역 |
|---------|------|------|
| **직접 매칭 (quick-win)** | ~12 | 그룹 2 (3) + 그룹 4 stage 4 (6) + stage 11 (1) + cli buddy 직접 (3) — Step 3 매트릭스의 우선 검토 대상 |
| **간접 매칭 (영역 reference)** | ~25 | 그룹 1 일부 + plugin/agents/ source pool + harness pattern |
| **(불확실 — 명명 부족)** | ~20 | 추가 탐색 필요 (각 프로젝트 README 검토) |
| **charter scope 외** | ~30 | 도메인 특화 / 무관 |
| **이미 buddy 가 사용** | 2 | `cli-wrapper`, `context7` |
| **buddy README 명시 source** | 2 | `gstack`, `awesome-claude-skills` (mattpocock/skills 변형 가능) |

---

## 7. Step 3 진행 시 우선 검토 항목 (~12)

매트릭스 작성 시 *모든 92 프로젝트* 깊이 분석은 비효율. 우선순위:

| 우선순위 | 프로젝트 | 매칭 |
|--------|---------|------|
| 1 | `ai-professional-replacement-legal-exploration_skill` | 그룹 2 review-legal-regulatory |
| 2 | `korean-legal-guide_skill` | 그룹 2 review-legal-regulatory |
| 3 | `patent-application-drafting_skill` | 그룹 2 review-legal-regulatory |
| 4 | `marketingskills` | 그룹 4 stage 11 |
| 5 | `designer-skills` | 그룹 4 stage 4 |
| 6 | `ui-design-brain` | 그룹 4 stage 4 |
| 7 | `ui-ux-pro-max-skill` | 그룹 4 stage 4 |
| 8 | `make-interfaces-feel-better` | 그룹 4 stage 4 |
| 9 | `agent-manager` | cli buddy 진짜 목적 |
| 10 | `multi-agent-shogun` | cli buddy 진짜 목적 |
| 11 | `claude-code-organizer` | cli buddy + plugin buddy 5단계 비전 |
| 12 | `claude-task-master` | cli buddy task DAG (roadmap §5) |

추가로 그룹 1 (29 미구현 skill) 의 매칭 후보는 *Step 3 매트릭스 작성 시 발화* — 본 인벤토리는 *명명 / 경로 기반 추정* 만.

---

## 8. 다음 액션

1. 본 외부 자산 인벤토리 commit
2. **Step 3 — 매트릭스 작성** 진입:
   - 위 §7 의 12 우선순위 프로젝트 *상세 검토* (각 프로젝트 README + skill 목록)
   - missing-skills (그룹 1 + 2 신규 1 + 그룹 4) 와의 매핑
   - 호환성 / 궁합 점검 (PROCEDURE 양식 / 명명 / cascade 정합)
   - per-skill 결정: (a) 그대로 차용 / (b) 수정 차용 / (c) 참고만 + 신규 / (d) 무관 + 신규
3. **그룹 4 명명 / scope 결정** (사용자 D-A 권장 옵션 — Step 2 후 결정) — Step 3 매트릭스에서 외부 자산 단서 기반 제안
4. Step 4 — `docs/superpowers/plans/<date>-skill-completion-plan.md` 작성
5. Step 5 — 실제 skill 작성 (skill 별 PR 단위)
