package analytics

import (
	"context"
	"fmt"
	"sort"
	"strings"
)
func (a *SQLAdapter) QueryFeedbackCorpus(ctx context.Context, q FeedbackQuery) (FeedbackResult, error) {
	from, to := rangeArgs(q.TimeRange)

	clauses := []string{"occurred_at >= ?", "occurred_at < ?"}
	args := []any{from, to}
	if q.Source != "" {
		clauses = append(clauses, "source = ?")
		args = append(args, string(q.Source))
	}
	whereSQL := strings.Join(clauses, " AND ")

	var total int64
	if err := a.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM feedback_items WHERE "+whereSQL, args...,
	).Scan(&total); err != nil {
		return FeedbackResult{}, fmt.Errorf("analytics: feedback total: %w", err)
	}

	res := FeedbackResult{TotalItems: total}

	if q.TopicModeling {
		topics, err := a.queryFeedbackTopics(ctx, whereSQL, args, q.SentimentAnalysis)
		if err != nil {
			return FeedbackResult{}, err
		}
		res.Topics = topics
	}

	if q.Source == FeedbackNPSComment || q.Source == "" {
		bands, err := a.queryNPSBands(ctx, whereSQL, args)
		if err != nil {
			return FeedbackResult{}, err
		}
		if len(bands) > 0 {
			res.NPSSegments = bands
		}
	}
	return res, nil
}

func (a *SQLAdapter) queryFeedbackTopics(ctx context.Context, whereSQL string, args []any, withSentiment bool) ([]FeedbackTopic, error) {
	rows, err := a.db.QueryContext(ctx,
		"SELECT COALESCE(topic, '(unclassified)'), COALESCE(sentiment, ''), body FROM feedback_items WHERE "+whereSQL,
		args...)
	if err != nil {
		return nil, fmt.Errorf("analytics: feedback topics: %w", err)
	}
	defer rows.Close()

	type acc struct {
		count     int64
		sentiment map[string]int64
		quotes    []string
	}
	by := map[string]*acc{}
	for rows.Next() {
		var topic, sent, body string
		if err := rows.Scan(&topic, &sent, &body); err != nil {
			return nil, err
		}
		a := by[topic]
		if a == nil {
			a = &acc{sentiment: map[string]int64{}}
			by[topic] = a
		}
		a.count++
		if withSentiment && sent != "" {
			a.sentiment[sent]++
		}
		if len(a.quotes) < 3 && body != "" {
			a.quotes = append(a.quotes, body)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	topics := make([]FeedbackTopic, 0, len(by))
	for name, info := range by {
		t := FeedbackTopic{Topic: name, ItemCount: info.count, VerbatimQuotes: info.quotes}
		if withSentiment && len(info.sentiment) > 0 {
			t.Sentiment = info.sentiment
		}
		topics = append(topics, t)
	}
	sort.Slice(topics, func(i, j int) bool { return topics[i].ItemCount > topics[j].ItemCount })
	return topics, nil
}

func (a *SQLAdapter) queryNPSBands(ctx context.Context, whereSQL string, args []any) (map[string]NPSBand, error) {
	rows, err := a.db.QueryContext(ctx,
		"SELECT nps_score, COALESCE(topic, '(unclassified)') FROM feedback_items WHERE "+whereSQL+" AND nps_score IS NOT NULL",
		args...)
	if err != nil {
		return nil, fmt.Errorf("analytics: feedback nps bands: %w", err)
	}
	defer rows.Close()

	bands := map[string]*NPSBand{
		"promoter":  {TopTopics: []string{}},
		"passive":   {TopTopics: []string{}},
		"detractor": {TopTopics: []string{}},
	}
	topicCount := map[string]map[string]int64{
		"promoter":  {},
		"passive":   {},
		"detractor": {},
	}
	for rows.Next() {
		var score int64
		var topic string
		if err := rows.Scan(&score, &topic); err != nil {
			return nil, err
		}
		band := classifyNPS(score)
		bands[band].Count++
		topicCount[band][topic]++
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	out := map[string]NPSBand{}
	hadAny := false
	for band, b := range bands {
		if b.Count == 0 {
			continue
		}
		hadAny = true
		type kv struct {
			topic string
			n     int64
		}
		ranked := make([]kv, 0, len(topicCount[band]))
		for t, n := range topicCount[band] {
			ranked = append(ranked, kv{t, n})
		}
		sort.Slice(ranked, func(i, j int) bool { return ranked[i].n > ranked[j].n })
		top := make([]string, 0, 3)
		for i, r := range ranked {
			if i >= 3 {
				break
			}
			top = append(top, r.topic)
		}
		out[band] = NPSBand{Count: b.Count, TopTopics: top}
	}
	if !hadAny {
		return nil, nil
	}
	return out, nil
}

// classifyNPS maps a 0-10 NPS score to the standard 3-band split.
func classifyNPS(score int64) string {
	switch {
	case score >= 9:
		return "promoter"
	case score >= 7:
		return "passive"
	default:
		return "detractor"
	}
}

// ─── tiny stats helpers ────────────────────────────────────────────────────
