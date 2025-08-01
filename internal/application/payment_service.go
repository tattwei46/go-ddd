package application

import (
	"context"
	"fmt"

	"go-ddd/internal/domain/audit"
	"go-ddd/internal/domain/payment"
)

type PaymentApplicationService struct {
	paymentService *payment.Service
	auditService   *audit.Service
}

func NewPaymentApplicationService(paymentService *payment.Service, auditService *audit.Service) *PaymentApplicationService {
	return &PaymentApplicationService{
		paymentService: paymentService,
		auditService:   auditService,
	}
}

func (s *PaymentApplicationService) CreatePayment(ctx context.Context, amount float64, currency, description, userID string) (*payment.Payment, error) {
	amountVO, err := payment.NewAmount(amount, currency)
	if err != nil {
		return nil, fmt.Errorf("invalid amount: %w", err)
	}

	p, err := s.paymentService.CreatePayment(ctx, amountVO, description)
	if err != nil {
		return nil, fmt.Errorf("failed to create payment: %w", err)
	}

	paymentData := map[string]interface{}{
		"id":          p.ID().String(),
		"amount":      p.Amount().Value(),
		"currency":    p.Amount().Currency(),
		"description": p.Description(),
		"status":      p.Status().String(),
		"created_at":  p.CreatedAt(),
	}

	if err := s.auditService.RecordPaymentCreated(ctx, p.ID().String(), userID, paymentData); err != nil {
		return nil, fmt.Errorf("failed to record audit: %w", err)
	}

	return p, nil
}

func (s *PaymentApplicationService) ProcessPayment(ctx context.Context, paymentID string, userID string) error {
	id := payment.PaymentIDFromString(paymentID)

	p, err := s.paymentService.GetPayment(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get payment: %w", err)
	}

	oldStatus := p.Status().String()

	err = s.paymentService.ProcessPayment(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to process payment: %w", err)
	}

	if err := s.auditService.RecordPaymentStatusChange(ctx, paymentID, userID, oldStatus, "processing"); err != nil {
		return fmt.Errorf("failed to record audit: %w", err)
	}

	return nil
}

func (s *PaymentApplicationService) CompletePayment(ctx context.Context, paymentID string, userID string) error {
	id := payment.PaymentIDFromString(paymentID)

	p, err := s.paymentService.GetPayment(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get payment: %w", err)
	}

	oldStatus := p.Status().String()

	err = s.paymentService.CompletePayment(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to complete payment: %w", err)
	}

	if err := s.auditService.RecordPaymentStatusChange(ctx, paymentID, userID, oldStatus, "completed"); err != nil {
		return fmt.Errorf("failed to record audit: %w", err)
	}

	return nil
}

func (s *PaymentApplicationService) GetPaymentAuditHistory(ctx context.Context, paymentID string) ([]*audit.AuditEntry, error) {
	return s.auditService.GetAuditHistory(ctx, audit.EntityTypePayment, paymentID)
}
