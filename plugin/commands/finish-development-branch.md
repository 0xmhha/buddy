---
description: §5 build-feature 완료 후 PR 생성까지의 5-stage chain (pre-flight sync + quality-gate + changelog + docs-sync + PR + mergeable verify). 자동화 git 안전 원칙 (force 금지 / safe 대체 / STOP 우선) 준수. Iron Law mergeable 검증.
argument-hint: "<commit/branch 변경 요약 1줄> [--base=<branch>] [--dry-run] [--skip-sync] [--skip-changelog] [--draft]"
disable-model-invocation: true
---

# /buddy:finish-development-branch

§5 build-feature → PR 생성 sub-orchestrator. 5 stage: (0) pre-flight sync (fetch + safe merge only) → (1) quality-gate → (2) changelog → (3) docs-sync → (4) PR 생성 (force 없는 push) → (5) mergeable=CLEAN 검증 (Iron Law). force-류 명령 절대 사용 안 함 — conflict / reject 시 STOP + 사용자 처리. dry-run 권장.

## 실행 지시

`Skill` 도구로 `router` skill 을 호출하라. 다음 컨텍스트를 전달한다:

- mode: `single`
- target PROCEDURE: `finish-development-branch`
- 사용자 인자:
    ```
    $ARGUMENTS
    ```
