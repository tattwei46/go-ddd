package repository

import (
	"context"
	"errors"
	"sync"

	"go-ddd/internal/domain/audit"
)

type AuditMemoryRepository struct {
	mu      sync.RWMutex
	entries map[string]*audit.AuditEntry
}

func NewAuditMemoryRepository() *AuditMemoryRepository {
	return &AuditMemoryRepository{
		entries: make(map[string]*audit.AuditEntry),
	}
}

func (r *AuditMemoryRepository) Save(ctx context.Context, entry *audit.AuditEntry) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.entries[entry.ID().String()] = entry
	return nil
}

func (r *AuditMemoryRepository) FindByID(ctx context.Context, id audit.AuditID) (*audit.AuditEntry, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	entry, exists := r.entries[id.String()]
	if !exists {
		return nil, errors.New("audit entry not found")
	}

	return entry, nil
}

func (r *AuditMemoryRepository) FindByEntityID(ctx context.Context, entityType audit.EntityType, entityID string) ([]*audit.AuditEntry, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*audit.AuditEntry
	for _, entry := range r.entries {
		if entry.EntityType() == entityType && entry.EntityID() == entityID {
			result = append(result, entry)
		}
	}

	return result, nil
}

func (r *AuditMemoryRepository) FindByFilter(ctx context.Context, filter audit.AuditFilter) ([]*audit.AuditEntry, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*audit.AuditEntry
	for _, entry := range r.entries {
		if r.matchesFilter(entry, filter) {
			result = append(result, entry)
		}
	}

	return result, nil
}

func (r *AuditMemoryRepository) matchesFilter(entry *audit.AuditEntry, filter audit.AuditFilter) bool {
	if filter.EntityType != nil && entry.EntityType() != *filter.EntityType {
		return false
	}

	if filter.EntityID != nil && entry.EntityID() != *filter.EntityID {
		return false
	}

	if filter.Action != nil && entry.Action() != *filter.Action {
		return false
	}

	if filter.UserID != nil && entry.UserID() != *filter.UserID {
		return false
	}

	if filter.FromDate != nil && entry.Timestamp().Before(*filter.FromDate) {
		return false
	}

	if filter.ToDate != nil && entry.Timestamp().After(*filter.ToDate) {
		return false
	}

	return true
}
