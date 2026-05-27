package usage

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"
)

func TestValidateOpenCodeSchemaAcceptsRequiredMetadataTablesAndColumns(t *testing.T) {
	ctx := context.Background()
	sourcePath := filepath.Join(t.TempDir(), "opencode.db")
	createOpenCodeFixture(t, sourcePath)

	source := openDB(t, sourcePath)
	defer source.Close()

	if err := ValidateOpenCodeSchema(ctx, source); err != nil {
		t.Fatalf("expected OpenCode schema to be supported: %v", err)
	}
}

func TestValidateOpenCodeSchemaReportsUnsupportedSchemaForMissingRequiredTable(t *testing.T) {
	ctx := context.Background()
	sourcePath := filepath.Join(t.TempDir(), "opencode.db")
	source := openDB(t, sourcePath)
	defer source.Close()
	execSQL(t, source, `create table project (id text primary key, worktree text not null, name text, time_created integer not null, time_updated integer not null)`)

	err := ValidateOpenCodeSchema(ctx, source)
	var schemaErr UnsupportedOpenCodeSchemaError
	if !errors.As(err, &schemaErr) {
		t.Fatalf("expected unsupported OpenCode schema error, got %v", err)
	}
	if schemaErr.Missing != "table session" {
		t.Fatalf("expected missing session table, got %q", schemaErr.Missing)
	}
}

func TestValidateOpenCodeSchemaReportsUnsupportedSchemaForMissingRequiredColumn(t *testing.T) {
	ctx := context.Background()
	sourcePath := filepath.Join(t.TempDir(), "opencode.db")
	source := openDB(t, sourcePath)
	defer source.Close()
	execSQL(t, source, `create table project (id text primary key, name text, time_created integer not null, time_updated integer not null)`)
	execSQL(t, source, `create table session (id text primary key, project_id text not null, title text not null, time_created integer not null, time_updated integer not null, time_archived integer)`)
	execSQL(t, source, `create table message (id text primary key, session_id text not null, time_created integer not null, time_updated integer not null, data text not null)`)

	err := ValidateOpenCodeSchema(ctx, source)
	var schemaErr UnsupportedOpenCodeSchemaError
	if !errors.As(err, &schemaErr) {
		t.Fatalf("expected unsupported OpenCode schema error, got %v", err)
	}
	if schemaErr.Missing != "column project.worktree" {
		t.Fatalf("expected missing project.worktree column, got %q", schemaErr.Missing)
	}
}

func TestOpenCodeUsageSyncStoresProjectsAgentSessionsAndModelCalls(t *testing.T) {
	ctx := context.Background()
	sourcePath := filepath.Join(t.TempDir(), "opencode.db")
	storePath := filepath.Join(t.TempDir(), "analytics.db")
	createOpenCodeFixture(t, sourcePath)

	if err := SyncOpenCode(ctx, sourcePath, storePath); err != nil {
		t.Fatalf("expected Usage Sync to succeed: %v", err)
	}

	store := openDB(t, storePath)
	defer store.Close()

	var projectName, projectPath, projectSourceUpdated string
	if err := store.QueryRowContext(ctx, `select name, path, source_updated_at from projects where source = 'opencode' and source_id = 'proj_1'`).Scan(&projectName, &projectPath, &projectSourceUpdated); err != nil {
		t.Fatalf("expected synced Project: %v", err)
	}
	if projectName != "Agent Dash" || projectPath != "/work/agent-dash" || projectSourceUpdated != "2000" {
		t.Fatalf("expected Project metadata, got name=%q path=%q source_updated_at=%q", projectName, projectPath, projectSourceUpdated)
	}

	var sessionTitle, sessionStatus, sessionSourceUpdated string
	if err := store.QueryRowContext(ctx, `select title, status, source_updated_at from agent_sessions where source = 'opencode' and source_id = 'ses_1'`).Scan(&sessionTitle, &sessionStatus, &sessionSourceUpdated); err != nil {
		t.Fatalf("expected synced Agent Session: %v", err)
	}
	if sessionTitle != "Implement usage sync" || sessionStatus != "active" || sessionSourceUpdated != "4000" {
		t.Fatalf("expected Agent Session metadata, got title=%q status=%q source_updated_at=%q", sessionTitle, sessionStatus, sessionSourceUpdated)
	}

	var provider, model, status string
	var inputTokens, outputTokens, reasoningTokens, cacheReadTokens, cacheWriteTokens int
	var actualCost float64
	if err := store.QueryRowContext(ctx, `select provider, model, status, input_tokens, output_tokens, reasoning_tokens, cache_read_tokens, cache_write_tokens, actual_cost from model_calls where source = 'opencode' and source_id = 'msg_1'`).Scan(&provider, &model, &status, &inputTokens, &outputTokens, &reasoningTokens, &cacheReadTokens, &cacheWriteTokens, &actualCost); err != nil {
		t.Fatalf("expected synced Model Call: %v", err)
	}
	if provider != "opencode" || model != "claude-opus-4-5" || status != "stop" || inputTokens != 10 || outputTokens != 20 || reasoningTokens != 3 || cacheReadTokens != 4 || cacheWriteTokens != 5 || actualCost != 0.25 {
		t.Fatalf("expected Model Call Usage Metadata, got provider=%q model=%q status=%q tokens=%d/%d/%d/%d/%d cost=%f", provider, model, status, inputTokens, outputTokens, reasoningTokens, cacheReadTokens, cacheWriteTokens, actualCost)
	}
}

func TestOpenCodeUsageSyncUpsertsRepeatedSyncsAndChangedSourceRecords(t *testing.T) {
	ctx := context.Background()
	sourcePath := filepath.Join(t.TempDir(), "opencode.db")
	storePath := filepath.Join(t.TempDir(), "analytics.db")
	createOpenCodeFixture(t, sourcePath)

	if err := SyncOpenCode(ctx, sourcePath, storePath); err != nil {
		t.Fatalf("expected initial Usage Sync to succeed: %v", err)
	}
	if err := SyncOpenCode(ctx, sourcePath, storePath); err != nil {
		t.Fatalf("expected repeated Usage Sync to succeed: %v", err)
	}

	store := openDB(t, storePath)
	defer store.Close()
	assertCount(t, store, "agent_sessions", 1)
	assertCount(t, store, "model_calls", 1)

	source := openDB(t, sourcePath)
	defer source.Close()
	execSQL(t, source, `update session set title = 'Updated usage sync', time_updated = 7000 where id = 'ses_1'`)
	execSQL(t, source, `update message set time_updated = 8000, data = '{"role":"assistant","modelID":"claude-sonnet-4-5","providerID":"opencode","cost":0.5,"tokens":{"input":11,"output":21,"reasoning":4,"cache":{"read":5,"write":6}},"finish":"tool-calls"}' where id = 'msg_1'`)

	if err := SyncOpenCode(ctx, sourcePath, storePath); err != nil {
		t.Fatalf("expected changed source Usage Sync to succeed: %v", err)
	}
	assertCount(t, store, "agent_sessions", 1)
	assertCount(t, store, "model_calls", 1)

	var title, sessionSourceUpdated string
	if err := store.QueryRowContext(ctx, `select title, source_updated_at from agent_sessions where source_id = 'ses_1'`).Scan(&title, &sessionSourceUpdated); err != nil {
		t.Fatalf("expected updated Agent Session: %v", err)
	}
	if title != "Updated usage sync" || sessionSourceUpdated != "7000" {
		t.Fatalf("expected changed Agent Session metadata, got title=%q source_updated_at=%q", title, sessionSourceUpdated)
	}

	var model, callSourceUpdated string
	var actualCost float64
	if err := store.QueryRowContext(ctx, `select model, actual_cost, source_updated_at from model_calls where source_id = 'msg_1'`).Scan(&model, &actualCost, &callSourceUpdated); err != nil {
		t.Fatalf("expected updated Model Call: %v", err)
	}
	if model != "claude-sonnet-4-5" || actualCost != 0.5 || callSourceUpdated != "8000" {
		t.Fatalf("expected changed Model Call metadata, got model=%q actual_cost=%f source_updated_at=%q", model, actualCost, callSourceUpdated)
	}
}

func TestOpenCodeUsageSyncIncludesFailedAndAbortedModelCallsWithUsage(t *testing.T) {
	ctx := context.Background()
	sourcePath := filepath.Join(t.TempDir(), "opencode.db")
	storePath := filepath.Join(t.TempDir(), "analytics.db")
	createOpenCodeFixture(t, sourcePath)

	source := openDB(t, sourcePath)
	defer source.Close()
	execSQL(t, source, `insert into message (id, session_id, time_created, time_updated, data) values ('msg_failed', 'ses_1', 7000, 7000, '{"role":"assistant","modelID":"claude-opus-4-5","providerID":"opencode","cost":0,"tokens":{"input":1,"output":0,"reasoning":0,"cache":{"read":0,"write":0}},"finish":"error"}')`)
	execSQL(t, source, `insert into message (id, session_id, time_created, time_updated, data) values ('msg_aborted', 'ses_1', 8000, 8000, '{"role":"assistant","modelID":"claude-opus-4-5","providerID":"opencode","cost":0.1,"tokens":{"input":0,"output":0,"reasoning":0,"cache":{"read":0,"write":0}},"finish":"abort"}')`)
	execSQL(t, source, `insert into message (id, session_id, time_created, time_updated, data) values ('msg_empty_error', 'ses_1', 9000, 9000, '{"role":"assistant","modelID":"claude-opus-4-5","providerID":"opencode","finish":"error"}')`)

	if err := SyncOpenCode(ctx, sourcePath, storePath); err != nil {
		t.Fatalf("expected Usage Sync to succeed: %v", err)
	}

	store := openDB(t, storePath)
	defer store.Close()
	assertCount(t, store, "model_calls", 3)
}

func TestOpenCodeUsageSyncStoresMetadataOnly(t *testing.T) {
	ctx := context.Background()
	sourcePath := filepath.Join(t.TempDir(), "opencode.db")
	storePath := filepath.Join(t.TempDir(), "analytics.db")
	createOpenCodeFixture(t, sourcePath)

	source := openDB(t, sourcePath)
	defer source.Close()
	execSQL(t, source, `create table account (id text primary key, email text not null, access_token text not null, refresh_token text not null)`)
	execSQL(t, source, `insert into account (id, email, access_token, refresh_token) values ('acct_1', 'secret@example.com', 'access-secret', 'refresh-secret')`)
	execSQL(t, source, `update message set data = '{"role":"assistant","modelID":"claude-opus-4-5","providerID":"opencode","cost":0.25,"tokens":{"input":10,"output":20,"reasoning":3,"cache":{"read":4,"write":5}},"finish":"stop","text":"assistant response secret","parts":[{"type":"tool","input":"tool input secret","output":"tool output secret"}],"path":{"cwd":"/work/agent-dash"}}' where id = 'msg_1'`)

	if err := SyncOpenCode(ctx, sourcePath, storePath); err != nil {
		t.Fatalf("expected Usage Sync to succeed: %v", err)
	}

	store := openDB(t, storePath)
	defer store.Close()
	for _, table := range []string{"projects", "agent_sessions", "model_calls"} {
		rows, err := store.QueryContext(ctx, `select name from pragma_table_info(?)`, table)
		if err != nil {
			t.Fatalf("expected Analytics Store table info for %s: %v", table, err)
		}
		for rows.Next() {
			var column string
			if err := rows.Scan(&column); err != nil {
				t.Fatalf("expected column name: %v", err)
			}
			if column == "data" || column == "prompt" || column == "response" || column == "tool_input" || column == "tool_output" || column == "access_token" || column == "refresh_token" {
				t.Fatalf("expected metadata-only Analytics Store, found column %q on %s", column, table)
			}
		}
		if err := rows.Close(); err != nil {
			t.Fatalf("expected table info rows to close: %v", err)
		}
	}
}

func createOpenCodeFixture(t *testing.T, path string) {
	t.Helper()
	db := openDB(t, path)
	defer db.Close()

	execSQL(t, db, `create table project (id text primary key, worktree text not null, name text, time_created integer not null, time_updated integer not null)`)
	execSQL(t, db, `create table session (id text primary key, project_id text not null, title text not null, time_created integer not null, time_updated integer not null, time_archived integer)`)
	execSQL(t, db, `create table message (id text primary key, session_id text not null, time_created integer not null, time_updated integer not null, data text not null)`)
	execSQL(t, db, `insert into project (id, worktree, name, time_created, time_updated) values ('proj_1', '/work/agent-dash', 'Agent Dash', 1000, 2000)`)
	execSQL(t, db, `insert into session (id, project_id, title, time_created, time_updated, time_archived) values ('ses_1', 'proj_1', 'Implement usage sync', 3000, 4000, null)`)
	execSQL(t, db, `insert into message (id, session_id, time_created, time_updated, data) values ('msg_1', 'ses_1', 5000, 6000, '{"role":"assistant","modelID":"claude-opus-4-5","providerID":"opencode","cost":0.25,"tokens":{"input":10,"output":20,"reasoning":3,"cache":{"read":4,"write":5}},"finish":"stop"}')`)
}

func openDB(t *testing.T, path string) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("failed to open sqlite database: %v", err)
	}
	return db
}

func execSQL(t *testing.T, db *sql.DB, statement string) {
	t.Helper()
	if _, err := db.Exec(statement); err != nil {
		t.Fatalf("failed to execute SQL %q: %v", statement, err)
	}
}

func assertCount(t *testing.T, db *sql.DB, table string, expected int) {
	t.Helper()
	var actual int
	if err := db.QueryRow(`select count(*) from ` + table).Scan(&actual); err != nil {
		t.Fatalf("expected %s count: %v", table, err)
	}
	if actual != expected {
		t.Fatalf("expected %s count %d, got %d", table, expected, actual)
	}
}
