---
description: LSP rename + 호출 그래프 cross-check + grep 누락 검증 + test baseline / 후 검증. pure rename = 단일 commit (기능 변경 동시 X). public API 시 deprecation alias chain.
argument-hint: "<old name → new name 또는 file>"
disable-model-invocation: true
---

# /buddy:refactor-with-rename-trace

광범위 영향 refactor (rename / 분해 / dead code 제거) 의 안전 절차. LSP 인식 못한 영역 (config / template / 외부) 식별 + 단일 refactor = 단일 commit.
