package main

import (
	"encoding/json"
	"io"
)

// writeJSONIndented pretty-prints v as JSON. Used by `buddy agent run`
// so the terminal output mirrors the file-output JSON layout.
func writeJSONIndented(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}
