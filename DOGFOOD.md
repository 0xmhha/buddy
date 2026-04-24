# DOGFOOD — buddy v0.1 첫 사용 가이드

> 이 문서대로 따라하면 10분 안에 buddy를 본인 머신에 켜고, 며칠 동안 평소처럼
> Claude Code를 쓰면서 hook 신뢰성·통계 데이터를 모을 수 있어. 며칠 후 회고는
> [`docs/dogfood-feedback-template.md`](./docs/dogfood-feedback-template.md) 사용.

대상 독자: buddy를 처음 설치하는 사용자. macOS / Linux. Claude Code가 이미
설치되어 있고 `~/.claude/settings.json`이 존재한다고 가정.

---

## 0. 안전 장치 (먼저 읽기)

### 0.1 `~/.claude/settings.json` 수동 백업
buddy install이 자동으로 `settings.json.buddy.bak`를 만들지만, **신뢰성 1순위**
이므로 수동 사본 하나 더 떠두자.

```bash
cp ~/.claude/settings.json ~/.claude/settings.json.pre-buddy.$(date +%Y%m%d)
```

문제 생기면 위 사본을 직접 `cp`로 덮어 복원.

### 0.2 첫 사용은 `--with-cliwrap` 없이
첫 dogfood는 **cliwrap 없이** 시작 권장. 이유: 실패 모드를 단순하게 격리하기
위해. cli-wrapper supervision은 round 2에서 옵트인.

`--with-cliwrap`을 쓰려면 별도 binary가 필요해 (자동 설치 안 됨):
```bash
go install github.com/0xmhha/cli-wrapper/cmd/cliwrap@latest
which cliwrap   # PATH에 있어야 함
```
첫 사용에서는 그냥 건너뛰자.

### 0.3 무엇을 켜는가
buddy는 `~/.claude/settings.json`의 hook command를 `buddy hook-wrap …`으로
**감싸기만** 한다. hook의 동작은 그대로, 실행 결과만 buddy 옆구리(SQLite)에
기록. 백업 → wrapping → 복원 모두 atomic.

---

## 1. 설치

```bash
# 1) 빌드
cd /path/to/buddy
make build

# 2) 절대 경로 변수
BUDDY_BIN="$PWD/bin/buddy"
echo "$BUDDY_BIN"

# 3) install (default — without cliwrap)
$BUDDY_BIN install --buddy-binary "$BUDDY_BIN"
```

성공 시 친구 메시지:
```
buddy: 등록 완료. 이제 옆에서 보고 있을게.
```

이미 wrap된 상태에서 다시 실행하면:
```
buddy: 이미 등록되어 있어. 변화 없음.
```

### 1.1 백업·wrapping 결과 확인
```bash
ls -la ~/.claude/settings.json.buddy.bak    # 자동 백업 파일
diff ~/.claude/settings.json.buddy.bak ~/.claude/settings.json
```

`diff` 출력에서 hook command가 다음 형태로 바뀌어 있으면 성공:
```
"command": "<원래 명령>"
  ↓
"command": "/abs/path/to/buddy hook-wrap <HookName>:<matcher> --event <Event> -- <원래 명령>"
```

### 1.2 buddy 작업 디렉터리
기본 `~/.buddy/` 가 사용된다. install은 이 디렉터리를 **미리 만들지는 않는다**
— daemon이 처음 뜰 때 자동 생성. 직접 다른 경로를 쓰고 싶으면 `--buddy-dir`
지정 (이후 모든 명령에 같은 `--db`를 줘야 함).

---

## 2. daemon 시작

`buddy daemon`은 outbox(임시 큐)를 `hook_events` + `hook_stats` 테이블로
드레인하는 백그라운드 프로세스. **이걸 안 띄우면 stats/doctor가 못 읽는다.**

```bash
$BUDDY_BIN daemon start
$BUDDY_BIN daemon status   # running (pid …)
```

`daemon start` 직후 메시지는 부모 프로세스가 detach 직후라 `pid -1`로 표시될
수 있다 (cosmetic). 실제 pid는 `daemon status`로 확인.

### 2.1 daemon이 떠야 doctor도 의미 있음
SQLite 테이블은 daemon이 처음 뜰 때 마이그레이션된다. **daemon 한 번도 안 띄우고**
`buddy doctor`를 실행하면 다음 에러가 뜬다:
```
어, 몇 가지 봐줄 게 있어.

  • DB를 못 열었어 (~/.buddy/buddy.db): SQL logic error: no such table: hook_outbox (1)
```
→ 정상. `daemon start` 먼저.

---

## 3. 사용 — 평소대로 Claude Code 쓰기

여기까지 왔으면 끝. **Claude Code를 평소처럼** 며칠 사용하면 buddy가 hook 호출을
자동 기록한다.

가끔 들여다볼 명령:

```bash
# health 한눈에
$BUDDY_BIN doctor

# 1시간 통계
$BUDDY_BIN stats --window 1h

# tool별로 쪼개 보기
$BUDDY_BIN stats --by-tool --window 1h

# 어색한 게 있으면 raw event tail
$BUDDY_BIN events --limit 50

# 실시간 (Ctrl-C로 종료)
$BUDDY_BIN events --follow
```

처음 24시간은 데이터가 적어 `stats`가 다음 메시지를 낼 수 있다:
```
아직 기록된 hook 통계가 없어. daemon을 띄우고 좀 기다려봐.
```

특정 hook만 보려면:
```bash
$BUDDY_BIN stats --hook PreToolUse --window 24h
$BUDDY_BIN events --hook PostToolUse --limit 30
```

`stats --window` 허용값: `5m | 1h | 24h`.

---

## 4. 트러블슈팅

| 증상 | 친구 메시지 / 확인 | 대응 |
|------|------------------|------|
| `buddy: 실행 중인 daemon이 없어.` | `daemon stop` 했는데 status가 not running | `daemon start` 다시 |
| `outbox에 N개 쌓였어. daemon 한 번 봐줘.` | `doctor`가 backlog 알림 | `daemon stop && daemon start`, 또는 임시로 `daemon run --batch 2000` |
| `DB를 못 열었어 … no such table` | daemon이 한 번도 안 떴음 | `daemon start` 먼저 |
| `DB를 못 열었어 … out of memory (14)` | `--db` 경로의 부모 디렉터리가 없음 | `mkdir -p $(dirname <db>)` |
| `buddy: 바이너리 경로에 공백이 있어.` | install 시 `BUDDY_BIN`에 공백 포함 | 공백 없는 경로로 옮긴 뒤 재실행 |
| `buddy: ~/.claude/settings.json 이 안 보여.` | Claude Code 미설치 또는 경로 다름 | `--claude-dir`로 명시 |
| hook이 평소보다 느린 느낌 | `stats --by-tool`로 어떤 tool인지 확인 | hook 자체 이슈일 가능성 — buddy는 5-10ms 오버헤드만 |
| settings.json 손상 의심 | `diff ~/.claude/settings.json ~/.claude/settings.json.buddy.bak` | `cp ~/.claude/settings.json.buddy.bak ~/.claude/settings.json` |

### 4.1 daemon이 한 번도 안 떠있던 시점에 hook이 호출되면?
괜찮다. hook-wrap는 outbox(SQLite)에 **append-only**로 쓰고 끝낸다. daemon은
나중에 outbox를 드레인. 단, 첫 hook 호출 전에 `daemon start`로 DB 마이그레이션
(테이블 생성)이 한 번 일어나야 outbox row를 받아줄 수 있다 → **install 직후
바로 daemon start** 가 깔끔한 순서.

---

## 5. 며칠 후 회고

며칠(권장: 3~7일) 사용 후 [`docs/dogfood-feedback-template.md`](./docs/dogfood-feedback-template.md)
복사해서 채우자.

수집할 데이터 항목:
```bash
# 24시간 hook 호출 수
$BUDDY_BIN stats --window 24h | wc -l

# 가장 자주 트리거된 hook + tool
$BUDDY_BIN stats --by-tool --window 24h

# 느려지는 추세 확인 (p95 변화)
$BUDDY_BIN stats --window 1h    # 최근
$BUDDY_BIN stats --window 24h   # 24h 평균과 비교

# silent failure 후보
$BUDDY_BIN events --limit 100 | grep -i 'fail\|error\|timeout'
```

회고 템플릿에 더해, **buddy 자체에 대한 마찰 포인트**를 솔직히 적자:
- 이 가이드에서 막힌 곳
- 친구 톤 메시지가 어색했던 곳
- 있어야 할 명령/플래그가 없었던 곳
- doctor가 잡지 못한 이슈

이 마찰 포인트들이 M5(`config / threshold tuning / 페르소나 polish`)의 직접적인
입력.

---

## 6. 안전 종료

```bash
# 1) daemon 멈추기
$BUDDY_BIN daemon stop
# → buddy: daemon에 종료 신호 보냈어 (pid …).

# 2) settings.json 원복
$BUDDY_BIN uninstall
# → buddy: 해제 완료. 백업에서 복원했어.
```

복원 후 검증:
```bash
diff ~/.claude/settings.json.pre-buddy.<날짜> ~/.claude/settings.json
# 출력 없으면 완전히 원상복구됨
```

`uninstall`은:
- `settings.json.buddy.bak`이 있으면 → 백업에서 복원 (가장 신뢰)
- 백업이 없거나 손상 → wrap된 hook command만 unwrap (fallback)
- 등록된 게 없으면 → `buddy: 등록된 게 없어. 그대로 둘게.`

`~/.buddy/` 디렉터리 (DB·로그)는 `uninstall`이 건드리지 않는다. 통계 보존
용도. 완전히 지우려면 수동:
```bash
rm -rf ~/.buddy
```

---

## 부록 A. cli-wrapper 통합 (round 2 옵션)

첫 사용에서 buddy가 안정적으로 돌면, daemon supervision을 cli-wrapper에
맡기고 싶을 수 있어. 그때:

```bash
# 1) cliwrap 설치
go install github.com/0xmhha/cli-wrapper/cmd/cliwrap@latest

# 2) 일단 uninstall
$BUDDY_BIN uninstall

# 3) cliwrap.yaml도 함께 생성
$BUDDY_BIN install --with-cliwrap --buddy-binary "$BUDDY_BIN"
# → buddy: cliwrap.yaml 도 써뒀어 (~/.buddy/cliwrap.yaml).

# 4) cliwrap이 daemon을 supervise (재기동 자동)
cliwrap --config ~/.buddy/cliwrap.yaml run buddy-daemon
```

cli-wrapper supervision은 daemon이 죽으면 자동 재기동·로그 회전을 해 주지만,
첫 사용 단계에서는 **두 layer가 동시에 실패할 때 디버깅이 어려워지므로** 옵트인
권장.

---

## 부록 B. 임시 디렉터리에서 dry-run

본인 `~/.claude`를 건드리기 전에 가짜 디렉터리에서 한 번 돌려보고 싶다면:

```bash
# 1) 가짜 settings.json
mkdir -p /tmp/buddy-trial-claude
echo '{"hooks":{"PreToolUse":[{"matcher":"Bash","hooks":[{"type":"command","command":"echo pre"}]}]}}' \
  > /tmp/buddy-trial-claude/settings.json

# 2) install → wrap된 결과 확인
$BUDDY_BIN install \
  --claude-dir /tmp/buddy-trial-claude \
  --buddy-dir /tmp/buddy-trial-buddy \
  --buddy-binary "$BUDDY_BIN"
cat /tmp/buddy-trial-claude/settings.json
ls /tmp/buddy-trial-claude/settings.json.buddy.bak

# 3) daemon (별도 db/pid)
$BUDDY_BIN daemon start \
  --db /tmp/buddy-trial-buddy/buddy.db \
  --pid /tmp/buddy-trial-buddy/daemon.pid
$BUDDY_BIN doctor \
  --db /tmp/buddy-trial-buddy/buddy.db \
  --pid /tmp/buddy-trial-buddy/daemon.pid
# → 모두 정상이야.

# 4) cleanup
$BUDDY_BIN daemon stop \
  --db /tmp/buddy-trial-buddy/buddy.db \
  --pid /tmp/buddy-trial-buddy/daemon.pid
$BUDDY_BIN uninstall \
  --claude-dir /tmp/buddy-trial-claude \
  --buddy-binary "$BUDDY_BIN"
rm -rf /tmp/buddy-trial-claude /tmp/buddy-trial-buddy
```

이 가이드의 모든 명령은 위 dry-run 시나리오로 검증된 것.
