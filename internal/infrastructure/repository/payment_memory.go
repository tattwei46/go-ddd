package repository

import (
	"context"
	"errors"
	"sync"

	"go-ddd/internal/domain/payment"
)

type PaymentMemoryRepository struct {
	mu       sync.RWMutex
	payments map[string]*payment.Payment
}

func NewPaymentMemoryRepository() *PaymentMemoryRepository {
	return &PaymentMemoryRepository{
		payments: make(map[string]*payment.Payment),
	}
}

func (r *PaymentMemoryRepository) Save(ctx context.Context, p *payment.Payment) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.payments[p.ID().String()] = p
	return nil
}

func (r *PaymentMemoryRepository) FindByID(ctx context.Context, id payment.PaymentID) (*payment.Payment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	p, exists := r.payments[id.String()]
	if !exists {
		return nil, errors.New("payment not found")
	}

	return p, nil
}

func (r *PaymentMemoryRepository) FindAll(ctx context.Context) ([]*payment.Payment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	payments := make([]*payment.Payment, 0, len(r.payments))
	for _, p := range r.payments {
		payments = append(payments, p)
	}

	return payments, nil
}

func (r *PaymentMemoryRepository) Update(ctx context.Context, p *payment.Payment) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.payments[p.ID().String()]; !exists {
		return errors.New("payment not found")
	}

	r.payments[p.ID().String()] = p
	return nil
}

func (r *PaymentMemoryRepository) Delete(ctx context.Context, id payment.PaymentID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.payments[id.String()]; !exists {
		return errors.New("payment not found")
	}

	delete(r.payments, id.String())
	return nil
}
