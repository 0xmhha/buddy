---
name: classify-review-risks
description: "[패턴 라이브러리] structural code review에서 반복적으로 놓치는 risk 11 category 분류 (SQL safety, LLM trust boundary, conditional side effects 등). anti-pattern scan checklist. 직접 invoke보다 orchestrator가 import해 사용. 트리거: '리뷰 리스크 분류' / '리뷰 risk 11개' / 'category 매핑' / '리뷰 카테고리'. 참조 위치: critique-plan §11 리뷰 섹션, review-engineering, review-architecture."
type: skill
---

# Snippet: 코드 리뷰 안티패턴 카테고리


> 구조적 카테고리 taxonomy + diff scan flow. ship 통합, 전문 subagent, adversarial pass, 리포트 작성은 별도 스킬에서.

## 이 snippet을 사용하는 경우
- 자체 pre-merge 리뷰 스킬 구축
- 수동 코드 리뷰 시 참조 체크리스트
- CI hook에 카테고리 기반 스캔 추가

## 카테고리

### 카테고리 1: SQL & 데이터 안전성

**살펴볼 것:**
- SQL에 string interpolation — `.to_i`/`.to_f`도 포함. 파라미터화 쿼리 사용.
- TOCTOU race: 원자적 `WHERE` + `update_all`이어야 할 check-then-set 패턴.
- 직접 DB 쓰기에서 model validation 우회 (Rails `update_column`, Django `QuerySet.update()`, Prisma raw query).
- N+1 쿼리: 루프/뷰에서 eager loading (`.includes()`, `joinedload()`, `include`) 누락.

**Bad:**
```ruby
User.where("id = #{params[:id]}")              # interpolation
order.update_column(:status, "paid")           # validation 우회
posts.each { |p| puts p.author.name }          # N+1
```

**Good:**
```ruby
User.where(id: params[:id])                    # parameterized
order.update!(status: "paid")                  # validated
Post.includes(:author).each { |p| ... }        # eager loaded
```

**Grep 패턴:**
- `where\(["'].*#\{` — `where`의 string interpolation
- `update_column|update_all|raw\(` — validation 우회
- `\.each.*\.\w+\.\w+` — 루프의 N+1 가능성

---

### 카테고리 2: LLM 출력 신뢰 경계

**살펴볼 것:**
- LLM 생성 값 (이메일, URL, 이름)이 포맷 검증 없이 DB 쓰기나 mailer 전달. 지속 전에 `EMAIL_REGEXP`, `URI.parse`, `.strip` 가드 추가.
- 구조화된 도구 출력 (배열, 해시)이 DB 쓰기 전 타입/shape 체크 없이 수용됨.
- LLM 생성 URL이 allowlist 없이 fetch — URL이 내부 네트워크를 가리키면 SSRF 위험. Hostname parse, blocklist(`169.254.*`, `10.*`, `127.*`, `localhost`) 체크.
- LLM 출력이 sanitization 없이 지식 베이스나 vector DB에 저장 — stored prompt injection 위험.

**Bad:**
```python
email = llm_response["email"]
User.create(email=email)                       # 포맷 체크 없음
requests.get(llm_response["url"])              # SSRF 위험
```

**Good:**
```python
email = llm_response["email"].strip()
if not EMAIL_RE.match(email): raise InvalidEmail
host = urlparse(url).hostname
if host in BLOCKLIST: raise UnsafeURL
```

**Grep 패턴:**
- `llm|claude|openai|gpt.*\.(create|insert|save|update)`
- `requests\.get|httpx\.get|fetch\(.*llm|.*ai_response`

---

### 카테고리 3: 조건부 사이드 이펙트

**사이드 이펙트에 해당하는 것:** DB 쓰기, 네트워크 호출, 파일 I/O, signal send, 캐시 invalidation, 함수 밖에서 관찰 가능한 모든 것.

**흔한 gotcha:**
- 조건이 race 가능한(read-check-write) 조건문 내 사이드 이펙트.
- 고유 DB index 없는 find-or-create — 동시 호출이 중복 생성.
- 원자적 `WHERE old_status = ? UPDATE SET new_status` 안 쓰는 상태 전이 — 동시 업데이트가 skip되거나 이중 적용.
- 사용자 제어 데이터에 안전하지 않은 HTML 렌더링 (`html_safe`, `dangerouslySetInnerHTML`, `v-html`, `|safe`) (XSS).

**Bad:**
```ruby
user = User.find_by(email: e) || User.create!(email: e)   # race: dup rows
order.update!(status: "paid") if order.pending?           # race: double charge
```

**Good:**
```ruby
User.find_or_create_by!(email: e)                         # email에 고유 index로
Order.where(id: id, status: "pending").update_all(status: "paid")  # atomic
```

---

### 카테고리 4: Shell & Code Injection

**살펴볼 것:**
- `shell=True` + f-string/`.format()` interpolation을 함께 쓰는 Python 서브프로세스 호출. 인자 배열 사용.
- 변수 interpolation 있는 `os.system()`. 배열 인자 기반 호출로 교체.
- 샌드박스 없이 LLM 생성 또는 사용자 제공 코드의 `eval()` / `exec()`.
- Node의 shell-style child process 호출 + template literal — `execFile(cmd, [args])` 스타일 선호.

**Bad:** `subprocess.run(f"ls {path}", shell=True)`
**Good:** `subprocess.run(["ls", path])`

**Grep 패턴:** `shell=True`, `os\.system\(`, `\beval\(|\bexec\(`, `child_process` + template literal

---

### 카테고리 5: Enum & 값 완전성

Diff가 새 enum 값, status string, tier 이름, type 상수를 도입하면 — 이게 **diff 밖 코드를 읽어야 하는** 유일한 카테고리.

**절차:**
1. **모든 consumer를 통해 새 값 trace.** Sibling 값 grep (예: tier에 `revise` 추가 → `quick`, `lfg`, `mega` grep). 각 매치 read.
2. **Allowlist / 필터 배열 체크.** `%w[quick lfg mega]` 리스트는 업데이트돼야.
3. **`case`/`if-elsif` 체인 체크.** 새 값이 잘못된 default로 fall through하나?
4. **흔한 miss:** 프론트엔드 dropdown에 추가됐지만 백엔드 모델/compute 메서드가 지속 안 함.

---

### 카테고리 6: 경계의 타입 강제 변환

- 언어 경계 (Ruby→JSON→JS, Python→JSON→TS) 넘나드는 값에서 numeric vs string이 flip 가능 — hash/digest 입력은 normalize 필수.
- 직렬화 전 `.to_s` 안 하는 해시 입력: `{ cores: 8 }`과 `{ cores: "8" }`은 다른 해시 생성.

---

## Diff 리뷰 Flow

### Step 1: Fresh base fetch
```bash
git fetch origin <base> --quiet
git diff origin/<base>
```
Stale 로컬 base는 false positive 생성. 항상 먼저 fetch.

### Step 2: 크리티컬 pass
카테고리 1–5를 모든 변경 hunk에 실행. Blocking.

### Step 3: Informational pass
카테고리 6 + async/sync 혼용, 시간 윈도우 안전성, view/프론트엔드 perf. Non-blocking 하지만 flag.

### Step 4: Confidence calibration
모든 finding은 1–10 confidence 점수:

| 점수 | 의미 | 표시 |
|-------|---------|---------|
| 9–10 | 코드 읽어 확인; 구체적 버그 시연 | 정상 표시 |
| 7–8 | 고 confidence 패턴 매치 | 정상 표시 |
| 5–6 | False positive 가능 | "Medium confidence, verify" caveat와 함께 |
| 3–4 | 의심되지만 괜찮을 수 | 부록으로 억제 |
| 1–2 | 추측 | 심각도가 P0일 때만 |

**포맷:** `[SEVERITY] (confidence: N/10) file:line — description`

### Step 5: Enum cross-reference
새 enum 값에 대해 sibling grep, consumer 읽기 (카테고리 5의 Step 5).

---

## Pass/Fail 게이트

**PASS:** Confidence ≥ 7의 CRITICAL finding 없음. Informational finding은 허용.

**FAIL:** 다음 중 하나라도:
- 카테고리 1–5의 CRITICAL finding (confidence ≥ 7)
- DB 쓰기나 네트워크 호출하는 LLM 신뢰 경계 위반
- Unhandled consumer 있는 enum 추가
- Confidence ≥ 5인 SQL injection 또는 shell injection

**ASK_HUMAN (escalate):**
- Confidence 5–6의 CRITICAL finding (false positive 가능)
- Fix가 사용자 가시 동작을 바꾸는 race condition
- 새 allowlist 필요한 신뢰 경계 fix (누가 리스트 소유?)
- 5개 초과 파일을 건드리는 enum 완전성 fix

---

## 출력 템플릿

```
## Review Result: [PASS / FAIL / ASK_HUMAN]

### SQL & Data Safety: [PASS/FAIL/N/A]
- [file:line] (confidence: N/10) — finding
  Recommended fix: ...

### LLM Trust Boundary: [PASS/FAIL/N/A]
- ...

### Conditional Side Effects: [PASS/FAIL/N/A]
- ...

### Shell & Code Injection: [PASS/FAIL/N/A]
- ...

### Enum & Value Completeness: [PASS/FAIL/N/A]
- ...

### Type Coercion: [PASS/FAIL/N/A]
- ...

### Recommendations
- ...
```

이슈 없음: `Pre-Landing Review: No issues found.`
