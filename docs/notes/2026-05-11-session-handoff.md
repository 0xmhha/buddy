# Session Handoff — Cycle 1 Close + 1.1.1 Patch (2026-05-11)

> **목적**: 다른 세션에서 *처음 3분 안에* 어디까지 와있는지 + *다음 한 시간 안에* 무엇을 진행할지 파악.
> **선행 핸드오프**: `docs/HANDOFF.md` (장기 트랙 상태). N-1 closure 의 상세는 `docs/superpowers/decisions/2026-05-09-buddy-commands-disable-model-invocation.md` (ADR-001) 가 canonical.
> **본 세션 범위**: dogfood validation Cycle 1 종료 + Cycle 1 fix 반영 + 1.1.1 patch release + Cycle 2 structural verification

---

## §1. 본 세션 한 줄 요약

`/buddy:concretize-idea "<test>"` 실 invocation 으로 시작된 dogfood validation Cycle 1 을 **minimum-viable subset 통과** 로 종료, 7 issue 중 B1-B5 (immediate-fix) 를 반영해 **plugin v1.1.1 patch release** 까지 완료. Cycle 2 는 structural verification (5/5 single-skill 본문 정합 + 4/4 cascade chain) 만 수행하고 live dispatch 는 별 세션으로 deferred.

---

## §2. 본 세션 추가 commits (5개, 모두 origin/main 푸시 완료)

| # | commit | 범위 | files |
|---|--------|------|-------|
| 1 | `24d3b1c` | dogfood cycle-1 result + HANDOFF/tasks sync | `docs/notes/2026-05-10-dogfood-result-cycle-1.md` (신규) + `docs/HANDOFF.md` + `docs/tasks.md` |
| 2 | `d30ccdb` | Cycle 1 fix B1-B5 — concretize-idea bracket / fallback 제거 + validate-idea Q1 phrasing + scenarios §3.3 canned 표 | `plugin/skills/concretize-idea/PROCEDURE.md` + `plugin/skills/validate-idea/PROCEDURE.md` + `docs/notes/2026-05-10-dogfood-validation-scenarios.md` |
| 3 | `f490d21` | plugin patch 1.1.0 → 1.1.1 + CHANGELOG [1.1.1] entry | `plugin/.claude-plugin/plugin.json` + `.claude-plugin/marketplace.json` + `CHANGELOG.md` |
| 4 | `7e1cc0c` | `.claude/settings.local.json` gitignore | `.gitignore` |
| 5 | `60dcef4` | Cycle 2 structural verification addendum (§10.1-§10.3) + tasks A-4.9 ✅ / A-4.10 신규 | `docs/notes/2026-05-10-dogfood-result-cycle-1.md` + `docs/tasks.md` |

`origin/main`: `436c702 → 60dcef4`. push 완료, 작업 디렉터리 clean.

---

## §3. Cycle 1 발견 issue (7개)

전체 목록은 `docs/notes/2026-05-10-dogfood-result-cycle-1.md` §7. 본 세션에서 fix 한 것 (B1-B5) + deferred (B6, B7) 분리:

### Fixed (commit `d30ccdb`)

| ID | sev | 위치 | fix |
|----|-----|------|-----|
| B1 | minor | `dogfood-validation-scenarios.md` §3 | §3.3 canned business scenarios 표 추가 (S1 한국어 SaaS / S2 SMB 회계 / S3 i18n release) |
| B2 | minor | `validate-idea/PROCEDURE.md` L171 | Q1 phrasing "내일 사라지면 진짜로 화날" → "내일 사라지면 진짜로 화내는 사람" |
| B3 | **major** | `concretize-idea/PROCEDURE.md` Stage 4, Stage 6 | bracket notation 제거 + "신규 작성 필요" 안내 삭제 — *cascade flow 가 둘 갈래로 분기 가능* 하던 위험 해소 |
| B4 | minor | 동 PROCEDURE Stage 3-5 본문 | "`analyze-competition-and-substitutes` skill 미존재 시 orchestrator가 수행" fallback prose 삭제 → 직접 skill invoke 명시 |
| B5 | minor | 동 PROCEDURE Stage 6 본문 | "`map-customer-segments` skill 미존재 시" fallback prose 삭제 → 동일 |

### Deferred

| ID | sev | 위치 | 본질 | trigger |
|----|-----|------|------|---------|
| B6 | minor | PROCEDURE 양식 통일 | concretize-idea `## 산출물 형식` vs validate-idea `## Output: 디자인 문서` — 148 skill 양식 inconsistency. self-check 명칭도 일부에만 `§6` | 별 design + skeleton template + lint script |
| B7 | conceptual | router smart-skip | `validate-idea` Q4 smart-skip rule (이전 답 cover 시 skip) 가 대화 메모리 의존. `/buddy:run` 재invocation 시 session state 유실 | cli-buddy spec §6.2 feature-management-mcp 또는 별 design |

---

## §4. 다음 세션 진입점 — 우선순위

### (a) GitHub release tag v1.1.1 (사용자 액션, 1 분)

CHANGELOG `[1.1.1]` entry 가 이미 commit `f490d21` 에 포함됨. 다음 명령으로 release 생성:

```bash
gh release create v1.1.1 \
  --title "v1.1.1 — Cycle 1 dogfood fix" \
  --notes-file <(awk '/^## \[1\.1\.1\]/,/^## \[1\.1\.0\]/' CHANGELOG.md)
```

권장 — patch 라도 fresh install 가시성 + git history reproducibility.

### (b) Cycle 2 live dispatch (별 세션, 30-50K token 예상)

`docs/notes/2026-05-10-dogfood-validation-scenarios.md` §4 의 5 single-skill + §5 cascade 4 stage 실 invocation. canned scenarios §3.3 사용:

| 시나리오 | 입력 |
|---------|------|
| §4.1 | `/buddy:decide-target-market "S1 한국어 SaaS target market 결정"` |
| §4.2 | `/buddy:review-legal-regulatory "S1 한국 cluster 가정"` |
| §4.3 | `/buddy:audit-test-coverage-meaningful "test suite trust score"` |
| §4.4 | `/buddy:optimize-conversion-funnel "AARRR funnel 분기 review"` |
| §4.5 | `/buddy:archive-product "product EOL plan"` |
| §5 cascade | `/buddy:concretize-idea "S2 SMB 회계"` → `decide-target-market` → `review-legal-regulatory` → `design-system` |

기대 결과는 scenarios doc §4.1-§4.5 / §5.1-§5.2 에 정의. *발견 issue 는 cycle-2 result doc* 으로 정리, *fix 는 별 PR* — Cycle 1 패턴 따름.

### (c) cli buddy 본격 spec lock-in (의사결정 필요)

`docs/cli-buddy-spec.md` (371줄) 가 Draft. 본격 구현 (W3-2~6) 진입 전:

1. spec §5 "plugin buddy embedding layer = Claude Code subprocess" 결정 confirm 필요 — 반대 시 ADR 추가
2. spec lock-in 후 ADR-{N} 작성 (`docs/superpowers/decisions/`) + supersede 정책 따라 §6.2 의 *향후 ADR 후보* 표 갱신
3. lock-in 후 `docs/roadmap.md` 의 v0.2/v0.3/v1.0 outline rewrite (W5-1~4) trigger — ADR-002 의 4-pattern matrix 따름

### (d) PROCEDURE 양식 통일 (B6 fix, 별 design)

148 skill 본문 통일 — `## 산출물 형식` / `## Output` / `## 검증 (self-check)` / `## §6` 등 section 명칭 inconsistency. 다음 cycle 의 highest-leverage fix 후보. 단독 design + skeleton template (e.g. `plugin/skills/.template/PROCEDURE.md`) + GitHub Actions lint script.

---

## §5. Trigger 발화 시 작업 (사용자 명시 발화 대기)

| 잠재 작업 | trigger 발화 예시 |
|----------|------------------|
| Korea cluster 3 skill | "target market = Korea", "한국 시장 진출 결정" |
| USA / EU cluster | "target market = USA", "EU GDPR 대응" |
| feature-management-mcp | "cli buddy spec lock-in", "feature 재사용성 측정 필요" |
| analytics-mcp 구현 (spec `docs/superpowers/specs/2026-05-10-analytics-mcp-spec.md` 존재) | "production traffic 회수됨", "AARRR funnel 자동화" |
| `cmd/buddy/main.go` 685줄 분할 (W6-1) | cli buddy 신규 명령 추가 직전 |
| Go CLI v0.2 dogfood feedback 회수 | "며칠 써보니..." (사용자 페이스) |

---

## §6. 알아둬야 할 것 (Gotcha)

1. **두 트랙 동시 존재**: `docs/two-tracks-charter.md` 가 정체성 SSoT. plugin buddy (active) vs cli buddy (부분 구현). 작업 시작 시 어느 트랙인지 명시 확인.
2. **응답 형식**: `docs/response-format-guide.md` 의 paper-style flow (current state → problem → methods → pros/cons → decision + rationale → quantitative verification) 적용. 사용자 CLAUDE.md 의 *Fact-based Answer* 양식도 함께 준수.
3. **commit 규칙**: English Conventional Commits, no Co-Authored-By, project-relative paths only (사용자 memory).
4. **PROCEDURE batch 금지**: validate-idea 등 forcing-question PROCEDURE 는 *한 번에 하나 질문 + 응답 대기* 강제 — anti-sycophancy 의도. mock 응답으로 자동 진행 시 진단 가치 zero.
5. **placeholder path resolution**: router 의 `${CLAUDE_PLUGIN_ROOT}` 경로는 Claude Code runtime 이 substitute. 첫 시도 실패 시 router SKILL.md §"Path resolution" 의 Bash fallback.
6. **disable-model-invocation**: 모든 `plugin/commands/*.md` frontmatter 의 `disable-model-invocation: true` (ADR-001). 신규 command 추가 시 반드시 포함 — 미준수 시 baseline context 에 description 누적 등장.
7. **Skill Completion Cycle 종료 후 stale 위험**: bracket notation 같은 *작성 전 placeholder* 가 작성 후에도 남아있는 경우 — 본 세션 B3 가 예시. PROCEDURE 본문 갱신 누락 의 systematic check 가 B6 fix 의 부분 목표.

---

## §7. 누적 트랙 상태 (2026-05-11 현재)

### plugin buddy (active)

- Total: **148 procedures / 99 commands** (Skill Completion Cycle 100% — v1.1.0), patch v1.1.1 = content fix only
- Charter §3 scope 12 stage 100% cover
- Deferred (trigger 발화 대기): Korea cluster 3 / feature-management-mcp / USA EU cluster
- Spec 확정 + 미구현: analytics-mcp (`docs/superpowers/specs/2026-05-10-analytics-mcp-spec.md`, 349줄)

### cli buddy (partial)

- v0.1.0 = hook reliability monitor (한 sub-feature)
- TUI / agent runtime / plugin buddy 내재화 layer 모두 미구현
- `docs/cli-buddy-spec.md` (371줄, Draft) — lock-in 후 W3-2~6 진입 가능

### ADR (3개 + 향후 5+ 예상)

1. ADR-001 (`disable-model-invocation: true × 57 commands`) — Accepted
2. ADR-002 (roadmap × charter gap, 4-pattern matrix) — Accepted
3. ADR-003 (superpowers external attribution policy) — Accepted

향후 후보 (cli buddy / B6 / B7 등): `docs/superpowers/decisions/README.md` §"향후 ADR 후보" 표.

---

## §8. References

- `docs/HANDOFF.md` — 장기 트랙 상태 SSoT (Last updated: 2026-05-10)
- `docs/two-tracks-charter.md` — 정체성 SSoT
- `docs/response-format-guide.md` — paper-style 응답 형식
- `docs/cli-buddy-spec.md` — cli buddy Draft spec (371줄)
- `docs/superpowers/specs/2026-05-10-analytics-mcp-spec.md` — analytics-mcp Draft spec (349줄)
- `docs/superpowers/decisions/README.md` — ADR Index (3 ADR)
- `docs/notes/2026-05-10-dogfood-validation-scenarios.md` — Cycle 검증 scenario doc
- `docs/notes/2026-05-10-dogfood-result-cycle-1.md` — Cycle 1 결과 (§1-§11)
- `docs/tasks.md` — 모든 ID (A-4.5 ~ A-4.10 본 세션 갱신)
- `CHANGELOG.md` — [1.1.1] entry 본 세션 추가
