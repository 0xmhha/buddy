---
description: SLO error budget burn rate (multi-window 1h/6h/1d/3d × multi-burn 14.4×/6×/3×/1×) + release gate 4 단계 (>50/25-50/<25/0% 소진) + alert fatigue 회피 (fast burn 만 page).
argument-hint: "<SLO 정의 또는 분기>"
disable-model-invocation: true
---

# /buddy:audit-error-budget

design-observability SLO 의 budget consumption 측정 + release decision driver. Google SRE multi-window multi-burn-rate alert 패턴. analyze-actor-failure-rate / chaos-test / analyze-cost-anomaly 의 burn 통합.
