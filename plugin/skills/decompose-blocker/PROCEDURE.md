# Decompose Blocker — 막힘 상태 해소를 위한 문제 분해 + 행동 후보 도출


코드 작업 중 *stuck 상태*(어디서 봐야 할지 entry point 불분명 / 가설 안 떠오름 / 옵션 사이 선택 불가)일 때, 문제를 *기계적으로 분해*하여 *비용-정보 매트릭스* 기반의 행동 후보를 도출하는 절차. 직진 시도하다 막힌 상태에서 *한 발 물러서 탐색 공간 자체를 정리*하는 것이 핵심.

핵심 차별점:
- `diagnose-bug`이 *bug 증상이 명확해진 후*의 root cause 분석이라면, `decompose-blocker`은 *증상조차 모호한* stuck 상태에서 *어디부터 볼지*를 결정한다 — 진단 진입 *전* 단계.
- `verify-best-alternative`가 *이미 옵션 후보가 보이는* 결정의 다관점 검토라면, `decompose-blocker`은 *옵션이 안 보이는* stuck에서 *후보를 도출*한다.
- `critique-plan`이 *plan이 있을 때*의 strategic 비판이라면, `decompose-blocker`은 *plan을 만들기 전*의 entry point 결정.
- `concretize-idea`(§1)가 *새 제품 아이디어*의 PRD화라면, `decompose-blocker`은 *기존 작업 중 막힘*의 ad-hoc 해소.

**용어 안내**:

| 용어 | 정의 |
|------|------|
| **Stuck** | 다음 행동이 *결정 불가*한 상태. 두 종류: *too-wide*(탐색공간 너무 넓음) / *too-narrow*(선택지가 안 보임). |
| **분해 축** | 문제 공간을 나누는 차원 (기본 4D = WHERE/WHEN/WHAT/WHY, 도메인에 따라 binary search, 5 Whys, fishbone 등 채택 가능). |
| **모름의 지도** | 사용자가 *답할 수 없다*고 표명한 차원의 목록. dead-end가 아니라 *분해의 한 결과* — "여기 정보 없음"이 곧 첫 행동 후보의 단서. |
| **비용-정보 매트릭스** | 각 행동 후보에 (수행 비용 × 얻는 정보량)을 점수화. 저비용·고정보 우선. |
| **행동 단위** | 행동 후보의 종류 — `information-gathering`(로그/grep/실측) / `hypothesis-test`(가설 검증 실험) / `code-modification`(수정). 한 사이클에 한 단위로 통일 권장. |

## 1. 목적

stuck 상태인 사용자(또는 AI 자체)가 *진단·결정·구현* 어느 단계로든 진입할 entry point를 찾도록 돕는다. 본 스킬은 **분해**(문제 공간 정리)와 **행동 후보 도출**(다음 1~3 행동) 두 가지를 산출한다 — *직접 fix 수행은 하지 않는다*.

## 2. 사용 시점

다음 상황에서 호출하라:

- 사용자가 "어디서 봐야 할지 모르겠어", "어떻게 시작해야 할지 모르겠어"와 같이 *entry point 불명*을 표명
- AI 자체가 코드 분석 중 *너무 많은 가능성*에 직면해 다음 행동 결정 불가
- 선택지가 *0개로 보임* — 옵션 도출 자체부터 필요
- 가설이 다수 떠올랐는데 *어떤 가설을 먼저 검증할지* 우선순위 불명
- 시간 압박은 있는데 *직진 시도가 막힘*

다음 상황에서는 **호출하지 마라**:

- bug repro 가능 + 증상 명확 → `diagnose-bug`
- 옵션 후보 2~5개 있고 비교 필요 → `verify-best-alternative`
- 작성된 plan/spec critique → `critique-plan`
- 새 기능 idea의 PRD화 → `concretize-idea`(§1)
- 작은 코드 변경 + 방향 명확 → 그냥 진행, decompose-blocker는 *과한 ceremony*

## 3. 입력

### 필수 입력

- **사용자 발화 (stuck 상태 묘사)** — 가급적 그대로 인용. paraphrase 금지.
- **컨텍스트 식별** — 어느 코드/시스템/문제 도메인. 한 줄로.

### 선택 입력

- **이전 시도** — 이미 해봤지만 실패/막힘에 도달한 행동들
- **시간 압박** — 시급도 (이번 시간 / 오늘 / 이번 주 / 무관)
- **권한·도구 제약** — production access, 인프라 권한 등 (행동 후보 가능성에 영향)

### 입력 부족 시 forcing question (최대 3개)

다음 중 답이 모호하면 사용자에게 *3개 이하*로 묻는다:

- "stuck의 정체가 *탐색공간 too wide*(어디부터 볼지 모름)인지 *선택지 too narrow*(옵션 안 보임)인지?"
- "원하는 행동 후보 단위가 *정보수집*(로그/grep)인지 *가설검증*(실험)인지 *수정*(코드 변경)인지?"
- "이 사이클 종료 시점이 *행동 후보 제시*까지인지 *실행 후 follow-up*까지인지?"

3개 답을 모두 못 하더라도 *default 가정*으로 진행 가능 (§4 원칙 8 참조).

## 4. 핵심 원칙 (Principles)

1. **Fact / 추측 / 모름 3분류** — 사용자 발화에서 *반드시 분리*. "504 timeout이 난다"(fact)와 "코드 문제 같다"(추측)와 "어느 endpoint인지 모름"(모름)을 같은 단락에 섞지 말 것.

2. **분해 축 명시 선언** — 4D(WHERE/WHEN/WHAT/WHY) 기본. 도메인에 따라 *binary search of code path*, *5 Whys*, *fishbone* 등 선택 가능. 무엇을 선택했는지 *반드시 명시*.

3. **모름의 지도가 1급 시민** — 사용자가 "모름"이라 답한 차원은 *dead-end가 아니라* 첫 행동 후보의 단서. "여기 정보 없음"이 곧 "정보 수집 행동이 필요"의 신호.

4. **비용-정보 매트릭스 강제** — 모든 행동 후보에 *비용*(시간/리스크)과 *정보 yield*를 점수화. 저비용·고정보 우선. 비용 미평가 후보는 사용자가 선택 못 함 → 제시 금지.

5. **사용자 질문 ≤ 3** — 한 사이클당 최대 3개. *저비용·high-information* 축부터. 더 필요하면 다음 사이클로 분리.

6. **행동 후보 단위 통일** — 한 사이클에 *한 단위*(information-gathering OR hypothesis-test OR code-modification). 단위 섞으면 사용자가 비교 못 함.

7. **stuck 종류 분리** — *too-wide*(탐색공간 too wide)와 *too-narrow*(선택지 too narrow)는 *반대 문제*. 같은 절차로 풀면 둘 중 하나에 부적합 — 분리 처리.

8. **default 가정 + 명시** — 입력이 모호해도 default 가정으로 진행. 단, *어떤 가정을 했는지* 산출물에 명시 — 사용자가 가정에 동의 안 하면 재시작.

9. **가설 압축 후 행동** — Step 6에서 모인 정보로 가설을 *3개 이하로 압축*. 압축 안 하고 모든 가설을 다 검증하라 권유 금지.

10. **종료 조건 default = 행동 후보 제시** — 한 사이클의 끝은 *사용자가 다음 행동을 선택 가능한 상태*. 실행·follow-up은 별도 호출.

## 5. 실행 단계 (Steps)

### Step 1. Stuck 정체 식별

사용자 발화 직후, *stuck의 종류*를 결정한다:

- **too-wide**: "어디부터 볼지 모르겠어" / "옵션이 너무 많아" → 탐색공간 압축이 우선
- **too-narrow**: "아무것도 안 떠올라" / "다 시도해봤어" → 옵션 발산이 우선
- **mixed**: 둘 다 — 우선 *too-wide*를 먼저 (압축 후 발산이 더 효율)

판정 모호 시 forcing question 1번으로 확정.

### Step 2. Fact / 추측 / 모름 3분류

사용자 발화에서 추출:

```
Fact (확신도 None): ...        # 사용자가 *관찰한* 것
추측 (확신도 Low~Mid): ...     # 사용자가 *유추한* 것
모름: ...                      # 사용자가 *답할 수 없다*고 표명한 것
```

원문 표현 그대로 인용. paraphrase는 의미 손실을 만든다.

### Step 3. 분해 축 결정

도메인을 보고 분해 축 선택:

| 도메인 | 권장 축 |
|--------|---------|
| 시스템·서버·인프라 문제 | 4D (WHERE/WHEN/WHAT/WHY) |
| 코드 path·로직 문제 | binary search of code path |
| 비즈니스·UX·근본 원인 | 5 Whys |
| 다요인 복합 문제 | fishbone (Cause-and-Effect) |
| 옵션 선택 stuck | trade-off matrix (cost × value × risk) |

축은 *2개 이상 혼용 가능*. 선택 근거는 산출물에 명시.

### Step 4. 모름의 지도 (Known Unknowns)

각 분해 축에 사용자가 *답할 수 있는지* 표기:

```
| 축 | 답 가능성 | 비용 |
|----|----------|------|
| WHERE | △ (로그 검색 필요) | 중 |
| WHEN | ○ (대시보드 즉시) | 저 |
| WHAT | ○ (이미 확정) | 0 |
| WHY | × (지금은 불가) | — |
```

저비용·답 가능 축부터 압축한다.

### Step 5. 사용자 질문 (≤ 3)

Step 4 결과를 보고 사용자에게 *3개 이하* 질문. 각 질문은:

- 한 축에 집중
- *왜 이 질문이 필요한지* 한 줄 근거 동반
- "모름"도 valid answer임을 명시

질문 구조 예시:
```
**Q1 (WHEN)** 최근 1주일 중 [증상]이 *몰린 시간대*가 있나요? (모니터링 대시보드 한 번 보면 됩니다)
**Q2 (WHERE-거시)** [증상]이 모든 endpoint에서 비슷하게 나나요, 아니면 특정 경로 의심?
**Q3 (인프라 경계)** [관련 인프라 컴포넌트]가 있나요? 누가 [증상]을 내고 있는지 확인 위함.
```

### Step 6. 응답 반영 → 가설 압축

사용자 응답으로 분해 공간을 압축한다:

- 답 가능 축 → 사실 확정
- 모름 축 → 다음 사이클 후보 또는 즉시 정보수집 행동 후보
- *3개 이하 가설*로 압축. 각 가설에 *증거*와 *반증 방법* 1줄씩.

```
가설 H1 (확신도 Mid-High): [statement]
  증거: [사용자 응답에서 어느 부분]
  반증: [어떤 정보가 있으면 H1을 버릴 수 있는가]
```

### Step 7. 행동 후보 도출 (비용-정보 매트릭스)

가설별로 *검증 비용·정보 yield* 매트릭스 작성:

```
| # | 행동 | 단위 | 비용 | 얻는 정보 | 우선 |
|---|------|------|------|----------|------|
| A1 | [구체 행동] | info-gather | 저 (15분) | WHERE+WHEN 확정 | ★ |
| A2 | [구체 행동] | hypothesis-test | 중 (배포 1회) | H1 검증 | A1 후 |
```

행동 단위는 *§3 입력의 user 선택* 또는 *default information-gathering* 통일. 단위 섞지 말 것.

### Step 8. 사용자 confirm + 다음 행동 선택

사용자에게 *상위 2~3 행동*을 비용 순으로 제시, *다음 1개를 선택*하도록 요청. 종료 조건:

- 사용자가 행동 선택 → 본 사이클 완료
- 사용자가 "이 후보들 다 별로다" → Step 3 회귀 (다른 분해 축 시도)
- 가설 압축이 충분히 안 됨 → 다음 사이클 (Step 5부터)

본 사이클은 *행동 후보 제시*에서 끝. 실행은 *호출자*가 다음 스킬(`diagnose-bug` / `iterate-fix-verify` / `build-with-tdd` 등)로 dispatch.

## 6. 출력 템플릿

다음 yaml 구조로 결과를 호출자에게 반환:

```yaml
stuck_type: too-wide | too-narrow | mixed
domain_context: "<1줄>"

input_classification:
  facts:
    - "<원문 인용>"
  guesses:
    - statement: "<원문 인용>"
      confidence: low | mid
  unknowns:
    - dimension: WHERE | WHEN | WHAT | WHY | ...
      user_can_answer: yes | no | with-effort
      answer_cost: low | mid | high

decomposition:
  axes_chosen: ["4D", "binary-search", "5-whys"]  # 1개 이상
  rationale: "<도메인 근거 1줄>"

user_questions_asked:
  - id: Q1
    axis: <axis name>
    question: "<원문>"
    rationale: "<왜 이 질문>"
  # 최대 3개

hypotheses_after_questions:
  - id: H1
    statement: "<...>"
    confidence: low | mid | high
    evidence: "<응답 어느 부분>"
    disproof: "<반증 방법>"
  # 최대 3개

action_candidates:
  unit: information-gathering | hypothesis-test | code-modification
  candidates:
    - id: A1
      action: "<구체>"
      cost: low | mid | high
      info_yield: low | mid | high
      priority: 1
  # 비용 정렬

user_selected_action: <action_id> | "rejected_all" | "needs_more_decomposition"
next_skill_dispatch: "diagnose-bug" | "iterate-fix-verify" | "build-with-tdd" | "verify-best-alternative" | null

default_assumptions:
  - "<입력 모호 시 가정한 default>"
```

## 7. 자매 스킬

### 앞 단계 (선행 스킬)

- (진입점) — 사용자가 stuck 상태를 표명하면 직접 호출
- `concretize-idea`(§1) — 새 기능 idea level의 모호함이면 §1 stage로 회귀

### 페어 (동시 동작 가능)

- `consult-codex` — Step 6 가설 압축에서 *외부 second opinion* 필요 시
- `verify-best-alternative` — Step 7 행동 후보가 *상호 배타적 alternatives*라면 호출

### 후속 단계 (다음 스킬)

사용자가 선택한 행동 단위에 따라:

- **information-gathering 선택** → 사용자가 *직접 실행* (로그/grep). 결과 모이면 본 스킬 재호출 또는 `diagnose-bug` dispatch
- **hypothesis-test 선택** → `diagnose-bug`(가설 검증 phase) 호출
- **code-modification 선택** → `iterate-fix-verify` 또는 `build-with-tdd` 호출
- *모든 후보 거절* → Step 3 회귀 (다른 분해 축)

### 호출 흐름 예시

```
사용자: "504 timeout 가끔 나는데 어디부터 봐야 할지 모르겠어"
    → decompose-blocker
        → Step 1 (too-wide 판정)
        → Step 2 (fact/추측/모름 3분류)
        → Step 3 (4D 분해 채택)
        → Step 4 (모름의 지도)
        → Step 5 (3개 질문)
        → Step 6 (응답 반영 → H1, H2 압축)
        → Step 7 (information-gathering 단위, A1·A2·A3 후보)
        → Step 8 (사용자 A1 선택)
    → 사용자가 nginx access.log 검색 (직접)
    → 결과 가지고 diagnose-bug 호출 또는 decompose-blocker 재호출
```

## 8. Anti-patterns

다음은 본 스킬 적용 중 자주 나타나는 안티패턴과 교정 방법이다.

1. **분해 축 prior 미선언** — "그냥 떠오르는 대로 질문". 매 호출마다 절차가 달라져 재현성 0. 교정: Step 3에서 *반드시* 축 선택 + 근거 명시.

2. **행동 후보 단위 혼용** — A1은 로그 검색, A2는 코드 수정, A3은 가설 검증이 한 표에 섞임. 사용자가 비교 못 함. 교정: Step 7에서 *한 단위로 통일*. 단위 변경 원하면 별도 사이클.

3. **사용자 질문 무한정** — Step 5에서 5개·10개 던짐. 사용자 피로 + 분해 부족. 교정: *3개 상한*. 더 필요하면 다음 사이클.

4. **stuck 종류 구분 없이 동일 절차** — too-wide와 too-narrow를 같은 4D 분해로 풀면, too-narrow는 *발산이 필요*한데 *압축* 시도해 dead-end. 교정: Step 1에서 *명시적 분기*. too-narrow면 `verify-best-alternative`처럼 다양성 강제 모드로 전환.

5. **종료 조건 모호** — 한 사이클이 끝났는지 사용자가 알 수 없음. "또 뭐 필요해?" 무한 반복. 교정: Step 8에서 *행동 후보 제시 = 종료* default 명시. follow-up 원하면 사용자가 명시 요청.

6. **Fact·추측·모름 미분리** — 사용자 발화를 *통째로* 다음 step에 던짐. AI가 추측을 fact로 오해 → 잘못된 가설로 직진. 교정: Step 2에서 *반드시* 3분류 표 작성.

7. **비용 미평가 후보 제시** — "A, B, C 중 골라봐"만 던지고 *각각의 시간/리스크* 미명시. 사용자가 선택 정보 부족. 교정: Step 7 매트릭스에 *반드시* 비용·정보 yield 둘 다 점수화.

8. **가설 압축 없이 첫 후보 직진** — Step 6에서 10개 가설을 다 검증하라 권유. 사용자 정보 과부하. 교정: *3개 이하 압축* + 각 가설에 *증거·반증* 명시.

9. **모름을 dead-end 처리** — 사용자가 "모름" 답한 차원을 *무시*하고 다른 축으로만 진행. 모름이 종종 *가장 중요한 첫 행동의 단서*. 교정: 모름의 지도를 *1급 시민*으로 — Step 7 행동 후보의 우선순위 1번이 종종 "모름 차원 정보 수집".

10. **default 가정 미명시** — 입력 모호한데 가정을 *조용히 적용*. 사용자가 가정에 동의 안 하면 전체 사이클 무용. 교정: 출력 yaml의 `default_assumptions` 필드에 *반드시* 명시 — 사용자가 거절하면 재시작.

## 9. 체크리스트 (Step별 자가 점검)

각 Step 종료 시:

- [ ] Step 1: stuck 종류(too-wide/too-narrow/mixed) 명시했는가?
- [ ] Step 2: fact/추측/모름 3분류 표 작성했는가?
- [ ] Step 3: 분해 축 선택 + 근거 명시했는가?
- [ ] Step 4: 모름의 지도(축별 답 가능성·비용)를 표로 작성했는가?
- [ ] Step 5: 사용자 질문이 *3개 이하*인가? 각 질문에 근거 동반했는가?
- [ ] Step 6: 가설을 *3개 이하*로 압축했는가? 각 가설에 증거·반증 명시?
- [ ] Step 7: 행동 후보의 *단위 통일*? 비용·정보 yield 둘 다 점수화?
- [ ] Step 8: 종료 조건(사용자가 행동 선택 또는 회귀)을 명시적으로 도달?
- [ ] 출력 yaml의 `default_assumptions` 필드에 입력 모호 시 가정 명시?

9개 중 하나라도 No이면 종료 선언 금지.

## 10. 종료 조건

decompose-blocker 호출이 종료되는 조건:

- 출력 yaml이 호출자에게 반환됨 (모든 필수 필드 채워짐)
- 사용자가 *행동 후보 1개를 선택* OR *모든 후보 거절 + Step 3 회귀 결정* 중 하나
- §9 체크리스트 9개 항목 모두 충족

종료 후 호출자는 사용자 선택에 따라 `diagnose-bug` / `iterate-fix-verify` / `build-with-tdd` / `verify-best-alternative` 중 하나로 dispatch한다. 본 스킬 자체는 *진단·수정을 수행하지 않는다*.
