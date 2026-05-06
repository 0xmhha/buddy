---
name: guide-setup-wizard
description: "[패턴 라이브러리] auto-detect → picker → verify pattern으로 credential/config setup flow 설계. 가능한 option만 노출 + verification 후 success 선언. 직접 invoke보다 orchestrator가 import해 사용. 트리거: '설정 마법사' / '초기 설정 가이드' / 'setup wizard' / '셋업 도와줘'. 참조 위치: detect-install-type 페어, 부트스트랩 1회, 신규 사용자 onboarding."
type: skill
---

# Snippet: Interactive Setup Wizard Pattern


> 대화형 설정 wizard 구조 — 도구별 Keychain/CDP plumbing 제외.

## 이 snippet을 사용하는 경우

- API key / OAuth 설정 흐름
- Credential 마이그레이션 (한 프로바이더 → 다른 프로바이더)
- 여러 옵션이 존재하는 도구 설정
- 모든 "first-run setup" UX

## 패턴: 4-Phase Wizard

핵심 아이디어: 사용자에게 모든 옵션을 쏟아붓지 말 것. 먼저 사용자 환경에서 실제로 가능한 게 뭔지 감지하고, 작은 picker를 제시하고, 선택이 동작하는지 검증한 뒤에 성공 선언.

### Phase 1: Auto-Detect

Prompt **전에** 환경을 스캔한다. 목표는 불가능한 옵션을 제거하고 이미 설정된 경우 short-circuit.

이 phase에서 답할 두 질문:

1. **설정이 이미 완료됐나?** 그렇다면 그렇게 말하고 중단. wizard 재실행 금지.
2. **여기서 실제로 사용 가능한 source는 뭔가?** (설치된 앱, 설정된 프로바이더, env var, 사용 가능한 auth 백엔드 등)

```bash
# 예시: 이미 설정됐으면 short-circuit
if already_configured; then
  echo "Not needed — credentials already loaded from <source>. Stop."
  exit 0
fi

# 예시: 사용 가능한 것만 열거
detected_sources=$(scan_environment_for_credential_sources)
```

아무것도 감지되지 않으면, 0개 옵션의 picker를 제시하지 말고 가능한 가장 작은 다음 스텝과 함께 시끄럽게 실패("X를 설치한 뒤 재실행").

### Phase 2: Interactive Picker

Phase 1에서 감지된 것**만** 옵션별 한 줄 설명과 함께 제시. 이 머신에서 사용자가 실제로 고를 수 없는 옵션은 절대 표시하지 말 것.

지원할 가치가 있는 두 가지 상호작용 모드:

- **Default (interactive):** 감지된 source를 나열하는 picker UI나 prompt 기반 메뉴 오픈. 목록이 길면 search/filter 허용.
- **Direct (one-shot):** 사용자가 이미 커맨드라인에서 source 이름을 지정했다면 (예: `setup-credential github`), picker를 skip하고 그 선택으로 바로 Phase 3로.

UX 규칙:

- 도메인/프로바이더 이름과 짧은 상태("configured", "stale", "ready")만 표시. 비밀 값은 표시하지 말 것.
- 파괴적 액션(기존 credential 제거)은 명백한 confirm 스텝 필요.
- Picker가 외부에서 열리면(브라우저 탭, GUI), 사용자에게 명시적으로 알림: "Picker가 열렸습니다 — 거기서 선택하신 뒤 완료되면 알려주세요."

### Phase 3: Verify

진행하기 전에 credential이 **실제로 동작하는지** 테스트. 이게 "wizard 종료"와 "wizard 성공"을 가르는 단계다.

```bash
# credential을 실제로 사용하는 가장 작은 현실적 호출 실행
verify_output=$(probe_with_credential 2>&1)
if probe_succeeded "$verify_output"; then
  echo "Verified: <what was checked>"
else
  echo "Verification failed: <reason>. Run <next-step> to retry."
  exit 1
fi
```

도메인별 검증 예시:

- API key → 한 번의 저렴한 인증 read (`/me`, `/whoami`, 아이템 하나 list).
- OAuth 토큰 → scope가 워크플로우가 필요로 하는 것과 일치하는지 확인.
- Cookie/session import → 넘어온 세션 도메인과 개수 list.
- Config file → 재파싱하고 필수 필드가 있는지 assert.

"import가 에러를 출력하지 않았으니까" 검증을 skip하지 말 것. credential 설정 중 silent failure는 전체 워크플로우에서 가장 비싼 실패 모드다.

### Phase 4: Confirm + Next Steps

사용자에게 한 줄 요약으로 알림:

1. **무엇이 설정됐는지** (어떤 source, 어떤 scope/도메인).
2. **이제 뭘 할 수 있게 됐는지** 이전에 못하던 것.
3. **이를 활용할 정확한 다음 명령**.

~3줄로 유지. 사용자는 이제 wizard를 떠나 도구를 사용할 준비가 됐다.

## 안티패턴

- 환경에 상관없이 모든 이론적 옵션을 쏟아붓기 (2개만 설치됐는데 12개 auth 프로바이더 picker 제시).
- 검증 없이 설정 성공 표시. ("Imported." OK, 근데 실제로 인증되나?)
- Silent failure: 검증을 돌렸지만 결과를 surface하지 않음.
- 이미 설정됐는데 전체 wizard 재실행. Phase 1 상단에 항상 short-circuit.
- 도구가 스스로 감지할 수 있는 옵션을 사용자에게 고르라고 요청.
- Picker UI에 비밀 값 표시. 이름과 개수만.
