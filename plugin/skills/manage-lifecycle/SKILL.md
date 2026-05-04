---
name: manage-lifecycle
description: This skill should be used when the user wants to "deprecate a feature", "sunset a product", "migrate customers away from X", "archive a feature", "spin off a feature as a product", or is managing the end-of-life of an existing feature or product. Orchestrates §9 Lifecycle Management phase.
---

# manage-lifecycle — §9 Lifecycle Management Orchestrator

§9 라이프사이클 단계의 진입점. Feature/product 노후화 → Deprecation notice + migration playbook + EOL.

**진입 조건**: Feature 또는 product 노후화 감지 (usage decline, maintenance cost > value, strategic pivot).
**산출물**: Deprecation notice, customer migration playbook, EOL documentation.

> **상용 장기운영 전용**: 1년 이상 운영하면 불가피하게 등장. 이 단계를 사전에 설계하지 않으면 고객 이탈과 법적 분쟁이 발생한다.

---

## Stage 흐름

```
manage-lifecycle (§9 phase orchestrator)
├── stage 1: assess-lifecycle-status   (노후화 판단 기준 평가)
├── stage 2: deprecate-feature         (점진적 sunset 계획)
├── stage 3: migrate-customers         (강제 마이그레이션 플레이북)
├── stage 4: archive-product           (EOL 흐름 실행)
└── stage 5: [spin-off-feature]        🆕 별도 제품으로 분리
```

---

## 실행 절차

### Stage 1: Lifecycle Status 평가

노후화 판단 기준:
```yaml
deprecation_triggers:
  usage: "월간 활성 사용자 < 5% of peak"
  maintenance: "버그 수정 비용 > feature value"
  strategic: "제품 방향 변경 또는 피벗"
  technical: "기술 부채로 안전한 운영 불가"
  compliance: "규제 변경으로 법적 운영 불가"
```

`summarize-retro`와 `monitor-regressions`의 데이터를 기반으로 판단한다.

### Stage 2: Feature Deprecation

점진적 sunset 계획을 수립한다:

```
Deprecation timeline (최소 90일):
  Day 0:   내부 결정 확정 + 영향 범위 분석
  Day 1:   고객 공지 (email + in-app banner + docs)
  Day 30:  신규 사용 차단 (readonly mode 또는 signup 불가)
  Day 60:  마이그레이션 지원 window 마감
  Day 90:  EOL (feature 제거 또는 readonly archival)
```

`guard-destructive-commands` skill을 invoke해 feature 제거 전 위험 명령을 가드한다.

### Stage 3: 고객 마이그레이션

강제 마이그레이션이 필요한 경우 플레이북을 작성한다:
- 마이그레이션 대상 고객 수 + 데이터 규모
- 대체 솔루션 (신규 feature / 외부 서비스)
- 자동 마이그레이션 스크립트 (가능한 경우)
- 수동 마이그레이션 가이드
- 지원 채널 (Slack / email / 1:1 call)
- 환불 정책 (subscription 고객 대상)

### Stage 4: EOL 실행

`write-changelog`로 EOL release note를 작성한다.
`sync-release-docs`로 docs에서 deprecated feature 참조를 제거한다.

EOL 체크리스트:
- [ ] 고객 데이터 export 완료 (요청 시)
- [ ] API endpoint 비활성화 (404 또는 410 Gone 반환)
- [ ] 코드베이스에서 feature 제거 (`auto-create-pr` invoke)
- [ ] 모니터링 alert 제거
- [ ] 문서 아카이브

### Stage 5: Spin-off

Feature를 별도 제품으로 분리하는 경우:
- 분리 근거 문서화 (ADR)
- 새 레포지토리 생성
- IP / 라이선스 검토 (`review-license-and-ip-risk` invoke)
- 고객 transition 계획 (기존 고객이 새 제품으로 이동)

---

## 참조

- Architecture spec: `docs/superpowers/specs/2026-05-04-lifecycle-orchestrator-architecture.md` §4 §9
