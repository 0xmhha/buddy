#!/usr/bin/env python3
"""buddy knowledge embedder — sentence-transformers wrapper.

Stdin protocol:  JSON Lines, one {"id": int, "text": str} per line.
Stdout protocol: JSON Lines, one {"id": int, "embedding": [float, ...]} per line.
Errors:          single {"error": str} JSON object on stderr; exit code != 0.

Model: sentence-transformers/all-MiniLM-L6-v2 (384-dim, ~80MB on disk).
Override with BUDDY_EMBED_MODEL env var to use a different model.

Invoked by internal/knowledge/embedder.go via os/exec. The Go side keeps
the process alive for one batch and reads streamed JSON Lines until the
input pipe closes — model load happens once per batch, not per chunk.
"""
from __future__ import annotations

import json
import os
import sys

DEFAULT_MODEL = "sentence-transformers/all-MiniLM-L6-v2"


def fail(msg: str, code: int = 2) -> None:
    sys.stderr.write(json.dumps({"error": msg}) + "\n")
    sys.stderr.flush()
    sys.exit(code)


def main() -> None:
    try:
        from sentence_transformers import SentenceTransformer  # type: ignore
    except ImportError as exc:
        fail(
            f"sentence-transformers not installed in this Python env "
            f"({sys.executable}): {exc}. "
            f"Install with: pip install sentence-transformers"
        )

    model_name = os.environ.get("BUDDY_EMBED_MODEL", DEFAULT_MODEL)
    try:
        model = SentenceTransformer(model_name)
    except Exception as exc:  # noqa: BLE001 — surface load failure verbatim
        fail(f"failed to load model {model_name}: {exc}")

    # Stream input → output. Buffer flush after each line so the Go caller
    # sees results incrementally rather than at EOF.
    for line in sys.stdin:
        line = line.strip()
        if not line:
            continue
        try:
            req = json.loads(line)
        except json.JSONDecodeError:
            continue  # malformed input line; skip silently

        text = req.get("text", "")
        if not isinstance(text, str) or not text:
            continue

        try:
            emb = model.encode(text, normalize_embeddings=True).tolist()
        except Exception as exc:  # noqa: BLE001
            fail(f"encode failed for id={req.get('id')}: {exc}", code=3)

        out = {"id": req.get("id"), "embedding": emb}
        sys.stdout.write(json.dumps(out) + "\n")
        sys.stdout.flush()


if __name__ == "__main__":
    main()
