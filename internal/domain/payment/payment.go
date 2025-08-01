package payment

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type PaymentID struct {
	value string
}

func NewPaymentID() PaymentID {
	return PaymentID{value: uuid.New().String()}
}

func PaymentIDFromString(id string) PaymentID {
	return PaymentID{value: id}
}

func (id PaymentID) String() string {
	return id.value
}

type Amount struct {
	value    float64
	currency string
}

func NewAmount(value float64, currency string) (Amount, error) {
	if value < 0 {
		return Amount{}, errors.New("amount cannot be negative")
	}
	if currency == "" {
		return Amount{}, errors.New("currency cannot be empty")
	}
	return Amount{value: value, currency: currency}, nil
}

func (a Amount) Value() float64 {
	return a.value
}

func (a Amount) Currency() string {
	return a.currency
}

type PaymentStatus int

const (
	PaymentStatusPending PaymentStatus = iota
	PaymentStatusProcessing
	PaymentStatusCompleted
	PaymentStatusFailed
	PaymentStatusCancelled
)

func (s PaymentStatus) String() string {
	switch s {
	case PaymentStatusPending:
		return "pending"
	case PaymentStatusProcessing:
		return "processing"
	case PaymentStatusCompleted:
		return "completed"
	case PaymentStatusFailed:
		return "failed"
	case PaymentStatusCancelled:
		return "cancelled"
	default:
		return "unknown"
	}
}

type Payment struct {
	id          PaymentID
	amount      Amount
	status      PaymentStatus
	description string
	createdAt   time.Time
	updatedAt   time.Time
}

func NewPayment(amount Amount, description string) *Payment {
	now := time.Now()
	return &Payment{
		id:          NewPaymentID(),
		amount:      amount,
		status:      PaymentStatusPending,
		description: description,
		createdAt:   now,
		updatedAt:   now,
	}
}

func (p *Payment) ID() PaymentID {
	return p.id
}

func (p *Payment) Amount() Amount {
	return p.amount
}

func (p *Payment) Status() PaymentStatus {
	return p.status
}

func (p *Payment) Description() string {
	return p.description
}

func (p *Payment) CreatedAt() time.Time {
	return p.createdAt
}

func (p *Payment) UpdatedAt() time.Time {
	return p.updatedAt
}

func (p *Payment) Process() error {
	if p.status != PaymentStatusPending {
		return errors.New("payment can only be processed from pending status")
	}
	p.status = PaymentStatusProcessing
	p.updatedAt = time.Now()
	return nil
}

func (p *Payment) Complete() error {
	if p.status != PaymentStatusProcessing {
		return errors.New("payment can only be completed from processing status")
	}
	p.status = PaymentStatusCompleted
	p.updatedAt = time.Now()
	return nil
}

func (p *Payment) Fail() error {
	if p.status == PaymentStatusCompleted {
		return errors.New("completed payment cannot be failed")
	}
	p.status = PaymentStatusFailed
	p.updatedAt = time.Now()
	return nil
}

func (p *Payment) Cancel() error {
	if p.status == PaymentStatusCompleted || p.status == PaymentStatusProcessing {
		return errors.New("payment cannot be cancelled in current status")
	}
	p.status = PaymentStatusCancelled
	p.updatedAt = time.Now()
	return nil
}