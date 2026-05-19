# Architecture Decision Records — Index

> **목적**: buddy 의 lock-in 결정의 단일 색인. 새 ADR 작성 시 본 문서 갱신 강제.
>
> ADR 양식 reference: [Michael Nygard's ADR template](https://adr.github.io/madr/) — 7 표준 섹션 (Status / Date / Deciders / Tags / Related / Context / Decision / Consequences / Alternatives / Verification / Trigger to revisit / References).

---

## ADR 목록

| ID | Date | Title | Status | Tags |
|----|------|-------|--------|------|
| [ADR-001](./2026-05-09-buddy-commands-disable-model-invocation.md) | 2026-05-09 | Disable model invocation for all `plugin/commands/*.md` | Accepted | plugin-architecture, context-cost, routing |
| [ADR-002](./2026-05-10-roadmap-charter-gap.md) | 2026-05-10 | roadmap.md v0.2/v0.3/v1.0 outline × two-tracks-charter cli buddy gap | Accepted | roadmap, charter, cli-buddy, naming, scope-alignment |
| [ADR-003](./2026-05-10-superpowers-attribution.md) | 2026-05-10 | `superpowers` external project attribution policy | Accepted | attribution, naming-collision, charter, skill-catalog |
| [ADR-004](./2026-05-11-plugin-version-reset.md) | 2026-05-11 | Plugin track version reset to v0.x.x (was v1.1.1) | Accepted | versioning, release, semver, plugin-track, charter-alignment |
| [ADR-005](./2026-05-11-cli-buddy-spec-lock-in.md) | 2026-05-11 | cli buddy spec scope lock-in (Draft → Accepted) | Accepted | cli-buddy, agent-runtime, embedding, charter-alignment, spec-lock-in |
| [ADR-006](./2026-05-19-procedure-form-allowlist-policy.md) | 2026-05-19 | B6 PROCEDURE form: bulk-allowlist intentional deviations + --strict for new skills | Accepted | b6-lint, procedure-form, allowlist, ci-gate, plugin-v1.0-entry |
| [ADR-007](./2026-05-19-router-no-cross-invocation-state.md) | 2026-05-19 | Router does not maintain cross-invocation conversational state (supersedes B7) | Accepted | router, session-state, b7-supersede, plugin-v1.0-entry, scope-discipline |

---

## ADR 작성 정책

### 작성 trigger

다음 케이스 중 하나라도 해당 시 ADR 작성:

| trigger | 예시 |
|---------|------|
| **다년 락인 결정** | 언어 / 프레임워크 / DB 선택 (e.g. Go pivot — v0.1 spec §3 에 documented, ADR 미작성 — historical) |
| **책임 경계 변경** | plugin buddy / cli buddy 트랙 분리 ([`charter`](../../two-tracks-charter.md)), N-1 closure ([ADR-001](./2026-05-09-buddy-commands-disable-model-invocation.md)) |
| **외부 자산 차용** | superpowers attribution ([ADR-003](./2026-05-10-superpowers-attribution.md)) |
| **silent conflict 해소** | roadmap × charter gap ([ADR-002](./2026-05-10-roadmap-charter-gap.md)) |
| **명명 / naming convention 결정** | `{verb}-{region}-{topic}` (charter §2.4.2 region-cross-cutting) |
| **policy 변경** | 컴플라이언스 / privacy / IP 정책 |

### ADR 양식 (표준 7 섹션)

```markdown
# ADR-{N} — {Title}

> **Status**: Proposed / Accepted / Deprecated / Superseded by ADR-{M}
> **Date**: YYYY-MM-DD
> **Deciders**: {who}
> **Tags**: {comma-separated}
> **Related**: {다른 ADR / 문서 link}

## 1. Context
(왜 이 결정이 필요했는지 — 배경 + 발견 + silent conflict 가능성)

## 2. Decision
(무엇을 결정했는지 — 명확한 statement)

## 3. Alternatives considered (rejected)
(고려된 대안 + rejection rationale)

## 4. Consequences
### Positive
### Negative
### Neutral

## 5. Verification
(결정 적용을 어떻게 측정 / 검증)

## 6. Trigger to revisit
(어떤 조건에서 본 ADR 을 재평가할 것인가)

## 7. References
(연관 docs / 외부 source)
```

### 작성 후 강제

새 ADR 작성 시:

1. **본 README.md 의 ADR 목록 표 갱신** (행 추가)
2. **CHANGELOG `[Unreleased]` 또는 다음 release entry 의 Decisions 섹션 추가**
3. **연관 docs 의 reference 업데이트** (HANDOFF.md / charter / spec 등)
4. **commit message** 에 `docs(decisions): add ADR-{N} {title}` 양식

---

## ADR 변경 / Supersede 정책

### Status 전환

```
Proposed → Accepted → (Deprecated 또는 Superseded by ADR-M)
```

- **Proposed**: draft, 사용자 confirm 대기
- **Accepted**: lock-in, 영속 reference
- **Deprecated**: 결정 자체 폐기 (대체 ADR 없음)
- **Superseded by ADR-{M}**: 신규 ADR 이 같은 영역 결정 변경

### Supersede 규칙

기존 ADR 을 supersede 시:

1. 신규 ADR 의 Status = `Accepted` (Proposed 단계 거침)
2. 기존 ADR 의 Status = `Superseded by ADR-{M}` 갱신
3. 본 README 의 표에서 *기존 ADR 의 Status* 업데이트
4. 결정 변경 사유는 신규 ADR 의 §1 Context 에 명시

### Trigger to revisit 발화

각 ADR 의 §6 *Trigger to revisit* 명시. trigger 발화 시:

1. 별도 PR 또는 issue 로 *재평가* 진행
2. 재평가 결과:
   - 변경 없음 → 기존 ADR Status 유지 + revisit 기록 (ADR commit history 만)
   - 변경 결정 → 신규 ADR + 기존 supersede

---

## 향후 ADR 후보

charter §6.2 의 cli buddy 진화 시 예상 ADR 후보:

| 예상 ADR 후보 | 트리거 |
|------------|--------|
| ADR-{N} TUI vs web (Q-1) 최종 결정 | cli buddy W3-2 진입 후 dogfood feedback |
| ADR-{N} agent runtime model 상세 (Q-2 ~ Q-3 / Q-5) | cli buddy W3-3 진입 시 |
| ADR-{N} task vs agent naming (P3 from ADR-002) | roadmap rewrite 시 |
| ADR-{N} AGENTS.md naming collision (P3 from ADR-002) | v1.0 plugin model 결정 시 |
| ~~ADR-{N} PROCEDURE skeleton enforcement (B6 follow-up)~~ | **Closed by ADR-006 (2026-05-19)** — bulk allowlist + `--strict` CI gate |
| ~~ADR-{N} router smart-skip session state (B7)~~ | **Closed by ADR-007 (2026-05-19)** — supersede, not router responsibility |
| ADR-{N} skill MCP exposure | cli buddy 의 §4.1 option (b)/(d) 재검토 시 |
| ADR-{N} cli buddy agent sandbox security (Q-6) | W3-3 agent runtime 의 *임의 명령 실행* 위험 surface 시 |

---

## References

- [`docs/two-tracks-charter.md`](../../two-tracks-charter.md) — plugin buddy / cli buddy 정체성
- [`docs/HANDOFF.md`](../../HANDOFF.md) — 세션 인계 + decision history
- [`CHANGELOG.md`](../../../CHANGELOG.md) — release 별 Decisions 섹션
