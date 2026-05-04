# Audit Security — 출력 포맷 상세

> 본 파일은 `audit-security` 스킬의 detailed reference입니다. 메인 SKILL.md에서 참조하며, on-demand 로드됩니다.

## Finding 테이블

```
SECURITY FINDINGS
=================
#   Sev    Conf   Status      Category              Finding                          File:Line
--  ----   ----   ------      --------              -------                          ---------
1   CRIT   9/10   VERIFIED    Secrets               AWS key in git history           .env:3 (commit abc123)
2   CRIT   9/10   VERIFIED    CI/CD                 pull_request_target + checkout   .github/ci.yml:12
3   HIGH   8/10   VERIFIED    Supply Chain          postinstall script in prod dep   node_modules/foo/package.json
4   HIGH   9/10   UNVERIFIED  Webhook               Webhook without signature verify api/webhooks.ts:24
5   HIGH   8/10   VERIFIED    LLM                   user input in system prompt      llm/chat.ts:47
```

## Finding별 포맷

```
## Finding N: [Title] — [File:Line]

* Severity:    CRITICAL | HIGH | MEDIUM | LOW
* Confidence:  N/10
* Status:      VERIFIED | UNVERIFIED | TENTATIVE
* Category:    Secrets | Supply Chain | CI/CD | LLM | OWASP A01-A10 | STRIDE | Webhook | Infrastructure
* Description: What is wrong, in one sentence.
* Exploit scenario: Step-by-step attack path. Concrete. Named threat actor
  if relevant.
* Impact:      What the attacker gains. Data, money, access.
* Recommendation: Specific fix with example code or config.
```

## Incident Response 플레이북

Leak된 시크릿에 대해 모든 finding이 플레이북과 함께 ship:

1. **Revoke** 즉시 provider에서 자격증명
2. **Rotate** — 새 자격증명 발급
3. **Scrub history** — `git filter-repo` 또는 BFG Repo-Cleaner
4. **Force-push** 깨끗한 히스토리 (팀과 조율)
5. **노출 window audit** — 언제 commit? 언제 감지? Repo public? 어디든 mirror?
6. **Abuse 체크** — Leak된 자격증명의 provider audit 로그 리뷰
7. **감지 업데이트** — Leak된 패턴에 `.gitleaks.toml` 규칙 추가

## Remediation 로드맵

상위 5 finding에 대해 결정 artifact 생산:

1. Context: 취약점, 심각도, exploit 시나리오
2. 추천: Fix now / Mitigate / Accept risk / Defer
3. 옵션:
   - **A) Fix now** — 구체 코드 변경, 노력 추정
   - **B) Mitigate** — full fix 없이 리스크 줄이는 workaround
   - **C) 리스크 수락** — 결정 문서화, 리뷰 날짜 설정
   - **D) Defer** — 날짜와 함께 보안 backlog에 추가

## Persistence

리포트를 이 schema로 `~/<your-app>/security/reports/{YYYY-MM-DD}-{HHMMSS}.json`에 저장:

```json
{
  "version": "2.0.0",
  "date": "ISO-8601-datetime",
  "mode": "daily | comprehensive",
  "scope": "full | infra | code | supply-chain | owasp",
  "diff_mode": false,
  "categories_run": ["secrets", "supply_chain", "ci_cd", "llm", "owasp", "stride", "webhook", "infrastructure"],
  "attack_surface": {
    "code": { "public_endpoints": 0, "authenticated": 0, "admin": 0, "api": 0, "uploads": 0, "integrations": 0, "background_jobs": 0, "websockets": 0 },
    "infrastructure": { "ci_workflows": 0, "webhook_receivers": 0, "container_configs": 0, "iac_configs": 0, "deploy_targets": 0, "secret_management": "unknown" }
  },
  "findings": [{
    "id": 1,
    "severity": "CRITICAL",
    "confidence": 9,
    "status": "VERIFIED",
    "category": "Secrets",
    "fingerprint": "sha256-of-category-file-title",
    "title": "...",
    "file": "...",
    "line": 0,
    "commit": "...",
    "description": "...",
    "exploit_scenario": "...",
    "impact": "...",
    "recommendation": "...",
    "playbook": "...",
    "verification": "independent | self-verified"
  }],
  "supply_chain_summary": {
    "direct_deps": 0, "transitive_deps": 0,
    "critical_cves": 0, "high_cves": 0,
    "install_scripts": 0,
    "lockfile_present": true, "lockfile_tracked": true,
    "tools_skipped": []
  },
  "filter_stats": {
    "candidates_scanned": 0,
    "hard_exclusion_filtered": 0,
    "confidence_gate_filtered": 0,
    "verification_filtered": 0,
    "reported": 0
  },
  "totals": { "critical": 0, "high": 0, "medium": 0, "low": 0, "tentative": 0 },
  "trend": {
    "prior_report_date": null,
    "resolved": 0, "persistent": 0, "new": 0,
    "direction": "first_run | improving | degrading | stable"
  }
}
```

`~/<your-app>/security/`가 `.gitignore`에 있는지 확인 — 보안 리포트는 로컬 유지.
