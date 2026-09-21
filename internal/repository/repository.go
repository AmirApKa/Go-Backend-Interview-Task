package repository

import (
	"context"
	"sync"
	"time"
)

type AuditLog struct {
	ID           int64
	MaskedCard   string
	Status       string
	ResponseData string
	CreatedAt    time.Time
}

type Repository interface {
	SaveAuditLog(ctx context.Context, log *AuditLog) error
}

type MemoryRepository struct {
	mu   sync.Mutex
	logs []AuditLog
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		logs: make([]AuditLog, 0),
	}
}

func (r *MemoryRepository) SaveAuditLog(ctx context.Context, log *AuditLog) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	log.ID = int64(len(r.logs) + 1)
	log.CreatedAt = time.Now()
	r.logs = append(r.logs, *log)
	return nil
}
