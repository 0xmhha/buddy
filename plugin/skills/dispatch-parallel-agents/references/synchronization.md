# Actor-Track 동기화 전략

actor-track 간 contract 의존성을 관리하는 방법.

---

## 의존성 유형

### Type A: Contract File 의존
한 actor가 생성한 파일(API 명세, 타입 정의 등)을 다른 actor가 소비.

```
backend → docs/api-contract.yaml → frontend
backend → src/types/user.ts      → frontend (공유 타입)
```

**처리**: backend track 완료 → artifact 확인 → frontend track 시작.

### Type B: Interface 합의 의존
두 actor가 동시에 진행하지만 경계 인터페이스를 먼저 합의해야 함.

```
frontend + backend 모두 시작 전 →
  POST /auth/signup 요청/응답 스키마 합의 (design-api-contract)
```

**처리**: §3 design-system에서 API contract 먼저 확정 → 두 actor 동시 시작.

### Type C: 공유 인프라 의존
DB 스키마, queue 토픽 등 인프라가 먼저 준비되어야 함.

```
infra track → DB 스키마 마이그레이션 → backend track 시작
```

**처리**: infra track 먼저 완료 → backend track 시작.

---

## 동기화 체크리스트 (SYNC 선언)

`docs/actor-track-plan.yaml`의 sync_points 섹션:

```yaml
sync_points:
  - id: SYNC-1
    type: contract-file
    produced_by: backend
    artifact: docs/api-contract.yaml
    consumed_by: [frontend]
    check: "test -f docs/api-contract.yaml && jq '.paths' docs/api-contract.yaml > /dev/null"
    blocked_tasks: [FE-1]  # frontend FE-1은 이 sync 완료 전 시작 금지

  - id: SYNC-2
    type: shared-infra
    produced_by: infra
    artifact: "DB migration 완료"
    consumed_by: [backend, integration]
    check: "go run ./cmd/migrate status | grep 'up to date'"
    blocked_tasks: [BE-1, INT-1]
```

---

## 병렬 실행 가능 그룹 결정

의존성 없는 track은 동시에 시작 가능:

```
signup-email-password 예시:

Phase 1 (동시 시작 가능):
  ├── infra track    — DB 스키마 마이그레이션
  └── (backend/frontend는 SYNC-2 대기)

Phase 2 (infra 완료 후):
  ├── backend track  — POST /auth/signup 구현 (SYNC-2 해제)
  └── integration track — SendGrid 연동 (SYNC-2 해제)

Phase 3 (SYNC-1 완료 후):
  └── frontend track — 폼 구현 (backend api-contract 소비)
```

---

## 동기화 확인 명령

Claude가 다음 track을 시작하기 전 수동으로 확인:

```bash
# SYNC-1 확인
test -f docs/api-contract.yaml && echo "SYNC-1 완료" || echo "SYNC-1 대기 중"

# SYNC-2 확인  
go run ./cmd/migrate status | grep -q "up to date" && echo "SYNC-2 완료"

# 모든 sync 한 번에 확인
for sync in docs/api-contract.yaml docs/db-migration-complete.flag; do
  [[ -f "$sync" ]] && echo "✓ $sync" || echo "✗ $sync (대기 중)"
done
```

---

## 충돌 없는 병합을 위한 파일 분리 원칙

각 actor-track이 수정하는 파일 영역이 겹치지 않도록:

| Actor | 소유 디렉토리/파일 |
|-------|-----------------|
| frontend | `src/components/`, `src/pages/`, `src/hooks/` |
| backend | `internal/`, `cmd/`, `pkg/` |
| integration | `pkg/integrations/`, `config/integrations.yaml` |
| infra | `migrations/`, `terraform/`, `k8s/` |

경계 파일 (공유 — 순차 수정 필요):
- `docs/api-contract.yaml` — backend가 먼저 생성, frontend가 읽기만
- `src/types/` — backend가 먼저 정의, frontend가 import만
- `go.mod`, `package.json` — 충돌 발생 시 수동 병합
