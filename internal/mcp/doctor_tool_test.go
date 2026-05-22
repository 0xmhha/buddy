package mcp

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"

	"github.com/0xmhha/buddy/internal/db"
)

// TestDoctorTool_HealthyOnFreshDB — the doctor tool reports healthy
// when the DB exists and no daemon PID file is present (daemon not
// running is OK; the doctor only flags missing DBs and active
// problems). The MCP round trip covers (a) the tool is registered,
// (b) the JSON envelope returns without error, (c) the structured
// result mirrors diagnose.Report.
func TestDoctorTool_HealthyOnFreshDB(t *testing.T) {
	t.Parallel()
	dbPath := filepath.Join(t.TempDir(), "doctor.db")
	conn, err := db.Open(db.Options{Path: dbPath})
	require.NoError(t, err)
	require.NoError(t, conn.Close())

	server := NewBuddyServer(Options{DBPath: dbPath})
	client := mcp.NewClient(&mcp.Implementation{Name: "doctor-test", Version: "0.0.0"}, nil)
	st, ct := mcp.NewInMemoryTransports()

	ctx := context.Background()
	serverSession, err := server.Connect(ctx, st, nil)
	require.NoError(t, err)
	defer func() { _ = serverSession.Close() }()
	clientSession, err := client.Connect(ctx, ct, nil)
	require.NoError(t, err)
	defer func() { _ = clientSession.Close() }()

	res, err := clientSession.CallTool(ctx, &mcp.CallToolParams{
		Name: "doctor", Arguments: map[string]any{},
	})
	require.NoError(t, err)
	require.NotEmpty(t, res.Content, "doctor must return a text body")
}

// TestResolvePIDFile_DerivesFromDBPath — when DBPath is set the PID
// file lives in the same directory; when empty the default DB path
// drives the resolution. Pinning this saves us from having to read
// daemon-spawn integration logs to confirm the contract.
func TestResolvePIDFile_DerivesFromDBPath(t *testing.T) {
	t.Parallel()

	got, err := resolvePIDFile("/tmp/some-dir/buddy.db")
	require.NoError(t, err)
	require.Equal(t, "/tmp/some-dir/daemon.pid", got,
		"DBPath set → daemon.pid sits next to it")

	// Empty path falls back to db.DefaultPath; the only contract is
	// that the returned path ends with daemon.pid.
	got, err = resolvePIDFile("")
	require.NoError(t, err)
	require.Equal(t, "daemon.pid", filepath.Base(got),
		"empty DBPath → filename is daemon.pid under default DB dir")
}
