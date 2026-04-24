// Package diagnose produces a one-shot, read-only health snapshot of buddy.
//
// Doctor never blocks the daemon: it opens the SQLite store read-only and
// inspects existing state (outbox backlog, recent hook_stats) plus the
// daemon PID file. All policy thresholds live in DefaultThresholds (Decision 2,
// v0.1-spec §6.2). Output language is friend-tone Korean (Decision 3, §6.3).
//
// All functions in this package are pure: no os.Exit, no cobra, no globals.
// The CLI layer (cmd/buddy) owns process exit codes and IO routing.
package diagnose

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/wm-it-22-00661/buddy/internal/daemon"
	"github.com/wm-it-22-00661/buddy/internal/db"
)

// IssueKind identifies the category of a diagnostic, for testability and for
// downstream tooling that may want to filter or color issues.
type IssueKind string

const (
	KindDBOpen   IssueKind = "db-open"
	KindDaemon   IssueKind = "daemon"
	KindBacklog  IssueKind = "backlog"
	KindSlow     IssueKind = "slow"
	KindFailRate IssueKind = "fail-rate"
)

// Diagnostic is a single issue found by Check.
type Diagnostic struct {
	Kind    IssueKind
	Message string
}

// Report is the full result of a doctor run. Healthy is true iff Issues is empty.
type Report struct {
	Issues  []Diagnostic
	Healthy bool
}

// Thresholds are the per-check policy knobs. Zero values mean "use defaults".
// Defaults are locked in by Decision 2 (v0.1-spec §6.2).
type Thresholds struct {
	HookTimeoutMs   int64 // 30000 — informational only here; M3 measurement
	HookSlowMs      int64 // 5000  — p95 over this in last 5min triggers
	HookFailRatePct int   // 20    — failure rate over this in last 5min triggers
	OutboxBacklog   int   // 1000  — pending outbox rows over this triggers
}

// DefaultThresholds returns the spec-locked defaults.
func DefaultThresholds() Thresholds {
	return Thresholds{
		HookTimeoutMs:   30_000,
		HookSlowMs:      5_000,
		HookFailRatePct: 20,
		OutboxBacklog:   1_000,
	}
}

// withDefaults fills any zero field with its DefaultThresholds counterpart.
// We don't replace the whole struct on a single zero field because callers may
// legitimately pass a partial override.
func (t Thresholds) withDefaults() Thresholds {
	d := DefaultThresholds()
	if t.HookTimeoutMs == 0 {
		t.HookTimeoutMs = d.HookTimeoutMs
	}
	if t.HookSlowMs == 0 {
		t.HookSlowMs = d.HookSlowMs
	}
	if t.HookFailRatePct == 0 {
		t.HookFailRatePct = d.HookFailRatePct
	}
	if t.OutboxBacklog == 0 {
		t.OutboxBacklog = d.OutboxBacklog
	}
	return t
}

// Options configure a Check run. Zero-valued fields fall back to spec defaults.
type Options struct {
	DBPath     string
	PIDFile    string
	Thresholds Thresholds
}

// topN limits how many slow / fail-rate issues we surface so a noisy run does
// not drown the overall report. Picked over CLI flags for v0.1 simplicity.
const topN = 5

// Check runs all diagnostics in a fixed order and returns a Report.
// The order is: DB → daemon → backlog → slow p95 → high failure rate.
// If the DB cannot be opened the function bails after the single DBOpen issue:
// no other check is meaningful without the store.
func Check(opts Options) (Report, error) {
	thr := opts.Thresholds.withDefaults()

	dbPath, err := resolveDBPath(opts.DBPath)
	if err != nil {
		return Report{}, fmt.Errorf("resolve db path: %w", err)
	}
	pidFile := resolvePIDFile(opts.PIDFile, dbPath)

	rep := Report{}

	conn, err := db.Open(db.Options{Path: dbPath, ReadOnly: true})
	if err != nil {
		return dbOpenReport(dbPath, err), nil
	}
	defer conn.Close()

	// modernc/sqlite defers actual file access to first query — so a missing
	// or corrupt DB surfaces here, not at Open. Any DB-side failure during the
	// snapshot collapses into a single KindDBOpen issue: partial diagnostics
	// would mislead more than they'd help.
	rep.Issues = append(rep.Issues, daemonIssues(pidFile)...)

	backlogIssues, err := outboxBacklogIssues(conn, thr.OutboxBacklog)
	if err != nil {
		return dbOpenReport(dbPath, err), nil
	}
	rep.Issues = append(rep.Issues, backlogIssues...)

	slowIssues, err := slowHookIssues(conn, thr.HookSlowMs)
	if err != nil {
		return dbOpenReport(dbPath, err), nil
	}
	rep.Issues = append(rep.Issues, slowIssues...)

	failIssues, err := failRateIssues(conn, thr.HookFailRatePct)
	if err != nil {
		return dbOpenReport(dbPath, err), nil
	}
	rep.Issues = append(rep.Issues, failIssues...)

	rep.Healthy = len(rep.Issues) == 0
	return rep, nil
}

// dbOpenReport builds the single-issue, not-healthy report for any DB-level
// failure (open or query). Centralised so the friend-tone wording stays
// identical across call sites.
func dbOpenReport(dbPath string, err error) Report {
	return Report{
		Healthy: false,
		Issues: []Diagnostic{{
			Kind:    KindDBOpen,
			Message: fmt.Sprintf("DB를 못 열었어 (%s): %v", dbPath, err),
		}},
	}
}

// Render writes a friend-tone Korean summary of the report to w.
// Healthy reports get a single line; otherwise we lead with a header and bullet
// each issue. No trailing blank line — callers that need spacing add their own.
func (r Report) Render(w io.Writer) {
	if r.Healthy {
		_, _ = io.WriteString(w, "모두 정상이야.\n")
		return
	}
	_, _ = io.WriteString(w, "어, 몇 가지 봐줄 게 있어.\n\n")
	for _, issue := range r.Issues {
		_, _ = io.WriteString(w, "  • "+issue.Message+"\n")
	}
}

// --- internals ---

func resolveDBPath(p string) (string, error) {
	if p != "" {
		return p, nil
	}
	return db.DefaultPath()
}

func resolvePIDFile(p, dbPath string) string {
	if p != "" {
		return p
	}
	return filepath.Join(filepath.Dir(dbPath), "daemon.pid")
}

func daemonIssues(pidFile string) []Diagnostic {
	st, err := daemon.CheckStatus(pidFile)
	if err != nil {
		// A pid-file read error is itself a daemon-level signal: the file is
		// there but we can't tell who owns it. Surface it under KindDaemon so
		// the user sees a single coherent thread per concern.
		return []Diagnostic{{
			Kind: KindDaemon,
			Message: fmt.Sprintf(
				"daemon 상태를 못 읽었어 (%s): %v", pidFile, err),
		}}
	}
	if !st.Running {
		return []Diagnostic{{
			Kind:    KindDaemon,
			Message: "daemon이 실행 중이 아니야. 'buddy daemon start'로 띄울 수 있어.",
		}}
	}
	return nil
}

func outboxBacklogIssues(conn *sql.DB, threshold int) ([]Diagnostic, error) {
	var n int
	row := conn.QueryRow("SELECT COUNT(*) FROM hook_outbox WHERE consumed_at IS NULL")
	if err := row.Scan(&n); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	if n <= threshold {
		return nil, nil
	}
	return []Diagnostic{{
		Kind: KindBacklog,
		Message: fmt.Sprintf(
			"outbox에 %s개 쌓였어. daemon 한 번 봐줘 (buddy daemon status).",
			formatThousands(n)),
	}}, nil
}

// slowHookEntry is one row from the slow-hook query, kept compact for sorting.
type slowHookEntry struct {
	hookName string
	toolName string
	tsBucket int64
	p95      int64
}

// slowHookIssues finds (hook, tool) pairs whose latest 5-min bucket has p95
// above threshold. We pick the latest bucket per (hook, tool) so a single old
// outlier doesn't keep paging the user; a stale bucket simply won't appear in
// the latest tick window.
func slowHookIssues(conn *sql.DB, thresholdMs int64) ([]Diagnostic, error) {
	rows, err := conn.Query(`
		SELECT hook_name, tool_name, ts_bucket, p95_ms
		  FROM hook_stats h
		 WHERE window_min = 5
		   AND p95_ms IS NOT NULL
		   AND p95_ms > ?
		   AND ts_bucket = (
		         SELECT MAX(ts_bucket) FROM hook_stats
		          WHERE hook_name = h.hook_name
		            AND tool_name = h.tool_name
		            AND window_min = 5
		       )
		 ORDER BY p95_ms DESC`, thresholdMs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []slowHookEntry
	for rows.Next() {
		var e slowHookEntry
		if err := rows.Scan(&e.hookName, &e.toolName, &e.tsBucket, &e.p95); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(entries) > topN {
		entries = entries[:topN]
	}

	out := make([]Diagnostic, 0, len(entries))
	for _, e := range entries {
		out = append(out, Diagnostic{
			Kind: KindSlow,
			Message: fmt.Sprintf(
				"'%s' hook이 좀 느려졌어. p95가 %s (기준 %s).",
				displayName(e.hookName, e.toolName),
				humanDur(e.p95), humanDur(thresholdMs)),
		})
	}
	return out, nil
}

// failRateEntry is one row from the failure-rate query.
type failRateEntry struct {
	hookName string
	toolName string
	count    int64
	failures int64
}

func failRateIssues(conn *sql.DB, thresholdPct int) ([]Diagnostic, error) {
	rows, err := conn.Query(`
		SELECT hook_name, tool_name, count, failures
		  FROM hook_stats h
		 WHERE window_min = 5
		   AND count > 0
		   AND (failures * 100) / count > ?
		   AND ts_bucket = (
		         SELECT MAX(ts_bucket) FROM hook_stats
		          WHERE hook_name = h.hook_name
		            AND tool_name = h.tool_name
		            AND window_min = 5
		       )
		 ORDER BY (failures * 100) / count DESC, failures DESC`, thresholdPct)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []failRateEntry
	for rows.Next() {
		var e failRateEntry
		if err := rows.Scan(&e.hookName, &e.toolName, &e.count, &e.failures); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Stable ordering for deterministic test output: the SQL ORDER BY already
	// pins the primary key (rate desc, fail count desc), but ties on (hook,tool)
	// could surface in either order across SQLite versions. Force a final tiebreaker.
	sort.SliceStable(entries, func(i, j int) bool {
		ri := (entries[i].failures * 100) / entries[i].count
		rj := (entries[j].failures * 100) / entries[j].count
		if ri != rj {
			return ri > rj
		}
		if entries[i].failures != entries[j].failures {
			return entries[i].failures > entries[j].failures
		}
		if entries[i].hookName != entries[j].hookName {
			return entries[i].hookName < entries[j].hookName
		}
		return entries[i].toolName < entries[j].toolName
	})

	if len(entries) > topN {
		entries = entries[:topN]
	}

	out := make([]Diagnostic, 0, len(entries))
	for _, e := range entries {
		pct := (e.failures * 100) / e.count
		out = append(out, Diagnostic{
			Kind: KindFailRate,
			Message: fmt.Sprintf(
				"'%s' hook 실패율이 %d%% 야. 최근 %d번 중 %d번 실패.",
				displayName(e.hookName, e.toolName), pct, e.count, e.failures),
		})
	}
	return out, nil
}

// displayName combines hook + tool when tool is present, so users see the same
// granularity hook_stats actually keys on. e.g. "PreToolUse:Bash".
func displayName(hookName, toolName string) string {
	if toolName == "" {
		return hookName
	}
	return hookName + ":" + toolName
}

// humanDur renders ms as a human-friendly duration. Boundaries:
//
//	[0, 1000)      → "<n>ms"
//	[1000, 60000)  → "<n.n>s"
//	[60000, ∞)     → "<n.n>m"
//
// We fix one decimal for s/m so columns line up in lists.
func humanDur(ms int64) string {
	if ms < 0 {
		ms = 0
	}
	switch {
	case ms < 1000:
		return strconv.FormatInt(ms, 10) + "ms"
	case ms < 60_000:
		return formatOneDecimal(float64(ms)/1000.0) + "s"
	default:
		return formatOneDecimal(float64(ms)/60_000.0) + "m"
	}
}

// formatOneDecimal trims to one decimal place without scientific notation,
// avoiding fmt's default %g rounding surprises near boundaries (e.g. 59.999).
func formatOneDecimal(v float64) string {
	s := strconv.FormatFloat(v, 'f', 1, 64)
	// Strip a trailing ".0" only if it would still leave a digit; we keep "1.0"
	// rather than "1" to make the unit suffix unambiguous in friend-tone copy.
	return s
}

// formatThousands inserts commas as thousands separators. Pure helper, no
// locale awareness — v0.1 only ships ko/en, both of which are fine with commas.
func formatThousands(n int) string {
	if n < 0 {
		return "-" + formatThousands(-n)
	}
	s := strconv.Itoa(n)
	if len(s) <= 3 {
		return s
	}
	var b strings.Builder
	pre := len(s) % 3
	if pre > 0 {
		b.WriteString(s[:pre])
		if len(s) > pre {
			b.WriteByte(',')
		}
	}
	for i := pre; i < len(s); i += 3 {
		b.WriteString(s[i : i+3])
		if i+3 < len(s) {
			b.WriteByte(',')
		}
	}
	return b.String()
}
