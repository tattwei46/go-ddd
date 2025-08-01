package payment

import (
	"testing"
	"time"
)

func TestNewAmount(t *testing.T) {
	tests := []struct {
		name     string
		value    float64
		currency string
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "valid amount and currency",
			value:    100.50,
			currency: "USD",
			wantErr:  false,
		},
		{
			name:     "zero amount",
			value:    0,
			currency: "EUR",
			wantErr:  false,
		},
		{
			name:     "negative amount",
			value:    -10.50,
			currency: "USD",
			wantErr:  true,
			errMsg:   "amount cannot be negative",
		},
		{
			name:     "empty currency",
			value:    100.50,
			currency: "",
			wantErr:  true,
			errMsg:   "currency cannot be empty",
		},
		{
			name:     "large amount",
			value:    999999.99,
			currency: "JPY",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			amount, err := NewAmount(tt.value, tt.currency)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
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

			if amount.Value() != tt.value {
				t.Errorf("expected value %f, got %f", tt.value, amount.Value())
			}

			if amount.Currency() != tt.currency {
				t.Errorf("expected currency %q, got %q", tt.currency, amount.Currency())
			}
		})
	}
}

func TestPaymentStatus_String(t *testing.T) {
	tests := []struct {
		name   string
		status PaymentStatus
		want   string
	}{
		{
			name:   "pending status",
			status: PaymentStatusPending,
			want:   "pending",
		},
		{
			name:   "processing status",
			status: PaymentStatusProcessing,
			want:   "processing",
		},
		{
			name:   "completed status",
			status: PaymentStatusCompleted,
			want:   "completed",
		},
		{
			name:   "failed status",
			status: PaymentStatusFailed,
			want:   "failed",
		},
		{
			name:   "cancelled status",
			status: PaymentStatusCancelled,
			want:   "cancelled",
		},
		{
			name:   "unknown status",
			status: PaymentStatus(999),
			want:   "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.status.String()
			if got != tt.want {
				t.Errorf("expected %q, got %q", tt.want, got)
			}
		})
	}
}

func TestPayment_Process(t *testing.T) {
	tests := []struct {
		name           string
		initialStatus  PaymentStatus
		wantErr        bool
		errMsg         string
		expectedStatus PaymentStatus
	}{
		{
			name:           "process from pending",
			initialStatus:  PaymentStatusPending,
			wantErr:        false,
			expectedStatus: PaymentStatusProcessing,
		},
		{
			name:          "process from processing",
			initialStatus: PaymentStatusProcessing,
			wantErr:       true,
			errMsg:        "payment can only be processed from pending status",
		},
		{
			name:          "process from completed",
			initialStatus: PaymentStatusCompleted,
			wantErr:       true,
			errMsg:        "payment can only be processed from pending status",
		},
		{
			name:          "process from failed",
			initialStatus: PaymentStatusFailed,
			wantErr:       true,
			errMsg:        "payment can only be processed from pending status",
		},
		{
			name:          "process from cancelled",
			initialStatus: PaymentStatusCancelled,
			wantErr:       true,
			errMsg:        "payment can only be processed from pending status",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			amount, _ := NewAmount(100.0, "USD")
			payment := NewPayment(amount, "test payment")
			payment.status = tt.initialStatus
			oldUpdatedAt := payment.updatedAt

			time.Sleep(1 * time.Millisecond) // Ensure time difference

			err := payment.Process()

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
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

			if payment.Status() != tt.expectedStatus {
				t.Errorf("expected status %v, got %v", tt.expectedStatus, payment.Status())
			}

			if !payment.UpdatedAt().After(oldUpdatedAt) {
				t.Errorf("expected updated_at to be updated")
			}
		})
	}
}

func TestPayment_Complete(t *testing.T) {
	tests := []struct {
		name           string
		initialStatus  PaymentStatus
		wantErr        bool
		errMsg         string
		expectedStatus PaymentStatus
	}{
		{
			name:           "complete from processing",
			initialStatus:  PaymentStatusProcessing,
			wantErr:        false,
			expectedStatus: PaymentStatusCompleted,
		},
		{
			name:          "complete from pending",
			initialStatus: PaymentStatusPending,
			wantErr:       true,
			errMsg:        "payment can only be completed from processing status",
		},
		{
			name:          "complete from completed",
			initialStatus: PaymentStatusCompleted,
			wantErr:       true,
			errMsg:        "payment can only be completed from processing status",
		},
		{
			name:          "complete from failed",
			initialStatus: PaymentStatusFailed,
			wantErr:       true,
			errMsg:        "payment can only be completed from processing status",
		},
		{
			name:          "complete from cancelled",
			initialStatus: PaymentStatusCancelled,
			wantErr:       true,
			errMsg:        "payment can only be completed from processing status",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			amount, _ := NewAmount(100.0, "USD")
			payment := NewPayment(amount, "test payment")
			payment.status = tt.initialStatus
			oldUpdatedAt := payment.updatedAt

			time.Sleep(1 * time.Millisecond)

			err := payment.Complete()

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
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

			if payment.Status() != tt.expectedStatus {
				t.Errorf("expected status %v, got %v", tt.expectedStatus, payment.Status())
			}

			if !payment.UpdatedAt().After(oldUpdatedAt) {
				t.Errorf("expected updated_at to be updated")
			}
		})
	}
}

func TestPayment_Fail(t *testing.T) {
	tests := []struct {
		name           string
		initialStatus  PaymentStatus
		wantErr        bool
		errMsg         string
		expectedStatus PaymentStatus
	}{
		{
			name:           "fail from pending",
			initialStatus:  PaymentStatusPending,
			wantErr:        false,
			expectedStatus: PaymentStatusFailed,
		},
		{
			name:           "fail from processing",
			initialStatus:  PaymentStatusProcessing,
			wantErr:        false,
			expectedStatus: PaymentStatusFailed,
		},
		{
			name:           "fail from failed",
			initialStatus:  PaymentStatusFailed,
			wantErr:        false,
			expectedStatus: PaymentStatusFailed,
		},
		{
			name:           "fail from cancelled",
			initialStatus:  PaymentStatusCancelled,
			wantErr:        false,
			expectedStatus: PaymentStatusFailed,
		},
		{
			name:          "fail from completed",
			initialStatus: PaymentStatusCompleted,
			wantErr:       true,
			errMsg:        "completed payment cannot be failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			amount, _ := NewAmount(100.0, "USD")
			payment := NewPayment(amount, "test payment")
			payment.status = tt.initialStatus
			oldUpdatedAt := payment.updatedAt

			time.Sleep(1 * time.Millisecond)

			err := payment.Fail()

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
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

			if payment.Status() != tt.expectedStatus {
				t.Errorf("expected status %v, got %v", tt.expectedStatus, payment.Status())
			}

			if !payment.UpdatedAt().After(oldUpdatedAt) {
				t.Errorf("expected updated_at to be updated")
			}
		})
	}
}

func TestPayment_Cancel(t *testing.T) {
	tests := []struct {
		name           string
		initialStatus  PaymentStatus
		wantErr        bool
		errMsg         string
		expectedStatus PaymentStatus
	}{
		{
			name:           "cancel from pending",
			initialStatus:  PaymentStatusPending,
			wantErr:        false,
			expectedStatus: PaymentStatusCancelled,
		},
		{
			name:           "cancel from failed",
			initialStatus:  PaymentStatusFailed,
			wantErr:        false,
			expectedStatus: PaymentStatusCancelled,
		},
		{
			name:           "cancel from cancelled",
			initialStatus:  PaymentStatusCancelled,
			wantErr:        false,
			expectedStatus: PaymentStatusCancelled,
		},
		{
			name:          "cancel from processing",
			initialStatus: PaymentStatusProcessing,
			wantErr:       true,
			errMsg:        "payment cannot be cancelled in current status",
		},
		{
			name:          "cancel from completed",
			initialStatus: PaymentStatusCompleted,
			wantErr:       true,
			errMsg:        "payment cannot be cancelled in current status",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			amount, _ := NewAmount(100.0, "USD")
			payment := NewPayment(amount, "test payment")
			payment.status = tt.initialStatus
			oldUpdatedAt := payment.updatedAt

			time.Sleep(1 * time.Millisecond)

			err := payment.Cancel()

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
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

			if payment.Status() != tt.expectedStatus {
				t.Errorf("expected status %v, got %v", tt.expectedStatus, payment.Status())
			}

			if !payment.UpdatedAt().After(oldUpdatedAt) {
				t.Errorf("expected updated_at to be updated")
			}
		})
	}
}

func TestNewPayment(t *testing.T) {
	tests := []struct {
		name        string
		amount      Amount
		description string
	}{
		{
			name:        "create payment with valid data",
			amount:      mustCreateAmount(100.50, "USD"),
			description: "Test payment",
		},
		{
			name:        "create payment with empty description",
			amount:      mustCreateAmount(50.00, "EUR"),
			description: "",
		},
		{
			name:        "create payment with zero amount",
			amount:      mustCreateAmount(0, "JPY"),
			description: "Zero amount payment",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			before := time.Now()
			payment := NewPayment(tt.amount, tt.description)
			after := time.Now()

			if payment == nil {
				t.Fatal("expected payment to be created")
			}

			if payment.ID().String() == "" {
				t.Error("expected payment ID to be set")
			}

			if payment.Amount().Value() != tt.amount.Value() {
				t.Errorf("expected amount value %f, got %f", tt.amount.Value(), payment.Amount().Value())
			}

			if payment.Amount().Currency() != tt.amount.Currency() {
				t.Errorf("expected currency %q, got %q", tt.amount.Currency(), payment.Amount().Currency())
			}

			if payment.Description() != tt.description {
				t.Errorf("expected description %q, got %q", tt.description, payment.Description())
			}

			if payment.Status() != PaymentStatusPending {
				t.Errorf("expected status %v, got %v", PaymentStatusPending, payment.Status())
			}

			if payment.CreatedAt().Before(before) || payment.CreatedAt().After(after) {
				t.Error("expected created_at to be set to current time")
			}

			if payment.UpdatedAt().Before(before) || payment.UpdatedAt().After(after) {
				t.Error("expected updated_at to be set to current time")
			}

			if !payment.CreatedAt().Equal(payment.UpdatedAt()) {
				t.Error("expected created_at and updated_at to be equal for new payment")
			}
		})
	}
}

func TestPaymentIDFromString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "valid UUID string",
			input:    "123e4567-e89b-12d3-a456-426614174000",
			expected: "123e4567-e89b-12d3-a456-426614174000",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "arbitrary string",
			input:    "test-payment-id",
			expected: "test-payment-id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			paymentID := PaymentIDFromString(tt.input)
			if paymentID.String() != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, paymentID.String())
			}
		})
	}
}

func mustCreateAmount(value float64, currency string) Amount {
	amount, err := NewAmount(value, currency)
	if err != nil {
		panic(err)
	}
	return amount
}
