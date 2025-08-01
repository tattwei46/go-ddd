package audit

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type AuditID struct {
	value string
}

func NewAuditID() AuditID {
	return AuditID{value: uuid.New().String()}
}

func (id AuditID) String() string {
	return id.value
}

type EntityType string

const (
	EntityTypePayment EntityType = "payment"
)

type ActionType string

const (
	ActionTypeCreated   ActionType = "created"
	ActionTypeUpdated   ActionType = "updated"
	ActionTypeDeleted   ActionType = "deleted"
	ActionTypeProcessed ActionType = "processed"
	ActionTypeCompleted ActionType = "completed"
	ActionTypeFailed    ActionType = "failed"
	ActionTypeCancelled ActionType = "cancelled"
)

type AuditEntry struct {
	id         AuditID
	entityType EntityType
	entityID   string
	action     ActionType
	oldData    map[string]interface{}
	newData    map[string]interface{}
	userID     string
	timestamp  time.Time
	metadata   map[string]string
}

func NewAuditEntry(entityType EntityType, entityID string, action ActionType, userID string) *AuditEntry {
	return &AuditEntry{
		id:         NewAuditID(),
		entityType: entityType,
		entityID:   entityID,
		action:     action,
		oldData:    make(map[string]interface{}),
		newData:    make(map[string]interface{}),
		userID:     userID,
		timestamp:  time.Now(),
		metadata:   make(map[string]string),
	}
}

func (a *AuditEntry) ID() AuditID {
	return a.id
}

func (a *AuditEntry) EntityType() EntityType {
	return a.entityType
}

func (a *AuditEntry) EntityID() string {
	return a.entityID
}

func (a *AuditEntry) Action() ActionType {
	return a.action
}

func (a *AuditEntry) OldData() map[string]interface{} {
	return a.oldData
}

func (a *AuditEntry) NewData() map[string]interface{} {
	return a.newData
}

func (a *AuditEntry) UserID() string {
	return a.userID
}

func (a *AuditEntry) Timestamp() time.Time {
	return a.timestamp
}

func (a *AuditEntry) Metadata() map[string]string {
	return a.metadata
}

func (a *AuditEntry) SetOldData(data interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	var dataMap map[string]interface{}
	if err := json.Unmarshal(jsonData, &dataMap); err != nil {
		return err
	}

	a.oldData = dataMap
	return nil
}

func (a *AuditEntry) SetNewData(data interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	var dataMap map[string]interface{}
	if err := json.Unmarshal(jsonData, &dataMap); err != nil {
		return err
	}

	a.newData = dataMap
	return nil
}

func (a *AuditEntry) AddMetadata(key, value string) {
	a.metadata[key] = value
}

type AuditFilter struct {
	EntityType *EntityType
	EntityID   *string
	Action     *ActionType
	UserID     *string
	FromDate   *time.Time
	ToDate     *time.Time
}
