package repository

import (
	"context"
	"database/sql"
	"sync"
	"time"
)

// AuditLog یک رکورد از یک تلاش استعلام (موفق یا ناموفق) است.
type AuditLog struct {
	ID                 int64
	RequestID          string
	MaskedCard         string
	Status             string // SUCCESS, VALIDATION_ERROR, EXTERNAL_ERROR, INTERNAL_ERROR
	ExternalHTTPStatus sql.NullInt64
	ExternalResponse   sql.NullString
	ErrorMessage       sql.NullString
	ClientIP           string
	DurationMs         int64
	CreatedAt          time.Time
}

type Repository interface {
	SaveAuditLog(ctx context.Context, log *AuditLog) error
}

// MemoryRepository پیاده‌سازی درون‌حافظه‌ای Repository است؛
// فقط برای تست‌های واحد نگه داشته شده تا لازم نباشه در تست به پایگاه داده واقعی وصل شویم.
type MemoryRepository struct {
	mu   sync.Mutex
	Logs []AuditLog
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{Logs: make([]AuditLog, 0)}
}

func (r *MemoryRepository) SaveAuditLog(ctx context.Context, log *AuditLog) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	log.ID = int64(len(r.Logs) + 1)
	log.CreatedAt = time.Now()
	r.Logs = append(r.Logs, *log)
	return nil
}
