# Archive Product — EOL (End of Life) checklist + data export + tombstone

## 1. 목적

product 전체의 *end-of-life* 절차. **EOL 결정 → customer migration → data export → service shutdown → tombstone → legal/compliance** 6 단계 통합.

`deprecate-feature` 가 *single feature* 면, 본 skill 은 *product 전체*. 가장 큰 lifecycle 결정.


## Input Requirements

| Input | Required | Type | Source | 미제공 시 |
|-------|----------|------|--------|----------|
| EOL 결정 | ✅ | decision | 사용자 의사결정 | "제품 EOL을 확정하나요?" |
| 제품 정보 | ✅ | knowledge | 사용자 도메인 지식 | "아카이브할 제품명과 범위를 알려주세요." |

## Output Contract

| Output | Type | Format | Consumers |
|--------|------|--------|-----------|
| EOL documentation (data export + tombstone + legal + knowledge) | artifact | structured document | (아카이브) |

## 2. 사용 시점

- §9 manage-lifecycle 의 *product retirement* 결정 후
- M&A 후 acquired product retire
- pivot 으로 기존 product line 종료
- 사업성 (`assess-business-viability`) 결과 *no-go* 후 신규 line 으로 전환
- compliance 강제 종료 (regulatory exit)

## 3. 입력

### 필수
- archive 대상 product
- 사용자 base + tier 분포
- contractual obligation list (enterprise SLA)

### 선택
- 후속 product (있으면 — migration target)
- legal review (data retention / customer notification 의무)

## 4. Stage 흐름

### Stage 1: EOL 결정 + 공식 발표

| 결정 트리거 | 검토 |
|----------|------|
| 사용자 base 충분히 작음 | 운영 비용 vs ARR |
| Strategic 변화 | 후속 product 와 cannibalize |
| Regulatory | compliance 강제 |
| 보안 / 안정성 | fix 비용 비현실적 |

→ EOL 결정 문서 (ADR) + executive 승인.

### Stage 2: Timeline (보통 6~12 month)

| month | activity |
|-------|---------|
| -12 | 내부 결정 + ADR |
| -9 | enterprise customer 1:1 outreach |
| -6 | public 발표 + migration plan |
| -3 | sign-up 차단 + final reminder |
| -1 | data export deadline 안내 |
| 0 | service shutdown |
| +1~+3 | data retention / final export window |
| +6 | data destruction / GDPR right-to-erasure |

### Stage 3: Customer migration

`migrate-customers` (§9) 호출:
- 후속 product 있으면 → migration plan
- 후속 없으면 → 경쟁사 추천 (윤리적)
- enterprise tier → 1:1 dedicated migration

### Stage 4: Data export

사용자 자기 데이터 *download* 가능:

| format | 용도 |
|--------|-----|
| JSON / CSV | 일반 사용자 |
| SQL dump | 기술 사용자 |
| archive (zip / tar) | 통합 export |

→ export window 충분 (3 month+ 후 EOL). GDPR / 개인정보보호법 의 *data portability* 권리.

### Stage 5: Service shutdown

| 단계 | 시점 |
|------|-----|
| Sign-up 차단 | -3 month |
| New transaction 차단 | -1 month |
| Login 차단 | EOL day |
| API access 차단 | EOL day |
| Database read-only | EOL day~+3 month |
| Database destroy | +6 month (GDPR) |

→ 각 단계 *명시 communication*.

### Stage 6: Tombstone

shutdown 후 *접근 시 안내*:

| layer | 내용 |
|------|-----|
| Domain redirect | 후속 product 또는 정보 page |
| Tombstone page | EOL 안내 + data export link (window 안) + 후속 안내 |
| Email auto-reply | inquiry 시 EOL 안내 |
| Documentation | 마지막 version archive (legal / 학술 가치) |

### Stage 7: Legal / compliance

| 영역 | 의무 |
|------|-----|
| Data retention | region 법령 (GDPR 30~90일 / KISA / etc) |
| Customer notification | 30일+ 사전 (계약 의무) |
| Tax / 회계 | revenue recognition 종료 |
| Trademark | 유지 / 폐기 결정 |
| Open source | 코드 archive (GitHub archive 모드) |

### Stage 8: Knowledge preservation

product 의 *학습 자산* 영속화:
- ADR archive
- postmortem / retro 종합
- key decision history → `persist-learning-jsonl` (구현됨)
- 후속 product 입력 (lessons learned)

## 5. 산출물 형식

```markdown
## Product Archive Plan — {product}

### EOL 결정
- 트리거: ...
- ADR: ...

### Timeline
| month | activity |

### Customer migration
- target: 후속 / 경쟁사 추천
- enterprise 1:1 plan

### Data export
- format / window / GDPR 의무

### Service shutdown phases
| phase | timing |

### Tombstone
- domain redirect / page / email / docs

### Legal / compliance
- retention / notification / tax / trademark / OSS

### Knowledge preservation
- ADR / postmortem / learnings → 후속
```

## 6. 검증

- [ ] EOL 결정 ADR + executive 승인?
- [ ] Timeline 6~12 month 단계?
- [ ] Customer migration plan (후속 / 경쟁사 / 1:1)?
- [ ] Data export 3 format + window?
- [ ] Service shutdown 6 단계?
- [ ] Tombstone 4 layer?
- [ ] Legal / compliance 5 영역?
- [ ] Knowledge preservation 영속화?

## 7. 다음 phase

- `migrate-customers` cascade
- `spin-off-feature` — archive 후 *부분 기능* 만 별도 product 로 전환 시
- `persist-learning-jsonl` (구현됨) — 학습 영속화

## 8. 참조

- Google EOL playbook (Reader / Stadia / etc 사례)
- GDPR Right to Data Portability (Article 20)
- 한국 개인정보보호법 — 보존 / 파기 의무
- buddy `deprecate-feature` 와 책임 분리 (single feature vs entire product)
