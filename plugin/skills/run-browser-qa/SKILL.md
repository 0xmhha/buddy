---
name: run-browser-qa
description: "[패턴 라이브러리] browser automation QA 패턴 — snapshot diff, form testing, responsive checks, dialog handling, accessibility-tree interaction. 직접 invoke보다 orchestrator가 import해 사용. 트리거: '브라우저 QA' / 'playwright 패턴' / 'E2E QA' / '스냅샷 diff'. 참조 위치: classify-qa-tiers Exhaustive tier, 배포 전 검증, audit-live-devex."
type: skill
---

# Snippet: Browser QA Workflow Patterns


> 브라우저 자동화 워크플로우 패턴 — 도구 무관 (Playwright, Puppeteer, Selenium, browse 등에 적용).

## 이 snippet을 사용하는 경우
- Playwright / Puppeteer / Selenium 등 위에 프로젝트별 QA 스킬 구축
- "좋은" QA 워크플로우가 어떤 모양인지 참조
- 팀 간 스크린샷 증거 패턴 표준화

## 패턴 1: Snapshot Diff (Before/After 검증)

### 워크플로우
```
1. snapshot_before = take_snapshot(page)
2. perform_action(page, ...)        # click, fill, navigate, etc.
3. snapshot_after = take_snapshot(page)
4. diff = compute_diff(before, after)
5. assert(diff matches expected_changes)
```

### 이유
*의도치 않은* 사이드 이펙트 캐치. 버튼 클릭 시 1개만 바뀔 거라고 기대했는데 50개가 바뀌면 diff가 노출한다.

### Snapshot 타입이 중요
- **DOM snapshot**: 구조적 변경 (element 추가/제거)
- **Accessibility tree**: 의미적 변경 (보조 기술에 가시)
- **Visual snapshot (스크린샷)**: 픽셀 변경 (CSS 회귀 캐치)
- **Network requests**: 사이드 이펙트 변경 (새 API 호출 등)

Assertion에 맞는 snapshot 타입 사용.

## 패턴 2: Accessibility Tree → @ref 상호작용

### IP
대부분의 브라우저 자동화 도구는 CSS selector나 XPath로 element를 선택하며, 이 둘은 페이지가 변경되면 깨진다. 대신 **accessibility tree** 사용:

1. Accessibility-tree snapshot — `{role, name, ref}` 노드 트리 생성
2. `role` + `name`으로 타겟 element 찾기 (의미적, 구조적 아님)
3. 그것의 `@ref` 획득 (안정적 불투명 식별자)
4. `@ref`에 대해 action 수행

### 일반화 이유
- Accessibility tree를 노출하는 모든 브라우저 자동화 도구에서 동작
- CSS class rename, 레이아웃 shift, 라이브러리 업그레이드에 탄력적
- 스크린 리더와 AI 에이전트가 페이지를 보는 방식 거울

### Pseudocode
```
snapshot = page.accessibility_tree()
btn_ref = snapshot.find(role="button", name="Submit").ref
page.click(btn_ref)
```

## 패턴 3: 스크린샷 증거를 위한 Annotation

버그 리포트 시, annotated 스크린샷이 raw보다 10배 유용.

### Annotation 타입
- **Box outline**: 문제 element 강조
- **Arrow**: 작은 디테일 가리키기
- **Text caption**: 무엇이 잘못인지 설명 (스크린샷 아래)
- **Before/after side-by-side**: 회귀 표시

### 워크플로우
```
1. screenshot = take_screenshot(page)
2. element_box = get_bounding_box(target_element)
3. annotated = draw_box(screenshot, element_box, color="red")
4. annotated = draw_caption(annotated, "Error message overflows container")
5. save(annotated, f"bug-{date}-{slug}.png")
```

## 패턴 4: 반응형 테스팅

### 테스트할 표준 breakpoint
- Mobile: 375x667 (iPhone SE)
- Tablet: 768x1024 (iPad)
- Desktop: 1440x900
- Wide: 1920x1080

### 워크플로우
```
for breakpoint in [(375, 667), (768, 1024), (1440, 900), (1920, 1080)]:
  page.set_viewport(*breakpoint)
  page.reload()
  screenshot = take_screenshot(page)
  save(screenshot, f"responsive-{breakpoint[0]}x{breakpoint[1]}.png")
  check_no_horizontal_scroll(page)
  check_clickable_elements_visible(page)
```

### Breakpoint별 흔한 버그
- Mobile: 탭 타겟 < 44x44px
- Tablet: 레이아웃이 예상외로 붕괴
- Desktop: max-width 누락으로 지나치게 넓은 텍스트
- Wide: 이미지 스케일링 이슈

## 패턴 5: 폼 제출

### 테스트할 서브 패턴
1. **Happy path**: 유효 입력 → 성공
2. **Validation**: 각 필드의 validation 룰
3. **Error states**: 서버 측 에러가 올바로 렌더링
4. **Empty submission**: 필수 필드 강제
5. **Edge values**: 경계 조건 (max length, 특수 문자)

### 폼당 워크플로우
```
1. snapshot before
2. fill all fields
3. click submit
4. wait for response
5. snapshot after
6. assert success indicator OR error indicator visible
7. assert URL changed (if expected)
```

## 패턴 6: 파일 업로드

```
1. accessibility tree로 file input 찾기
2. set_files(input_ref, [path/to/test-file.pdf])
3. wait for upload progress (often XHR or fetch)
4. assert file appears in UI (filename visible)
5. assert no error
```

## 패턴 7: Dialog 처리 (Alert / Confirm / Prompt)

### Trigger 전 항상 handler 등록
```
page.on("dialog", lambda d: d.accept(text="user input"))
page.click(button_that_opens_dialog)
```

### 왜 before?
Dialog는 blocking. Handler가 준비 안 됐으면 클릭이 hang.

### 두 경로 모두 테스트
- accept → post-accept 동작 assert
- dismiss → 액션 수행 안 됨 assert

## 패턴 8: 멀티 탭 테스팅

```
context = browser.new_context()
page1 = context.new_page()
page2 = context.new_page()

page1.goto("/")
page2.goto("/profile")

# Cross-tab 동작 테스트:
# - page1에서 login, page2 refresh → 여기도 로그인돼야 (session sync)
# - page1에서 logout, page2가 감지해야 (BroadcastChannel 또는 storage event)
```

## 안티패턴

- CSS selector / XPath를 주 element 식별자로 (깨지기 쉬움)
- 조건 대기 대신 고정 대기 시간(`sleep 2`)
- 단일 breakpoint 테스팅 (desktop 통과, mobile 실패)
- Annotation 없는 raw 스크린샷 (리뷰어가 뭐가 잘못인지 모름)
- Accessibility tree 무시 (나중에 a11y 테스트 추가할 때 다시 빌드하게 됨)
