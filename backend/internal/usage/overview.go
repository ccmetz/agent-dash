package usage

import (
	"context"
	"errors"
	"os"
	"time"
)

type Overview struct {
	Range  OverviewRange  `json:"range"`
	Totals OverviewTotals `json:"totals"`
	Daily  []DailyUsage   `json:"daily"`
}

type OverviewRange struct {
	Days  int    `json:"days"`
	Start string `json:"start"`
	End   string `json:"end"`
}

type OverviewTotals struct {
	Tokens        TokenTotals `json:"tokens"`
	ActualCost    float64     `json:"actualCost"`
	AgentSessions int         `json:"agentSessions"`
	ModelCalls    int         `json:"modelCalls"`
}

type TokenTotals struct {
	Total      int `json:"total"`
	Input      int `json:"input"`
	Output     int `json:"output"`
	Reasoning  int `json:"reasoning"`
	CacheRead  int `json:"cacheRead"`
	CacheWrite int `json:"cacheWrite"`
}

type DailyUsage struct {
	Date       string  `json:"date"`
	Tokens     int     `json:"tokens"`
	ActualCost float64 `json:"actualCost"`
}

func UsageOverview(ctx context.Context, storePath string, days int, end time.Time) (Overview, error) {
	if days <= 0 {
		days = 30
	}
	end = end.UTC()
	start := end.AddDate(0, 0, -days)
	overview := Overview{Range: OverviewRange{Days: days, Start: start.Format(time.RFC3339), End: end.Format(time.RFC3339)}, Daily: []DailyUsage{}}
	if _, err := os.Stat(storePath); errors.Is(err, os.ErrNotExist) {
		return overview, nil
	} else if err != nil {
		return overview, err
	}

	store, err := openAnalyticsStore(storePath)
	if err != nil {
		return overview, err
	}
	defer store.Close()
	if err := ensureAnalyticsStore(ctx, store); err != nil {
		return overview, err
	}

	startMillis := start.UnixMilli()
	endMillis := end.UnixMilli()
	if err := store.QueryRowContext(ctx, `
		select
			coalesce(sum(coalesce(total_tokens, input_tokens + output_tokens + reasoning_tokens + cache_read_tokens + cache_write_tokens)), 0),
			coalesce(sum(input_tokens), 0),
			coalesce(sum(output_tokens), 0),
			coalesce(sum(reasoning_tokens), 0),
			coalesce(sum(cache_read_tokens), 0),
			coalesce(sum(cache_write_tokens), 0),
			coalesce(sum(actual_cost), 0),
			count(*)
		from model_calls
		where cast(source_created_at as integer) >= ? and cast(source_created_at as integer) <= ?
	`, startMillis, endMillis).Scan(
		&overview.Totals.Tokens.Total,
		&overview.Totals.Tokens.Input,
		&overview.Totals.Tokens.Output,
		&overview.Totals.Tokens.Reasoning,
		&overview.Totals.Tokens.CacheRead,
		&overview.Totals.Tokens.CacheWrite,
		&overview.Totals.ActualCost,
		&overview.Totals.ModelCalls,
	); err != nil {
		return overview, err
	}
	if err := store.QueryRowContext(ctx, `
		select count(distinct session_source_id)
		from model_calls
		where cast(source_created_at as integer) >= ? and cast(source_created_at as integer) <= ?
	`, startMillis, endMillis).Scan(&overview.Totals.AgentSessions); err != nil {
		return overview, err
	}

	rows, err := store.QueryContext(ctx, `
		select
			date(cast(source_created_at as integer) / 1000, 'unixepoch') usage_date,
			coalesce(sum(coalesce(total_tokens, input_tokens + output_tokens + reasoning_tokens + cache_read_tokens + cache_write_tokens)), 0),
			coalesce(sum(actual_cost), 0)
		from model_calls
		where cast(source_created_at as integer) >= ? and cast(source_created_at as integer) <= ?
		group by usage_date
		order by usage_date
	`, startMillis, endMillis)
	if err != nil {
		return overview, err
	}
	defer rows.Close()
	for rows.Next() {
		var day DailyUsage
		if err := rows.Scan(&day.Date, &day.Tokens, &day.ActualCost); err != nil {
			return overview, err
		}
		overview.Daily = append(overview.Daily, day)
	}
	return overview, rows.Err()
}
