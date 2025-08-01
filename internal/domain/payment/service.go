package payment

import (
	"context"
	"errors"
)

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{
		repository: repository,
	}
}

func (s *Service) CreatePayment(ctx context.Context, amount Amount, description string) (*Payment, error) {
	payment := NewPayment(amount, description)

	if err := s.repository.Save(ctx, payment); err != nil {
		return nil, err
	}

	return payment, nil
}

func (s *Service) GetPayment(ctx context.Context, id PaymentID) (*Payment, error) {
	return s.repository.FindByID(ctx, id)
}

func (s *Service) GetAllPayments(ctx context.Context) ([]*Payment, error) {
	return s.repository.FindAll(ctx)
}

func (s *Service) ProcessPayment(ctx context.Context, id PaymentID) error {
	payment, err := s.repository.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if payment == nil {
		return errors.New("payment not found")
	}

	if err := payment.Process(); err != nil {
		return err
	}

	return s.repository.Update(ctx, payment)
}

func (s *Service) CompletePayment(ctx context.Context, id PaymentID) error {
	payment, err := s.repository.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if payment == nil {
		return errors.New("payment not found")
	}

	if err := payment.Complete(); err != nil {
		return err
	}

	return s.repository.Update(ctx, payment)
}

func (s *Service) FailPayment(ctx context.Context, id PaymentID) error {
	payment, err := s.repository.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if payment == nil {
		return errors.New("payment not found")
	}

	if err := payment.Fail(); err != nil {
		return err
	}

	return s.repository.Update(ctx, payment)
}

func (s *Service) CancelPayment(ctx context.Context, id PaymentID) error {
	payment, err := s.repository.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if payment == nil {
		return errors.New("payment not found")
	}

	if err := payment.Cancel(); err != nil {
		return err
	}

	return s.repository.Update(ctx, payment)
}
