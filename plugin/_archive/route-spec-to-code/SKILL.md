---
name: route-spec-to-code
description: "[ARCHIVE / 참조 전용] LLM이 직접 invoke 금지. command-routing 패턴이 흡수했으므로 historical reference로만 유지. 원본 의도: spec에서 signal을 추출해 curated template을 선택하는 pattern-routing concept 설명 — spec-to-code generator에서 free-form generation을 줄이는 참조 패턴."
type: skill
---

# Snippet: Spec → Code Pattern Routing

> **⚠️ ARCHIVE NOTICE**
>
> 이 스킬은 **참조 전용 라우팅 패턴 reference**입니다. LLM이 직접 invoke하지 마세요.
> 라우팅 책임은 `autoplan`(orchestrator)이 가집니다.
> `validate-idea` / `review-scope` / `critique-plan`은 `autoplan`의 단계로 호출되거나 standalone 으로도 사용 가능한 dual-mode stage skill 입니다.
> stage skill 단독 dispatch는 사용자가 명시 의도를 보일 때만 — orchestrator로 escalate 금지 (User Sovereignty).
>
> 본 본문은 향후 라우팅 정책을 재설계할 때의 historical reference로만 유지됩니다.

> Spec → code pattern-routing 컨셉 — 어떤 spec-to-code 생성기에도 재사용 가능한 generic shape.

## 이 snippet을 사용하는 경우
- "design spec → 동작 코드" 생성기 구축 (UI, schema, API client, IaC, slides, queries)
- 자유 생성(free-form generation)의 대안으로서 curated template 라이브러리
- 출력을 known-good 패턴의 작은 집합으로 제약해 AI 환각(hallucination) 감소
- 참신함보다 출력 품질이 중요한 모든 파이프라인

## 컨셉

파이프라인은 4단계, 순서대로:

```
spec 도착 → signal 추출 → catalog의 template과 매칭 → instantiate → output
```

1. **Spec 도착.** 읽기 가능한 정도로 구조화된 모든 것: 이미지, 문서, JSON schema, prose 계획, 사용자 요청. 형태는 중요치 않다 — signal을 추출할 수 있기만 하면 된다.
2. **Signal 추출.** template 간 구별에 도움이 되는 소수의 고정 feature를 뽑는다. Signal은 *계산 비용이 싸야* 하고 *라우팅에 load-bearing*이어야 한다. (나쁜 signal: "예쁜가?" 좋은 signal: "list-of-items semantics를 담고 있는가?")
3. **Catalog의 template과 매칭.** free-form prompt가 아니라 decision table. 각 행은 "signal이 X처럼 보이면 template Y 사용"을 말한다. Catalog는 작고, curated되고, known-good이다.
4. **Instantiate.** spec의 내용으로 template을 채운다. Template은 구조를 제약; spec은 slot을 채운다.

## Curated catalog가 free-form generation보다 나은 이유

| Dimension | Free-form generation | Curated catalog |
|-----------|---------------------|-----------------|
| 품질 하한선 | run마다 크게 다름 | 모든 출력이 "최소 template-품질" |
| Slop 리스크 | 높음 — 모델이 안티패턴을 만들어냄 | 낮음 — 검증한 패턴만 출하 |
| 속도 | 매 호출마다 full reasoning | 라우팅은 싸고, fill만 생성적 |
| 디버깅성 | "왜 저게 나왔지?" | "틀린 template 선택 → 라우팅 고침" 또는 "맞는 template, 나쁜 fill → template 고침" |
| 온보딩 | 모든 기여자가 재발명 | catalog는 팀이 축적한 taste |

비용은 curation: 누군가 catalog를 유지해야 한다. 그 비용은 두 번째 사용자가 생성기를 돌리는 순간부터 회수된다.

## 똑똑한 라우팅: 어떤 signal을 쓸까

라우팅 결정은 `signals → template` 함수다. 다음 속성을 가진 signal을 고른다:
- **Discriminative** — 서로 다른 template은 서로 다른 signal 값에 대응
- **계산이 쌈** — 무거운 모델 호출 없이 추출 가능
- **안정적** — 같은 spec은 재실행 시 같은 signal을 낸다

흔한 signal family (자기 도메인에 맞는 몇 가지 고르기):

| Signal family | 예시 값 | 구별 대상 |
|---------------|---------|----------|
| 구조적 형태 | flat / nested / repeating / single | list vs detail vs form vs dashboard |
| 밀도 | sparse / medium / dense | hero vs grid vs table |
| Interaction surface | read-only / editable / streaming | static page vs form vs feed |
| Content type | text-heavy / media-heavy / numeric | article vs gallery vs report |
| 제약 출처 | user-provided / inferred / default | template에 얼마나 엄격할지 |

Decision tree 형태 (sketch, 처방이 아님):

```
1. 명시적 사용자 제약이 있나? → 따르고 라우팅 skip
2. Signal A가 특화 template에 매칭? → 사용
3. Signal B가 특화 template에 매칭? → 사용
4. 그 외 → generic fallback template
```

선택된 template *과 그 이유*를 출력에 기재한다. 이렇게 하면 잘못된 라우팅 결정을 spec 재실행 없이 진단할 수 있다.

## 패턴 카탈로그 설계 — 자기 catalog 만들기

Catalog의 좋은 template은 다음을 갖는다:
- **이름**: 구현이 아니라 형태를 묘사
  (`list-with-detail`, `dashboard-grid`, `editorial-article` — `Template3`이 아니라).
- **Trigger 조건**: vibe가 아니라 signal 용어로 기술.
- **완전하고 실행 가능한 예시** — template 자체가 source of truth이고 주변 prose가 아니다. 기여자가 예시를 copy해서 작동시키지 못한다면 엔트리가 깨진 것.
- **Non-goal 목록** — 이 template이 *아닌* 것. 과도한 확장을 막고 라우터에 skip할 때를 알려준다.

Template 추가 시점:
- 같은 hand-roll 출력을 3번 이상 출하 → 승격.
- 이 slot에서 free-form generation이 known-bad 패턴을 계속 만들어냄 → 올바른 패턴을 template으로 추가하고 라우팅.

Template 은퇴 시점:
- 라우팅이 N회 연속 skip.
- Trigger 조건이 더 잘 작동하는 새 template과 겹침.
- 기반 기술이 이동 (deprecated API, framework 재작성) — 포팅하거나 삭제; 절반만 업데이트된 template을 catalog에 남기지 말 것.

Catalog 위생 규칙:
- 사람이 한 번에 다 읽을 수 있을 만큼 catalog를 작게 유지. 그 이상으로 커지면 라우터가 잘못된 일을 하는 것 — 도메인 단위 catalog로 분할.
- 각 template은 동작하는 최단 버전이어야 한다. Template은 매 호출마다 라우터가 읽는다; bloat는 token과 명료도 모두에 비용.
- Template을 prose가 아니라 code로 취급. 테스트. 버전. 변경 리뷰.
