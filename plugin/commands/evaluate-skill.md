---
description: |
  Use when reviewing buddy skill quality, after authoring a new skill, or to decide skill improvement priority. Trigger phrases: "스킬 평가", "skill 점검", "evaluate skill", "PROCEDURE 검토", "이 스킬 점수", "skill 품질". Evaluates buddy skill with 26-item paired checklist (F1-F5 frontmatter + CE1-CE5 catalog entry + B1-B12 body + P1-P4 persona — auto-expands skill name to PROCEDURE.md + commands/<name>.md + skill-catalog entry). Outputs weighted score (0-100) per case PC1-PC5, grade (우수/합격/보강필요/재작성), priority-sorted actionable improvements. authoring-guide §1.6 paired evaluation SSoT.
argument-hint: "<skill-name (paired 자동 확장) 또는 명시적 경로 (단일 평가) — 예: 'start' / 'plugin/skills/start/PROCEDURE.md' / '--single plugin/commands/start.md'>"
disable-model-invocation: true
---

# /buddy:evaluate-skill

## 실행 지시

`Skill` 도구로 `router` skill 을 호출하라. 다음 컨텍스트를 전달한다:

- mode: `single`
- target PROCEDURE: `evaluate-skill`
- 사용자 인자:
    ```
    $ARGUMENTS
    ```
