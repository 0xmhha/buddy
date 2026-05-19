package knowledge

import (
	"math"
	"sort"
	"strings"
	"unicode"
)

// BM25 parameters per Robertson et al. — k1=1.5 / b=0.75 are the
// canonical text-retrieval defaults. Exposed as constants because
// every internal call uses the same values; the function-level args
// stayed off to keep the surface small.
const (
	bm25K1 = 1.5
	bm25B  = 0.75
)

// BM25Index is a tiny in-memory BM25 retriever. Build it from a Chunk
// corpus once, then call Search repeatedly. Per ADR-014 §Q3 the corpus
// is small (~10k chunks) so in-memory full scan beats any disk-backed
// inverted index in both code complexity and latency.
//
// Thread-safety: Search is safe for concurrent callers (read-only).
// Construction must finish before Search is invoked.
type BM25Index struct {
	chunks       []Chunk            // by index → ID lookup later
	docTokens    [][]string         // pre-tokenised, lowercased
	docLen       []int              // |d| for the length-norm term
	df           map[string]int     // document frequency per term
	avgDocLen    float64
	totalDocs    int
	idfCache     map[string]float64 // computed once at NewBM25Index
}

// NewBM25Index builds the index over the provided chunks. Empty corpus
// is allowed — Search returns nothing.
func NewBM25Index(chunks []Chunk) *BM25Index {
	idx := &BM25Index{
		chunks:    chunks,
		df:        map[string]int{},
		idfCache:  map[string]float64{},
		totalDocs: len(chunks),
	}
	if len(chunks) == 0 {
		return idx
	}

	idx.docTokens = make([][]string, len(chunks))
	idx.docLen = make([]int, len(chunks))
	var totalLen int
	for i, c := range chunks {
		toks := tokenizeBM25(c.Content)
		idx.docTokens[i] = toks
		idx.docLen[i] = len(toks)
		totalLen += len(toks)
		seen := map[string]struct{}{}
		for _, t := range toks {
			if _, ok := seen[t]; ok {
				continue
			}
			seen[t] = struct{}{}
			idx.df[t]++
		}
	}
	if len(chunks) > 0 {
		idx.avgDocLen = float64(totalLen) / float64(len(chunks))
	}
	// Pre-compute idf for every term seen — the typical Search hits
	// 1-10 query terms but caching every term saves the log() recompute
	// when the same query repeats.
	for term, df := range idx.df {
		idx.idfCache[term] = bm25IDF(idx.totalDocs, df)
	}
	return idx
}

// bm25IDF uses the +0.5 / +0.5 / +1 smoothing common in Lucene/Elastic.
// Always >= 0 for our corpus sizes (the +1 in the denominator prevents
// going negative for terms that appear in every doc).
func bm25IDF(N, df int) float64 {
	return math.Log(1 + (float64(N-df)+0.5)/(float64(df)+0.5))
}

// Search returns the top-k chunks ranked by BM25. k<=0 → top 10.
func (idx *BM25Index) Search(query string, k int) []ScoredChunk {
	if k <= 0 {
		k = 10
	}
	qTokens := tokenizeBM25(query)
	if len(qTokens) == 0 || idx.totalDocs == 0 {
		return nil
	}

	scores := make([]float64, idx.totalDocs)
	for i, docTokens := range idx.docTokens {
		dl := float64(idx.docLen[i])
		for _, qt := range qTokens {
			idf, ok := idx.idfCache[qt]
			if !ok {
				continue
			}
			tf := termFrequency(docTokens, qt)
			if tf == 0 {
				continue
			}
			// BM25 contribution per term.
			num := tf * (bm25K1 + 1)
			den := tf + bm25K1*(1-bm25B+bm25B*(dl/idx.avgDocLen))
			scores[i] += idf * (num / den)
		}
	}

	// Collect non-zero, sort by score desc, take top-k.
	out := make([]ScoredChunk, 0, k)
	type idxScore struct {
		i int
		s float64
	}
	var ranked []idxScore
	for i, s := range scores {
		if s > 0 {
			ranked = append(ranked, idxScore{i: i, s: s})
		}
	}
	sort.Slice(ranked, func(a, b int) bool {
		if ranked[a].s != ranked[b].s {
			return ranked[a].s > ranked[b].s
		}
		// Tie-break: prefer earlier chunks (id ASC). Deterministic
		// ordering matters for tests + hybrid RRF stability.
		return idx.chunks[ranked[a].i].ID < idx.chunks[ranked[b].i].ID
	})
	if len(ranked) > k {
		ranked = ranked[:k]
	}
	for _, r := range ranked {
		out = append(out, ScoredChunk{
			Chunk:   idx.chunks[r.i],
			Score:   r.s,
			Channel: "bm25",
		})
	}
	return out
}

// termFrequency counts occurrences of t in tokens. Cheap linear scan —
// inverted-index would beat this asymptotically but the corpora are
// small (~10k chunks × ~500 tokens each = ~5M op worst case per
// Search), still well under 100ms on commodity hardware.
func termFrequency(tokens []string, t string) float64 {
	var n float64
	for _, x := range tokens {
		if x == t {
			n++
		}
	}
	return n
}

// tokenizeBM25 is the shared tokeniser for both index build + query.
// Lowercases, splits on whitespace / punctuation, drops 1-char tokens
// (mostly noise: 'a', 'I', etc.). Korean: keeps multi-rune words
// because unicode.IsLetter covers Hangul.
//
// Future work: language-aware tokenisation (Korean morpheme splitter)
// would improve recall on prose mixing Korean + English. Deferred to
// v0.10.x — BM25's bag-of-words approximation tolerates the simpler
// split well enough for v0.10.0.
func tokenizeBM25(s string) []string {
	var out []string
	var b strings.Builder
	flush := func() {
		if b.Len() == 0 {
			return
		}
		tok := strings.ToLower(b.String())
		b.Reset()
		// Skip 1-char tokens — overwhelmingly noise.
		if len([]rune(tok)) <= 1 {
			return
		}
		out = append(out, tok)
	}
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
			continue
		}
		flush()
	}
	flush()
	return out
}
