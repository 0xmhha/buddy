package agent

import (
	"encoding/json"
	"io"
)

// writeJSON pretty-prints v to w. Kept in its own file so tests can shadow it
// (and the runtime.go stays focused on orchestration).
func writeJSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}
