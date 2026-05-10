---
description: cloud / SaaS 비용 spike anomaly detection (moving average ±2σ) + drill-down 5차원 (service/component/region/account/tag) + root cause 6 분류 + 3 단계 alert + 5-step recovery.
argument-hint: "<billing period 또는 service>"
disable-model-invocation: true
---

# /buddy:analyze-cost-anomaly

audit-cost-efficiency (§6, 구현됨) 와 책임 분리 — efficiency 는 최적화 일반, anomaly 는 spike 사건 대응. tagging 정책 활성 + alert fatigue 회피. conduct-postmortem 학습.
