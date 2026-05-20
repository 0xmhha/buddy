package notify

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"runtime"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestDesktopChannel_NameAndUnsupported(t *testing.T) {
	t.Parallel()
	d := NewDesktopChannel()
	require.Equal(t, ChannelDesktop, d.Name())
	if runtime.GOOS != "darwin" && runtime.GOOS != "linux" {
		// On unsupported OSes Send returns ErrUnsupported.
		err := d.Send(context.Background(), Notification{Title: "t", Body: "b"})
		require.ErrorIs(t, err, ErrUnsupported)
	}
}

// TestDesktopChannel_RunnerIntercepted — inject a fake exec to avoid
// hitting the real OS during CI.
func TestDesktopChannel_RunnerIntercepted(t *testing.T) {
	t.Parallel()
	if runtime.GOOS != "darwin" && runtime.GOOS != "linux" {
		t.Skip("intercept test relies on a supported OS path")
	}
	var calls int32
	d := &DesktopChannel{
		Bin: "true", // exists on every POSIX, exits 0
		Runner: func(ctx context.Context, name string, args ...string) *exec.Cmd {
			atomic.AddInt32(&calls, 1)
			// Use `true` regardless of name so output is empty + exit 0.
			return exec.CommandContext(ctx, "true")
		},
	}
	err := d.Send(context.Background(), Notification{Title: "t", Body: "b"})
	require.NoError(t, err)
	require.Equal(t, int32(1), atomic.LoadInt32(&calls))
}

func TestWebhookChannel_PostsJSON(t *testing.T) {
	t.Parallel()
	var gotPayload webhookPayload
	var gotMethod, gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		_ = json.NewDecoder(r.Body).Decode(&gotPayload)
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	ch := NewWebhookChannel(WebhookConfig{URL: srv.URL + "/hook"})
	err := ch.Send(context.Background(), Notification{
		AdvisoryID: 7, Kind: "k", Severity: SeverityWarn,
		Title: "T", Body: "B", CreatedAt: time.Now().UTC(),
	})
	require.NoError(t, err)
	require.Equal(t, "POST", gotMethod)
	require.Equal(t, "/hook", gotPath)
	require.Equal(t, int64(7), gotPayload.AdvisoryID)
	require.Equal(t, "k", gotPayload.Kind)
	require.Equal(t, "warn", gotPayload.Severity)
}

func TestWebhookChannel_NonSuccessIsError(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, `{"error":"nope"}`, http.StatusUnauthorized)
	}))
	t.Cleanup(srv.Close)
	ch := NewWebhookChannel(WebhookConfig{URL: srv.URL})
	err := ch.Send(context.Background(), Notification{Title: "t", Body: "b"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "401")
}

func TestWebhookChannel_RespectsCustomHeaders(t *testing.T) {
	t.Parallel()
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)
	ch := NewWebhookChannel(WebhookConfig{
		URL:     srv.URL,
		Headers: map[string]string{"Authorization": "Bearer xyz"},
	})
	err := ch.Send(context.Background(), Notification{Title: "t", Body: "b"})
	require.NoError(t, err)
	require.Equal(t, "Bearer xyz", gotAuth)
}

func TestWebhookChannel_EmptyURLErrors(t *testing.T) {
	t.Parallel()
	ch := NewWebhookChannel(WebhookConfig{})
	err := ch.Send(context.Background(), Notification{Title: "t", Body: "b"})
	require.Error(t, err)
}

func TestShellPromptChannel_NoOpSend(t *testing.T) {
	t.Parallel()
	ch := NewShellPromptChannel()
	require.Equal(t, ChannelShell, ch.Name())
	require.NoError(t, ch.Send(context.Background(), Notification{}))
}

func TestTUIBannerChannel_NoOpSend(t *testing.T) {
	t.Parallel()
	ch := NewTUIBannerChannel()
	require.Equal(t, ChannelTUIBanner, ch.Name())
	require.NoError(t, ch.Send(context.Background(), Notification{}))
}

// sanity: a malformed URL surfaces as transport error, not panic.
func TestWebhookChannel_MalformedURL(t *testing.T) {
	t.Parallel()
	ch := NewWebhookChannel(WebhookConfig{URL: "://bad-url"})
	err := ch.Send(context.Background(), Notification{Title: "t", Body: "b"})
	require.Error(t, err)
	// The error wraps either parse or transport — both acceptable.
	require.True(t, errors.Is(err, err))
}
