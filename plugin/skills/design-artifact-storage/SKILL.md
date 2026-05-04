---
name: design-artifact-storage
description: "patch / git_bundle / template / package 같은 immutable artifact의 저장·검증·배포 설계. SHA-256 hash, author signature, signed download URL, security scan 결과 보존, immutable revisioning. feature-management-saas-mcp의 patch artifact 핵심. 트리거: 'artifact 저장 설계' / 'patch hash 검증' / 'signed URL' / 'immutable artifact' / 'object storage' / 'patch download 보안' / 'author signature'. 입력: artifact 종류 + 저장 backend + license 정책 + 다운로드 보안 요구. 출력: artifact pipeline + storage schema + verification flow + signed URL 발급. 흐름: design-mcp-server → design-artifact-storage → build-with-tdd."
type: skill
---

# Design Artifact Storage — Immutable Patch / Bundle / Template 저장소

## 1. 목적

`feature-management-saas-mcp`의 patch artifact, git bundle, template, package 같은 **immutable artifact**의 저장·검증·배포 시스템을 설계한다.

핵심 보장 4개:
1. **Immutability** — 한 번 발행된 artifact는 변경 불가. 변경 시 새 revision.
2. **Integrity** — SHA-256 hash로 다운로드 후 검증 가능.
3. **Authenticity** — author signature로 발행자 검증.
4. **Authorization** — paid / private artifact는 entitlement 확인 후 short-TTL signed URL.

이 스킬을 통과한 storage는 patch download 전 hash + signature + security scan + license 4단계 검증을 강제한다.

## 2. 사용 시점 (When to invoke)

- `feature-management-saas-mcp` patch artifact module 구현 전
- 모델 weights / dataset 같은 ML artifact 배포
- 공유 template / starter kit 발행
- enterprise customer에게 deployable bundle 제공
- third-party plugin / extension marketplace
- backup / archive / immutable log 저장소
- 인수합병 due diligence 지원 (audit-friendly storage)

## 3. 입력 (Inputs)

### 필수
- artifact 종류 (patch / git_bundle / template / package / model_weights)
- 저장 backend (S3 / R2 / GCS / Azure Blob / 자체 MinIO)
- license / commercial 정책 (free / paid / private)
- 다운로드 보안 요구 (signed URL, IP 제한, audit log)

### 선택
- CDN 전면 (빈번한 다운로드 시)
- multi-region replication
- KMS 키 관리
- compliance frame (HIPAA, SOC2)

### 입력 부족 시 forcing question
- "artifact 평균 크기 얼마야? 10KB? 10MB? 1GB? backend 결정 영향."
- "다운로드 빈도 어때? 빈번하면 CDN. 드물면 직접 S3."
- "author signature가 self-signed야 CA 기반이야? key rotation 정책?"
- "private artifact는 multi-tenant? tenant-isolated bucket vs 단일 bucket + access control?"

## 4. 핵심 원칙 (Principles)

1. **Immutable artifact** — 발행 후 변경 금지. 변경 시 새 revision (artifact_id 새로 발급).
2. **Hash는 발행 시점 고정** — SHA-256 (또는 SHA-512). artifact metadata에 저장.
3. **Author signature는 hash 위에** — content가 아닌 hash에 서명. tamper detection.
4. **Signed URL은 short TTL** — 5분 / 15분. 유출 시 영향 최소화.
5. **Security scan은 발행 전 게이트** — high severity 발견 시 발행 차단.
6. **License는 metadata 별도** — artifact content 안에 묻지 말고 metadata에. 검색 + 변경 가능.
7. **Audit log 모든 download** — who / when / what / from where. 분쟁 대응.
8. **Versioning은 semver + revision** — feature@1.1.0의 patch_revision_2. 동일 version 다른 revision 허용 (메타 변경).

## 5. 단계 (Phases)

### Phase 1. Backend Selection
- **S3 / R2 / GCS / Azure Blob**: 일반. egress 비용 고려.
- **CloudFlare R2**: egress 무료. 글로벌 CDN 무료.
- **MinIO**: self-hosted, on-prem 요구.
- **IPFS**: content-addressable, 분산. 고도 사용 case.

### Phase 2. Schema Design
artifact metadata vs binary 분리:
- metadata: PostgreSQL / DynamoDB
- binary: object storage

```
PatchArtifact (metadata)
  id, implementation_id, version, revision
  source_commit_range
  storage_backend, storage_key
  size_bytes
  sha256, sha512
  signature, signed_by, signed_at
  license, commercial_policy
  visibility (public / private / org)
  security_scan_id, security_status
  apply_guide_url, migration_guide_url
  created_at
```

### Phase 3. Upload Pipeline
1. uploader가 binary + metadata 제출
2. license / visibility / commercial policy 검증
3. SHA-256 / SHA-512 계산
4. security scan 큐잉
5. scan 통과 시 발행 (publish state)
6. scan 실패 시 quarantine state
7. author signature (post-scan)

### Phase 4. Download Pipeline
1. 사용자 / agent download 요청 (`patch.download`)
2. authentication + authorization (token scope)
3. entitlement 검증 (free / paid 결제 / point 차감)
4. license / commercial policy 명시 동의 확인
5. security scan 결과 확인 (high severity 차단)
6. point 예약 (paid 시)
7. signed URL 발급 (5-15분 TTL)
8. 다운로드 성공 이벤트 수신 (webhook 또는 redirect)
9. point 확정 차감
10. audit log 기록

### Phase 5. Verification Flow (`patch.verify` MCP tool)
다운로드 전 검증:
- hash 일치 (재계산 가능)
- signature 유효 (CA / web of trust / self-signed 정책)
- security_status = passed
- license / commercial policy 합치
- artifact 존재 + 만료 안 됨

### Phase 6. Revision Management
artifact 교체 시:
- 기존 artifact 삭제 금지 (audit / 기존 다운로드 보존)
- 새 revision 생성 (revision_id 새로)
- 기존 revision은 deprecated 마킹 + replaced_by 링크
- 다운로드 history는 원본 보존

### Phase 7. Multi-Tenancy
- bucket 분리 vs 단일 bucket + path prefix
- tenant_id를 path에 포함 (또는 metadata)
- access control (signed URL 발급 시 tenant 검증)
- billing (storage + egress 별 tenant)

### Phase 8. Compliance
- 보존 기간 (retention)
- legal hold (소송 / 감사 시 삭제 차단)
- data residency (EU artifact는 EU region)
- encryption (at-rest, in-transit)

## 6. 출력 템플릿 (Output Format)

```yaml
artifact_storage_system:
  primary_backend: cloudflare_r2  # or s3, gcs, minio
  cdn: cloudflare  # or cloudfront
  egress_cost_strategy: r2_zero_egress | cloudfront_optimization

schema:
  metadata_db: postgres
  binary_store: r2_bucket
  table:
    PatchArtifact:
      fields:
        - id (uuid, pk)
        - implementation_id (fk)
        - version (semver)
        - revision (int, autoincrement per version)
        - storage_backend (enum)
        - storage_key (string)
        - size_bytes (bigint)
        - sha256 (string, 64-char)
        - sha512 (string, optional)
        - signature (text)
        - signed_by (string, public_key_id)
        - signed_at (timestamp)
        - license (string, SPDX)
        - commercial_policy (enum)
        - visibility (enum: public | org | private)
        - security_scan_id (fk)
        - security_status (enum: passed | quarantine | failed)
        - apply_guide_url (string)
        - migration_guide_url (string)
        - download_count (bigint)
        - state (enum: draft | published | quarantine | deprecated)
        - replaced_by (fk, optional)
        - created_at (timestamp)

upload_pipeline:
  steps:
    - validate_metadata (license, visibility, commercial)
    - compute_hash (sha256, sha512)
    - security_scan (async, gating)
    - on_pass:
        - sign (author hash signature)
        - state = published
    - on_fail:
        - state = quarantine
        - notify uploader
  rollback: yes

download_pipeline:
  steps:
    - authenticate (token verify)
    - authorize (scope: download:patches)
    - check_entitlement (free | paid via points | subscription)
    - verify_license_acknowledged (yes for paid)
    - check_security_status (block if quarantine)
    - reserve_points (if paid)
    - generate_signed_url (TTL 15 min)
    - record_download_event (audit)
    - on_download_success:
        - confirm_points_charge
        - increment_download_count
        - revenue_share_calc

verification_pipeline:
  tool: patch.verify (MCP)
  checks:
    - hash_recompute_match
    - signature_valid
    - security_status_passed
    - license_compatible (intended_use)
    - commercial_policy_satisfied
    - artifact_state_published

signing:
  algorithm: ed25519  # or RSA-PSS
  key_management:
    storage: KMS  # AWS KMS, GCP KMS, Vault
    rotation: yearly
    revocation: published_revocation_list
  pubkey_distribution: known_authors_endpoint

signed_url:
  ttl_seconds: 900  # 15 min
  include_ip_lock: optional  # high-security
  include_user_lock: yes  # user_id in URL signed payload
  one_time_use: yes  # nonce-based

multi_tenant:
  isolation: path_prefix  # tenant_id/feature_id/patch_id
  access_control: signed_url_includes_tenant
  billing:
    storage_per_tenant: yes
    egress_per_tenant: yes

revision_management:
  immutability: enforced
  on_replace:
    - new_revision_id
    - mark_old_deprecated
    - link_replaced_by
    - notify_existing_downloaders (optional)
  retention: forever  # legal_hold 가능

security:
  scan_pipeline:
    - dependency_vulnerability
    - secret_scan (gitleaks)
    - malware_scan (clamav)
    - sast (codeql / semgrep)
    - license_scan (scancode / fossa)
  scan_provider: snyk | github_advanced | self_hosted
  block_on_high_severity: yes
  rescan_frequency: weekly  # for published artifacts

compliance:
  retention_default: 7_years
  legal_hold:
    enabled: yes
    storage_class: glacier
  data_residency:
    eu_artifacts: eu_region
    kr_artifacts: kr_region
  encryption:
    at_rest: aes_256_kms
    in_transit: tls_1_3

audit_log:
  events:
    - artifact_uploaded
    - artifact_signed
    - artifact_security_scanned
    - artifact_published
    - artifact_quarantined
    - artifact_downloaded
    - artifact_deprecated
    - artifact_legal_hold_added
  fields: [timestamp, actor_id, action, artifact_id, ip, user_agent, success, error]
  retention: 7_years
  immutable: append_only (e.g. AWS CloudTrail)
```

## 7. 자매 스킬 (Sibling Skills)

- 앞 단계: `design-mcp-server` — `Skill` tool로 invoke (`patch.verify`, `patch.download` 정의)
- 페어: `design-billing-system` — paid artifact point 차감
- 페어: `audit-security` — security scan + supply chain
- 페어: `review-license-and-ip-risk` — license metadata 정렬
- 다음 단계: `build-with-tdd` — upload / download / verify 시나리오 TDD

## 8. Anti-patterns

1. **Mutable artifact** — 같은 artifact_id에 binary 교체. 기존 다운로더 hash mismatch.
2. **Long TTL signed URL** — 1일 TTL? 유출 시 1일간 자유 다운. 5-15분 강제.
3. **Hash 검증 client-side만** — server는 발급만? 분쟁 시 증거 약. server-side 재계산.
4. **Author signature 없음** — supply chain attack 방어 약. 강제.
5. **Security scan 사후 (발행 후)** — 발행 후 발견? 다운로더 영향. 발행 전 게이트.
6. **License를 binary에 hardcode** — 변경 시 binary 재발행. metadata 별도.
7. **Audit log 없음** — 분쟁 시 증거 없음. immutable append-only.
8. **Multi-tenant access 단일 bucket + 단일 token** — tenant 격리 약. signed URL에 tenant lock.
