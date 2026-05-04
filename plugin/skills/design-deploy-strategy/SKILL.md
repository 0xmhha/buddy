---
name: design-deploy-strategy
description: "production 배포 전략 설계 — canary / blue-green / rolling / recreate 선택, env 분리 (dev/staging/prod), secret 관리, IaC (Terraform/Pulumi/k8s), rollback playbook, post-deploy smoke test 자동화. 트리거: '배포 전략 설계' / 'canary 배포' / 'rollback playbook' / 'env 관리' / 'IaC 작성' / 'k8s manifest' / 'production 배포'. 입력: 서비스 type (web/api/worker/static), 트래픽 패턴, downtime 허용, infra provider, SLA. 출력: deploy strategy + env matrix + IaC spec + rollback playbook + smoke test. 흐름: automate-release-tagging → design-deploy-strategy → monitor-regressions."
type: skill
---

# Design Deploy Strategy — 배포 5 차원 설계

## 1. 목적

production 배포를 **5 차원**으로 설계한다:
1. **rollout strategy**: canary / blue-green / rolling / recreate
2. **env 분리**: dev / staging / prod (또는 더 많은 stage)
3. **infrastructure**: IaC (Terraform/Pulumi/k8s) + provider 선택
4. **rollback playbook**: 실패 감지 → 자동/수동 rollback 경로
5. **post-deploy smoke test**: 정상 동작 자동 검증

핵심 가치: **deploy를 black-box → reproducible artifact**, 실패 시 5분 내 rollback.

## 2. 사용 시점 (When to invoke)

- 첫 production 배포 전 strategy 결정
- 신규 service 추가 (deploy 흐름 통합)
- downtime 사고 후 strategy 재검토
- 트래픽 증가로 canary % 조정 필요
- multi-region 확장 (region별 rollout)
- migration deployment (DB schema + code)
- compliance 요구로 staged rollout 필수

## 3. 입력 (Inputs)

### 필수
- service type (web / API / worker / static / mobile / native)
- 트래픽 패턴 (peak QPS / 평균 / 시즌별)
- downtime 허용 (zero-downtime / 5분 / 1시간)
- infra provider (AWS / GCP / Azure / Vercel / Cloudflare / 자체)
- SLA (uptime 99.9% / 99.95% / 99.99%)

### 선택
- compliance 요구 (HIPAA / SOC2 / GDPR)
- multi-region 필요
- 기존 IaC (있으면 통합)
- secret 관리 도구 (Vault / SSM / Secret Manager)

### 입력 부족 시 forcing question
- "downtime 허용 얼마야? zero-downtime 요구면 blue-green / rolling 강제."
- "트래픽 peak QPS 얼마야? canary 5%가 절대값으로 의미 있어?"
- "DB schema migration 있어? code deploy와 함께? 아니면 별도?"
- "현재 secret 어디 보관해? .env? Vault? hardcoded면 stop."

## 4. 핵심 원칙 (Principles)

1. **Zero-downtime이 default** — 사용자 끊김은 신뢰 손상. recreate는 last resort.
2. **Canary는 5% → 50% → 100%** — gradual traffic shift. 각 단계 health check 통과 후 진행.
3. **Rollback은 5분 안에** — 자동화 강제. 수동 rollback은 사고 확대.
4. **Env 격리 강제** — dev / staging / prod 분리. shared DB / secret 절대 금지.
5. **Secret은 절대 IaC에 hardcode** — 모든 secret은 vault. IaC는 reference만.
6. **IaC 변경은 PR + review** — `terraform apply` 직접 금지. plan → review → apply.
7. **Smoke test는 즉시 실패 감지** — health endpoint + critical user flow + DB query 1개.
8. **Migration은 backward-compatible** — code N + DB N+1 동시 동작 가능해야 (zero-downtime migration).

## 5. 단계 (Phases)

### Phase 1. Service Profile
- service type 분류
- 트래픽 패턴 (RPS, latency, 시즌)
- 의존성 그래프 (downstream service)
- statefulness (stateless / stateful / quasi-stateful)

### Phase 2. Rollout Strategy 선택

| Strategy | 적합한 경우 | Trade-off |
|---|---|---|
| **Canary** | 큰 사용자 base, gradual risk | 복잡 setup, % traffic shift 도구 필요 |
| **Blue-Green** | full env duplicate 가능 | 비용 2x, 동시 활성 부담 |
| **Rolling** | k8s native, gradual replacement | 이전/신규 코드 동시 실행 (compat 필수) |
| **Recreate** | downtime OK, 단순 | downtime 발생, last resort |
| **Feature Flag** | code 배포 ≠ 활성화 분리 | flag 관리 복잡도 |

대부분의 SaaS는 **Canary + Feature Flag** 조합 권장.

### Phase 3. Env Matrix
- **dev**: 개발자 local + 공유 dev 환경
- **staging**: prod 미러 (DB / secret 별도)
- **prod**: 실 사용자
- **추가**: load-test / canary / preview / DR

각 env에:
- compute: container / VM / serverless
- DB: separate instance / schema
- secret: separate path
- domain: dev.example.com / staging.example.com / example.com
- TLS: wildcard or per-env

### Phase 4. IaC 설계

provider별:
- **Terraform**: multi-cloud, mature, state backend (S3 + DynamoDB)
- **Pulumi**: 일반 언어 (TS/Python/Go), state backend
- **k8s manifest + Helm/Kustomize**: k8s native
- **Cloudflare Workers**: wrangler.toml
- **Vercel**: vercel.json + GitHub integration

원칙:
- module별 분리 (network / compute / db / dns / auth)
- variable로 env 분기
- state는 remote backend (lock 강제)
- plan → review → apply (CI 게이트)

### Phase 5. Secret 관리
- **Vault** (HashiCorp) — multi-cloud, 복잡
- **AWS SSM Parameter Store / Secrets Manager** — AWS native
- **GCP Secret Manager** — GCP native
- **k8s Secret + sealed-secrets** — k8s native
- **doppler / 1Password Secrets Automation** — managed

원칙:
- IaC에 secret hardcode 금지
- runtime fetch (env var injection)
- rotation 정책 (90일 / 365일)
- audit log (누가 언제 access)
- least privilege (per-service role)

### Phase 6. Rollback Playbook
- **자동 rollback trigger**:
  - error rate > 5%
  - latency p99 > 2x baseline
  - health check fail 3회
  - canary metric degradation
- **rollback 명령** (5분 안에 완료):
  - canary: revert traffic to 0%
  - blue-green: switch back to blue
  - rolling: rollout undo (k8s `kubectl rollout undo`)
  - recreate: previous artifact redeploy

- **DB migration rollback**:
  - 가능하면 backward-compatible (no rollback 필요)
  - 불가하면 별도 down migration (테스트 필수)

### Phase 7. Smoke Test 자동화
배포 후 즉시 (1-3분):
- health endpoint (HTTP 200)
- critical user flow (login + 1 main action)
- DB query 1개 (read + write)
- external dependency ping

실패 시 자동 rollback trigger.

### Phase 8. Monitoring Wire-up
- log aggregation (Datadog / Sentry / ELK)
- metric (Prometheus / CloudWatch / Datadog)
- alert (PagerDuty / Opsgenie / Slack)
- distributed tracing (OpenTelemetry / Datadog APM / Jaeger)

`monitor-regressions` 스킬 페어로 회귀 감지.

## 6. 출력 템플릿 (Output Format)

```yaml
service_profile:
  name: api-server
  type: api
  language: typescript
  runtime: node-20
  peak_qps: 500
  p99_latency_ms: 200
  statefulness: stateless

rollout_strategy:
  primary: canary
  canary_progression: [5, 25, 50, 100]
  duration_per_step: "10 minutes"
  health_check_required: yes
  feature_flag_paired: yes

env_matrix:
  - name: dev
    compute: docker-compose (local)
    db: postgres (local)
    secret: .env.dev (gitignored)
    domain: dev.example.local
  - name: staging
    compute: ECS Fargate
    db: RDS Postgres (small)
    secret: AWS SSM /staging/*
    domain: staging.example.com
    autoscale_min: 1
    autoscale_max: 3
  - name: prod
    compute: ECS Fargate
    db: RDS Postgres (multi-AZ)
    secret: AWS Secrets Manager /prod/*
    domain: example.com
    autoscale_min: 3
    autoscale_max: 50
  - name: canary
    compute: ECS Fargate (separate target group)
    db: shared with prod (read-only feature flag)
    secret: AWS SSM /prod/*
    traffic_share: "5% via ALB weighted target group"

iac:
  tool: terraform
  state_backend: s3 + dynamodb
  modules:
    - network/vpc
    - network/alb
    - compute/ecs-cluster
    - compute/ecs-service
    - data/rds
    - dns/route53
    - secrets/ssm
  ci_gate:
    plan_on_pr: yes
    apply_on_merge: yes
    manual_approval_for_prod: yes

secret_management:
  tool: aws-secrets-manager + ssm
  rotation_policy: "90 days for credentials, 365 days for API keys"
  audit_log: cloudtrail
  least_privilege: per-service-iam-role

rollback_playbook:
  triggers:
    - error_rate_pct: 5
      window_min: 5
    - p99_latency_ms_factor: 2.0
      window_min: 5
    - health_check_failures: 3
      window_min: 2
  rollback_commands:
    canary: "alb-update-target-group --canary-weight 0"
    rolling: "kubectl rollout undo deployment/api-server"
  rollback_target_min: 5
  db_migration:
    backward_compatible: yes
    down_migration_tested: n/a

smoke_test:
  endpoint: https://example.com/health
  expected_status: 200
  critical_flows:
    - name: login
      command: "curl -X POST .../auth/login ..."
    - name: db_read
      command: "curl .../api/me"
  timeout_sec: 30
  failure_action: auto_rollback

monitoring:
  log: datadog
  metric: datadog
  alert:
    channel: pagerduty
    severity: critical
    on:
      - error_rate
      - latency_p99
      - apdex_score
  tracing: datadog_apm
  custom_metrics:
    - api.request.count
    - api.request.duration_ms
    - business.signup.count
```

## 7. 자매 스킬 (Sibling Skills)

- 앞 단계: `automate-release-tagging` (tag 발행 후) — `Skill` tool로 invoke
- 페어: `audit-security` (deploy 전 보안 audit)
- 페어: `setup-quality-gates` (CI 게이트가 deploy 전제조건)
- 다음 단계: `monitor-regressions` (배포 후 즉시 모니터링)
- 후속: `audit-live-devex` (배포 환경 라이브 검증)
- 운영: `triage-work-items` (운영 이슈 유입)

## 8. Anti-patterns

1. **Recreate to prod** — downtime 발생. zero-downtime 전략 default.
2. **Canary 5% → 100% 단번** — gradual progression 의미 약. 5/25/50/100 필수.
3. **Rollback 수동** — 5분 안에 못 함. 자동 trigger.
4. **Env 공유 DB / secret** — staging fail이 prod 영향. 격리 강제.
5. **Secret hardcode** — git 노출 위험. vault + runtime fetch.
6. **IaC apply 직접** — drift 위험. plan → review → apply.
7. **Smoke test 없이 deploy** — 실패 감지 늦음. health + critical flow 강제.
8. **Migration code와 동시 deploy non-backward-compatible** — rollback 불가. backward-compatible 강제.
