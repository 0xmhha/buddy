# §9 Retirement / Decommissioning (폐기·종료) — Stage Skills

- **Orchestrator (entry)**: `manage-lifecycle`
- **DoR → DoD**: usage/adoption data + 폐기 결정 → deprecation plan + migration plan + EOL documentation
- 정체성 원본: `engineering-phases.md` §2 Phase 9 (Mode B와 분류 정책) | 명사 원본: `se-lifecycle-naming.md` §1
- 분류 정책: **외부 사용자 영향 있음** → §9. **내부만 영향**(dead code 제거 등) → §1 Mode B(`assess-product-change`, scope small/medium).

## Stage skills

| Skill | Trigger | When to use | Not when (→ 대안) |
|---|---|---|---|
| `deprecate-feature` | command + dispatch | feature sunset — timeline + sunset notice 5 layer + telemetry + migration path | 내부 dead code 제거(외부 영향 X) → §1 `assess-product-change` |
| `migrate-customers` | command + dispatch | 대규모 customer migration — tier segmentation + batch + rollback (Strangler Fig) | feature 1건 sunset → `deprecate-feature` |
| `archive-product` | command + dispatch | product 전체 EOL — data export(GDPR Art.20) + tombstone + legal + knowledge preservation | 일부 feature만 종료 → `deprecate-feature` |
| `spin-off-feature` | command + dispatch | 기능을 별도 product/repo로 분리 — 적합성 5차원 + 코드 분리 5 패턴 | 종료(분리 아님) → `deprecate-feature`/`archive-product` |

## Disambiguation (노드 내)
- `deprecate-feature`(기능 1개 종료) vs `archive-product`(제품 전체 EOL) vs `spin-off-feature`(종료가 아닌 분리·존속).
- 외부 영향 유무가 §9 vs §1 Mode B 분기 기준 (위 분류 정책).

## Escalation
- 노드 내 2개+ 모호 → `../routing-rules.md` §3. DoR(usage data) 부재 → §8 `iterate-product` 선행. 내부 정리면 §1 Mode B.
