package transaction

import (
	"context"
	"fmt"
	"github.com/Masterminds/squirrel"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

type repository struct {
	dbPool *pgxpool.Pool
}

func newRepository(dbPool *pgxpool.Pool) *repository {
	return &repository{
		dbPool: dbPool,
	}
}

func (r *repository) ensurePartition(ctx context.Context, monthKey string) error {
	t, err := time.Parse("2006_01", monthKey)
	if err != nil {
		return fmt.Errorf("failed to parse monthKey: %w", err)
	}

	partitionName := fmt.Sprintf("transaction_%s", monthKey)
	startDate := t.Format("2006-01-02")
	endDate := t.AddDate(0, 1, 0).Format("2006-01-02")

	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS "%s"
		PARTITION OF transaction
		FOR VALUES FROM ('%s') TO ('%s');
	`, partitionName, startDate, endDate)

	_, err = r.dbPool.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to create partition %s: %w", partitionName, err)
	}

	return nil
}

func (r *repository) enrichedTransactionList(ctx context.Context, month int32, year int32, userID int64, bankID int64) ([]EnrichedTransaction, error) {
	queryBuilder := squirrel.
		Select(
			"t.id",
			"t.bank_id",
			"t.external_id",
			"t.user_id",
			"t.transaction_date",
			"t.amount",
			"t.category_id",
			"t.description",
			"t.type",
			"t.created_at",
			"b.name AS bank_name",
			"c.name AS category_name",
			"u.name AS user_name",
		).
		From("transaction t").
		LeftJoin("bank b ON t.bank_id = b.id").
		LeftJoin("category c ON t.category_id = c.id").
		LeftJoin("users u ON t.user_id = u.id").
		Where("EXTRACT(MONTH FROM t.transaction_date) = ?", month).
		Where("EXTRACT(YEAR FROM t.transaction_date) = ?", year).
		OrderBy("t.id").
		PlaceholderFormat(squirrel.Dollar)

	if userID > 0 {
		queryBuilder = queryBuilder.Where(squirrel.Eq{"t.user_id": userID})
	}
	if bankID > 0 {
		queryBuilder = queryBuilder.Where(squirrel.Eq{"t.bank_id": bankID})
	}

	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build SQL: %w", err)
	}

	var transactions []EnrichedTransaction
	if err = pgxscan.Select(ctx, r.dbPool, &transactions, query, args...); err != nil {
		return nil, fmt.Errorf("failed to select transactions: %w", err)
	}

	return transactions, nil
}

func (r *repository) getEnrichedTransaction(ctx context.Context, id int64) (*EnrichedTransaction, error) {
	queryBuilder := squirrel.
		Select(
			"t.id",
			"t.bank_id",
			"t.external_id",
			"t.user_id",
			"t.transaction_date",
			"t.amount",
			"t.category_id",
			"t.description",
			"t.type",
			"t.created_at",
			"b.name AS bank_name",
			"c.name AS category_name",
			"u.name AS user_name",
		).
		From("transaction t").
		LeftJoin("bank b ON t.bank_id = b.id").
		LeftJoin("category c ON t.category_id = c.id").
		LeftJoin("users u ON t.user_id = u.id").
		Where(squirrel.Eq{"t.id": id}).
		PlaceholderFormat(squirrel.Dollar)

	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build SQL: %w", err)
	}

	var transaction EnrichedTransaction
	if err = pgxscan.Get(ctx, r.dbPool, &transaction, query, args...); err != nil {
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}

	return &transaction, nil
}

// todo
func (r *repository) getTransaction(ctx context.Context, id int64) (*Transaction, error) {
	queryBuilder := squirrel.
		Select(
			"id",
			"external_id",
			"bank_id",
			"user_id",
			"amount",
			"category_id",
			"description",
			"type",
			"transaction_date",
			"created_at",
			"updated_at",
		).
		From(transactionTable).
		Where(squirrel.Eq{"id": id}).
		PlaceholderFormat(squirrel.Dollar)

	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build SQL: %w", err)
	}

	var transaction Transaction
	if err = pgxscan.Get(ctx, r.dbPool, &transaction, query, args...); err != nil {
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}

	return &transaction, nil
}

func (r *repository) saveTransaction(ctx context.Context, transactions []*Transaction) error {
	builder := squirrel.
		Insert(transactionTable).
		Columns("bank_id", "external_id", "user_id", "transaction_date", "amount", "category_id", "description", "type").
		PlaceholderFormat(squirrel.Dollar)

	for _, t := range transactions {
		builder = builder.Values(
			t.BankID,
			t.ExternalID,
			t.UserID,
			t.TransactionDate,
			t.Amount,
			t.CategoryID,
			t.Description,
			t.Type,
		)
	}

	query, args, err := builder.Suffix("ON CONFLICT ON CONSTRAINT uniq_transaction_external DO NOTHING").ToSql()
	if err != nil {
		return fmt.Errorf("failed to build SQL: %w", err)
	}

	_, err = r.dbPool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to execute insert: %w", err)
	}

	return nil
}

func (r *repository) updateTransaction(ctx context.Context, tx *EnrichedTransaction) (*Transaction, error) {
	query, args, err := squirrel.
		Update("transaction").
		Set("category_id", tx.CategoryID).
		Set("type", tx.Type).
		Where(squirrel.Eq{"id": tx.ID}).
		Suffix("RETURNING id, bank_id, external_id, user_id, transaction_date, amount, category_id, description, created_at, type").
		PlaceholderFormat(squirrel.Dollar).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("failed to build update SQL: %w", err)
	}

	var updatedTr Transaction
	err = pgxscan.Get(ctx, r.dbPool, &updatedTr, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to update transaction: %w", err)
	}

	return &updatedTr, nil
}
