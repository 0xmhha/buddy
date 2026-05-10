---
description: feature 단위 effort estimation — T-shirt sizing (XS/S/M/L/XL) + ideal-h × multiplier + 4-point uncertainty (best/expected/p90/worst). XL / 3× uncertainty 발견 시 분해 강제.
argument-hint: "<feature spec list 또는 backlog>"
disable-model-invocation: true
---

# /buddy:estimate-feature-effort

define-feature-spec 산출 → feature 별 T-shirt size + ideal-h × multiplier (1.0~2.5×) + 4-point PERT uncertainty. XL 또는 uncertainty 폭 3× 초과 시 split-work-into-features 분해 강제. score-feature-priority 의 effort 입력 + estimate-build-timeline (§4) task layer cascade.
