# Plugin Skills — 전이 그래프 (I/O Contract 기반)

> **자동 도출**: 각 스킬의 Output Contract Consumers를 따라가면 생성되는 연결 그래프.
> 총 **149 connections** across 153 skills.
>
> **기준 문서**: [`engineering-phases.md`](../plugin/skills/router/references/engineering-phases.md)

---

## §1. Phase 간 핵심 흐름 (Orchestrator 수준)

```mermaid
graph LR
  subgraph P1["§1 Problem/Opportunity"]
    CI[concretize-idea]
    APC[assess-product-change]
  end

  subgraph P2["§2 Feature Definition"]
    DF[define-features]
  end

  subgraph P3["§3 Technical Design"]
    DS[design-system]
  end

  subgraph P4["§4 Implementation Plan"]
    PB[plan-build]
  end

  subgraph P5["§5 Development"]
    BF[build-feature]
  end

  subgraph P6["§6 Verification"]
    VQ[verify-quality]
  end

  subgraph P7["§7 Release"]
    SR[ship-release]
  end

  subgraph P8["§8 Operations"]
    IP[iterate-product]
    GIT[generate-improvement-tasks]
  end

  CI --> DF
  APC -->|large| DF
  APC -->|medium| DS
  APC -->|small| BF
  DF --> DS
  DS --> PB
  PB --> BF
  BF --> VQ
  VQ --> SR
  SR --> IP
  GIT -->|§2 재진입| DF
```

---

## §2. Phase 2 내부 cascade (Feature Definition)

```mermaid
graph TD
  IA[identify-actors] --> MAUC[map-actor-use-cases]
  MAUC --> MUCSB[map-use-case-to-system-boundary]
  MUCSB --> CFUC[compose-feature-from-use-cases]
  CFUC --> DFS[define-feature-spec]
  CFUC --> SFP[score-feature-priority]
  CFUC --> MFD[map-feature-dependencies]
  DFS --> SFP
  DFS --> DAT[define-acceptance-test-plan]
  EFE[estimate-feature-effort] --> SFP
  EFE --> MFD
  SWF[split-work-into-features] --> DFS
  SWF --> SFP
  QFR[query-feature-registry] --> DFS

  DFS -->|§3 진입| DS[design-system]
  DFS -->|§4 진입| PB[plan-build]
```

---

## §3. Phase 3 내부 cascade (Technical Design)

```mermaid
graph TD
  DFF[decide-form-factor-app-vs-web] --> DTS[define-tech-stack]
  DTS --> DDM[design-data-model]
  DTS --> DAC[design-api-contract]
  DTS --> WADR[write-adr]
  DAC --> DES[design-event-schema]
  DAC --> GFAC[generate-from-api-contract]
  DAC --> TPAUC[test-per-actor-use-case]
  MUCTI[map-use-cases-to-infra] --> DST[derive-system-topology]
  DST --> DDS[design-deploy-strategy]
  DST --> DES
  DST --> DO[design-observability]
  DDS --> SCD[setup-canary-deploy]
  DDS --> SRR[setup-rollback-runbook]
  DO --> SIP[setup-incident-paging]
  CDS[consult-design-system] --> ADS[apply-design-system]
  DIP[design-interaction-pattern] --> ADS
  VBA[verify-best-alternative] --> WADR
  DDM --> GFAC

  subgraph "→ Phase 5"
    BF[build-feature]
  end
  DAM[design-auth-model] --> BF
  DBS[design-billing-system] --> BF
  DCH[design-claude-hooks] --> BF
  DEMB[design-embedding-search] --> BF
  DES --> BF
  DIS[design-i18n-strategy] --> BF
  DMC[design-mcp-server] --> BF
  DSM[design-secret-management] --> BF
  ADS --> BF
```

---

## §4. Phase 4 내부 cascade (Implementation Planning)

```mermaid
graph TD
  DFAT[decompose-feature-to-actor-tracks] --> DTTT[decompose-track-to-tasks]
  DFAT --> DPA[dispatch-parallel-agents]
  DTTT --> MTD[map-task-dependencies]
  MTD --> PPE[plan-parallel-execution]
  MTD --> EBT[estimate-build-timeline]
  PPE --> DPA
  PTT[publish-to-tracker] --> DPA
```

---

## §5. Phase 5-6-7 cascade (Development → Verification → Release)

```mermaid
graph LR
  subgraph P5["§5 Development"]
    BWTDD[build-with-tdd]
    DB[diagnose-bug]
    IFV[iterate-fix-verify]
    RRT[refactor-with-rename-trace]
    GFAC[generate-from-api-contract]
    GTFS[generate-tests-from-spec]
  end

  subgraph P6["§6 Verification"]
    VQ[verify-quality]
    AS[audit-security]
    MCH[measure-code-health]
    AUL[audit-ubiquitous-language]
    CRR[classify-review-risks]
    RE[review-engineering]
  end

  subgraph P7["§7 Release"]
    PLC[prepare-launch-checklist]
    SR[ship-release]
    ACP[auto-create-pr]
    ART[automate-release-tagging]
    SRD[sync-release-docs]
  end

  BWTDD --> VQ
  GFAC --> BWTDD
  GTFS --> BWTDD
  AUL --> RRT
  CRR --> RE
  ACP --> RE
  AS --> PLC
  MCH --> PLC
  ART --> SRD
  PLC --> SR

  %% Backtrack
  VQ -.->|fail| IFV
```

---

## §6. Phase 8 내부 cascade (Operations)

```mermaid
graph TD
  MR[monitor-regressions] --> HI[handle-incident]
  SIP[setup-incident-paging] --> HI
  SRR[setup-rollback-runbook] --> HI
  HI --> CP[conduct-postmortem]
  CP --> GIT[generate-improvement-tasks]

  DAE[design-ab-experiment] --> AAE[analyze-ab-experiment]
  AAE --> GIT
  AUF[analyze-user-funnel] --> GIT
  AFA[analyze-feature-adoption] --> GIT
  AUC[analyze-user-cohort] --> GIT
  AAFR[analyze-actor-failure-rate] --> GIT
  ACA[analyze-cost-anomaly] --> GIT
  ACFC[analyze-customer-feedback-corpus] --> GIT
  OCF[optimize-conversion-funnel] --> DAE
  PGE[plan-growth-experiment] --> DAE
  SR[summarize-retro] --> GIT

  GIT -->|§2 재진입| DF[define-features]
```

---

## §7. 연결 통계

| 항목 | 값 |
|------|---|
| 총 스킬 | 153 (PROCEDURE.md) |
| 총 I/O 연결 | 149 |
| Broken reference | 0 |
| Dead-end (output이 없는 비-leaf 스킬) | 0 |
| Cross-phase 연결 | 43 |
| Phase 내부 연결 | 106 |
