package application

import (
	"context"
	"errors"
	"testing"

	"go-ddd/internal/domain/audit"
	"go-ddd/internal/domain/payment"
)

func TestPaymentApplicationService_CreatePayment(t *testing.T) {
	tests := []struct {
		name        string
		amount      float64
		currency    string
		description string
		userID      string
		wantErr     bool
	}{
		{
			name:        "successful payment creation",
			amount:      100.50,
			currency:    "USD",
			description: "Test payment",
			userID:      "user-123",
			wantErr:     false,
		},
		{
			name:        "successful payment creation with zero amount",
			amount:      0,
			currency:    "EUR",
			description: "Zero amount payment",
			userID:      "user-456",
			wantErr:     false,
		},
		{
			name:        "invalid amount - negative",
			amount:      -10.50,
			currency:    "USD",
			description: "Invalid payment",
			userID:      "user-123",
			wantErr:     true,
		},
		{
			name:        "invalid currency - empty",
			amount:      100.50,
			currency:    "",
			description: "Invalid currency payment",
			userID:      "user-123",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			paymentSvc, auditSvc := createTestServices()

			service := NewPaymentApplicationService(paymentSvc, auditSvc)

			ctx := context.Background()
			result, err := service.CreatePayment(ctx, tt.amount, tt.currency, tt.description, tt.userID)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got none")
					return
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if result == nil {
				t.Error("expected payment to be returned")
				return
			}

			if result.Amount().Value() != tt.amount {
				t.Errorf("expected amount %f, got %f", tt.amount, result.Amount().Value())
			}

			if result.Amount().Currency() != tt.currency {
				t.Errorf("expected currency %q, got %q", tt.currency, result.Amount().Currency())
			}

			if result.Description() != tt.description {
				t.Errorf("expected description %q, got %q", tt.description, result.Description())
			}

			if result.Status() != payment.PaymentStatusPending {
				t.Errorf("expected status %v, got %v", payment.PaymentStatusPending, result.Status())
			}
		})
	}
}

func TestPaymentApplicationService_ProcessPayment(t *testing.T) {
	tests := []struct {
		name          string
		setupPayment  bool
		paymentStatus payment.PaymentStatus
		wantErr       bool
	}{
		{
			name:          "successful payment processing",
			setupPayment:  true,
			paymentStatus: payment.PaymentStatusPending,
			wantErr:       false,
		},
		{
			name:         "payment not found",
			setupPayment: false,
			wantErr:      true,
		},
		{
			name:          "payment in wrong status for processing",
			setupPayment:  true,
			paymentStatus: payment.PaymentStatusCompleted,
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			paymentSvc, auditSvc := createTestServices()
			service := NewPaymentApplicationService(paymentSvc, auditSvc)

			var paymentID string
			if tt.setupPayment {
				amount, _ := payment.NewAmount(100.0, "USD")
				p := payment.NewPayment(amount, "test payment")
				if tt.paymentStatus != payment.PaymentStatusPending {
					p.Process() // Move to processing first if needed
					if tt.paymentStatus == payment.PaymentStatusCompleted {
						p.Complete()
					}
				}

				// Save the payment through the service
				ctx := context.Background()
				createdPayment, _ := paymentSvc.CreatePayment(ctx, amount, "test payment")
				paymentID = createdPayment.ID().String()

				// Update status if needed
				if tt.paymentStatus == payment.PaymentStatusCompleted {
					paymentSvc.ProcessPayment(ctx, createdPayment.ID())
					paymentSvc.CompletePayment(ctx, createdPayment.ID())
				}
			} else {
				paymentID = "non-existent-payment"
			}

			ctx := context.Background()
			err := service.ProcessPayment(ctx, paymentID, "user-123")

			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got none")
					return
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
		})
	}
}

func TestPaymentApplicationService_CompletePayment(t *testing.T) {
	tests := []struct {
		name          string
		setupPayment  bool
		paymentStatus payment.PaymentStatus
		wantErr       bool
	}{
		{
			name:          "successful payment completion",
			setupPayment:  true,
			paymentStatus: payment.PaymentStatusProcessing,
			wantErr:       false,
		},
		{
			name:         "payment not found",
			setupPayment: false,
			wantErr:      true,
		},
		{
			name:          "payment in wrong status for completion",
			setupPayment:  true,
			paymentStatus: payment.PaymentStatusPending,
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			paymentSvc, auditSvc := createTestServices()
			service := NewPaymentApplicationService(paymentSvc, auditSvc)

			var paymentID string
			if tt.setupPayment {
				amount, _ := payment.NewAmount(100.0, "USD")
				ctx := context.Background()
				createdPayment, _ := paymentSvc.CreatePayment(ctx, amount, "test payment")
				paymentID = createdPayment.ID().String()

				// Set up the payment in the required status
				if tt.paymentStatus == payment.PaymentStatusProcessing {
					paymentSvc.ProcessPayment(ctx, createdPayment.ID())
				}
			} else {
				paymentID = "non-existent-payment"
			}

			ctx := context.Background()
			err := service.CompletePayment(ctx, paymentID, "user-456")

			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got none")
					return
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
		})
	}
}

func TestPaymentApplicationService_GetPaymentAuditHistory(t *testing.T) {
	tests := []struct {
		name      string
		paymentID string
		wantErr   bool
	}{
		{
			name:      "get audit history successfully",
			paymentID: "payment-123",
			wantErr:   false,
		},
		{
			name:      "get audit history for non-existent payment",
			paymentID: "non-existent-payment",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			paymentSvc, auditSvc := createTestServices()
			service := NewPaymentApplicationService(paymentSvc, auditSvc)

			ctx := context.Background()
			result, err := service.GetPaymentAuditHistory(ctx, tt.paymentID)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got none")
					return
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// Result should be a slice (empty or not), never nil
			if result == nil {
				t.Error("expected result slice, got nil")
			}
		})
	}
}

// Create a simple test setup using the actual services with in-memory repositories
func createTestServices() (*payment.Service, *audit.Service) {
	paymentRepo := &mockPaymentRepository{
		payments: make(map[string]*payment.Payment),
	}
	auditRepo := &mockAuditRepository{
		entries: make(map[string]*audit.AuditEntry),
	}

	paymentService := payment.NewService(paymentRepo)
	auditService := audit.NewService(auditRepo)

	return paymentService, auditService
}

type mockPaymentRepository struct {
	payments map[string]*payment.Payment
}

func (m *mockPaymentRepository) Save(ctx context.Context, p *payment.Payment) error {
	m.payments[p.ID().String()] = p
	return nil
}

func (m *mockPaymentRepository) FindByID(ctx context.Context, id payment.PaymentID) (*payment.Payment, error) {
	p, exists := m.payments[id.String()]
	if !exists {
		return nil, errors.New("payment not found")
	}
	return p, nil
}

func (m *mockPaymentRepository) FindAll(ctx context.Context) ([]*payment.Payment, error) {
	var result []*payment.Payment
	for _, p := range m.payments {
		result = append(result, p)
	}
	return result, nil
}

func (m *mockPaymentRepository) Update(ctx context.Context, p *payment.Payment) error {
	if _, exists := m.payments[p.ID().String()]; !exists {
		return errors.New("payment not found")
	}
	m.payments[p.ID().String()] = p
	return nil
}

func (m *mockPaymentRepository) Delete(ctx context.Context, id payment.PaymentID) error {
	if _, exists := m.payments[id.String()]; !exists {
		return errors.New("payment not found")
	}
	delete(m.payments, id.String())
	return nil
}

type mockAuditRepository struct {
	entries map[string]*audit.AuditEntry
}

func (m *mockAuditRepository) Save(ctx context.Context, entry *audit.AuditEntry) error {
	m.entries[entry.ID().String()] = entry
	return nil
}

func (m *mockAuditRepository) FindByID(ctx context.Context, id audit.AuditID) (*audit.AuditEntry, error) {
	entry, exists := m.entries[id.String()]
	if !exists {
		return nil, errors.New("audit entry not found")
	}
	return entry, nil
}

func (m *mockAuditRepository) FindByEntityID(ctx context.Context, entityType audit.EntityType, entityID string) ([]*audit.AuditEntry, error) {
	result := make([]*audit.AuditEntry, 0)
	for _, entry := range m.entries {
		if entry.EntityType() == entityType && entry.EntityID() == entityID {
			result = append(result, entry)
		}
	}
	return result, nil
}

func (m *mockAuditRepository) FindByFilter(ctx context.Context, filter audit.AuditFilter) ([]*audit.AuditEntry, error) {
	result := make([]*audit.AuditEntry, 0)
	for _, entry := range m.entries {
		result = append(result, entry)
	}
	return result, nil
}
