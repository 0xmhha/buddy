---
description: secret store + rotation + access audit + 4-layer leak detection (pre-commit / repo scan / runtime / public) + 5-step incident response.
argument-hint: "<제품 이름 또는 secret 목록>"
disable-model-invocation: true
---

# /buddy:design-secret-management

API key / DB / OAuth / TLS 등 secret 의 저장 + 회전 + audit + 누수 탐지 통합 설계. runtime + dev secret 분리. plaintext-in-repo 0건 강제.
