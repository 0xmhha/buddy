---
name: ship-release
description: This skill should be used when the user wants to "ship the release", "deploy to production", "create a release", "run beta program", "setup canary deploy", "prepare launch", or has passed quality gates and is ready to release. Orchestrates §7 Release & Beta phase.
---

# ship-release — §7 Release & Beta Orchestrator

§7 라이프사이클 단계의 진입점. Quality gate pass → Tagged release + canary/UAT + GA.

**진입 조건**: §6 quality gate 통과 (QA report + security sign-off 완료).
**산출물**: Tagged release, release notes, canary/UAT pass, GA deployment.
**다음 phase**: Production traffic 발생 → `iterate-product` (§8).

---

## §7 내부 구조 (Q7=(b): Beta/UAT 분리)

```
ship-release (§7 phase orchestrator)
├── §7.1 Release Preparation
│   ├── stage 1: setup-quality-gates    (release gate 확인 — 이미 §5에서 설정했으면 skip)
│   ├── stage 2: write-changelog        (version bump + CHANGELOG 업데이트)
│   ├── stage 3: sync-release-docs      (doc drift audit + auto-update)
│   └── stage 4: auto-create-pr         (PR 생성 + 리뷰 요청)
├── §7.2 Beta / UAT (Q7=(b): 별도 sub-phase)
│   ├── stage 5: [run-uat]              🆕 UAT orchestration
│   └── stage 6: [run-beta-program]     🆕 클로즈드 베타 + 피드백 수집
└── §7.3 GA Release
    ├── stage 7: automate-release-tagging  (semver auto-decision + git tag)
    ├── stage 8: guard-destructive-commands  (배포 전 위험 명령 가드)
    └── stage 9: compose-safety-mode       (max safety mode 합성)
```

> 🆕 = 신규 작성 필요. 현재 orchestrator가 직접 수행.

---

## 실행 절차

### §7.1 Release Preparation

**Stage 1: Quality Gate 확인**

`setup-quality-gates` skill을 invoke해 release gate가 설정되어 있는지 확인한다.
§5에서 이미 설정했으면 gate 통과 여부만 확인한다:
- [ ] typecheck 통과
- [ ] lint 통과
- [ ] all tests 통과
- [ ] security scan 통과
- [ ] no secret leak

**Stage 2: Changelog**

`write-changelog` skill을 invoke해:
- semver 결정 (breaking → MAJOR, feat → MINOR, fix → PATCH)
- CHANGELOG 릴리즈 섹션 작성
- user-facing change summary

**Stage 3: Doc Sync**

`sync-release-docs` skill을 invoke해 code change 대비 docs drift를 감지하고 auto-update한다.

**Stage 4: PR 생성**

`auto-create-pr` skill을 invoke해 commit → branch push → PR 생성을 자동화한다.

### §7.2 Beta / UAT

**Stage 5: UAT**

UAT를 수행한다 (`run-uat` skill 미존재 시 orchestrator가 직접 수행):
- UAT 시나리오 목록 (feature spec의 acceptance criteria 기반)
- 각 시나리오 수동 실행 및 결과 기록
- go/no-go 판단

**Stage 6: 베타 프로그램**

클로즈드 베타가 필요하면:
- 베타 사용자 그룹 정의 (early adopter 5-20명)
- 베타 기간 설정 (1-2주 권장)
- 피드백 수집 채널 (형식: Slack DM / Google Form / email)
- go/no-go 기준 (critical bug 0개, user satisfaction ≥ 4/5)

### §7.3 GA Release

**Stage 7: Release Tagging**

`automate-release-tagging` skill을 invoke해:
- merged PR set → semver 결정
- git tag 생성 (v{MAJOR}.{MINOR}.{PATCH})
- release note 발행

**Stage 8-9: 배포 안전망**

`guard-destructive-commands` skill을 invoke해 배포 전 위험 명령을 가드한다.
`compose-safety-mode` skill로 max safety mode를 합성한다:
- `guard-destructive-commands` + `freeze-edit-scope` 동시 활성화

---

## Go-Live Readiness Checklist

- [ ] §6 quality gate 전체 통과
- [ ] CHANGELOG 작성 완료
- [ ] Docs sync 완료
- [ ] PR 승인됨
- [ ] UAT 통과 (critical issue 0)
- [ ] 베타 피드백 수집 완료 (해당하면)
- [ ] Release tag 생성됨
- [ ] 모니터링 alert 설정됨

---

## 다음 phase

- `/buddy:iterate-product` — §8 Operate & Iterate (production traffic 발생 후)

---

## 참조

- Architecture spec: `docs/superpowers/specs/2026-05-04-lifecycle-orchestrator-architecture.md` §4 §7, §§8 Q7=(b)
