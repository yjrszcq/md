package middleware

import (
	"fmt"
	"time"
)

// Migration represents a database migration
type Migration struct {
	Version     int
	Description string
	SQL         string
}

// All migrations in order - append new migrations to this list
var migrations = []Migration{
	{
		Version:     1,
		Description: "Add AI config and conversation tables",
		SQL: `
CREATE TABLE IF NOT EXISTS t_ai_config
(
	id varchar(50) PRIMARY KEY NOT NULL,
	user_id varchar(50) NOT NULL,
	base_url text NOT NULL DEFAULT '',
	api_key text NOT NULL DEFAULT '',
	model text NOT NULL DEFAULT '',
	system_prompts text NOT NULL DEFAULT '[]',
	current_prompt_id varchar(50) NOT NULL DEFAULT '',
	system_prompt_enabled boolean NOT NULL DEFAULT false,
	doc_context_enabled boolean NOT NULL DEFAULT false,
	panel_enabled boolean NOT NULL DEFAULT false,
	create_time bigint NOT NULL,
	update_time bigint NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS "ai_config_user_id"
ON "t_ai_config" (
  "user_id" ASC
);

CREATE TABLE IF NOT EXISTS t_ai_conversation
(
	id varchar(50) PRIMARY KEY NOT NULL,
	user_id varchar(50) NOT NULL,
	title text NOT NULL DEFAULT '新对话',
	content text NOT NULL DEFAULT '[]',
	create_time bigint NOT NULL,
	update_time bigint NOT NULL
);

CREATE INDEX IF NOT EXISTS "ai_conversation_user_id"
ON "t_ai_conversation" (
  "user_id" ASC
);
`,
	},
}

// RunMigrations checks and applies pending database migrations
func RunMigrations() error {
	// Ensure version table exists
	_, err := Db.Exec(`
		CREATE TABLE IF NOT EXISTS t_db_version (
			version int PRIMARY KEY NOT NULL,
			applied_at bigint NOT NULL
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create version table: %w", err)
	}

	// Get current version
	var dbVersion int
	err = Db.Get(&dbVersion, "SELECT COALESCE(MAX(version), 0) FROM t_db_version")
	if err != nil {
		return fmt.Errorf("failed to get db version: %w", err)
	}

	// Apply pending migrations
	for _, m := range migrations {
		if m.Version <= dbVersion {
			continue
		}

		Log.Info(fmt.Sprintf("Applying migration %d: %s", m.Version, m.Description))

		_, err = Db.Exec(m.SQL)
		if err != nil {
			return fmt.Errorf("migration %d failed: %w", m.Version, err)
		}

		// Record migration
		_, err = Db.Exec(
			"INSERT INTO t_db_version (version, applied_at) VALUES ($1, $2)",
			m.Version,
			time.Now().UnixMilli(),
		)
		if err != nil {
			return fmt.Errorf("failed to record migration %d: %w", m.Version, err)
		}

		Log.Info(fmt.Sprintf("Migration %d applied successfully", m.Version))
	}

	return nil
}
