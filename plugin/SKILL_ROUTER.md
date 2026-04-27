# Buddy Plugin — Skill Router

> **Lazy-load 문서.** 평상시 컨텍스트에 자동 포함되지 않는다.
> [`SKILLS.md`](./SKILLS.md)의 description만 보고 skill 라우팅이 결정되는 경우에는 이 문서를 읽지 않는다 — 토큰을 아낀다.
>
> **이 문서를 읽어야 할 때 (4가지):**
> 1. `SKILLS.md`에서 후보 skill이 2개 이상이고 우선순위가 명확하지 않을 때
> 2. 동일 trigger 경로(command / hook / dispatch)에 여러 skill이 매핑되어 있을 때
> 3. Skill 호출 순서·체이닝이 필요한 워크플로우일 때
> 4. 새 skill을 추가하면서 기존 skill과의 라우팅 충돌을 검토할 때

---

## 1. 라우팅 우선순위 (general)

여러 skill이 후보일 때 적용 순서:

1. **사용자 명시 트리거 (command)** — `/buddy:<name>`은 항상 최우선. dispatch 후보를 무시.
2. **Hook auto-trigger** — Claude Code hook이 fire한 skill은 사용자 의도와 동급.
3. **Description-based dispatch** — 사용자 발화·상황과 description 매칭이 가장 강한 skill.
4. **Tie-breaker** — 2개 이상이 동등하면 §2 도메인 우선순위 표를 따른다.

---

## 2. 도메인 우선순위 표

> Skill 카테고리 간 충돌 시 어떤 카테고리가 우선하는지 정의. 동일 카테고리 내 충돌은 §3에서 케이스별 처리.

| Priority | Category | Rationale |
|----------|----------|-----------|
| _(추가 예정)_ | _(예: process > implementation > review)_ | _(다른 세션 작업 중)_ |

---

## 3. 알려진 라우팅 충돌 / 케이스별 결정

> Skill 카탈로그가 자라면서 발견된 구체 충돌 사례. 각 사례는 *조건 → 선택* 형태.

_(다른 세션이 채우는 중. 새 충돌 발견 시 한 항목씩 추가.)_

예시 형식:
```
### 케이스: <간단 설명>
- 조건: <어떤 상황>
- 후보: <skill A> vs <skill B>
- 선택: <A> — 이유: <한 줄>
```

---

## 4. 새 skill 추가 시 router 검토 체크리스트

새 skill을 `SKILLS.md`에 등재한 직후 다음을 확인 — 위반하면 이 문서에 항목 추가:

- [ ] 동일 command 이름이 이미 등재되어 있지 않다 (`/buddy:<name>` 충돌 X)
- [ ] Description이 다른 skill의 description과 의미상 90% 이상 겹치지 않는다
- [ ] 같은 hook event(PreToolUse 등)에 매핑된 skill이 이미 있을 때, 실행 순서가 명시되었다
- [ ] 사용자 발화 패턴 1~2개로 이 skill이 dispatch 되는지 mental sim 통과

---

## 5. 참조

- Skill 카탈로그 본문 → [`SKILLS.md`](./SKILLS.md)
- Plugin manifest → [`.claude-plugin/plugin.json`](./.claude-plugin/plugin.json)
- Plugin scaffold spec → [`docs/superpowers/specs/2026-04-24-buddy-plugin-architecture-design.md`](../docs/superpowers/specs/2026-04-24-buddy-plugin-architecture-design.md)
