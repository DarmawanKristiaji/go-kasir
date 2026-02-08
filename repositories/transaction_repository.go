package repositories

import (
	"database/sql"
	"fmt"

	"kasir-api/models"
)

type TransactionRepository struct {
	db *sql.DB
}

func NewTransactionRepository(db *sql.DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

func (repo *TransactionRepository) CreateTransaction(items []models.CheckoutItem) (*models.Transaction, error) {
	if len(items) == 0 {
		return nil, fmt.Errorf("items cannot be empty")
	}

	tx, err := repo.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	totalAmount := 0
	details := make([]models.TransactionDetail, 0)

	for _, item := range items {
		if item.Quantity <= 0 {
			return nil, fmt.Errorf("invalid quantity for product %d", item.ProductID)
		}

		var productPrice, stock int
		var productName string

		err := tx.QueryRow("SELECT name, price, stock FROM products WHERE id = $1", item.ProductID).Scan(&productName, &productPrice, &stock)
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("product id %d not found", item.ProductID)
		}
		if err != nil {
			return nil, err
		}

		if stock < item.Quantity {
			return nil, fmt.Errorf("stock not enough for product %d", item.ProductID)
		}

		subtotal := productPrice * item.Quantity
		totalAmount += subtotal

		_, err = tx.Exec("UPDATE products SET stock = stock - $1 WHERE id = $2", item.Quantity, item.ProductID)
		if err != nil {
			return nil, err
		}

		details = append(details, models.TransactionDetail{
			ProductID:   item.ProductID,
			ProductName: productName,
			Quantity:    item.Quantity,
			Subtotal:    subtotal,
		})
	}

	var transactionID int
	err = tx.QueryRow("INSERT INTO transactions (total_amount) VALUES ($1) RETURNING id", totalAmount).Scan(&transactionID)
	if err != nil {
		return nil, err
	}

	for i := range details {
		details[i].TransactionID = transactionID
		_, err = tx.Exec("INSERT INTO transaction_details (transaction_id, product_id, quantity, subtotal) VALUES ($1, $2, $3, $4)",
			transactionID, details[i].ProductID, details[i].Quantity, details[i].Subtotal)
		if err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &models.Transaction{
		ID:          transactionID,
		TotalAmount: totalAmount,
		Details:     details,
	}, nil
}

func (repo *TransactionRepository) GetTodaySummary() (*models.ReportSummary, error) {
	var totalRevenue, totalTransaksi int

	err := repo.db.QueryRow(`
		SELECT COALESCE(SUM(total_amount), 0), COALESCE(COUNT(*), 0)
		FROM transactions
		WHERE created_at::date = CURRENT_DATE
	`).Scan(&totalRevenue, &totalTransaksi)
	if err != nil {
		return nil, err
	}

	var topName sql.NullString
	var topQty sql.NullInt64
	_ = repo.db.QueryRow(`
		SELECT p.name, COALESCE(SUM(td.quantity), 0) as qty
		FROM transaction_details td
		JOIN transactions t ON t.id = td.transaction_id
		JOIN products p ON p.id = td.product_id
		WHERE t.created_at::date = CURRENT_DATE
		GROUP BY p.name
		ORDER BY qty DESC
		LIMIT 1
	`).Scan(&topName, &topQty)

	summary := &models.ReportSummary{
		TotalRevenue:   totalRevenue,
		TotalTransaksi: totalTransaksi,
		ProdukTerlaris: models.ReportTopProduct{
			Nama:       "",
			QtyTerjual: 0,
		},
	}

	if topName.Valid && topQty.Valid {
		summary.ProdukTerlaris.Nama = topName.String
		summary.ProdukTerlaris.QtyTerjual = int(topQty.Int64)
	}

	return summary, nil
}
