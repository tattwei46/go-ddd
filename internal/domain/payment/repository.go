package payment

import "context"

type Repository interface {
	Save(ctx context.Context, payment *Payment) error
	FindByID(ctx context.Context, id PaymentID) (*Payment, error)
	FindAll(ctx context.Context) ([]*Payment, error)
	Update(ctx context.Context, payment *Payment) error
	Delete(ctx context.Context, id PaymentID) error
}