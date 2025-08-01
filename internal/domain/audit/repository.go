package audit

import "context"

type Repository interface {
	Save(ctx context.Context, entry *AuditEntry) error
	FindByID(ctx context.Context, id AuditID) (*AuditEntry, error)
	FindByFilter(ctx context.Context, filter AuditFilter) ([]*AuditEntry, error)
	FindByEntityID(ctx context.Context, entityType EntityType, entityID string) ([]*AuditEntry, error)
}