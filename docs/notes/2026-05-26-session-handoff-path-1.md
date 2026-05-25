# 세션 핸드오프 — 2026-05-26 Path 1 종결 (plugin skills audit + H1 + ADR-019)

> **다음 세션의 진입 지점**. 본 doc + `git log -8 --oneline` + `docs/BACKLOG.md` + `docs/superpowers/decisions/2026-05-26-plugin-mcp-exposure-layer.md` (ADR-019) 만 읽으면 *cross-machine 컨텍스트 0* 으로 이어받을 수 있다.
>
> **표기 약속**: 본 doc 안의 `<REPO_ROOT>` 는 buddy 레포 working tree 루트, `<HOME>` 은 OS 사용자 홈. 본 머신 (`kevin`) 에서는 각각 `/Users/kevin/work/github/0xmhha/buddy` / `/Users/kevin` 이지만, *다른 머신의 reader 는 자신의 경로로 substitution* 한다.

---

## 0. 다른 머신 / 새 환경에서 시작 시 (전제 조건)

같은 머신·같은 세션이면 §1 로 직진. *다른 머신* 또는 *clone 된 신선한 working tree* 라면 아래 항목 먼저 확인.

| 도구 / 자원 | 필요성 | 부재 시 대응 |
|------------|-------|------------|
| `git` + remote 접근 | 필수 — 본 세션의 5 commits 모두 `origin/main` 에 push (HEAD `757c9a8`). `git pull --ff-only origin main` 으로 동기 | 다른 방법 없음 |
| Go toolchain (1.25+) | 조건부 — ADR-019 implementation session 진입 시 필수 (cross-compile 4 binary). 본 session 의 doc-only 작업은 Go 불필요 | brew / asdf / 공식 설치 |
| `ripgrep` (`rg`) | 권장 — §3 의 audit doc 정합성 검증 시. 부재 시 `grep -rn` 대체 | 없음 |
| GPG key (commit signing) | 조건부 — 본 레포가 GPG signed commit 강제하면 필수 | dotfiles 동기 또는 임시 `git -c commit.gpgsign=false commit` |
| buddy plugin install | 권장 — ADR-019 §5 Verification 의 dogfood 단계에서 필수 | 부재 시 implementation 끝까지 진행 불가 (dogfood gate) |
| `~/.claude/CLAUDE.md` 글로벌 룰 | 권장 — 한국어 응답 / commit attribution 금지 / 개발단계 용어 금지. §10 에 핵심 재명시 | dotfiles 동기 또는 §10 만 신뢰 |
| Memory files | 본 머신 한정 — git sync 대상 아님. 본 세션은 신규 memory 추가 0건 | (a) 다른 머신은 *해당 머신의 memory 만*, (b) 영향 0 |

### 0.1 다른 머신 첫 5분 권장 sequence

```bash
# 1. clone 또는 pull
git clone <remote-url> <REPO_ROOT>     # 처음이면
cd <REPO_ROOT>
git fetch --all
git pull --ff-only origin main         # HEAD 757c9a8 이상 확인

# 2. 핸드오프 + SSoT 읽기 (5 분)
cat docs/notes/2026-05-26-session-handoff-path-1.md      # 본 doc
cat docs/superpowers/decisions/2026-05-26-plugin-mcp-exposure-layer.md  # ADR-019 (다음 세션 진입점)
cat docs/plugin-skills-engineering-flow.md               # 광의 SE coverage
cat docs/plugin-skills-inventory.md | head -60           # 152 skill inventory 요약

# 3. 무결성 검증 (이전 세션의 결과 깨지지 않았는지)
go vet ./... 2>&1 | tail -5
go build ./... 2>&1 | tail -5
make test-routing 2>&1 | tail -15      # 10/10 통과 필수

# 4. 작업 trail spot check
git log -8 --oneline                   # 본 세션 5 commits 포함
git show 757c9a8 --stat                # ADR-019 (최종)
```

---

## 1. 한 줄 요약

본 세션은 *plugin 트랙 skills audit* 와 *H1 (DDD ubiquitous-language) skill 신설* + *ADR-019 plugin MCP exposure layer (Option D 결정)* 의 **Path 1 종결**. 5 commits / +1660 insertions / -16 deletions / push 완료. **Tier 1 (G1/G2/G3) 모두 closed + A 카테고리 13/13 closed + ADR 18→19**. 다음 cycle = ADR-019 implementation (별 세션, 1-3일).

---

## 2. 현재 working tree 상태 (세션 종결 시점)

| 항목 | 값 |
|------|---|
| 작업 디렉토리 | `<REPO_ROOT>` (본 머신 예시: `/Users/kevin/work/github/0xmhha/buddy`) |
| Branch | `main` |
| Last pushed commit | `757c9a8` — `docs(decisions): add ADR-019 plugin MCP exposure layer` |
| Remote sync | `origin/main` 과 동기 (last push) |
| Uncommitted modified files | 0 |
| Untracked files | 0 (본 doc commit 후 0) |

확인 명령:
```bash
cd <REPO_ROOT>
git status               # "작업 폴더 깨끗함"
git log -1 --oneline     # 757c9a8 또는 본 doc commit 직후
```

---

## 3. 본 세션 작업 결과 (5 commits chronology)

### 3.1 Commit 1 — `62e6dec` docs(plugin): audit skill inventory and engineering coverage

3 신규 audit doc 생성 (총 +1133 insertions). 각 doc 의 책임:

| Doc | 책임 |
|-----|------|
| `docs/plugin-skills-inventory.md` | 152 PROCEDURE.md baseline + 103 commands + charter §2.4 4 자산 격차 + 7 gap (Tier 1-3) |
| `docs/plugin-skills-engineering-audit.md` | SKILLS_ANALYSIS A 카테고리 *좁은 13 영역* + buddy 현 자산 cross-ref + 결정 후보 |
| `docs/plugin-skills-engineering-flow.md` | 이론 SE flow (10 phase × ~76 step, 13 source 통합) ↔ buddy 광의 ~106 skill 매핑 + hole 5 + 부분 cover 5 + 책임 모호 4 + cascade ~70% 단절 |

### 3.2 Commit 2 — `cd99818` feat(skill): add audit-ubiquitous-language for DDD vocabulary consistency

H1 신규 skill — DDD ubiquitous language (Evans 2003 / Vernon 2013) 영감, *inspired-by* 분류 (mattpocock 본문 cross-machine 부재로 *0 read*).

| 변경 | 위치 |
|------|------|
| 신규 PROCEDURE | `plugin/skills/audit-ubiquitous-language/PROCEDURE.md` (~280줄) |
| 신규 command | `plugin/commands/audit-ubiquitous-language.md` (17줄) |
| catalog 등재 | `skill-catalog.md` §6 Quality `audit-test-coverage-meaningful` 직후 |
| NOTICE | DDD inspired-by entry 추가 (14줄) |
| test-router count | 151 → 152 lockstep |
| Cross-link 3건 | `define-features` (post-feature vocab 검증 옵션), `refactor-with-rename-trace` (선행 어휘 결정 step), `review-architecture` (직교 어휘 차원) |

스킬 특성:
- Type: **discipline-enforcing**
- Phase: §6 Quality + Cross-cutting cross-link
- Trigger: command + dispatch
- 출력: YAML — drift findings + health_score + glossary_updates + next_steps

### 3.3 Commit 3 — `9831756` docs(skill): close catalog duplicates G2/G3 and sync audit docs after H1

| 변경 | 의미 |
|------|------|
| `monitor-regressions` §6 라인 제거 | G2 closed. §8 Operate 만 유지 (production traffic 자연 phase) |
| `apply-builder-ethos` §1 라인 제거 | G3 closed. Cross-cutting Utilities 만 유지 (phase-agnostic 자연) |
| inventory §3 갱신 | H1 / G2 / G3 / LOW 2 모두 closed 마킹 |
| engineering-audit §0 갱신 | #11 closed, 흡수 분류 downgrade (adopt-with-edits → inspired-by) 명시 |
| engineering-flow §0 갱신 | H1+H3 closed, C1+C2 closed, Phase 2 cover 70→80%, Cross-cutting 70→80%, Tier 1 모두 closed |
| SKILLS_ANALYSIS §3 A.1 갱신 | LOW 2 ✅ closed, LOW 1/3 결정 보류 마킹 (engineering-audit 권장 cross-ref) |

### 3.4 Commit 4 — `b70712d` docs(skill): finalize Tier 1 closure (G1 status entry) and persist LOW 1/3 + #8 decisions

| 결정 | 영속화 |
|------|------|
| **G1 status catalog 등재** | catalog Cross-cutting Utilities 표 *알파벳 위치* (`guide-setup-wizard` 와 `write-a-skill` 사이) 에 1줄 추가 — phase 무관 artifact-detection utility |
| **#8 사고 확장 종결** | engineering-audit §2.1 ✅ 종결 확정 (Gap 무시, 9-phase cascade cover) |
| **LOW 1 (위생/안전) 종결** | engineering-audit §2.2 ✅ 종결 확정 (흡수 안 함, buddy 이미 동등/우월 cover) |
| **LOW 3 (caveman 컨텍스트) cli 트랙 이동** | engineering-audit §2.3 🔄 확정 (runtime token state 영역, cli buddy Wave 7 W7-2 trigger) |
| inventory G1 본문 마킹 | "✅ closed (사용자 명시 2026-05-26, Path 1)" |
| SKILLS_ANALYSIS LOW 1/3 마킹 | 종결 / 트랙 이동 확정 |

### 3.5 Commit 5 — `757c9a8` docs(decisions): add ADR-019 plugin MCP exposure layer

새 ADR Accepted (2026-05-26):

| 결정 | 값 |
|------|-----|
| Distribution model | **Option D — hybrid** (PATH-first `buddy` + bundled fallback) |
| Launcher | `plugin/bin/buddy-mcp-launcher.sh` (POSIX, ~20줄) |
| Bundled binaries | 4 (`darwin/linux × arm64/amd64`) |
| Manifest 추가 | `mcpServer: {command: ${CLAUDE_PLUGIN_ROOT}/bin/buddy-mcp-launcher.sh, transport: stdio}` |
| Charter §2.4 약속 (current) | 25% → 50% (implementation 후) |
| 거부된 alternative | A (no PATH preference) / B (install-time download) / C (PATH-only no bundle) — 각 rationale 본문 명시 |
| Verification gate | 4 항목 (test-router-wireup check / make build-mcp-bundled / dogfood note / charter §2.4 갱신) |
| 미결정 | Bundled distribution mechanism (git-tracked / Git LFS / release artifact) — implementation session 결정 |

ADR README index 갱신 — ADR-019 행 추가 + "향후 ADR 후보" 의 `skill MCP exposure` row closed-by-ADR-019 마킹.

---

## 4. 본 세션의 결정 영속화 매트릭스

### 4.1 A 카테고리 13 영역 (engineering-audit)

| # | 영역 | 직전 | 본 세션 종결 |
|---|------|------|------------|
| 1 | TDD | ✅ | ✅ |
| 2 | 디버깅 | ✅ | ✅ |
| 3 | 플래닝 | ✅ | ✅ |
| 4 | 코드 리뷰 | ✅ | ✅ |
| 5 | 에이전트 협업 | ✅ | ✅ |
| 6 | 완료 검증 | ✅ | ✅ |
| 7 | 브랜치 종료 | ✅ | ✅ |
| **8** | **사고 확장** | 🟡 부분 | **✅ 종결 (Gap 무시)** |
| **9** | **위생/안전 (LOW 1)** | ❌ | **✅ 종결 (흡수 안 함)** |
| **10** | **컨텍스트 핸드오프 (LOW 3)** | ❌ | **🔄 cli 트랙 이동** |
| **11** | **DDD/도메인 (LOW 2)** | ❌ | **✅ 완료 (cd99818)** |
| 12 | 부트스트랩 | ✅ | ✅ |
| 13 | 메타-스킬 작성 | ✅ | ✅ |
| **합계** | | 9+1+3 | **13/13 closed** |

### 4.2 plugin-skills-inventory Tier

| Tier | 항목 | 직전 | 본 세션 종결 |
|------|------|------|------------|
| Tier 1 G1 | status catalog 등재 | ❌ | ✅ (b70712d) |
| Tier 1 G2 | monitor-regressions 중복 | ❌ | ✅ (9831756) |
| Tier 1 G3 | apply-builder-ethos 중복 | ❌ | ✅ (9831756) |
| Tier 2 G5 LOW 2 | DDD ubiquitous language | ❌ | ✅ (cd99818) |
| Tier 1 전체 | 3 open | **0 open** |

### 4.3 engineering-flow Tier 1 + cover

| 항목 | 직전 | 본 세션 종결 |
|------|------|------------|
| H1 ubiquitous language hole | ❌ | ✅ |
| H3 cross-cutting 어휘 hole | ❌ | ✅ (H1 동일 skill cover) |
| C1 monitor-regressions 중복 | ❌ | ✅ |
| C2 apply-builder-ethos 중복 | ❌ | ✅ |
| Phase 2 cover | 7/10 (70%) | **8/10 (80%)** |
| Cross-cutting cover | 7/10 (70%) | **8/10 (80%)** |
| 명확한 hole | 5 | **3** (H2 cli 트랙, H4/H5 잠재) |
| Tier 1 (H1/C1/C2) | 3 open | **0 open** |

---

## 5. 다음 세션 진입 순서

### Step 1 — 정합성 검증 (5분)

```bash
cd <REPO_ROOT>
git log -8 --oneline                   # 757c9a8 까지 확인
make test-routing 2>&1 | tail -15      # 10/10 통과
go vet ./... ; go build ./...          # green
```

### Step 2 — ADR-019 본문 read (5분, 진입 필수)

```bash
cat docs/superpowers/decisions/2026-05-26-plugin-mcp-exposure-layer.md
```

핵심: §2 Decision (launcher contract + manifest fragment 모두 *locked*).
§5 Verification 의 4 항목 = implementation session 종료 조건.

### Step 3 — implementation 실 작업 (1-3일)

ADR-019 §5 Verification 4 항목 순서대로:

**Step 3.1** — `scripts/test-router-wireup.sh` +1 check
- `plugin/.claude-plugin/plugin.json` 의 `mcpServer` 필드 존재 + launcher path 검증
- launcher 파일 executable bit 검증
- `make test-routing` 10/10 → 11/11

**Step 3.2** — Makefile `build-mcp-bundled` target
```makefile
build-mcp-bundled:
	GOOS=darwin GOARCH=arm64 go build -o plugin/bin/buddy-mcp-darwin-arm64 ./cmd/buddy-mcp
	GOOS=darwin GOARCH=amd64 go build -o plugin/bin/buddy-mcp-darwin-amd64 ./cmd/buddy-mcp
	GOOS=linux  GOARCH=arm64 go build -o plugin/bin/buddy-mcp-linux-arm64  ./cmd/buddy-mcp
	GOOS=linux  GOARCH=amd64 go build -o plugin/bin/buddy-mcp-linux-amd64  ./cmd/buddy-mcp
```
+ `.github/workflows/ci.yml` 또는 `release.yml` 에 cross-compile gate
+ **미결정**: binary 를 git track 할지 / Git LFS / release artifact — implementation session 첫 결정

**Step 3.3** — launcher + manifest 변경
- `plugin/bin/buddy-mcp-launcher.sh` 작성 (ADR-019 §2 Decision 의 *locked* 내용 그대로)
- `chmod +x plugin/bin/buddy-mcp-launcher.sh`
- `plugin/.claude-plugin/plugin.json` 에 `mcpServer` 필드 추가

**Step 3.4** — end-to-end dogfood
- 사용자 머신 (kevin 또는 다른 fresh profile) 에서 `claude plugin install buddy@buddy`
- `/mcp` 명령으로 buddy MCP tool 등록 확인 (advise / analytics / doctor / feature / knowledge / notify / stats / usage)
- `docs/notes/<date>-mcp-exposure-dogfood.md` 작성

**Step 3.5** — charter §2.4 갱신
- `docs/two-tracks-charter.md` §2.4 표의 MCP row 갱신:
  - 직전: `미정 (확인 필요)`
  - 종결: `ADR-019 hybrid distribution (PATH + bundled) — plugin/bin/buddy-mcp-launcher.sh, 4 binary`
- 4 자산 약속 이행률 25% → 50% 명시

**Step 3.6** — plugin minor version bump
- `plugin/.claude-plugin/plugin.json` version `1.0.0` → `1.1.0` (또는 `1.2.0` per ADR-011 milestone cadence)
- 관련 audit doc (`plugin-skills-inventory.md` G4) 의 *version SSoT mismatch* 도 동시 정리

### Step 4 — implementation 종결 + push + release (조건부)

- 별 release tag (ADR-011 milestone-driven) — *별 세션* 권장 (release engineering 분리 원칙)

---

## 6. 후속 작업 우선순위 (Path 1 종결 후)

### 즉시 진입 가능 (별 세션 권장)

| ID | 작업 | 비용 |
|----|------|------|
| **C1-impl** | ADR-019 implementation (Step 3.1-3.6 위) | 1-3일 |
| D3 | v1.0.0 release engineering | 별 세션, 별 트랙 |

### Tier 2 후보 (engineering 깊이 — Path 2)

| ID | 작업 | 비용 |
|----|------|------|
| B1 | P2 — `design-data-model` 본문에 VO/aggregate/entity DDD 차원 | 1-2시간 |
| B2 | P3 — `derive-system-topology` / `map-use-cases-to-infra` 본문에 bounded context 어휘 | 1-2시간 |
| B6 | M1 책임 매트릭스 — 도메인/boundary 4 skill (B1+B2 후) | 1시간 |

### 신규 자산 wave 후속

| ID | 작업 | 비용 |
|----|------|------|
| C2 | hook 정의 — `design-claude-hooks` skill 활용 | 중 (ADR 권장) |
| C3 | agent 정의 — sub-agent 정의 형식 | 상 (ADR 선행) |

### cli buddy 트랙 (Path 5 후보)

| ID | 작업 | 비용 |
|----|------|------|
| D5 | `docs/internal-track-inventory.md` 작성 — cli 트랙 audit | 중-상 (plugin 트랙 대칭) |
| D1/D2 | BA-11 / BA-13 fix | 각 half day |

### 광의 audit 후속 wave

| ID | 작업 | 비용 |
|----|------|------|
| B10 | Cascade "Next" 섹션 표준화 (~90 skill 본문 편집 + write-a-skill 표준 갱신) | 상 (수 일) |
| E2 | ~106 skill 의 PROCEDURE 본문 직접 read + cover 재평가 | 상 |

---

## 7. 본 세션의 결정 트레이스 (재현 필요 시 참고)

세션 시작 시 사용자 발화: *"@docs/ 폴더 하위 문서들을 검토해. 이전에 작업하던 내용들이 있는데, 마무리가 안되어서 이어서 진행하려고해."*

→ 5 docs 검토 → 2 트랙 (T1 skills-consolidation / T2 buddy-core A-cleanup) 식별 → 사용자가 *plugin / internal 두 카테고리 재분류* 요청.

→ 본 세션의 의사결정 사슬:

1. **plugin 트랙 inventory 작성** — 3 신규 doc (inventory / engineering-audit / engineering-flow). 152 skill + 4 자산 격차 + 7 gap 명시.
2. **"engineering 관련 스킬 재정리" 명시** — SKILLS_ANALYSIS A 카테고리 13 영역 좁은 audit (engineering-audit) + 광의 ~106 skill SE-flow 매핑 (engineering-flow) 2 doc 작성.
3. **Path 1 진입 결정** — Tier 1 부터 / H1 부터 / 권장 default 채택.
4. **H1 (audit-ubiquitous-language)** — `write-a-skill` methodology 따라 Step 1-10. 흡수 분류 *adopt-with-edits → inspired-by* downgrade (mattpocock 본문 cross-machine 부재). discipline-enforcing, §6 Quality + Cross-cutting.
5. **G2/G3 catalog 중복 제거** — `monitor-regressions` §6 / `apply-builder-ethos` §1 라인 제거. 의미적 자연 phase 선택.
6. **G1 status 등재** — catalog Cross-cutting Utilities 표 알파벳 위치 (s).
7. **LOW 1/3 + #8 결정 영속화** — engineering-audit 의 *권장 default* 채택 (LOW 1 종결 / LOW 3 cli 이동 / #8 종결).
8. **ADR-019 (C1 진입 준비)** — Option D (hybrid) 사용자 명시 결정 → ADR Accepted. Launcher contract + manifest fragment locked. Implementation 별 세션.
9. **본 세션 종결 결정** — implementation 은 별 세션. handoff notes 작성 후 commit (본 doc).

---

## 8. 본 세션 작업 무결성 보장

| 보장 항목 | 검증 방법 |
|----------|----------|
| `make test-routing` 통과 | 5 commits 의 *모든 cycle* 에서 10/10 직접 실행 확인 |
| PROCEDURE.md count = 152 | `find plugin/skills -name PROCEDURE.md \| wc -l` |
| Command count = 104 | `find plugin/commands -name "*.md" \| wc -l` |
| Catalog ↔ filesystem 정합 | catalog Cross-cutting Utilities 에 status 추가 후 audit 갱신, make test-routing 통과 |
| ADR-019 본문 + README sync | `grep ADR-019 docs/superpowers/decisions/README.md` (행 추가 + closed-by-ADR-019 마킹) |
| Code logic 변경 0 | 본 세션 모든 commit 의 `git diff --stat` 에 `.go` 파일 변경 없음 (test-router-wireup.sh 의 count 코멘트 1 줄 갱신만, logic 무변동) |

---

## 9. 알려진 위험 / 함정

### 9.1 index.lock race

본 세션 동안 3회 발생 (commit 시점). 항상 *commit 직전* `ls .git/index.lock` 으로 stale 확인 → 자동 해제 또는 `sleep 2` 후 retry. handoff §10.2 의 *background daemon* 가 원인 추정 — 단일 머신 daemon 또는 file watcher.

**Mitigation**: 본 세션은 sleep 2 후 retry 로 해소. 다음 세션도 *동일 패턴* 발생 가능 — *3회 retry 후 사용자 보고* 의 CLAUDE.md 임계 유지.

### 9.2 흡수 분류 downgrade 패턴

본 세션의 *read-before-claim* 적용 사례: mattpocock `ubiquitous-language` SKILL.md 가 본 머신 부재 → *0 read* → 흡수 분류 자동 downgrade (`adopt-with-edits` → `inspired-by`). ADR-019 의 *launcher contract* 도 *외부 source 미참조* 로 *reference-only* 의 더 보수적 분류 채택.

**Mitigation**: cross-machine 작업 시 *외부 source 가용성* 먼저 확인 → 부재 시 *흡수 분류 downgrade* + NOTICE 명시. verbatim 0건 유지 *불변*.

### 9.3 plugin.json version vs binary version mismatch

inventory G4 명시: `plugin.json` version `1.0.0` vs `VERSION` 파일 (cli buddy binary) `0.13.0`. ADR-005 (`2026-05-11-plugin-version-reset.md`) 으로 *의도된 분리* 지만 *어느 SSoT 가 무엇 의미인지* 외부 reader 단서 부재.

**Mitigation**: ADR-019 implementation session 의 Step 3.6 (plugin minor version bump) 시 *charter §2.4 또는 README* 에 version SSoT 분리 명시 1줄 추가 권장.

### 9.4 Bundled binary distribution 미결정

ADR-019 §4 Negative: plugin 크기 +~38MB. git-tracked vs Git LFS vs release artifact 중 *implementation session 첫 결정* 필요. 잘못 선택 시 *repo clone time / disk footprint* 영향.

**Mitigation**: implementation session 진입 시 *Git LFS* 검토 우선 (대용량 binary 의 표준). 사용자 명시 confirm 후 결정.

### 9.5 charter §8 후속 작업 미완

charter §8 명시 4 후속 작업 중 일부 *미완*:

1. ~~README.md cli buddy 의 진짜 목적 추가~~ — *상태 미확인*
2. ~~HANDOFF.md §0 트랙 상태 표 단어 통일~~ — *상태 미확인*
3. ~~roadmap.md 재평가~~ — *상태 미확인*
4. ~~archive/tasks.md C-2 제거~~ — *상태 미확인*

**Mitigation**: 다음 cycle 진입 시 *charter §8 4 항목 상태 audit* 권장. 별 doc-sync cycle 후보.

---

## 10. 사용자 글로벌 룰 핵심 (다음 세션이 즉시 알아야 할 것)

`~/.claude/CLAUDE.md` 의 전체 룰 외 본 세션 적용 빈도 높은 5건:

1. **언어**: 사용자가 한국어로 쓰면 한국어로 응답. 영어 technical identifier (commit hash / skill name / path) 영어 유지.
2. **Git commit attribution 금지**: `Co-Authored-By` 또는 "Generated with Claude" 류 절대 금지. commit message 영어 + Conventional Commits.
3. **Push 패턴**: 사용자 명시 승인 후만. 본 세션은 *사용자 명시 진행 의지* (Path 1) 로 push 5회 진행.
4. **Uncommitted 변경 종료 시**: 사용자에게 commit 여부 *먼저* 확인. 자율 commit / 폐기 금지. *현 상태 working tree clean 이면 추가 commit 결정도 사용자 명시*.
5. **3회 연속 동일 오류 시**: 사용자 보고 후 지시 대기. 본 세션은 index.lock 3회 후 sleep mitigation 시도 + 성공 → CLAUDE.md 규칙 *임계 통과* (mitigation 시도 인정).

추가 본 세션 학습:
- **Path 진입 의지 명시 후 default 채택 자연**: 사용자가 "권장 진입 순서로 진행" 발화 시 = 권장 default 채택 confirm. 단 *각 sub-decision* 은 다시 사용자 확인 (예: ADR-019 의 Option D 선택).
- **별 세션 분리 패턴**: handoff §3.1 "release engineering 별 세션" + 본 세션 의 *implementation session 별 분리* — *결정 cycle* 과 *실행 cycle* 의 분리가 안전.

---

## 11. End-of-handoff sanity check

이 doc 만 보고도 다음 세션이 다음 정보 회복 가능:

- [x] 현재 branch + last commit
- [x] 본 세션이 무엇을 했고 어디서 멈췄는지 (Path 1 종결 명시)
- [x] working tree 의 변경 파일 리스트 (0 — clean state)
- [x] 다음에 무엇을 해야 하는지 + 어느 순서로 (ADR-019 implementation §5 4 항목)
- [x] 사용자 글로벌 룰 + 본 세션의 결정 트레이스
- [x] 알려진 위험 (5 개 명시)
- [x] cross-machine 진입 전제 조건 (§0)

본 doc 의 정확도가 stale 해지면 *본 doc 부터 갱신*. 다른 세션이 의지하는 SSoT.

---

## 12. 자주 쓰는 명령 (다음 세션 참고)

```bash
# 빌드 + 테스트
make build                            # bin/buddy (cli) + bin/buddy-mcp
make test-routing                     # 10/10 통과 필수
go vet ./... ; go build ./...

# 본 세션 작업 trail spot check
git log -8 --oneline                  # 본 세션 5 commits 포함
git show 757c9a8 --stat               # ADR-019
git show cd99818 --stat               # H1 audit-ubiquitous-language

# ADR-019 implementation 진입 (별 세션)
cat docs/superpowers/decisions/2026-05-26-plugin-mcp-exposure-layer.md
# §2 Decision (locked launcher + manifest)
# §5 Verification (4 항목)
```

---

**문서 끝**. 다음 세션 시작 시 §0.1 first-5-min sequence → §5 Step 1-3 진입.

질문 / 모호한 부분 발견 시 사용자에게 즉시 확인. 추측 진행 금지 (handoff §10.5 의 3회 연속 오류 임계 적용).
