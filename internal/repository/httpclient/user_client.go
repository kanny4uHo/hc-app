package httpclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

const getUserInfoByLoginPath = "/internal/api/v1/user/by_login/"
const getUserInfoByIDPath = "/internal/api/v1/user/by_id/"

type UserClient struct {
	baseUserHost string
	httpClient   *http.Client
}

var ErrUserNotFound = fmt.Errorf("user not found")

type UserInfo struct {
	UserID       int    `json:"user_id"`
	Login        string `json:"username"`
	Email        string `json:"email"`
	PasswordHash string `json:"password_hash"`
}

func NewUserClient(baseHost string, httpClient *http.Client) *UserClient {
	return &UserClient{
		baseUserHost: baseHost,
		httpClient:   httpClient,
	}
}

func (c *UserClient) GetUserInfo(_ context.Context, login string) (UserInfo, error) {
	response, err := c.httpClient.Get(c.baseUserHost + getUserInfoByLoginPath + login)
	if err != nil {
		return UserInfo{}, fmt.Errorf("failed to get user info by http and login %s: %w", login, err)
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		if response.StatusCode == http.StatusNotFound {
			return UserInfo{}, ErrUserNotFound
		}

		return UserInfo{}, fmt.Errorf("failed to get user info, status: %d", response.StatusCode)
	}

	var userInfo UserInfo

	err = json.NewDecoder(response.Body).Decode(&userInfo)

	if err != nil {
		return UserInfo{}, fmt.Errorf("failed to decode user info")
	}

	return userInfo, nil
}

func (c *UserClient) GetUserInfoByID(_ context.Context, userID uint64) (UserInfo, error) {
	requestPath := fmt.Sprintf("%s%s%d", c.baseUserHost, getUserInfoByIDPath, userID)
	response, err := c.httpClient.Get(requestPath)
	if err != nil {
		return UserInfo{}, fmt.Errorf("failed to get user info by id %s: %w", userID, err)
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		if response.StatusCode == http.StatusNotFound {
			return UserInfo{}, ErrUserNotFound
		}

		return UserInfo{}, fmt.Errorf("failed to get user info, status: %d", response.StatusCode)
	}

	var userInfo UserInfo

	err = json.NewDecoder(response.Body).Decode(&userInfo)

	if err != nil {
		return UserInfo{}, fmt.Errorf("failed to decode user info")
	}

	return userInfo, nil
}
