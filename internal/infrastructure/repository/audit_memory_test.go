package repository

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"go-ddd/internal/domain/audit"
)

func TestAuditMemoryRepository_Save(t *testing.T) {
	tests := []struct {
		name    string
		entry   *audit.AuditEntry
		wantErr bool
	}{
		{
			name:    "save valid audit entry",
			entry:   audit.NewAuditEntry(audit.EntityTypePayment, "payment-123", audit.ActionTypeCreated, "user-456"),
			wantErr: false,
		},
		{
			name:    "save audit entry with metadata",
			entry:   createAuditEntryWithMetadata("payment-789", "user-101", map[string]string{"source": "api"}),
			wantErr: false,
		},
		{
			name:    "save audit entry with data",
			entry:   createAuditEntryWithData("payment-abc", "user-xyz"),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewAuditMemoryRepository()
			ctx := context.Background()

			err := repo.Save(ctx, tt.entry)

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

			// Verify entry was saved
			savedEntry, err := repo.FindByID(ctx, tt.entry.ID())
			if err != nil {
				t.Errorf("failed to find saved audit entry: %v", err)
				return
			}

			if savedEntry.ID().String() != tt.entry.ID().String() {
				t.Errorf("expected audit ID %q, got %q", tt.entry.ID().String(), savedEntry.ID().String())
			}
		})
	}
}

func TestAuditMemoryRepository_FindByID(t *testing.T) {
	tests := []struct {
		name       string
		setupEntry bool
		entryID    audit.AuditID
		wantErr    bool
		errMsg     string
	}{
		{
			name:       "find existing audit entry",
			setupEntry: true,
			wantErr:    false,
		},
		{
			name:       "find non-existent audit entry",
			setupEntry: false,
			entryID:    audit.AuditID{},
			wantErr:    true,
			errMsg:     "audit entry not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewAuditMemoryRepository()
			ctx := context.Background()

			var expectedEntry *audit.AuditEntry
			var searchID audit.AuditID

			if tt.setupEntry {
				expectedEntry = audit.NewAuditEntry(audit.EntityTypePayment, "payment-123", audit.ActionTypeCreated, "user-456")
				repo.Save(ctx, expectedEntry)
				searchID = expectedEntry.ID()
			} else {
				searchID = tt.entryID
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
				t.Error("expected audit entry to be found")
				return
			}

			if result.ID().String() != expectedEntry.ID().String() {
				t.Errorf("expected audit ID %q, got %q", expectedEntry.ID().String(), result.ID().String())
			}
		})
	}
}

func TestAuditMemoryRepository_FindByEntityID(t *testing.T) {
	tests := []struct {
		name          string
		setupEntries  []auditEntrySetup
		entityType    audit.EntityType
		entityID      string
		expectedCount int
	}{
		{
			name: "find entries for existing entity",
			setupEntries: []auditEntrySetup{
				{entityID: "payment-123", action: audit.ActionTypeCreated},
				{entityID: "payment-123", action: audit.ActionTypeProcessed},
				{entityID: "payment-456", action: audit.ActionTypeCreated},
			},
			entityType:    audit.EntityTypePayment,
			entityID:      "payment-123",
			expectedCount: 2,
		},
		{
			name: "find entries for non-existent entity",
			setupEntries: []auditEntrySetup{
				{entityID: "payment-123", action: audit.ActionTypeCreated},
			},
			entityType:    audit.EntityTypePayment,
			entityID:      "payment-999",
			expectedCount: 0,
		},
		{
			name:          "find entries in empty repository",
			setupEntries:  []auditEntrySetup{},
			entityType:    audit.EntityTypePayment,
			entityID:      "payment-123",
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewAuditMemoryRepository()
			ctx := context.Background()

			// Setup entries
			for _, setup := range tt.setupEntries {
				entry := audit.NewAuditEntry(audit.EntityTypePayment, setup.entityID, setup.action, "user-123")
				repo.Save(ctx, entry)
			}

			result, err := repo.FindByEntityID(ctx, tt.entityType, tt.entityID)

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if len(result) != tt.expectedCount {
				t.Errorf("expected %d entries, got %d", tt.expectedCount, len(result))
				return
			}

			// Verify all returned entries match the entity
			for _, entry := range result {
				if entry.EntityType() != tt.entityType {
					t.Errorf("expected entity type %v, got %v", tt.entityType, entry.EntityType())
				}
				if entry.EntityID() != tt.entityID {
					t.Errorf("expected entity ID %q, got %q", tt.entityID, entry.EntityID())
				}
			}
		})
	}
}

func TestAuditMemoryRepository_FindByFilter(t *testing.T) {
	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)

	tests := []struct {
		name          string
		setupEntries  []auditEntrySetup
		filter        audit.AuditFilter
		expectedCount int
	}{
		{
			name: "filter by entity type",
			setupEntries: []auditEntrySetup{
				{entityID: "payment-123", action: audit.ActionTypeCreated},
				{entityID: "payment-456", action: audit.ActionTypeProcessed},
			},
			filter: audit.AuditFilter{
				EntityType: &[]audit.EntityType{audit.EntityTypePayment}[0],
			},
			expectedCount: 2,
		},
		{
			name: "filter by entity ID",
			setupEntries: []auditEntrySetup{
				{entityID: "payment-123", action: audit.ActionTypeCreated},
				{entityID: "payment-456", action: audit.ActionTypeProcessed},
			},
			filter: audit.AuditFilter{
				EntityID: &[]string{"payment-123"}[0],
			},
			expectedCount: 1,
		},
		{
			name: "filter by action",
			setupEntries: []auditEntrySetup{
				{entityID: "payment-123", action: audit.ActionTypeCreated},
				{entityID: "payment-456", action: audit.ActionTypeProcessed},
				{entityID: "payment-789", action: audit.ActionTypeCreated},
			},
			filter: audit.AuditFilter{
				Action: &[]audit.ActionType{audit.ActionTypeCreated}[0],
			},
			expectedCount: 2,
		},
		{
			name: "filter by user ID",
			setupEntries: []auditEntrySetup{
				{entityID: "payment-123", action: audit.ActionTypeCreated, userID: "user-111"},
				{entityID: "payment-456", action: audit.ActionTypeProcessed, userID: "user-222"},
			},
			filter: audit.AuditFilter{
				UserID: &[]string{"user-111"}[0],
			},
			expectedCount: 1,
		},
		{
			name: "filter by date range - wide range",
			setupEntries: []auditEntrySetup{
				{entityID: "payment-123", action: audit.ActionTypeCreated},
				{entityID: "payment-456", action: audit.ActionTypeProcessed},
			},
			filter: audit.AuditFilter{
				FromDate: &yesterday,
			},
			expectedCount: 2,
		},
		{
			name: "filter with multiple criteria",
			setupEntries: []auditEntrySetup{
				{entityID: "payment-123", action: audit.ActionTypeCreated, userID: "user-111"},
				{entityID: "payment-123", action: audit.ActionTypeProcessed, userID: "user-111"},
				{entityID: "payment-456", action: audit.ActionTypeCreated, userID: "user-222"},
			},
			filter: audit.AuditFilter{
				EntityID: &[]string{"payment-123"}[0],
				UserID:   &[]string{"user-111"}[0],
			},
			expectedCount: 2,
		},
		{
			name: "filter with no matches",
			setupEntries: []auditEntrySetup{
				{entityID: "payment-123", action: audit.ActionTypeCreated},
			},
			filter: audit.AuditFilter{
				EntityID: &[]string{"payment-999"}[0],
			},
			expectedCount: 0,
		},
		{
			name: "empty filter returns all",
			setupEntries: []auditEntrySetup{
				{entityID: "payment-123", action: audit.ActionTypeCreated},
				{entityID: "payment-456", action: audit.ActionTypeProcessed},
			},
			filter:        audit.AuditFilter{},
			expectedCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewAuditMemoryRepository()
			ctx := context.Background()

			// Setup entries
			for _, setup := range tt.setupEntries {
				userID := setup.userID
				if userID == "" {
					userID = "default-user"
				}

				entry := audit.NewAuditEntry(audit.EntityTypePayment, setup.entityID, setup.action, userID)
				repo.Save(ctx, entry)
			}

			result, err := repo.FindByFilter(ctx, tt.filter)

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if len(result) != tt.expectedCount {
				t.Errorf("expected %d entries, got %d", tt.expectedCount, len(result))
				return
			}

			// Verify all returned entries match the filter
			for _, entry := range result {
				if !repo.matchesFilter(entry, tt.filter) {
					t.Errorf("entry %q does not match filter", entry.ID().String())
				}
			}
		})
	}
}

func TestAuditMemoryRepository_MatchesFilter(t *testing.T) {
	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)
	tomorrow := now.Add(24 * time.Hour)

	entry := audit.NewAuditEntry(audit.EntityTypePayment, "payment-123", audit.ActionTypeCreated, "user-456")

	tests := []struct {
		name     string
		filter   audit.AuditFilter
		expected bool
	}{
		{
			name:     "empty filter matches",
			filter:   audit.AuditFilter{},
			expected: true,
		},
		{
			name: "matching entity type",
			filter: audit.AuditFilter{
				EntityType: &[]audit.EntityType{audit.EntityTypePayment}[0],
			},
			expected: true,
		},
		{
			name: "non-matching entity type",
			filter: audit.AuditFilter{
				EntityType: &[]audit.EntityType{"different-type"}[0],
			},
			expected: false,
		},
		{
			name: "matching entity ID",
			filter: audit.AuditFilter{
				EntityID: &[]string{"payment-123"}[0],
			},
			expected: true,
		},
		{
			name: "non-matching entity ID",
			filter: audit.AuditFilter{
				EntityID: &[]string{"payment-999"}[0],
			},
			expected: false,
		},
		{
			name: "matching action",
			filter: audit.AuditFilter{
				Action: &[]audit.ActionType{audit.ActionTypeCreated}[0],
			},
			expected: true,
		},
		{
			name: "non-matching action",
			filter: audit.AuditFilter{
				Action: &[]audit.ActionType{audit.ActionTypeDeleted}[0],
			},
			expected: false,
		},
		{
			name: "matching user ID",
			filter: audit.AuditFilter{
				UserID: &[]string{"user-456"}[0],
			},
			expected: true,
		},
		{
			name: "non-matching user ID",
			filter: audit.AuditFilter{
				UserID: &[]string{"user-999"}[0],
			},
			expected: false,
		},
		{
			name: "matching date range",
			filter: audit.AuditFilter{
				FromDate: &yesterday,
				ToDate:   &tomorrow,
			},
			expected: true,
		},
		{
			name: "date range - entry too old",
			filter: audit.AuditFilter{
				FromDate: &tomorrow,
			},
			expected: false,
		},
		{
			name: "date range - entry too new",
			filter: audit.AuditFilter{
				ToDate: &yesterday,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewAuditMemoryRepository()

			result := repo.matchesFilter(entry, tt.filter)

			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestAuditMemoryRepository_ConcurrentAccess(t *testing.T) {
	repo := NewAuditMemoryRepository()
	ctx := context.Background()

	const numGoroutines = 10
	const entriesPerGoroutine = 5

	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines*entriesPerGoroutine)

	// Test concurrent writes
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(routineID int) {
			defer wg.Done()

			for j := 0; j < entriesPerGoroutine; j++ {
				entityID := fmt.Sprintf("payment-%d-%d", routineID, j)
				entry := audit.NewAuditEntry(audit.EntityTypePayment, entityID, audit.ActionTypeCreated, "user-123")
				if err := repo.Save(ctx, entry); err != nil {
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

	// Verify all entries were saved by using FindByFilter
	filter := audit.AuditFilter{
		EntityType: &[]audit.EntityType{audit.EntityTypePayment}[0],
	}
	entries, err := repo.FindByFilter(ctx, filter)
	if err != nil {
		t.Errorf("failed to find all entries: %v", err)
		return
	}

	expectedCount := numGoroutines * entriesPerGoroutine
	if len(entries) != expectedCount {
		t.Errorf("expected %d entries, got %d", expectedCount, len(entries))
	}

	// Test concurrent reads
	var readWg sync.WaitGroup
	readErrors := make(chan error, len(entries))

	for _, entry := range entries {
		readWg.Add(1)
		go func(entryID audit.AuditID) {
			defer readWg.Done()

			_, err := repo.FindByID(ctx, entryID)
			if err != nil {
				readErrors <- err
			}
		}(entry.ID())
	}

	readWg.Wait()
	close(readErrors)

	// Check for read errors
	for err := range readErrors {
		t.Errorf("concurrent read error: %v", err)
	}
}

type auditEntrySetup struct {
	entityID  string
	action    audit.ActionType
	userID    string
	timestamp *time.Time
}

func createAuditEntryWithMetadata(entityID, userID string, metadata map[string]string) *audit.AuditEntry {
	entry := audit.NewAuditEntry(audit.EntityTypePayment, entityID, audit.ActionTypeCreated, userID)
	for key, value := range metadata {
		entry.AddMetadata(key, value)
	}
	return entry
}

func createAuditEntryWithData(entityID, userID string) *audit.AuditEntry {
	entry := audit.NewAuditEntry(audit.EntityTypePayment, entityID, audit.ActionTypeUpdated, userID)

	oldData := map[string]interface{}{"status": "pending"}
	newData := map[string]interface{}{"status": "processing"}

	entry.SetOldData(oldData)
	entry.SetNewData(newData)

	return entry
}
