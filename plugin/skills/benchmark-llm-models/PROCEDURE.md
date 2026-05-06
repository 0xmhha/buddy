---
name: benchmark-llm-models
description: "[패턴 라이브러리] multi-provider LLM benchmark 패턴 (Claude/GPT/Gemini) — dry-run auth verification, provider selection UX, judge cost transparency. 직접 invoke보다 orchestrator가 import해 사용. 트리거: 'Claude/GPT/Gemini 벤치' / '모델 비교' / 'LLM 벤치마크' / 'model benchmark'. 참조 위치: monitor-regressions 페어, 모델 선택 검증."
type: skill
---

# Snippet: Dry-Run Pre-Flight + Cost Transparency


> Pre-flight + cost transparency 패턴. 실제 벤치마크 실행(Claude / GPT / Gemini 등)은 도구별 구현.

## 이 snippet을 사용하는 경우
- 하나의 워크플로우에서 여러 LLM 벤더를 호출하는 모든 도구
- 사전에 사용자가 확인해야 하는 실제 API 비용이 발생하는 모든 연산
- 벤더마다 별도 로그인 상태를 갖는 multi-CLI 워크플로우의 사전 인증 체크

---

## 패턴 1: Dry-Run 인증 체크

비용 유발 액션을 취하기 **전에** 인증/가용성 probe를 실행한다. 사용자가 항상 보는 별도의 필수 단계로 취급한다.

```
<binary> --prompt "unused, dry-run" --models claude,gpt,gemini --dry-run
```

dry-run 출력에는 프로바이더별 "Adapter availability" 섹션이 포함돼야 한다:
- `OK` — 인증 + 도달 가능
- `NOT READY` — 한 줄짜리 해결 힌트 포함 (예: `claude login`, `codex login`, `gemini login`, `export GOOGLE_API_KEY`)

dry-run 이후 의사결정 로직:
1. **모든 프로바이더 NOT READY** → STOP. 진행하지 않는다. 무엇을 고쳐야 할지 사용자에게 알린다.
2. **최소 하나라도 OK** → 프로바이더 선택(패턴 2)으로 진행. 미인증 프로바이더는 **clean하게 skip**되며 배치를 abort하지 않는다.

별도 dry-run 단계가 필요한 이유:
- 배치 중간의 인증 실패는 이미 이전 프로바이더에 쓴 돈을 낭비한다
- 사용자는 전체 실행에 commit하기 전에 상태를 본다
- "clean skip, abort 금지" 규칙 덕에 부분 결과도 여전히 가치가 있다

---

## 패턴 2: 멀티 프로바이더 선택

dry-run 이후 어떤 프로바이더를 포함할지 사용자에게 묻는다. 기본은 최대 coverage; 필요하면 좁힌다.

질문 프레이밍:
> "어떤 모델을 포함할까요? 위 dry-run에서 어떤 게 인증됐는지 확인할 수 있습니다. 미인증 프로바이더는 clean하게 skip됩니다 — 배치를 abort하지 않습니다."

권장 옵션 (프로바이더 무관 형태):
- **A) 인증된 모든 프로바이더** — 가장 풍부한 비교, 권장 기본값
- **B) 단 하나의 프로바이더** — cross-model signal을 잃는다고 명시 (가치 감소)
- **C) 일부만 선택** — 사용자가 다음 턴에 지정

규칙:
- wrapper에 **모델 이름을 하드코딩 금지**. 사용자 선택을 바이너리에 그대로 전달.
- **"configured"가 아니라 "all-authed"가 기본값**. 미인증은 조용히 skip.
- 각 옵션의 완성도/가치 등급을 보여 사용자가 좁힐 때 잃는 것을 이해하게 한다.

---

## 패턴 3: Judge 비용 투명성

LLM-as-judge로 품질 점수를 매길 때, judge 자체도 실제 비용이 든다. invoke **전에** 예상치를 보여주고 명시적 opt-in을 요구한다.

사전 확인: judge 백엔드 가용성 확인 (예: `ANTHROPIC_API_KEY` env var 또는 저장된 credential). 사용 불가면 질문 자체를 건너뛰고 judge 플래그를 생략한다 — 사용자가 yes라고 답할 수 없는 질문은 묻지 않는다.

질문 프레이밍 (가용 시):
> "품질 judge가 각 모델 출력을 [모델명] 기준 0-10 척도로 채점합니다. 실행당 ~$0.05 추가. latency와 비용뿐 아니라 출력 품질이 중요하면 권장."

옵션:
- **A) Judge 사용 (+ ~$X)** — 품질이 중요하면 권장
- **B) Judge skip** — 속도/비용/토큰만, 가치 낮은 비교

규칙:
- **judge 자동 활성화 금지.** 실제 비용이 발생하므로 사용자가 매 run마다 opt-in해야 한다.
- **달러 추정치를 질문 안에 직접** 표시, 다른 곳에 묻히지 않도록.
- judge가 실행되면 최종 요약에 "Highest quality" 승자를 명시적으로 라벨링.

---

## 패턴 4: 비용 가드레일 (횡단적)

judge만이 아닌 전체 워크플로우에 적용:

- **비용은 run마다 가시적.** 결과 테이블이 프로바이더별 비용을 보여준다. 재실행 여부를 결정하기 전에 사용자가 본다.
- **조용한 재시도 금지.** 프로바이더 에러(auth/timeout/rate_limit) 발생 시 해결 경로와 함께 call out. 루프 돌며 다시 쓰지 않는다.
- **Baseline은 opt-in으로 저장.** 차후 run이 diff할 수 있도록 JSON으로 저장 제안(model drift / 품질 회귀 감지). 기본 저장 금지 — 사용자가 선택.

---

## 안티패턴

- 벤치마크를 먼저 돌리고 중간에 인증 실패 발견 — 이미 쓴 돈 낭비
- invoke 후까지 비용을 숨김 (예: 결과 테이블에서만 요금 표시)
- all-or-nothing 프로바이더 선택 (가장 느린 프로바이더를 기다리거나 실패 시 재실행 강제)
- "품질이 중요하니까" LLM judge 자동 활성화 — 사용자가 매 비용에 동의해야 함
- 프로바이더 목록 하드코딩 — 신규 벤더마다 config 변경이 아닌 코드 변경이 됨
- "configured"를 "authed"로 취급 — credential은 만료 가능; 항상 live probe

---

## 최소 워크플로우 스켈레톤

1. **Dry-run** 모든 후보 프로바이더. 가용성 + 해결 힌트 출력.
2. **인증된 게 없으면 STOP.** 있으면 사용자에게 어떤 subset을 실행할지 묻는다.
3. **Judge 묻기** (judge 백엔드 가용 시만). 질문 안에 $ 비용 표시.
4. **실행** 선택된 프로바이더와 judge 플래그로. 출력 stream.
5. **요약**: 가장 빠름 / 가장 저렴 / 품질 최고 / 종합 최고 (judge skip 시 caveat).
6. **Baseline 저장 제안** (트렌드 추적용 opt-in JSON dump).
