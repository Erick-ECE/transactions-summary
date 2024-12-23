package usecases

import (
	"fmt"

	"transactions-summary/internal/entities"
	"transactions-summary/internal/interfaces"
)

// GenerateSummary processes transactions and creates a summary.
type GenerateSummary struct {
	TransactionRepo interfaces.TransactionRepository
}

// NewGenerateSummary creates a new GenerateSummary use case.
func NewGenerateSummary(repo interfaces.TransactionRepository) *GenerateSummary {
	return &GenerateSummary{
		TransactionRepo: repo,
	}
}

// Execute calculates the summary from all transactions in the database.
func (uc *GenerateSummary) Execute(accountId string, transactions []entities.Transaction) (*entities.SummaryResult, string, error) {
	// Calculate summary data
	totalCredit := 0.0
	totalDebit := 0.0
	monthlyData := make(map[string]*entities.MonthlySummary)

	// Process each transaction
	for _, transaction := range transactions {

		// Get month name (e.g., "July")
		monthName := transaction.TransactionDate.Format("January")

		// Initialize monthly summary if not present
		if _, exists := monthlyData[monthName]; !exists {
			monthlyData[monthName] = &entities.MonthlySummary{
				Month: monthName,
			}
		}

		// Update monthly data
		monthlySummary := monthlyData[monthName]
		monthlySummary.NumTransactions++

		if transaction.Type == "credit" {
			totalCredit += transaction.Amount
			monthlySummary.TotalCredits += transaction.Amount
		} else if transaction.Type == "debit" {
			totalDebit += transaction.Amount
			monthlySummary.TotalDebits += transaction.Amount
		}
	}

	// Calculate averages for each month
	var monthlySummaries []entities.MonthlySummary
	for _, summary := range monthlyData {
		if summary.NumTransactions > 0 {
			// Calculate averages
			creditCount := float64(countCredits(transactions, summary.Month))
			debitCount := float64(countDebits(transactions, summary.Month))

			if creditCount > 0 {
				summary.AverageCredit = summary.TotalCredits / creditCount
			}
			if debitCount > 0 {
				summary.AverageDebit = summary.TotalDebits / debitCount
			}
		}
		monthlySummaries = append(monthlySummaries, *summary)
	}

	account, err := uc.TransactionRepo.GetAccount(accountId)
	if err != nil {
		return nil, "", fmt.Errorf("could not retrieve account %s: %v", accountId, err)
	}

	return &entities.SummaryResult{
		TotalCredit:      totalCredit,
		TotalDebit:       totalDebit,
		MonthlySummaries: monthlySummaries,
	}, account.Email, nil
}

// Helper functions for counting transactions by type
func countCredits(transactions []entities.Transaction, month string) int {
	count := 0
	for _, transaction := range transactions {
		if transaction.TransactionDate.Format("January") == month && transaction.Type == "credit" {
			count++
		}
	}
	return count
}

func countDebits(transactions []entities.Transaction, month string) int {
	count := 0
	for _, transaction := range transactions {
		if transaction.TransactionDate.Format("January") == month && transaction.Type == "debit" {
			count++
		}
	}
	return count
}
