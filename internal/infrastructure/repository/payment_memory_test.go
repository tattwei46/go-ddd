package repository

import (
	"context"
	"sync"
	"testing"

	"go-ddd/internal/domain/payment"
)

func TestPaymentMemoryRepository_Save(t *testing.T) {
	tests := []struct {
		name    string
		payment *payment.Payment
		wantErr bool
	}{
		{
			name:    "save valid payment",
			payment: mustCreatePayment(100.50, "USD", "Test payment"),
			wantErr: false,
		},
		{
			name:    "save payment with zero amount",
			payment: mustCreatePayment(0, "EUR", "Zero amount payment"),
			wantErr: false,
		},
		{
			name:    "save payment with empty description",
			payment: mustCreatePayment(50.00, "JPY", ""),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewPaymentMemoryRepository()
			ctx := context.Background()

			err := repo.Save(ctx, tt.payment)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// Verify payment was saved
			savedPayment, err := repo.FindByID(ctx, tt.payment.ID())
			if err != nil {
				t.Errorf("failed to find saved payment: %v", err)
				return
			}

			if savedPayment.ID().String() != tt.payment.ID().String() {
				t.Errorf("expected payment ID %q, got %q", tt.payment.ID().String(), savedPayment.ID().String())
			}
		})
	}
}

func TestPaymentMemoryRepository_FindByID(t *testing.T) {
	tests := []struct {
		name         string
		setupPayment bool
		paymentID    payment.PaymentID
		wantErr      bool
		errMsg       string
	}{
		{
			name:         "find existing payment",
			setupPayment: true,
			wantErr:      false,
		},
		{
			name:         "find non-existent payment",
			setupPayment: false,
			paymentID:    payment.PaymentIDFromString("non-existent-id"),
			wantErr:      true,
			errMsg:       "payment not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewPaymentMemoryRepository()
			ctx := context.Background()

			var expectedPayment *payment.Payment
			var searchID payment.PaymentID

			if tt.setupPayment {
				expectedPayment = mustCreatePayment(100.50, "USD", "Test payment")
				repo.Save(ctx, expectedPayment)
				searchID = expectedPayment.ID()
			} else {
				searchID = tt.paymentID
			}

			result, err := repo.FindByID(ctx, searchID)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got none")
					return
				}
				if err.Error() != tt.errMsg {
					t.Errorf("expected error message %q, got %q", tt.errMsg, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if result == nil {
				t.Error("expected payment to be found")
				return
			}

			if result.ID().String() != expectedPayment.ID().String() {
				t.Errorf("expected payment ID %q, got %q", expectedPayment.ID().String(), result.ID().String())
			}
		})
	}
}

func TestPaymentMemoryRepository_FindAll(t *testing.T) {
	tests := []struct {
		name          string
		setupPayments int
		expectedCount int
	}{
		{
			name:          "find all payments - empty repository",
			setupPayments: 0,
			expectedCount: 0,
		},
		{
			name:          "find all payments - single payment",
			setupPayments: 1,
			expectedCount: 1,
		},
		{
			name:          "find all payments - multiple payments",
			setupPayments: 5,
			expectedCount: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewPaymentMemoryRepository()
			ctx := context.Background()

			// Setup payments
			var expectedPayments []*payment.Payment
			for i := 0; i < tt.setupPayments; i++ {
				p := mustCreatePayment(float64(100+i), "USD", "Test payment")
				expectedPayments = append(expectedPayments, p)
				repo.Save(ctx, p)
			}

			result, err := repo.FindAll(ctx)

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if len(result) != tt.expectedCount {
				t.Errorf("expected %d payments, got %d", tt.expectedCount, len(result))
				return
			}

			// Verify all expected payments are present
			paymentMap := make(map[string]*payment.Payment)
			for _, p := range result {
				paymentMap[p.ID().String()] = p
			}

			for _, expected := range expectedPayments {
				if _, exists := paymentMap[expected.ID().String()]; !exists {
					t.Errorf("expected payment %q not found in results", expected.ID().String())
				}
			}
		})
	}
}

func TestPaymentMemoryRepository_Update(t *testing.T) {
	tests := []struct {
		name         string
		setupPayment bool
		wantErr      bool
		errMsg       string
	}{
		{
			name:         "update existing payment",
			setupPayment: true,
			wantErr:      false,
		},
		{
			name:         "update non-existent payment",
			setupPayment: false,
			wantErr:      true,
			errMsg:       "payment not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewPaymentMemoryRepository()
			ctx := context.Background()

			var testPayment *payment.Payment

			if tt.setupPayment {
				testPayment = mustCreatePayment(100.50, "USD", "Original description")
				repo.Save(ctx, testPayment)

				// Modify payment status to simulate an update
				testPayment.Process()
			} else {
				testPayment = mustCreatePayment(200.00, "EUR", "Non-existent payment")
			}

			err := repo.Update(ctx, testPayment)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got none")
					return
				}
				if err.Error() != tt.errMsg {
					t.Errorf("expected error message %q, got %q", tt.errMsg, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// Verify payment was updated
			updatedPayment, err := repo.FindByID(ctx, testPayment.ID())
			if err != nil {
				t.Errorf("failed to find updated payment: %v", err)
				return
			}

			if updatedPayment.Status() != testPayment.Status() {
				t.Errorf("expected status %v, got %v", testPayment.Status(), updatedPayment.Status())
			}
		})
	}
}

func TestPaymentMemoryRepository_Delete(t *testing.T) {
	tests := []struct {
		name         string
		setupPayment bool
		paymentID    payment.PaymentID
		wantErr      bool
		errMsg       string
	}{
		{
			name:         "delete existing payment",
			setupPayment: true,
			wantErr:      false,
		},
		{
			name:         "delete non-existent payment",
			setupPayment: false,
			paymentID:    payment.PaymentIDFromString("non-existent-id"),
			wantErr:      true,
			errMsg:       "payment not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewPaymentMemoryRepository()
			ctx := context.Background()

			var deleteID payment.PaymentID

			if tt.setupPayment {
				testPayment := mustCreatePayment(100.50, "USD", "Test payment")
				repo.Save(ctx, testPayment)
				deleteID = testPayment.ID()
			} else {
				deleteID = tt.paymentID
			}

			err := repo.Delete(ctx, deleteID)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got none")
					return
				}
				if err.Error() != tt.errMsg {
					t.Errorf("expected error message %q, got %q", tt.errMsg, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// Verify payment was deleted
			_, err = repo.FindByID(ctx, deleteID)
			if err == nil {
				t.Error("expected payment to be deleted, but it still exists")
			}
		})
	}
}

func TestPaymentMemoryRepository_ConcurrentAccess(t *testing.T) {
	repo := NewPaymentMemoryRepository()
	ctx := context.Background()

	const numGoroutines = 10
	const paymentsPerGoroutine = 5

	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines*paymentsPerGoroutine)

	// Test concurrent writes
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(routineID int) {
			defer wg.Done()

			for j := 0; j < paymentsPerGoroutine; j++ {
				p := mustCreatePayment(float64(routineID*100+j), "USD", "Concurrent test")
				if err := repo.Save(ctx, p); err != nil {
					errors <- err
					return
				}
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		t.Errorf("concurrent write error: %v", err)
	}

	// Verify all payments were saved
	payments, err := repo.FindAll(ctx)
	if err != nil {
		t.Errorf("failed to find all payments: %v", err)
		return
	}

	expectedCount := numGoroutines * paymentsPerGoroutine
	if len(payments) != expectedCount {
		t.Errorf("expected %d payments, got %d", expectedCount, len(payments))
	}

	// Test concurrent reads
	var readWg sync.WaitGroup
	readErrors := make(chan error, len(payments))

	for _, p := range payments {
		readWg.Add(1)
		go func(paymentID payment.PaymentID) {
			defer readWg.Done()

			_, err := repo.FindByID(ctx, paymentID)
			if err != nil {
				readErrors <- err
			}
		}(p.ID())
	}

	readWg.Wait()
	close(readErrors)

	// Check for read errors
	for err := range readErrors {
		t.Errorf("concurrent read error: %v", err)
	}
}

func mustCreatePayment(amount float64, currency, description string) *payment.Payment {
	amt, err := payment.NewAmount(amount, currency)
	if err != nil {
		panic(err)
	}
	return payment.NewPayment(amt, description)
}
