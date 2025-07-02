package repository

import (
	"context"
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"log"

	"healthcheckProject/internal/entity"
	"healthcheckProject/internal/service"
)

var selectByIDQuery = "SELECT id, username, email, password_hash, first_name, last_name FROM users WHERE id = $1 LIMIT 1"
var selectByEmailQuery = "SELECT id, username, email, password_hash, first_name, last_name FROM users WHERE username = $1 LIMIT 1"
var selectByLoginQuery = "SELECT id, username, email, password_hash, first_name, last_name FROM users WHERE email = $1 LIMIT 1"

func NewPgRepo(db *sql.DB) PgUserRepository {
	return PgUserRepository{
		sql: db,
	}
}

type PgUserRepository struct {
	sql *sql.DB
}

func (p PgUserRepository) WithinTransaction(ctx context.Context, f func(ctx context.Context) error) error {
	tx, err := p.sql.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}

	defer func() {
		txErr := tx.Commit()
		if txErr != nil {
			log.Printf("failed to commit transaction: %s", err)
		}
	}()

	if err = f(ctx); err != nil {
		return fmt.Errorf("failed to execute transaction: %w", err)
	}

	return nil
}

func (p PgUserRepository) AddUser(ctx context.Context, user entity.AddUserArgs) (entity.User, error) {
	passwordHash := md5.Sum([]byte(user.Password))

	var userID int64

	err := p.sql.QueryRowContext(
		ctx,
		`INSERT INTO users (username, email, password_hash, first_name, last_name) VALUES ($1, $2, $3, $4, $5) returning id`,
		user.Login,
		user.Meta.Email,
		hex.EncodeToString(passwordHash[:]),
		user.Meta.Name.First,
		user.Meta.Name.Last,
	).Scan(&userID)

	if err != nil {
		return entity.User{}, fmt.Errorf("failed to execute insert: %w", err)
	}

	return p.GetUserByID(ctx, userID)
}

func (p PgUserRepository) GetUserByID(ctx context.Context, id int64) (entity.User, error) {
	return p.selectOneUserByQuery(ctx, selectByIDQuery, id)
}

func (p PgUserRepository) GetUserByLogin(ctx context.Context, login string) (entity.User, error) {
	return p.selectOneUserByQuery(ctx, selectByLoginQuery, login)
}

func (p PgUserRepository) GetUserByEmail(ctx context.Context, email string) (entity.User, error) {
	return p.selectOneUserByQuery(ctx, selectByEmailQuery, email)
}

func (p PgUserRepository) UpdateUser(ctx context.Context, id int64, args service.UpdateUserArgs) error {
	if args.LastName != "" {
		_, err := p.sql.ExecContext(ctx, `UPDATE users set last_name = $1 where id = $2`, args.LastName, id)
		if err != nil {
			return fmt.Errorf("failed to execute update query - last name: %w", err)
		}
	}

	if args.FirstName != "" {
		_, err := p.sql.ExecContext(ctx, `UPDATE users set first_name = $1 where id = $2`, args.FirstName, id)
		if err != nil {
			return fmt.Errorf("failed to execute update query - first name: %w", err)
		}
	}

	if args.Email != "" {
		_, err := p.sql.ExecContext(ctx, `UPDATE users set email = $1 where id = $2`, args.Email, id)
		if err != nil {
			return fmt.Errorf("failed to execute update query - email: %w", err)
		}
	}

	return nil
}

func (p PgUserRepository) DeleteUser(ctx context.Context, id int64) error {
	_, err := p.sql.ExecContext(ctx, `DELETE FROM users WHERE id=$1`, id)
	if err != nil {
		return fmt.Errorf("failed to exec delete query: %w", err)
	}

	return nil
}

func (p PgUserRepository) selectOneUserByQuery(ctx context.Context, query string, args ...any) (entity.User, error) {
	var id int64
	var login string
	var email string
	var passwordHashFromDB string
	var firstName string
	var lastName string

	err := p.sql.QueryRowContext(ctx, query, args...).
		Scan(&id, &login, &email, &passwordHashFromDB, &firstName, &lastName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.User{}, nil
		}

		return entity.User{}, fmt.Errorf("failed to query user: %w", err)
	}

	return entity.User{
		ID:           id,
		Login:        login,
		PasswordHash: passwordHashFromDB,
		Meta: entity.UserMeta{
			Name: entity.UserName{
				First: firstName,
				Last:  lastName,
			},
			Email: email,
		},
	}, nil
}
