# Actor-Track → Subagent 매핑

§4 plan-build 산출물의 `system_boundary`를 기반으로 §5 build-feature에서
어떤 subagent_type을 dispatch할지 결정하는 기준표.

---

## 매핑 표

| system_boundary | 권장 subagent_type | 보조 | 선택 기준 |
|----------------|-------------------|------|----------|
| `frontend-spa` | `senior-frontend` | `feature-dev:code-architect` | React/Vue/Angular SPA |
| `frontend-mobile` | `senior-frontend` | `feature-dev:code-architect` | React Native / Flutter |
| `backend-api-service` | `senior-backend` | `feature-dev:code-architect` | REST/GraphQL API 서버 |
| `backend-auth-service` | `senior-backend` + `senior-secops` | — | 인증/인가 코드는 보안 리뷰 병행 |
| `backend-data-service` | `senior-backend` | `senior-data-engineer` | 데이터 집계/변환 서비스 |
| `external-saas` | `senior-devops` | `senior-backend` | 외부 API/SDK 통합 |
| `data-pipeline` | `senior-data-engineer` | `senior-backend` | ETL, 스트리밍 파이프라인 |
| `infra-cloud` | `senior-devops` | — | Terraform, Kubernetes, CDN |
| `infra-db` | `senior-backend` | `senior-data-engineer` | 마이그레이션, 인덱싱 |
| `infra-messaging` | `senior-devops` | `senior-backend` | Kafka, SQS, Pub/Sub |
| (기본) | `feature-dev:code-architect` | — | 위 목록에 없는 경우 |

---

## 결정 트리

```
system_boundary가 "frontend-*"?
  ├── YES → senior-frontend
  └── NO  → "backend-auth-*"?
              ├── YES → senior-backend (commit policy 강화: audit-security 병행)
              └── NO  → "backend-*"?
                          ├── YES → senior-backend
                          └── NO  → "external-saas" 또는 "infra-*"?
                                      ├── YES → senior-devops
                                      └── NO  → feature-dev:code-architect
```

---

## Agent prompt 필수 포함 항목

각 actor-track agent에게 전달하는 prompt에 반드시 포함:

```markdown
## 작업 컨텍스트
- Feature: {feature_id}
- Actor: {actor_id} (system_boundary: {system_boundary})
- Worktree: {worktree_path}
- Branch: actor/{actor_id}

## 구현 Tasks
{tasks[] 목록 — id, description, dependencies}

## Contract 의존성
- 기다려야 하는 artifact: {contracts_needed[]}
- 생성해야 하는 artifact: {contracts_produced[]}

## Commit Policy (필수 준수)
- 4-gate 통과 후만 commit: unit test + lint + typecheck + build ALL pass
- Conventional Commits: feat({scope}): {subject}
- WIP commit 금지
- `--no-verify` 금지

## 완료 기준
- [ ] 모든 task 구현 완료
- [ ] 위 4-gate 전부 통과
- [ ] contract 산출물 생성 완료
- [ ] actor/{actor_id} 브랜치에 commit
```

---

## 보안 강화 트랙 (backend-auth-*)

인증/인가 코드를 구현하는 track은 완료 후 `audit-security` skill을 추가 invoke:

```
build-feature Stage 5 (bug 대응) 이후:
→ 인증 관련 actor track이면: audit-security skill 호출
→ OWASP Top 10, secrets, JWT 취약점 검사
→ 이슈 발견 시 iterate-fix-verify loop
```
