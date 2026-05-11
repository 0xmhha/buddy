# ADR-004 — Plugin track version reset to v0.x.x (was v1.1.1)

> **Status**: Accepted
> **Date**: 2026-05-11
> **Deciders**: buddy maintainer
> **Tags**: versioning, release, semver, plugin-track, charter-alignment
> **Related**:
> - [`docs/two-tracks-charter.md`](../../two-tracks-charter.md) §2 plugin buddy / §3 cli buddy
> - [`docs/cli-buddy-spec.md`](../../cli-buddy-spec.md) — Draft (W3-2~6 미구현)
> - [`docs/superpowers/decisions/2026-05-10-roadmap-charter-gap.md`](./2026-05-10-roadmap-charter-gap.md) (ADR-002)

---

## 1. Context

### 1.1 현재 상태 (2026-05-11)

| 트랙 | 현재 version | release 상태 |
|------|------------|------------|
| **plugin buddy** | v1.1.1 (`plugin/.claude-plugin/plugin.json`, `.claude-plugin/marketplace.json`) | git tag publish **0건** (v1.0.0~v1.1.1 모두 *마켓플레이스 fetch 만 갱신*, GitHub Release 미생성) |
| **cli buddy** | v0.1.0 (`cmd/buddy/main.go`, Makefile, GitHub Release tag `v0.1.0` publish) | GitHub Release v0.1.0 (2026-04-26) |

두 트랙이 **독립 versioning stream** 으로 분기. plugin track 의 major 가 *premature `1`* 로 진입한 상태.

### 1.2 v1.0.0 milestone 정의 — charter 가 lock-in

`docs/two-tracks-charter.md` §4.1 의 의존 도식:

```
cli buddy → plugin buddy 내재화
```

charter §3.6 + cli-buddy-spec.md 상태:

> "cli buddy = 부분 구현 트랙 — v0.1.0 의 hook reliability monitor 가 *cli buddy 의 한 sub-feature*. 진짜 목표 (TUI / agent runtime / plugin buddy 내재화) 의 핵심 미구현"

cli-buddy-spec.md §9 Phase 분할:

| Phase | 비용 |
|-------|------|
| W3-1 spec | Done |
| W3-2 TUI | HIGH |
| W3-3 agent runtime | HIGH |
| W3-4 plugin buddy embedding | MED-HIGH |
| W3-5 v0.1.0 재배치 | MED |
| W3-6 reference agent | HIGH |

= **3~6 month estimate (single-dev cadence)**.

### 1.3 plugin track 의 v1.0.0 정의

charter + cli-buddy-spec 의 의존 chain 으로부터 derive:

> *plugin buddy v1.0.0 = cli buddy 의 plugin buddy 내재화 layer (W3-4) 가 정상 동작 + reference agent (W3-6) 가 production-proven 으로 끝낸 시점*.

이 정의 기준으로 현재 plugin v1.1.1 의 상태:
- ✅ 148 skill catalog (charter §3 scope 12 stage 100% cover)
- ✅ 99 commands (router single-dispatch)
- ❌ cli buddy 내재화 (W3-4) — 미구현
- ❌ reference agent (W3-6) — 미구현
- ❌ production-proven dogfood — Cycle 2 live deferred, real SaaS dogfood 미수행
- ❌ B6 PROCEDURE 양식 통일 — 일부 inconsistency 잔존
- ❌ B7 router smart-skip session state — 미해결

→ **v1.0.0 의 의미적 기준 미충족**. 현재 v1.1.1 은 *premature*.

### 1.4 외부 publish 영향 확인

- git tag: `v0.1.0` (cli buddy 전용 release) 외에는 *v1.0.0~v1.1.1 publish 0건*. `git tag -l` 확인.
- GitHub Releases: v0.1.0 (cli buddy) 외 *plugin track release page 0건*.
- marketplace.json: commit SHA 기반 fetch 이므로 *version downgrade 시 외부 install 사용자 영향 0* (publish 안 됨).

→ **downgrade migration 부담 0**.

---

## 2. Decision

### 2.1 plugin track 의 *major 를 v0 으로 reset*

- 현재 `plugin/.claude-plugin/plugin.json` + `.claude-plugin/marketplace.json` version `1.1.1` → **`0.2.0`** 으로 변경.
- 의미: *Skill Completion Cycle 100% + Cycle 1 dogfood fix* 까지의 자산을 *v0.2.0 baseline* 으로 lock-in.
- v0.1.0 reserved — *cli buddy track 의 binary release tag* 와 동일 token. 충돌 회피 위해 plugin track 의 *첫 publish 는 v0.2.0 부터*.

### 2.2 v1.0.0 milestone 정의 영속화

`plugin track v1.0.0` 진입 조건 (lock-in):

1. cli buddy track 의 W3-2~6 cascade 완료 (TUI / agent runtime / embedding / 재배치 / reference)
2. plugin buddy + cli buddy 통합 dogfood — 외부 SaaS 프로젝트 1건 이상 cycle (idea → ship-release) 완주
3. PROCEDURE 양식 inconsistency (B6) 해소 — 148 skill body 통일 + lint enforcement
4. router smart-skip session state (B7) 또는 supersede ADR 작성

위 4 조건 모두 만족 → plugin track v1.0.0 release. 미만족 시 v0.x.x 로 minor/patch bump.

### 2.3 향후 versioning 컨벤션

| 변경 종류 | bump |
|----------|------|
| skill 추가 / scope 확장 (charter 12 stage 외 영역) | `0.x.y` minor (예: `0.2.0` → `0.3.0`) |
| skill 본문 fix / cross-ref 정리 / content patch | `0.x.y` patch (예: `0.2.0` → `0.2.1`) |
| breaking change (PROCEDURE 양식 일괄 변경 등) | minor bump + ADR + migration notes |
| §2.2 4 조건 모두 만족 | **v1.0.0** |

### 2.4 두 트랙 version namespace 공유 (revised 2026-05-11 post-release)

| 트랙 | tag | 예시 |
|------|----------|------|
| plugin | `vX.Y.Z` | `v0.2.0`, `v0.3.0`, `v1.0.0` |
| cli | `vX.Y.Z` | `v0.1.0` (기존, 2026-04-26 published) |

**Rationale (revised)**: 초안에서는 `plugin-vX.Y.Z` / `cli-vX.Y.Z` prefix 분리를 제안했으나, 사용자 의사결정으로 *prefix 없는 공통 `vX.Y.Z` namespace* 채택. 두 트랙이 *같은 repo + 같은 release page* 를 공유하고, version 자체가 *시간순 cadence* 를 표현 — prefix 가 cognitive overhead.

**Trade-off**: 같은 `vX.Y.Z` namespace 라 release page 가 *artifact 종류 (plugin vs cli) 를 명시* 해야 함. 본 release 의 title pattern: `vX.Y.Z — {plugin buddy | Go CLI} {brief description}`. CHANGELOG entry 는 단일 `[X.Y.Z]` 으로 통일 + 본문 첫 줄에 *어느 artifact 의 release 인지* 명시.

기존 `v0.1.0` (Go CLI binary, 2026-04-26) 와 본 `v0.2.0` (plugin buddy, 2026-05-11) 는 *artifact 다름, version stream 연속 X*. 향후 minor/major bump 도 *artifact 별 독립* — 충돌 시 다음 사용 가능 version 사용.

---

## 3. Alternatives considered (rejected)

### 3.1 v1.1.1 그대로 유지 + 명시적 "pre-1.0 disclaimer"

`plugin.json` version 은 유지하되 README 에 *"v1.x 는 plugin track 만의 versioning, v1.0.0 ≠ project v1.0.0"* 명시.

**Rejected because**:
- semver 의 *major = stability* 신호와 silent conflict — 외부 사용자 / 본인 미래 세션이 v1.x 만 보고 *production-ready* 로 오인 가능
- README disclaimer 는 *해석* 의존, version number 자체가 *machine-readable signal* 인 게 더 정확
- marketplace 의 자동 fetch / dependency tool 이 version 만 읽음 — README 무시

### 3.2 v1.x → v0.11.1 (minor/patch 보존)

major 만 1 → 0, minor/patch 그대로 보존 (v1.1.1 → v0.11.1, v1.1.0 → v0.11.0, …, v1.0.0 → v0.10.0).

**Rejected because**:
- v0.11.1 같은 큰 minor 번호 가 *0.x track 의 미성숙* 의미와 부합 X (보통 0.x 는 *0.1.0 ~ 0.5.0 정도 의 작은 minor* 사용)
- v1.1.1 → v0.11.1 의 *11* 이 semver 적 의미 부재 (1.x.y 의 minor 11 이 아니라 *기존 1.1.x history 의 우회*)
- 사용자 시각에 *어색* + *역사 추적 비용 추가*

### 3.3 v0.1.0 으로 완전 reset

plugin track 도 *처음부터* v0.1.0 시작.

**Rejected because**:
- cli buddy binary 의 `v0.1.0` 과 *git tag token 충돌*. `git tag v0.1.0` 이미 존재 — `plugin-v0.1.0` prefix 로 회피 가능하나 *CHANGELOG entry 의 `[0.1.0]` 도 분리 안 됨*
- *Skill Completion Cycle 의 진척 (v1.0.0 → v1.1.0 의 +44 skill)* 가 *baseline 0* 으로 보여 *작업 양 underweighted*
- v0.2.0 으로 두면 *"v0.1.0 의 baseline (78 skill) → v0.2.0 의 baseline (148 skill)" 의 의미적 minor bump* 정합

### 3.4 v1.x 그대로 + v1.0.0 milestone 정의 조정

charter 의 v1.0.0 milestone 정의를 *현재 자산만으로 v1.0.0* 으로 완화.

**Rejected because**:
- charter 는 *사용자 발화 기반 lock-in 문서* — 임의 정의 완화 X
- cli buddy 의 *진짜 목표 (자동화 agent 관리)* 가 미구현 인데 v1.0.0 칭호 부여는 *사용자의 product 정체성 신호 와 silent conflict*

---

## 4. Consequences

### 4.1 Positive

- **semver major = stability 신호 정합** — v0.x 는 *pre-1.0 maturity* 로 외부 / 미래 self 가 즉시 식별
- **v1.0.0 milestone 의 명확한 entry condition lock-in** — §2.2 의 4 조건 으로 *언제 v1.0.0 진입 가능한지* 영속 record
- **두 트랙 version namespace 분리** (§2.4 prefix) — plugin / cli track 의 *release cadence 독립* 가능
- **외부 영향 0** — v1.x.x 가 publish 안 된 상태라 downgrade migration 부담 0

### 4.2 Negative

- **CHANGELOG history 의 v1.0.0 ~ v1.1.1 entry 가 *superseded* 표기로 잔존** — git history 에 의존한 archaeology 비용. 미래 사용자 가 "왜 v1.x 가 superseded 인가" 추측 시 본 ADR-004 로 도달.
- **마켓플레이스 cache 사용자 (만약 v1.1.1 fetch 한 적 있으면) 가 *downgrade 경험*** — 단 §1.4 verification 으로 *0건* 확인. 잠재적 / 가설적 risk.
- **버전 conversation 시 *"v0.2.0 = post-1.1.1?"* 식 혼란 가능** — 본 ADR 의 §1.3 표 가 reference.

### 4.3 Neutral

- buddy 자산 (skill catalog / command surface / hooks / agents / MCP) *변경 0*. version 만 변경.
- v0.2.0 이후 cycle 진입 시 *bump 정책* 은 §2.3 따름 — 자연스러운 cadence.

---

## 5. Verification

### 5.1 Static (적용 시점)

- `plugin/.claude-plugin/plugin.json` version `"0.2.0"`
- `.claude-plugin/marketplace.json` plugins[0].version `"0.2.0"`
- `CHANGELOG.md` `[0.2.0]` entry 작성 + 이전 `[1.0.0]~[1.1.1]` entry 의 *"superseded"* 표기
- `docs/HANDOFF.md` plugin track 상태 표 version 업데이트
- `docs/tasks.md` 트랙 상태 요약 release 컬럼 업데이트
- `README.md` plugin install 안내 + counts 업데이트
- `docs/superpowers/decisions/README.md` (ADR Index) 본 ADR-004 row 추가

### 5.2 Runtime (publish 후)

- git tag `v0.2.0` push → GitHub Release publish (CHANGELOG `[0.2.0]` entry 본문 사용). 기존 `v0.1.0` (Go CLI binary) 와 *같은 namespace 공유* — release title 로 artifact 종류 구분 (§2.4 revised)
- `claude plugin marketplace add 0xmhha/buddy` → marketplace 재fetch → `0.2.0` enabled

---

## 6. Trigger to revisit

- §2.2 의 4 조건 모두 만족 — 본 ADR 의 v1.0.0 정의 적용, plugin track v0.x.y → v1.0.0 bump
- semver 표준 자체가 변경 (예: SemVer 3.0 release) — 본 ADR 의 컨벤션 재검토
- 두 트랙 통합 packaging (binary 안에 plugin asset 포함) 결정 시 — version namespace (§2.4) 단일화

---

## 7. References

- [`docs/two-tracks-charter.md`](../../two-tracks-charter.md) §2 / §3 / §4 / §6
- [`docs/cli-buddy-spec.md`](../../cli-buddy-spec.md) §9 Phase 분할
- [`docs/superpowers/decisions/2026-05-10-roadmap-charter-gap.md`](./2026-05-10-roadmap-charter-gap.md) (ADR-002) — roadmap × charter gap
- [`CHANGELOG.md`](../../../CHANGELOG.md) `[0.2.0]` entry — 본 ADR 적용 결과
- semver 2.0 — https://semver.org/spec/v2.0.0.html
