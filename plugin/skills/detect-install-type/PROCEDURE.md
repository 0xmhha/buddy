---
name: detect-install-type
description: "[패턴 라이브러리] tool install type(global-git/local-git/vendored/package-manager/dev-symlink)을 detect하고 올바른 upgrade flow로 route. 직접 invoke보다 orchestrator가 import해 사용. 트리거: '설치 방식 감지' / '어떻게 설치됐어?' / 'install type' / '환경 감지'. 참조 위치: guide-setup-wizard 입력, 부트스트랩 1회, self-updating tool 작성 시."
type: skill
---

# Snippet: Install-Type Detection + Change Summary


> 설치 방식 감지 알고리즘 + pre/post 변경 요약 기법 — 도구 무관.

## 이 snippet을 사용하는 경우

- 자체 업그레이드가 필요한 모든 CLI (git-clone과 package-manager 설치가 경로가 다름)
- 도구가 dev-mode(editable checkout) vs 프로덕션 설치인지 감지
- "3가지 이상 방식으로 설치 가능, 올바른 명령 선택" 라우팅
- npm/pip로 설치된 디렉토리에 `git pull`하는 버그 회피

## 패턴 1: Install-Type 감지

가장 구체적 → 가장 generic 순으로 파일시스템을 probe. 첫 매치가 승자.

### 파일시스템 신호

| 신호 | 의미 |
|--------|------------------|
| `<install_dir>/.git` 존재 | `git clone`으로 설치됨 (in-place 업그레이드 가능) |
| `<install_dir>`이 symlink | Editable / dev-mode 설치, checkout을 가리킴 |
| `<install_dir>/package.json` (또는 동등물) 존재, `.git` 없음 | Vendored copy (snapshot, git history 없음) |
| `<install_dir>`이 package manager root (`node_modules`, `site-packages`, `~/.cargo/...`) 하에 존재 | Package-manager 관리 (`npm`/`pip`/`cargo`로 업그레이드 필수) |
| 디렉토리 전혀 없음 | 설치 안 됨 |

### Decision tree (bash sketch)

```bash
# 순서 중요: local-git 전 global-git, vendored 전 git-checkout,
# "missing" 전 vendored
if [ -L "$PRIMARY_DIR" ]; then
  INSTALL_TYPE="dev-symlink"           # editable checkout, 자동 업그레이드 금지
  INSTALL_DIR=$(readlink -f "$PRIMARY_DIR")
elif [ -d "$GLOBAL_DIR/.git" ]; then
  INSTALL_TYPE="global-git"            # git pull / reset --hard
  INSTALL_DIR="$GLOBAL_DIR"
elif [ -d "./<tool>/.git" ]; then
  INSTALL_TYPE="local-git"             # repo-local checkout, global과 동일 flow
  INSTALL_DIR="./<tool>"
elif [ -d "$GLOBAL_DIR" ]; then
  INSTALL_TYPE="vendored"              # snapshot copy, 통째로 교체
  INSTALL_DIR="$GLOBAL_DIR"
elif command -v <tool> >/dev/null && <tool> --installed-via 2>/dev/null | grep -q npm; then
  INSTALL_TYPE="package-manager"       # 패키지 매니저에 위임
else
  echo "ERROR: <tool> not found"; exit 1
fi
echo "Install type: $INSTALL_TYPE at $INSTALL_DIR"
```

### 타입별 업그레이드 라우팅

| Type | 업그레이드 명령 |
|------|-----------------|
| `global-git` / `local-git` | `cd $DIR && git stash && git fetch origin && git reset --hard origin/<branch> && ./setup` |
| `vendored` | `git clone --depth 1 <repo> /tmp/x && mv $DIR $DIR.bak && mv /tmp/x $DIR && cd $DIR && ./setup && rm -rf $DIR.bak` |
| `package-manager` | `npm i -g <pkg>@latest` (또는 `pip install --upgrade`, `cargo install --force`) — 디렉토리 직접 건드리지 말 것 |
| `dev-symlink` | 거부. "이건 editable checkout — 수동 pull" 출력. |

### 처리할 가치가 있는 edge case

- **한 머신에 여러 설치.** 주 설치 감지 후 *또한* 이차 copy(예: global + repo-local) 확인. 주 업그레이드 후 stale한 이차 설치를 resync하거나 삭제해 동작 drift 방지.
- **Install dir을 가리키는 symlink.** 경로 비교 전 `readlink -f`/`pwd -P`로 resolve — 문자열 등치 비교는 symlink 동등성을 놓친다.
- **Git 설치의 local 변경사항.** `git stash`를 먼저 해서 stash 출력을 캡처하고 "로컬 변경사항이 stash됐습니다" surface해 사용자가 pop할 수 있게.
- **Setup script 실패.** 이전 설치의 `.bak`을 유지. 실패 시 `.bak`에서 복원하고 수동 재시도 방법을 사용자에게 알림. 반쯤 업그레이드된 도구로 내버려두지 말 것.

## 패턴 2: Pre/Post 변경 요약

업그레이드 후 generic한 "업그레이드 완료"가 아니라 사용자에게 *실제로 무엇이 바뀌었는지* 보여준다. 이게 자체 업그레이드를 블랙박스가 아니라 안전하게 느끼게 만든다.

### 단계

1. **업그레이드 전:** 현재 버전을 변수로 읽기.
   ```bash
   OLD_VERSION=$(cat "$INSTALL_DIR/VERSION" 2>/dev/null || echo "unknown")
   ```
2. **업그레이드 실행** (위 라우팅 테이블의 타입별 명령).
3. **업그레이드 후:** 새 버전 읽기.
   ```bash
   NEW_VERSION=$(cat "$INSTALL_DIR/VERSION")
   ```
4. **changelog diff** 두 버전 간. `CHANGELOG.md`에서 반개구간 `(OLD, NEW]`에 속하는 모든 엔트리 찾기.
5. **5-7 bullet으로 테마별 요약.** 내부 refactor와 기여자 대상 churn은 skip — 사용자 대상만 유지. 전체 changelog 덤프 금지.

### 출력 형태

```
<tool> v{NEW} — upgraded from v{OLD}!

What's new:
- [bullet 1: user-facing change, plain language]
- [bullet 2]
- ...
```

### 이게 왜 중요한가

조용한 "업그레이드 완료"는 사용자에게 맹목 신뢰하거나 changelog를 직접 읽으라고 강요한다. 5-bullet 요약은 업그레이드를 "오 좋다, 써보고 싶다"의 순간으로 바꾸며, 이게 다음에 사용자가 업그레이드를 *원하게* 만드는 유일한 방법이다. 또한 실수로 다운그레이드도 감지 — 업그레이드 실행 후 `NEW < OLD`면 뭔가 잘못됐고 diff가 비었거나 음수.

### 선택: 업그레이드 후 migration

도구에 버전 간 drift 가능한 on-disk state(config schema, 디렉토리 레이아웃, cache 포맷)가 있으면, repo에 버전 marcked migration script(`v<X.Y.Z>.sh`)를 ship. `./setup` 후, 버전이 `> OLD_VERSION`이고 `<= NEW_VERSION`인 script들을 iterate해 멱등하게 실행, non-fatal 에러는 log. 재실행이 no-op이 되도록 migration을 멱등하게 유지.
