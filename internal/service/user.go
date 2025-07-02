package service

import (
	"context"
	"fmt"

	"healthcheckProject/internal/entity"
)

type UserService struct {
	userRepo UserRepo
}

func NewUserService(repo UserRepo) UserService {
	return UserService{
		userRepo: repo,
	}
}

func (s UserService) CreateUser(ctx context.Context, args entity.AddUserArgs) (entity.User, error) {
	var resultUser entity.User

	err := s.userRepo.WithinTransaction(ctx, func(ctx context.Context) error {
		userByLogin, err := s.userRepo.GetUserByLogin(ctx, args.Login)
		if err != nil {
			return fmt.Errorf("failed to get user by login: %w", err)
		}

		if !userByLogin.IsEmpty() {
			return InvalidArgumentError{
				Field:  "login",
				Value:  args.Login,
				Reason: "user already exists",
			}
		}

		userByEmail, err := s.userRepo.GetUserByEmail(ctx, args.Meta.Email)
		if err != nil {
			return fmt.Errorf("failed to get user by login: %w", err)
		}

		if !userByEmail.IsEmpty() {
			return InvalidArgumentError{
				Field:  "email",
				Value:  args.Meta.Email,
				Reason: "user already exists",
			}
		}

		resultUser, err = s.userRepo.AddUser(ctx, args)
		if err != nil {
			return fmt.Errorf("failed to add user: %w", err)
		}

		return nil
	})

	if err != nil {
		return resultUser, fmt.Errorf("add user transaction is failed: %w", err)
	}

	return resultUser, nil
}

func (s UserService) DeleteUser(ctx context.Context, id int64) (entity.User, error) {
	var resultUser entity.User

	err := s.userRepo.WithinTransaction(ctx, func(ctx context.Context) error {
		user, err := s.userRepo.GetUserByID(ctx, id)
		if err != nil {
			return fmt.Errorf("failed to get user by id: %w", err)
		}

		if user.IsEmpty() {
			return fmt.Errorf("%w: user %d is not found", ErrUserNotFound, id)
		}

		err = s.userRepo.DeleteUser(ctx, id)
		if err != nil {
			return fmt.Errorf("failed to delete user: %w", err)
		}

		resultUser = user
		return nil
	})

	if err != nil {
		return entity.User{}, fmt.Errorf("delete transaction is failed: %w", err)
	}

	return resultUser, nil
}

type UpdateUserArgs struct {
	FirstName string
	LastName  string
	Email     string
}

func (s UserService) UpdateUser(ctx context.Context, userID int64, args UpdateUserArgs) (entity.User, error) {
	var resultUser entity.User

	err := s.userRepo.WithinTransaction(ctx, func(ctx context.Context) error {
		user, err := s.userRepo.GetUserByID(ctx, userID)
		if err != nil {
			return fmt.Errorf("failed to get user by id: %w", err)
		}

		if user.IsEmpty() {
			return fmt.Errorf("%w: user is empty", ErrUserNotFound)
		}

		err = s.userRepo.UpdateUser(ctx, userID, args)
		if err != nil {
			return fmt.Errorf("failed to update user: %w", err)
		}

		resultUser, err = s.userRepo.GetUserByID(ctx, userID)
		if err != nil {
			return fmt.Errorf("failed to get user by id after update: %w", err)
		}

		return nil
	})

	if err != nil {
		return entity.User{}, fmt.Errorf("update transaction is failed: %w", err)
	}

	return resultUser, nil
}

func (s UserService) GetUser(ctx context.Context, id int64) (entity.User, error) {
	userByID, err := s.userRepo.GetUserByID(ctx, id)

	if err != nil {
		return entity.User{}, fmt.Errorf("failed get user by id err: %v", err)
	}

	if userByID.IsEmpty() {
		return entity.User{}, fmt.Errorf("%w: no user for id %d", ErrUserNotFound, id)
	}

	return userByID, nil
}
