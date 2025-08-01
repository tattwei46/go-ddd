package audit

import (
	"context"
	"errors"
)

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{
		repository: repository,
	}
}

func (s *Service) RecordAction(ctx context.Context, entityType EntityType, entityID string, action ActionType, userID string, oldData, newData interface{}) error {
	entry := NewAuditEntry(entityType, entityID, action, userID)
	
	if oldData != nil {
		if err := entry.SetOldData(oldData); err != nil {
			return err
		}
	}
	
	if newData != nil {
		if err := entry.SetNewData(newData); err != nil {
			return err
		}
	}
	
	return s.repository.Save(ctx, entry)
}

func (s *Service) GetAuditEntry(ctx context.Context, id AuditID) (*AuditEntry, error) {
	return s.repository.FindByID(ctx, id)
}

func (s *Service) GetAuditHistory(ctx context.Context, entityType EntityType, entityID string) ([]*AuditEntry, error) {
	return s.repository.FindByEntityID(ctx, entityType, entityID)
}

func (s *Service) GetAuditsByFilter(ctx context.Context, filter AuditFilter) ([]*AuditEntry, error) {
	return s.repository.FindByFilter(ctx, filter)
}

func (s *Service) RecordPaymentCreated(ctx context.Context, paymentID string, userID string, paymentData interface{}) error {
	return s.RecordAction(ctx, EntityTypePayment, paymentID, ActionTypeCreated, userID, nil, paymentData)
}

func (s *Service) RecordPaymentStatusChange(ctx context.Context, paymentID string, userID string, oldStatus, newStatus interface{}) error {
	var action ActionType
	
	switch newStatus {
	case "processing":
		action = ActionTypeProcessed
	case "completed":
		action = ActionTypeCompleted
	case "failed":
		action = ActionTypeFailed
	case "cancelled":
		action = ActionTypeCancelled
	default:
		return errors.New("unknown payment status")
	}
	
	oldData := map[string]interface{}{"status": oldStatus}
	newData := map[string]interface{}{"status": newStatus}
	
	return s.RecordAction(ctx, EntityTypePayment, paymentID, action, userID, oldData, newData)
}