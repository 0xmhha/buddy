# Buddy Plugin — Skill Catalog

> Plugin이 제공하는 skill 전체 카탈로그. 각 skill의 이름·트리거·용도·1줄 description.
> Claude는 이 문서만 보고도 대부분의 skill 라우팅 결정을 내릴 수 있어야 한다.
> 라우팅 결정이 모호할 때만 [`SKILL_ROUTER.md`](./SKILL_ROUTER.md)를 참조한다 (lazy-load — 토큰 절약).

---

## 1. 트리거 메커니즘

Buddy plugin의 skill은 세 경로로 활성화된다.

| 경로 | 형식 | 사용 시점 |
|------|------|----------|
| **사용자 명시 트리거 (command)** | `/buddy:<command-name> [args]` | 사용자가 의도적으로 호출 |
| **Plugin auto-trigger (hook)** | `~/.claude/settings.json`의 `hooks` 항목 | 특정 이벤트(PreToolUse 등)에 자동 |
| **Skill description-based dispatch** | `plugin/skills/<name>/SKILL.md` frontmatter `description` 매칭 | Claude가 상황 판단으로 자율 호출 |

세 경로 모두 동일한 skill 본문(`plugin/skills/<name>/SKILL.md`)을 실행한다.

---

## 2. Skill 목록

> 각 항목은 다른 세션에서 점증 추가 중. 본 카탈로그는 **항상 참조 가능**해야 하므로 한 항목당 한 줄을 넘지 않게 유지한다.
>
> 한 줄을 넘는 라우팅 정보(상황별 우선순위, 충돌 해소, 의사결정 트리)는 `SKILL_ROUTER.md`로 분리한다.

| Skill name | Path | Trigger 경로 | When to use (1줄) |
|------------|------|------------|------------------|
| _(추가 예정)_ | `plugin/skills/<name>/SKILL.md` | command / hook / dispatch | _(다른 세션 작업 중)_ |

---

## 3. 추가 / 수정 규칙

새 skill을 카탈로그에 등재할 때:

1. `plugin/skills/<name>/SKILL.md` 생성 (frontmatter `name`, `description` 필수).
2. 위 §2 표에 한 줄 추가 — `name` / `path` / `trigger` / `when to use(1줄)`.
3. **라우팅이 다른 skill과 겹치거나 우선순위가 필요한 경우에만** `SKILL_ROUTER.md`에 항목 추가.
4. command 트리거를 추가하면 `plugin/commands/buddy/<name>.md` 도 함께 등재.

---

## 4. 참조

- 라우팅 결정이 모호하거나 skill 간 충돌이 있을 때 → [`SKILL_ROUTER.md`](./SKILL_ROUTER.md)
- Plugin scaffold 전체 구조 → [`docs/superpowers/specs/2026-04-24-buddy-plugin-architecture-design.md`](../docs/superpowers/specs/2026-04-24-buddy-plugin-architecture-design.md)
- Plugin manifest → [`.claude-plugin/plugin.json`](./.claude-plugin/plugin.json)
