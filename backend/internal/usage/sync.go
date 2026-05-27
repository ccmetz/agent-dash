package usage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"

	_ "modernc.org/sqlite"
)

const openCodeSource = "opencode"

const selectOpenCodeProjectsSQL = `
	select
		id,
		coalesce(name, ''),
		worktree,
		time_created,
		time_updated
	from project
`

const upsertProjectSQL = `
	insert into projects (
		source,
		source_id,
		name,
		path,
		source_created_at,
		source_updated_at
	) values (?, ?, ?, ?, ?, ?)
	on conflict(source, source_id) do update set
		name = excluded.name,
		path = excluded.path,
		source_created_at = excluded.source_created_at,
		source_updated_at = excluded.source_updated_at
`

const selectOpenCodeAgentSessionsSQL = `
	select
		id,
		project_id,
		title,
		time_created,
		time_updated,
		time_archived
	from session
`

const upsertAgentSessionSQL = `
	insert into agent_sessions (
		project_source_id,
		source,
		source_id,
		title,
		status,
		source_created_at,
		source_updated_at
	) values (?, ?, ?, ?, ?, ?, ?)
	on conflict(source, source_id) do update set
		project_source_id = excluded.project_source_id,
		title = excluded.title,
		status = excluded.status,
		source_created_at = excluded.source_created_at,
		source_updated_at = excluded.source_updated_at
`

const selectOpenCodeModelCallsSQL = `
	select
		id,
		session_id,
		time_created,
		time_updated,
		data
	from message
`

const upsertModelCallSQL = `
	insert into model_calls (
		session_source_id,
		source,
		source_id,
		provider,
		model,
		status,
		input_tokens,
		output_tokens,
		reasoning_tokens,
		cache_read_tokens,
		cache_write_tokens,
		actual_cost,
		source_created_at,
		source_updated_at
	) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	on conflict(source, source_id) do update set
		session_source_id = excluded.session_source_id,
		provider = excluded.provider,
		model = excluded.model,
		status = excluded.status,
		input_tokens = excluded.input_tokens,
		output_tokens = excluded.output_tokens,
		reasoning_tokens = excluded.reasoning_tokens,
		cache_read_tokens = excluded.cache_read_tokens,
		cache_write_tokens = excluded.cache_write_tokens,
		actual_cost = excluded.actual_cost,
		source_created_at = excluded.source_created_at,
		source_updated_at = excluded.source_updated_at
`

func SyncOpenCode(ctx context.Context, sourcePath, storePath string) error {
	log.Printf("starting OpenCode Usage Sync source=%q analytics_store=%q", sourcePath, storePath)

	source, err := sql.Open("sqlite", sourcePath)
	if err != nil {
		return err
	}
	defer source.Close()

	store, err := sql.Open("sqlite", storePath)
	if err != nil {
		return err
	}
	defer store.Close()

	if err := ensureAnalyticsStore(ctx, store); err != nil {
		return err
	}
	projectCount, err := syncProjects(ctx, source, store)
	if err != nil {
		return err
	}
	log.Printf("synced OpenCode Projects count=%d", projectCount)

	agentSessionCount, err := syncAgentSessions(ctx, source, store)
	if err != nil {
		return err
	}
	log.Printf("synced OpenCode Agent Sessions count=%d", agentSessionCount)

	modelCallCount, err := syncModelCalls(ctx, source, store)
	if err != nil {
		return err
	}
	log.Printf("synced OpenCode Model Calls count=%d", modelCallCount)
	log.Printf("finished OpenCode Usage Sync source=%q analytics_store=%q", sourcePath, storePath)
	return nil
}

func ensureAnalyticsStore(ctx context.Context, db *sql.DB) error {
	statements := []string{
		`create table if not exists projects (
			id integer primary key,
			source text not null,
			source_id text not null,
			name text not null,
			path text not null,
			source_created_at text not null,
			source_updated_at text not null,
			unique(source, source_id)
		)`,
		`create table if not exists agent_sessions (
			id integer primary key,
			project_source_id text not null,
			source text not null,
			source_id text not null,
			title text not null,
			status text not null,
			source_created_at text not null,
			source_updated_at text not null,
			unique(source, source_id)
		)`,
		`create table if not exists model_calls (
			id integer primary key,
			session_source_id text not null,
			source text not null,
			source_id text not null,
			provider text not null,
			model text not null,
			status text not null,
			input_tokens integer not null,
			output_tokens integer not null,
			reasoning_tokens integer not null,
			cache_read_tokens integer not null,
			cache_write_tokens integer not null,
			actual_cost real not null,
			source_created_at text not null,
			source_updated_at text not null,
			unique(source, source_id)
		)`,
	}
	for _, statement := range statements {
		if _, err := db.ExecContext(ctx, statement); err != nil {
			return err
		}
	}
	return nil
}

func syncProjects(ctx context.Context, source, store *sql.DB) (int, error) {
	rows, err := source.QueryContext(ctx, selectOpenCodeProjectsSQL)
	if err != nil {
		return 0, fmt.Errorf("read OpenCode projects: %w", err)
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var id, name, path string
		var created, updated int64
		if err := rows.Scan(&id, &name, &path, &created, &updated); err != nil {
			return 0, err
		}
		if _, err := store.ExecContext(
			ctx,
			upsertProjectSQL,
			openCodeSource,
			id,
			name,
			path,
			fmt.Sprint(created),
			fmt.Sprint(updated),
		); err != nil {
			return 0, err
		}
		count++
	}
	return count, rows.Err()
}

func syncAgentSessions(ctx context.Context, source, store *sql.DB) (int, error) {
	rows, err := source.QueryContext(ctx, selectOpenCodeAgentSessionsSQL)
	if err != nil {
		return 0, fmt.Errorf("read OpenCode sessions: %w", err)
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var id, projectID, title string
		var created, updated int64
		var archived sql.NullInt64
		if err := rows.Scan(&id, &projectID, &title, &created, &updated, &archived); err != nil {
			return 0, err
		}
		status := "active"
		if archived.Valid {
			status = "archived"
		}
		if _, err := store.ExecContext(
			ctx,
			upsertAgentSessionSQL,
			projectID,
			openCodeSource,
			id,
			title,
			status,
			fmt.Sprint(created),
			fmt.Sprint(updated),
		); err != nil {
			return 0, err
		}
		count++
	}
	return count, rows.Err()
}

type messageData struct {
	Role       string  `json:"role"`
	ModelID    string  `json:"modelID"`
	ProviderID string  `json:"providerID"`
	Cost       float64 `json:"cost"`
	Finish     string  `json:"finish"`
	Tokens     struct {
		Input     int `json:"input"`
		Output    int `json:"output"`
		Reasoning int `json:"reasoning"`
		Cache     struct {
			Read  int `json:"read"`
			Write int `json:"write"`
		} `json:"cache"`
	} `json:"tokens"`
}

func syncModelCalls(ctx context.Context, source, store *sql.DB) (int, error) {
	rows, err := source.QueryContext(ctx, selectOpenCodeModelCallsSQL)
	if err != nil {
		return 0, fmt.Errorf("read OpenCode messages: %w", err)
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var id, sessionID, data string
		var created, updated int64
		if err := rows.Scan(&id, &sessionID, &created, &updated, &data); err != nil {
			return 0, err
		}

		var message messageData
		if err := json.Unmarshal([]byte(data), &message); err != nil {
			return 0, err
		}
		if message.Role != "assistant" || !messageHasUsage(message) {
			continue
		}

		if _, err := store.ExecContext(
			ctx,
			upsertModelCallSQL,
			sessionID,
			openCodeSource,
			id,
			message.ProviderID,
			message.ModelID,
			message.Finish,
			message.Tokens.Input,
			message.Tokens.Output,
			message.Tokens.Reasoning,
			message.Tokens.Cache.Read,
			message.Tokens.Cache.Write,
			message.Cost,
			fmt.Sprint(created),
			fmt.Sprint(updated),
		); err != nil {
			return 0, err
		}
		count++
	}
	return count, rows.Err()
}

func messageHasUsage(message messageData) bool {
	return message.Cost != 0 ||
		message.Tokens.Input != 0 ||
		message.Tokens.Output != 0 ||
		message.Tokens.Reasoning != 0 ||
		message.Tokens.Cache.Read != 0 ||
		message.Tokens.Cache.Write != 0
}
