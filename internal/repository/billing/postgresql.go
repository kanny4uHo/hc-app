package billing

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"healthcheckProject/internal/entity"
	"healthcheckProject/internal/service"
)

type UserAccountRepoImpl struct {
	db *sql.DB
}

var _ service.UserAccountRepo = (*UserAccountRepoImpl)(nil)

func NewUserAccountRepoImpl(db *sql.DB) *UserAccountRepoImpl {
	return &UserAccountRepoImpl{
		db: db,
	}
}

func (r *UserAccountRepoImpl) GetUserAccountByID(ctx context.Context, userID uint64) (entity.UserAccount, error) {
	var rowID int64
	var userIDRow int64
	var balanceRow int64

	err := r.db.QueryRowContext(ctx, "SELECT * FROM accounts WHERE user_id = $1 FOR UPDATE ", userID).
		Scan(&rowID, &userIDRow, &balanceRow)

	if err != nil {
		return entity.UserAccount{}, fmt.Errorf("failed to get user by id: %w", err)
	}

	return entity.UserAccount{
		ID:      userID,
		UserID:  userID,
		Balance: balanceRow,
	}, nil
}

func (r *UserAccountRepoImpl) SetBalance(ctx context.Context, userID uint64, newBalance int64) (int64, error) {
	_, err := r.db.ExecContext(ctx, "UPDATE accounts set balance = $1 WHERE user_id = $2 ", newBalance, userID)
	if err != nil {
		return 0, fmt.Errorf("could not update account: %w", err)
	}

	return newBalance, nil
}

func (r *UserAccountRepoImpl) SaveUser(ctx context.Context, account entity.UserAccount) error {
	_, err := r.db.ExecContext(
		ctx,
		`INSERT INTO accounts (user_id, balance) VALUES ($1, $2)`,
		account.UserID,
		account.Balance,
	)

	if err != nil {
		return fmt.Errorf("failed to execute insert: %w", err)
	}

	return nil
}

func (r *UserAccountRepoImpl) WithinTransaction(ctx context.Context, f func(ctx context.Context) error) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}

	defer func() {
		if err != nil {
			txErr := tx.Rollback()
			if txErr != nil {
				log.Printf("failed to rollback transaction: %s", err)
			}
		} else {
			txErr := tx.Commit()
			if txErr != nil {
				log.Printf("failed to commit transaction: %s", err)
			}
		}
	}()

	if err = f(ctx); err != nil {
		return fmt.Errorf("failed to execute transaction: %w", err)
	}

	return nil
}
