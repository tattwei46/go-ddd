package audit

import (
	"testing"
	"time"
)

func TestNewAuditEntry(t *testing.T) {
	tests := []struct {
		name       string
		entityType EntityType
		entityID   string
		action     ActionType
		userID     string
	}{
		{
			name:       "create payment audit entry",
			entityType: EntityTypePayment,
			entityID:   "payment-123",
			action:     ActionTypeCreated,
			userID:     "user-456",
		},
		{
			name:       "update payment audit entry",
			entityType: EntityTypePayment,
			entityID:   "payment-789",
			action:     ActionTypeUpdated,
			userID:     "user-101",
		},
		{
			name:       "process payment audit entry",
			entityType: EntityTypePayment,
			entityID:   "payment-abc",
			action:     ActionTypeProcessed,
			userID:     "user-xyz",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			before := time.Now()
			entry := NewAuditEntry(tt.entityType, tt.entityID, tt.action, tt.userID)
			after := time.Now()

			if entry == nil {
				t.Fatal("expected audit entry to be created")
			}

			if entry.ID().String() == "" {
				t.Error("expected audit ID to be set")
			}

			if entry.EntityType() != tt.entityType {
				t.Errorf("expected entity type %v, got %v", tt.entityType, entry.EntityType())
			}

			if entry.EntityID() != tt.entityID {
				t.Errorf("expected entity ID %q, got %q", tt.entityID, entry.EntityID())
			}

			if entry.Action() != tt.action {
				t.Errorf("expected action %v, got %v", tt.action, entry.Action())
			}

			if entry.UserID() != tt.userID {
				t.Errorf("expected user ID %q, got %q", tt.userID, entry.UserID())
			}

			if entry.Timestamp().Before(before) || entry.Timestamp().After(after) {
				t.Error("expected timestamp to be set to current time")
			}

			if entry.OldData() == nil {
				t.Error("expected old data to be initialized")
			}

			if entry.NewData() == nil {
				t.Error("expected new data to be initialized")
			}

			if entry.Metadata() == nil {
				t.Error("expected metadata to be initialized")
			}
		})
	}
}

func TestAuditEntry_SetOldData(t *testing.T) {
	tests := []struct {
		name    string
		data    interface{}
		wantErr bool
	}{
		{
			name: "set valid struct data",
			data: struct {
				ID     string  `json:"id"`
				Status string  `json:"status"`
				Amount float64 `json:"amount"`
			}{
				ID:     "payment-123",
				Status: "pending",
				Amount: 100.50,
			},
			wantErr: false,
		},
		{
			name: "set map data",
			data: map[string]interface{}{
				"id":     "payment-456",
				"status": "completed",
				"amount": 200.00,
			},
			wantErr: false,
		},
		{
			name:    "set string data",
			data:    "simple string",
			wantErr: true,
		},
		{
			name:    "set nil data",
			data:    nil,
			wantErr: false,
		},
		{
			name:    "set channel data (unmarshalable)",
			data:    make(chan int),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry := NewAuditEntry(EntityTypePayment, "test-id", ActionTypeCreated, "user-123")

			err := entry.SetOldData(tt.data)

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

			oldData := entry.OldData()
			// For nil data, old data should be empty map
			if oldData == nil && tt.data != nil {
				t.Error("expected old data to be set")
			}
		})
	}
}

func TestAuditEntry_SetNewData(t *testing.T) {
	tests := []struct {
		name    string
		data    interface{}
		wantErr bool
	}{
		{
			name: "set valid struct data",
			data: struct {
				ID     string  `json:"id"`
				Status string  `json:"status"`
				Amount float64 `json:"amount"`
			}{
				ID:     "payment-789",
				Status: "processing",
				Amount: 300.75,
			},
			wantErr: false,
		},
		{
			name: "set map data",
			data: map[string]interface{}{
				"id":     "payment-abc",
				"status": "failed",
				"amount": 150.25,
			},
			wantErr: false,
		},
		{
			name:    "set slice data",
			data:    []string{"item1", "item2", "item3"},
			wantErr: true,
		},
		{
			name:    "set function data (unmarshalable)",
			data:    func() {},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry := NewAuditEntry(EntityTypePayment, "test-id", ActionTypeUpdated, "user-456")

			err := entry.SetNewData(tt.data)

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

			newData := entry.NewData()
			if newData == nil {
				t.Error("expected new data to be set")
			}
		})
	}
}

func TestAuditEntry_AddMetadata(t *testing.T) {
	tests := []struct {
		name     string
		metadata map[string]string
	}{
		{
			name: "add single metadata",
			metadata: map[string]string{
				"source": "api",
			},
		},
		{
			name: "add multiple metadata",
			metadata: map[string]string{
				"source":    "web",
				"ip":        "192.168.1.1",
				"userAgent": "Mozilla/5.0",
			},
		},
		{
			name: "add empty values",
			metadata: map[string]string{
				"empty": "",
				"key":   "value",
			},
		},
		{
			name:     "add no metadata",
			metadata: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry := NewAuditEntry(EntityTypePayment, "test-id", ActionTypeCompleted, "user-789")

			for key, value := range tt.metadata {
				entry.AddMetadata(key, value)
			}

			metadata := entry.Metadata()

			if len(metadata) != len(tt.metadata) {
				t.Errorf("expected %d metadata entries, got %d", len(tt.metadata), len(metadata))
			}

			for key, expectedValue := range tt.metadata {
				if actualValue, exists := metadata[key]; !exists {
					t.Errorf("expected metadata key %q to exist", key)
				} else if actualValue != expectedValue {
					t.Errorf("expected metadata value %q for key %q, got %q", expectedValue, key, actualValue)
				}
			}
		})
	}
}

func TestActionType_Values(t *testing.T) {
	tests := []struct {
		name   string
		action ActionType
		want   string
	}{
		{
			name:   "created action",
			action: ActionTypeCreated,
			want:   "created",
		},
		{
			name:   "updated action",
			action: ActionTypeUpdated,
			want:   "updated",
		},
		{
			name:   "deleted action",
			action: ActionTypeDeleted,
			want:   "deleted",
		},
		{
			name:   "processed action",
			action: ActionTypeProcessed,
			want:   "processed",
		},
		{
			name:   "completed action",
			action: ActionTypeCompleted,
			want:   "completed",
		},
		{
			name:   "failed action",
			action: ActionTypeFailed,
			want:   "failed",
		},
		{
			name:   "cancelled action",
			action: ActionTypeCancelled,
			want:   "cancelled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := string(tt.action)
			if got != tt.want {
				t.Errorf("expected %q, got %q", tt.want, got)
			}
		})
	}
}

func TestEntityType_Values(t *testing.T) {
	tests := []struct {
		name       string
		entityType EntityType
		want       string
	}{
		{
			name:       "payment entity type",
			entityType: EntityTypePayment,
			want:       "payment",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := string(tt.entityType)
			if got != tt.want {
				t.Errorf("expected %q, got %q", tt.want, got)
			}
		})
	}
}

func TestAuditFilter(t *testing.T) {
	tests := []struct {
		name   string
		filter AuditFilter
	}{
		{
			name: "filter with all fields",
			filter: AuditFilter{
				EntityType: &[]EntityType{EntityTypePayment}[0],
				EntityID:   &[]string{"payment-123"}[0],
				Action:     &[]ActionType{ActionTypeCreated}[0],
				UserID:     &[]string{"user-456"}[0],
				FromDate:   &[]time.Time{time.Now().Add(-24 * time.Hour)}[0],
				ToDate:     &[]time.Time{time.Now()}[0],
			},
		},
		{
			name: "filter with partial fields",
			filter: AuditFilter{
				EntityType: &[]EntityType{EntityTypePayment}[0],
				UserID:     &[]string{"user-789"}[0],
			},
		},
		{
			name:   "empty filter",
			filter: AuditFilter{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := tt.filter

			// Test that filter fields can be accessed without panicking
			if filter.EntityType != nil && *filter.EntityType != EntityTypePayment {
				t.Error("entity type filter not working as expected")
			}

			if filter.EntityID != nil && *filter.EntityID == "" {
				t.Error("entity ID filter should not be empty when set")
			}

			if filter.Action != nil {
				// Action should be one of the valid action types
				validActions := []ActionType{
					ActionTypeCreated, ActionTypeUpdated, ActionTypeDeleted,
					ActionTypeProcessed, ActionTypeCompleted, ActionTypeFailed, ActionTypeCancelled,
				}
				found := false
				for _, validAction := range validActions {
					if *filter.Action == validAction {
						found = true
						break
					}
				}
				if !found {
					t.Error("invalid action type in filter")
				}
			}

			if filter.UserID != nil && *filter.UserID == "" {
				t.Error("user ID filter should not be empty when set")
			}

			if filter.FromDate != nil && filter.ToDate != nil {
				if filter.FromDate.After(*filter.ToDate) {
					t.Error("from date should not be after to date")
				}
			}
		})
	}
}
