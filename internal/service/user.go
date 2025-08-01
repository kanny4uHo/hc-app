package service

import (
	"context"
	"fmt"

	"healthcheckProject/internal/entity"
	"healthcheckProject/internal/utils"
)

type UserService struct {
	userRepo UserRepo
	authGate AuthGateway
}

func NewUserService(repo UserRepo, authGate AuthGateway) *UserService {
	return &UserService{
		userRepo: repo,
		authGate: authGate,
	}
}

func (s *UserService) CreateUser(ctx context.Context, args entity.AddUserArgs) (entity.User, error) {
	var resultUser entity.User

	err := s.userRepo.WithinTransaction(ctx, func(ctx context.Context) error {
		var err error

		registerUser, err := s.authGate.RegisterUser(ctx, args)
		if err != nil {
			return fmt.Errorf("failed to register user in authGate: %w", err)
		}

		resultUser, err = s.userRepo.AddUser(ctx, args, registerUser.PasswordHash)
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

func (s *UserService) DeleteUser(ctx context.Context, id int64) (entity.User, error) {
	user := utils.GetUser(ctx)
	if user.ID != id {
		return entity.User{}, fmt.Errorf("%w: user %d is not found (wrong auth)", ErrUserNotFound, id)
	}

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

func (s *UserService) UpdateUser(ctx context.Context, userID int64, args UpdateUserArgs) (entity.User, error) {
	user := utils.GetUser(ctx)
	if user.ID != userID {
		return entity.User{}, fmt.Errorf("%w: user %d is not found (wrong auth)", ErrUserNotFound, userID)
	}

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

func (s *UserService) GetUser(ctx context.Context, id int64) (entity.User, error) {
	user := utils.GetUser(ctx)
	if user.ID != id {
		return entity.User{}, fmt.Errorf("%w: user %d is not found (wrong auth)", ErrUserNotFound, id)
	}

	userByID, err := s.userRepo.GetUserByID(ctx, id)

	if err != nil {
		return entity.User{}, fmt.Errorf("failed get user by id err: %v", err)
	}

	if userByID.IsEmpty() {
		return entity.User{}, fmt.Errorf("%w: no user for id %d", ErrUserNotFound, id)
	}

	return userByID, nil
}

func (s *UserService) InternalGetUserByLogin(ctx context.Context, login string) (entity.User, error) {
	userByLogin, err := s.userRepo.GetUserByLogin(ctx, login)

	if err != nil {
		return entity.User{}, fmt.Errorf("failed get user by login: %w", err)
	}

	if userByLogin.IsEmpty() {
		return entity.User{}, fmt.Errorf("%w: user %s is not found", ErrUserNotFound, login)
	}

	return userByLogin, nil
}
