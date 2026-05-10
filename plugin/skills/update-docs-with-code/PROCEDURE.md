# Update Docs With Code — README / ADR / changelog 동기화

## 1. 목적

코드 변경 시 *대응 docs* 가 자동으로 따라오게. **README / ADR / CHANGELOG / handoff / skill-catalog** 5 영역의 동기화 절차 + automation hook.

doc drift = *코드와 docs 가 다른 진실* 표현. dogfood 시 *사용자 혼란* + AI agent 의 *잘못된 정보 인용*.

## 2. 사용 시점

- §5 build-feature 의 *각 commit* 후 — local doc 자동 동기화
- release 직전 (ship-release) — CHANGELOG entry 강제
- ADR 결정 시 — ADR 작성 + Index 갱신 강제
- handoff doc (HANDOFF.md) 의 *상태 표* 가 최신 commit 미반영 시
- skill-catalog 의 *count* (skills / commands) 가 실제와 drift 시

## 3. 입력

### 필수
- 변경 대상 (코드 영역) + 변경 종류 (feature / fix / refactor / docs / chore)
- 동기화 대상 docs list (영역 별)
- automation 도구 (있으면 — pre-commit hook / CI)

### 선택
- 이전 doc drift incident log (있으면 — 우선 영역 식별)
- LLM 보조 (AI agent 가 doc draft 작성)

## 4. Stage 흐름

### Stage 1: 변경 종류 → 동기화 대상 매트릭스

| 변경 종류 | README | ADR | CHANGELOG | HANDOFF | skill-catalog |
|--------|-------|----|-----------|---------|----------|
| 신규 feature | △ | (큰 결정 시) | ✓ | ✓ | (skill 추가 시) |
| 신규 skill (plugin buddy) | (count 갱신) | — | ✓ | ✓ | ✓ |
| Bug fix | — | — | ✓ | (incident postmortem 시) | — |
| Refactor (rename / 분리) | — | — | (public API 영향 시) | — | (영향 시) |
| 의사결정 (architecture) | — | ✓ | — | (요약 등재) | — |
| Doc drift fix | — | — | — | (자체) | — |

→ 변경 시점에 *체크리스트* 강제 — 누락 차단.

### Stage 2: README 동기화

| 영역 | 변경 trigger |
|------|---------|
| Tagline / 1줄 정체성 | 핵심 가치 변경 시 |
| Features 표 | 신규 영역 / 트랙 추가 |
| Counts (skills / commands) | plugin/skills/ + plugin/commands/ 변동 시 |
| Installation | 절차 변경 시 |
| Usage | 신규 명령 / 변경 |
| Roadmap | 마일스톤 변경 |
| Acknowledgments | 외부 자산 추가 시 (NOTICE 동시) |

→ count 같은 *기계적 갱신* 은 자동화 후보 (CI script).

### Stage 3: ADR (Architecture Decision Record)

큰 결정 (다년 락인 / 책임 경계 / 트랙 분리) 시:

```
docs/superpowers/decisions/{date}-{topic}.md
```

ADR 양식 (기존 buddy ADR-001 / 002 정합):
- Status / Date / Deciders / Tags / Related
- Context
- Decision
- Consequences (positive / negative)
- Alternatives considered
- Future revisits / triggers

→ ADR Index (`docs/superpowers/decisions/README.md` 또는 자체 list) 갱신.

### Stage 4: CHANGELOG

[Keep a Changelog](https://keepachangelog.com/) 양식:

```markdown
## [Unreleased]

### Added
- ...

### Changed
- ...

### Deprecated
- ...

### Removed
- ...

### Fixed
- ...

### Security
- ...

## [1.0.9] - 2026-05-15

...
```

→ commit 시 *unreleased* 에 1줄 추가. release 시 *unreleased → version* 묶음 이동.

### Stage 5: HANDOFF (세션 인계)

코드 변경 후 HANDOFF.md 의 *변동 영역* 갱신:
- §0 Last updated (날짜 + 핵심 변화 한 줄)
- 트랙 상태 표 (skills / commands count, phase 진척)
- §11 quick-wins (해결된 항목 strikethrough)
- §12 마지막 commit 출력 예시 (if changed)

### Stage 6: Plugin skill-catalog (plugin buddy 한정)

skill 추가 / 변경 시:
- §1~§9 phase 별 stage skill 표에 *cascade 정합 위치* 등재
- frontmatter description 변경 시 catalog 도 동기화
- archive 이동 시 _archive 표시

→ batch 단위 commit 시 *catalog 동시 갱신* 강제.

### Stage 7: Automation hooks

| 도구 | 적용 |
|------|------|
| pre-commit | CHANGELOG entry 강제 (Conventional Commits) |
| CI | README count 자동 산출 + drift 검증 |
| commitlint | commit message 양식 강제 |
| markdownlint | docs 양식 일관 |
| LLM 보조 | docs 변경 draft 자동 생성 |

→ automation 이 *모든 검증* 대체 X. *체크리스트* 가 backbone.

## 5. 산출물 형식

```markdown
## Doc Sync Log — {commit / batch}

### 변경 종류
- {feature / fix / refactor / chore / decision}

### 동기화 대상 (체크)
- [ ] README
- [ ] ADR
- [ ] CHANGELOG
- [ ] HANDOFF
- [ ] skill-catalog

### Drift 발견
- ...

### Automation 결과
- pre-commit / CI 통과 여부
```

## 6. 검증

- [ ] 변경 종류 → 동기화 대상 매트릭스 적용?
- [ ] README count 와 실제 plugin/skills/ + plugin/commands/ 일치?
- [ ] ADR 작성 (큰 결정 시)?
- [ ] CHANGELOG entry (release 영향 변경 시)?
- [ ] HANDOFF.md *Last updated* 갱신?
- [ ] skill-catalog 와 commands/*.md 1:1 매칭?
- [ ] Automation hook 활성?

## 7. 다음 phase

- `ship-release` 의 release prep 직전 — CHANGELOG 정정 + version bump
- `monitor-regressions` 의 doc drift incident 추적
- `summarize-retro` 의 분기 doc audit

## 8. 참조

- Keep a Changelog (keepachangelog.com)
- Conventional Commits (conventionalcommits.org)
- buddy 자체 ADR-001 / ADR-002 — ADR 양식 reference
- buddy `docs/two-tracks-charter.md` — *charter 변경 시 README + HANDOFF + roadmap 동시 갱신* 사례
