---
name: write-changelog
description: "[패턴 라이브러리] version bump + CHANGELOG release-summary format + voice rules + user-facing change summary 작성 패턴. 직접 invoke보다 orchestrator가 import해 사용. 트리거: 'CHANGELOG 작성' / '릴리스 노트' / '버전 범프' / 'release 준비'. 참조 위치: critique-plan 후 ship 단계, sync-release-docs 페어, 배포 전 게이트."
type: skill
---

# Snippet: Version Bump + CHANGELOG Format


> 버전 범프 + CHANGELOG 포맷 + voice 규칙. 전체 ship 워크플로우(테스트 실행, diff 리뷰, PR 생성, 배포)는 별도 스킬에서.

## 이 snippet을 사용하는 경우
- 버전 관리 + CHANGELOG 규율이 필요한 자체 릴리스 워크플로우 구축
- 팀 간 CHANGELOG 포맷 표준화
- AI 보조 릴리스 저작 참조 (사람 + 에이전트가 엔트리 작성)

---

## 버전 범프 로직

Source-of-truth 파일: repo root의 `VERSION` 파일 (4자리 `MAJOR.MINOR.PATCH.MICRO`) + 선택적으로 동기화 유지돼야 할 `package.json`의 version 필드.

### Step 1: 기존 버전 감지 + 상태 분류

범프 전, 베이스 브랜치의 `VERSION` 그리고 `package.json.version`과 비교해 상태를 분류한다. 네 가지 상태:

| 상태 | 의미 | 액션 |
|-------|---------|--------|
| **FRESH** | `VERSION`이 베이스 브랜치와 일치, `package.json`이 동의 | 범프 수행 |
| **ALREADY_BUMPED** | `VERSION`이 베이스와 다름, `package.json`이 `VERSION`과 일치 | 범프 skip, 현재 버전 재사용 |
| **DRIFT_STALE_PKG** | `VERSION`이 베이스와 다름, 그러나 `package.json`이 stale | `package.json`만 sync, 재범프 금지 |
| **DRIFT_UNEXPECTED** | `VERSION`이 베이스와 일치하지만 `package.json`이 불일치 | STOP — 수동 reconcile 필요 |

```bash
BASE_VERSION=$(git show origin/<base>:VERSION 2>/dev/null | tr -d '\r\n[:space:]' || echo "0.0.0.0")
CURRENT_VERSION=$(cat VERSION 2>/dev/null | tr -d '\r\n[:space:]' || echo "0.0.0.0")
# node/python/프로젝트 도구로 package.json.version을 PKG_VERSION으로 읽기
# 그다음 위 네 상태로 dispatch
```

멱등성 체크가 중요: 같은 브랜치에서 릴리스 flow를 재실행해도 이중 범프나 VERSION과 package.json의 조용한 desync가 발생하면 안 된다.

### Step 2: 범프 레벨 결정 (MAJOR / MINOR / PATCH / MICRO)

Diff에서 auto-decide. MINOR 또는 MAJOR일 때만 사용자에게 질문.

- **MICRO** (4번째 자리): < 50줄 변경, 자잘한 tweak, 오타, config
- **PATCH** (3번째 자리): 50+ 줄 변경, feature 신호 없음
- **MINOR** (2번째 자리): 어떤 feature 신호라도 있으면 **사용자에게 질문** (새 라우트/페이지, 새 DB migration, 새 소스와 함께 새 테스트 파일, `feat/...` 이름 브랜치), OR 500+ 줄 변경, OR 새 모듈/패키지 추가
- **MAJOR** (1번째 자리): **사용자에게 질문** — 마일스톤이나 breaking change에만

한 자리를 범프하면 그 오른쪽 모든 자리는 0으로 리셋. 예: `0.19.1.0` + PATCH → `0.19.2.0`.

### Step 3: 검증 + 쓰기

```bash
if ! printf '%s' "$NEW_VERSION" | grep -qE '^[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+$'; then
  echo "ERROR: NEW_VERSION ($NEW_VERSION) does not match MAJOR.MINOR.PATCH.MICRO. Aborting."
  exit 1
fi
echo "$NEW_VERSION" > VERSION
# package.json 존재하면 package.json.version에도 쓰기
```

`VERSION` 쓰기 후 `package.json` 쓰기가 실패하면, **조용히 계속 금지**. 다음 ship 호출의 멱등성 체크가 drift를 감지하고 `DRIFT_STALE_PKG`로 라우팅해 복구한다.

### Step 4: 브랜치 스코핑 규칙 적용

- ship하는 모든 feature 브랜치는 자신의 버전 범프와 CHANGELOG 엔트리를 가진다
- 엔트리는 이 브랜치 vs 베이스 브랜치의 모든 commit을 커버
- **이미 베이스 브랜치에 landing한 이전 버전의 CHANGELOG 엔트리 편집 금지** — 위에 다시 범프
- 베이스 브랜치를 feature 브랜치로 merge한다고 베이스의 버전을 채택하는 것은 아니다. 베이스가 `v0.13.8.0`인데 브랜치가 feature를 추가하면 `v0.13.9.0`로 범프, 새 엔트리를 위에 작성.

### Step 5: 시퀀스 검증

엔트리 이동/추가/제거하는 모든 CHANGELOG 편집 후 즉시 실행:

```bash
grep "^## \[" CHANGELOG.md
```

Commit 전에 전체 버전 시퀀스가 연속이고 gap/중복 없는지 확인. 버전이 빠졌다면 편집이 뭔가 깬 것 — 진행 전에 수정.

---

## CHANGELOG Release-Summary 포맷

각 `## [X.Y.Z]` 엔트리는 두 부분: release-summary(prose, 업그레이드 결정하는 사람용)와 itemized changes list(정확히 무엇이 바뀌었는지 알아야 할 에이전트용).

### 구조 (모든 `## [X.Y.Z]` 엔트리)

1. **두 줄짜리 볼드 헤드라인** (영문 기준 총 10–14 단어). 마케팅이 아니라 판결처럼 단호하게 내려앉아야. 오늘 ship해서 실제로 동작하는지 신경 쓰는 사람처럼 들려야.
2. **Lead paragraph** (3–5 문장). 무엇이 ship됐고 사용자에게 무엇이 바뀌었는지. 구체적, concrete, no hype.
3. **"The X numbers that matter" 섹션**:
   - 숫자의 출처를 명명하는 짧은 setup paragraph (실제 프로덕션 배포 OR 재현 가능한 벤치마크 — 파일이나 실행 명령 명명).
   - **BEFORE / AFTER / Δ** 컬럼의 3–6개 주요 메트릭 테이블.
   - 카테고리별 breakdown용 선택적 두 번째 테이블.
   - 가장 두드러진 숫자를 구체적 사용자 용어로 해석하는 1–2 문장.
4. **"What this means for [audience]" 마무리 paragraph** (2–4 문장) 메트릭을 실제 워크플로우 변화로 연결. 무엇을 할지로 끝내기.

목표 길이: 요약 ~250–350 단어. 한 viewport에 렌더링돼야.

### Voice 규칙 (generic 글쓰기 원칙, 프로젝트별 아님)

- **Em dash 금지.** 대신 쉼표, 마침표, "..." 사용.
- **AI 어휘 금지.** 회피: delve, crucial, robust, comprehensive, nuanced, multifaceted, furthermore, moreover, additionally, pivotal, landscape, tapestry, underscore, foster, showcase, intricate, vibrant, fundamental, significant, interplay, leverage.
- **금지 구절 없음.** 회피: "here's the kicker", "here's the thing", "plot twist", "let me break this down", "the bottom line", "make no mistake", "can't stress this enough".
- **실제 숫자, 실제 파일 이름, 실제 명령.** "빠름"이 아니라 "30K 페이지에서 ~30s". "테스트 런타임 개선"이 아니라 "test:evals가 12분에서 4분".
- **짧은 paragraph.** 한 문장 punch를 2–3 문장 run과 섞기.
- **사용자 결과에 연결.** "개선된 정밀도"보다 "에이전트가 ~3배 덜 읽는다"가 낫다.
- **품질에 직설적.** "Well-designed" 또는 "this is a mess." 판단 주변에서 빙빙 돌지 말 것.
- **숫자 지어내지 말 것.** 메트릭이 벤치마크나 프로덕션 데이터에 없으면 포함 금지. 물어보면 "no measurement yet"라고 말하기.

---

## Itemized Changes 섹션

Release summary 아래 `### Itemized changes`를 쓰고 적용되는 subsection만 이어서 작성 (빈 섹션 skip):

### Added
- 새 기능, 역량, 명령, 엔드포인트

### Changed
- 기존 기능의 변경 (동작, 기본값, 출력 포맷)

### Fixed
- 버그 픽스 (버그당 한 bullet, 증상과 fix 명명)

### Removed
- Deprecated 또는 삭제된 기능

### For contributors
- 사용자에게 보이지 않는 내부 전용 변경 (빌드 infra, eval tooling, refactor). 사용자 대상 섹션이 사용자 가치에 집중하도록 분리 유지.

### 기여자 크레딧 규칙

커뮤니티 기여가 포함되면 기여자 명명: `Contributed by @username`. 항상 크레딧. 기여자는 실제 일을 했다 — 매번 공개적으로 감사, 예외 없음.

---

## 자동 생성 절차 (엔트리 작성하는 에이전트용)

1. `CHANGELOG.md` 헤더를 읽어 포맷 학습.
2. 브랜치의 모든 commit 열거:
   ```bash
   git log <base>..HEAD --oneline
   ```
   Commit 개수 세기. 체크리스트로 사용.
3. 전체 diff 읽어 각 commit이 실제로 무엇을 바꿨는지 이해:
   ```bash
   git diff <base>...HEAD
   ```
4. 쓰기 전 **테마별로 commit 그룹화**. 흔한 테마: 새 기능, performance, 버그 fix, dead code, infra/tooling, refactoring.
5. 모든 그룹을 커버하는 엔트리 작성. 파일 헤더 뒤, 오늘 날짜로 삽입: `## [X.Y.Z.W] - YYYY-MM-DD`.
6. **Cross-check:** 모든 commit이 최소 하나의 bullet에 매핑돼야. Commit 하나라도 미표현이면 지금 추가. 브랜치에 K 테마에 걸친 N commit이 있으면 엔트리가 모든 K 테마를 반영해야.
7. **사용자에게 변경 설명을 요청하지 말 것.** Diff와 commit 히스토리에서 추론.

---

## 편집 전 검증 체크리스트

모든 CHANGELOG 편집 후:

- [ ] 브랜치가 이미 베이스 브랜치에 있는 이전 엔트리와 분리된 자기 엔트리를 가진다
- [ ] `VERSION` 파일이 베이스 브랜치의 `VERSION`보다 높다
- [ ] `package.json.version`이 `VERSION` 파일과 일치 (drift 없음)
- [ ] 엔트리가 `CHANGELOG.md`의 최상위 엔트리 (베이스 브랜치 최신 위)
- [ ] `grep "^## \[" CHANGELOG.md`가 gap/중복 없는 연속 버전 시퀀스 표시
- [ ] 브랜치의 모든 commit이 새 엔트리의 최소 하나 bullet에 매핑
- [ ] Release summary가 4개 구조적 파트(헤드라인, lead, 숫자 테이블, "what this means" 마무리) 보유
- [ ] Em dash 없음, 금지 AI 어휘 없음, 금지 구절 없음
- [ ] 해당 시 커뮤니티 기여자가 `Contributed by @username`으로 크레딧
