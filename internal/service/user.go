package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"healthcheckProject/internal/entity"
	"healthcheckProject/internal/utils"
)

type UserService struct {
	userRepo  UserRepo
	authGate  AuthGateway
	eventRepo EventRepo
	billing   BillingGateway
}

func NewUserService(
	repo UserRepo,
	authGate AuthGateway,
	eventRepo EventRepo,
	billing BillingGateway,
) *UserService {
	return &UserService{
		userRepo:  repo,
		authGate:  authGate,
		eventRepo: eventRepo,
		billing:   billing,
	}
}

func (s *UserService) CreateUser(ctx context.Context, args entity.AddUserArgs) (entity.User, error) {
	var resultUser entity.User

	err := s.userRepo.WithinTransaction(ctx, func(ctx context.Context) error {
		var err error

		start := time.Now()
		registerUser, err := s.authGate.RegisterUser(ctx, args)

		authRequestTiming := time.Now().Sub(start)
		log.Printf("register user in auth took %s", authRequestTiming.String())

		if err != nil {
			return fmt.Errorf("failed to register user in authGate: %w", err)
		}

		start = time.Now()

		resultUser, err = s.userRepo.AddUser(ctx, args, registerUser.PasswordHash)
		if err != nil {
			return fmt.Errorf("failed to add user: %w", err)
		}

		dbRequestTiming := time.Now().Sub(start)
		log.Printf("add user to db took %s", dbRequestTiming.String())

		return nil
	})

	if err != nil {
		return entity.User{}, fmt.Errorf("add user transaction is failed: %w", err)
	}

	start := time.Now()
	err = s.eventRepo.OnUserRegistered(ctx, resultUser)
	if err != nil {
		return entity.User{}, fmt.Errorf("add user event is failed: %w", err)
	}
	queueRequestTiming := time.Now().Sub(start)
	log.Printf("add user event to queue took %s", queueRequestTiming.String())

	return resultUser, nil
}

func (s *UserService) DeleteUser(ctx context.Context, id uint64) (entity.User, error) {
	user := utils.GetUser(ctx)
	if user.ID != id {
		return entity.User{}, fmt.Errorf("%w: user %d is not found (wrong auth)", ErrUserNotFound, id)
	}

	var resultUser entity.User

	err := s.userRepo.WithinTransaction(ctx, func(ctx context.Context) error {
		var err error
		resultUser, err = s.userRepo.GetUserByID(ctx, id)
		if err != nil {
			return fmt.Errorf("failed to get user by id: %w", err)
		}

		if resultUser.IsEmpty() {
			return fmt.Errorf("%w: user %d is not found", ErrUserNotFound, id)
		}

		err = s.userRepo.DeleteUser(ctx, id)
		if err != nil {
			return fmt.Errorf("failed to delete user: %w", err)
		}

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

func (s *UserService) UpdateUser(ctx context.Context, userID uint64, args UpdateUserArgs) (entity.User, error) {
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

func (s *UserService) GetUser(ctx context.Context, id uint64) (entity.User, error) {
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

	userAccountByID, err := s.billing.GetUserAccountByID(ctx, user.ID)
	if err != nil {
		return entity.User{}, fmt.Errorf("failed get userAccount by id err: %v", err)
	}

	userByID.UserAccount = userAccountByID

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

func (s *UserService) InternalGetUserByID(ctx context.Context, userID uint64) (entity.User, error) {
	userByID, err := s.userRepo.GetUserByID(ctx, userID)

	if err != nil {
		return entity.User{}, fmt.Errorf("failed get user by login: %w", err)
	}

	if userByID.IsEmpty() {
		return entity.User{}, fmt.Errorf("%w: user %d is not found", ErrUserNotFound, userID)
	}

	return userByID, nil
}
