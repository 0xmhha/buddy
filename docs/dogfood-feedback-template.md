# Buddy Dogfood Feedback — <YYYY-MM-DD ~ YYYY-MM-DD>

> 이 템플릿을 복사한 뒤 빈칸을 채워 PR 또는 이슈로 제출. 솔직한 마찰 포인트가
> M5 우선순위를 결정한다.

## Usage stats

- Days used: __
- buddy version (`buddy --version`): __
- Claude Code 사용 빈도 (대략): 매일 ___ 시간 / __ 세션
- Total hook 호출 (24h): `buddy events --limit 99999 | wc -l` (events 한 줄 = hook 호출 1번) → __
- Total hook 호출 (전체 기간): 위 값 × 일수 ≈ __

### Top 3 hooks by count
1. ___
2. ___
3. ___

### p95 latency (가장 느린 3개)
1. ___ ms — hook: ___
2. ___ ms — hook: ___
3. ___ ms — hook: ___

---

## What worked

- ...
- ...

## What was friction

- (예: install 직후 doctor 실행하니 에러 → daemon start 먼저 해야 함을 몰랐다)
- ...

## Bugs found

- (재현 절차 / 메시지 / 예상 vs 실제)
- `uninstall`이 daemon을 자동으로 안 끄는 게 의도적인지 알려줘 (orphan daemon 발견했나?): __
- ...

## Things buddy should know but didn't

> doctor가 잡아내야 했지만 못 잡은 이슈, stats가 보여줘야 하는데 없는 컬럼,
> events에서 확인하기 어려웠던 패턴 등.

- ...

## Friend tone — 실제 대화처럼 어색하지 않았나?

- 어색했던 메시지: "..." → 차라리 이렇게: "..."
- 너무 자주 떠서 noise가 된 메시지: ...
- 침묵해야 할 때 침묵했나? (yes/no)

## 명령어 / 플래그가 부족했던 것

- (예: `stats --since "2 days ago"` 같은 시점 지정이 필요했음)
- ...

## Wishes for M5

> M5는 config / threshold tuning / purge / 페르소나 polish. 위 마찰 포인트가
> config 한 줄로 해결될지, 새 명령이 필요할지 적자.

- config로 해결 가능: ...
- 새 명령 / 플래그 필요: ...
- 페르소나 메시지 catalog 정리 필요: ...

---

## 회고 한 줄
buddy를 (계속 쓰겠다 / 잠시 떼겠다 / 영영 떼겠다). 이유: ___
