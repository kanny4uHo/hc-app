package repository

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"sync"
	"sync/atomic"

	"healthcheckProject/internal/entity"
	"healthcheckProject/internal/service"
)

type UserMap struct {
	userMap map[int64]entity.User
	mutex   *sync.Mutex
	lastID  *atomic.Int64
}

func NewUserMap() UserMap {
	return UserMap{
		userMap: make(map[int64]entity.User),
		mutex:   &sync.Mutex{},
		lastID:  &atomic.Int64{},
	}
}

func (u UserMap) WithinTransaction(ctx context.Context, f func(ctx context.Context) error) error {
	u.mutex.Lock()
	defer u.mutex.Unlock()

	return f(ctx)
}

func (u UserMap) AddUser(_ context.Context, user entity.AddUserArgs) (entity.User, error) {
	newUserID := u.lastID.Add(1)

	passwordHash := md5.Sum([]byte(user.Password))

	newUser := entity.User{
		ID:           newUserID,
		Login:        user.Login,
		PasswordHash: hex.EncodeToString(passwordHash[:]),
		Meta:         user.Meta,
	}

	u.userMap[newUserID] = newUser

	return newUser, nil
}

func (u UserMap) GetUserByID(_ context.Context, id int64) (entity.User, error) {
	return u.userMap[id], nil
}

func (u UserMap) GetUserByLogin(_ context.Context, login string) (entity.User, error) {
	for _, user := range u.userMap {
		if user.Login == login {
			return user, nil
		}
	}

	return entity.User{}, nil
}

func (u UserMap) GetUserByEmail(_ context.Context, email string) (entity.User, error) {
	for _, user := range u.userMap {
		if user.Meta.Email == email {
			return user, nil
		}
	}

	return entity.User{}, nil
}

func (u UserMap) UpdateUser(_ context.Context, id int64, args service.UpdateUserArgs) error {
	userToUpdate, ok := u.userMap[id]
	if !ok {
		return fmt.Errorf("user %d not found", id)
	}

	if args.Email != "" {
		userToUpdate.Meta.Email = args.Email
	}

	if args.FirstName != "" {
		userToUpdate.Meta.Name.First = args.FirstName
	}

	if args.LastName != "" {
		userToUpdate.Meta.Name.Last = args.LastName
	}

	u.userMap[id] = userToUpdate

	return nil
}

func (u UserMap) DeleteUser(_ context.Context, id int64) error {
	delete(u.userMap, id)
	return nil
}
