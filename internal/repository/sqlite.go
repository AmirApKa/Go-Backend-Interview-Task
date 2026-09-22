package repository

import (
	"context"
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

const schema = `
CREATE TABLE IF NOT EXISTS audit_logs (
    id                    INTEGER PRIMARY KEY AUTOINCREMENT,
    request_id            TEXT NOT NULL UNIQUE,
    masked_card           TEXT NOT NULL,
    status                TEXT NOT NULL,
    external_http_status  INTEGER,
    external_response     TEXT,
    error_message         TEXT,
    client_ip             TEXT,
    duration_ms           INTEGER NOT NULL,
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON audit_logs(created_at);
`

type SQLiteRepository struct {
	db *sql.DB
}

func NewSQLiteRepository(dbPath string) (*SQLiteRepository, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open sqlite db: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping sqlite db: %w", err)
	}
	if _, err := db.Exec(schema); err != nil {
		return nil, fmt.Errorf("run migration: %w", err)
	}
	return &SQLiteRepository{db: db}, nil
}

func (r *SQLiteRepository) Close() error {
	return r.db.Close()
}

func (r *SQLiteRepository) SaveAuditLog(ctx context.Context, log *AuditLog) error {
	query := `
        INSERT INTO audit_logs
            (request_id, masked_card, status, external_http_status, external_response, error_message, client_ip, duration_ms)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?)
    `
	result, err := r.db.ExecContext(ctx, query,
		log.RequestID,
		log.MaskedCard,
		log.Status,
		log.ExternalHTTPStatus,
		log.ExternalResponse,
		log.ErrorMessage,
		log.ClientIP,
		log.DurationMs,
	)
	if err != nil {
		return fmt.Errorf("insert audit log: %w", err)
	}
	if id, err := result.LastInsertId(); err == nil {
		log.ID = id
	}
	return nil
}
