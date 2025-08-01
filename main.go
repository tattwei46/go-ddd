package main

import (
	"context"
	"fmt"
	"log"

	"go-ddd/internal/application"
	"go-ddd/internal/domain/audit"
	"go-ddd/internal/domain/payment"
	"go-ddd/internal/infrastructure/repository"
)

func main() {
	ctx := context.Background()
	
	paymentRepo := repository.NewPaymentMemoryRepository()
	auditRepo := repository.NewAuditMemoryRepository()
	
	paymentDomainService := payment.NewService(paymentRepo)
	auditDomainService := audit.NewService(auditRepo)
	
	paymentAppService := application.NewPaymentApplicationService(paymentDomainService, auditDomainService)
	
	fmt.Println("=== Payment Service with Audit Demo ===\n")
	
	userID := "user-123"
	
	fmt.Println("1. Creating a payment...")
	p, err := paymentAppService.CreatePayment(ctx, 100.50, "USD", "Online purchase", userID)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Created payment: ID=%s, Amount=%.2f %s, Status=%s\n\n", 
		p.ID().String(), p.Amount().Value(), p.Amount().Currency(), p.Status().String())
	
	paymentID := p.ID().String()
	
	fmt.Println("2. Processing the payment...")
	err = paymentAppService.ProcessPayment(ctx, paymentID, userID)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Payment processed successfully\n")
	
	fmt.Println("3. Completing the payment...")
	err = paymentAppService.CompletePayment(ctx, paymentID, userID)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Payment completed successfully\n")
	
	fmt.Println("4. Retrieving payment audit history...")
	auditEntries, err := paymentAppService.GetPaymentAuditHistory(ctx, paymentID)
	if err != nil {
		log.Fatal(err)
	}
	
	fmt.Printf("Found %d audit entries for payment %s:\n", len(auditEntries), paymentID)
	for i, entry := range auditEntries {
		fmt.Printf("  %d. Action: %s, User: %s, Time: %s\n", 
			i+1, entry.Action(), entry.UserID(), entry.Timestamp().Format("2006-01-02 15:04:05"))
		if len(entry.NewData()) > 0 {
			fmt.Printf("     New Data: %+v\n", entry.NewData())
		}
	}
	
	fmt.Println("\n=== Demo completed successfully! ===")
}