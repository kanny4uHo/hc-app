package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

const registerUserPath = "/internal/api/v1/register"

type AuthClient struct {
	baseUserHost string
	httpClient   *http.Client
}

func NewAuthClient(baseHost string, httpClient *http.Client) *AuthClient {
	return &AuthClient{
		baseUserHost: baseHost,
		httpClient:   httpClient,
	}
}

type RegisterUserRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type RegisterResponse struct {
	PasswordHash string `json:"password_hash"`
}

func (c *AuthClient) RegisterUser(_ context.Context, request RegisterUserRequest) (RegisterResponse, error) {
	jsonRequest, err := json.Marshal(request)
	if err != nil {
		return RegisterResponse{}, fmt.Errorf("failed to marshal request: %w", err)
	}

	buffer := bytes.NewBuffer(jsonRequest)

	response, err := c.httpClient.Post(c.baseUserHost+registerUserPath, "application/json", buffer)
	if err != nil {
		return RegisterResponse{}, fmt.Errorf("failed to register user by http and login %s: %w", request.Login, err)
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		if response.StatusCode == http.StatusNotFound {
			return RegisterResponse{}, ErrUserNotFound
		}

		return RegisterResponse{}, fmt.Errorf("failed to get user info, status: %d", response.StatusCode)
	}

	var userInfo RegisterResponse

	err = json.NewDecoder(response.Body).Decode(&userInfo)

	if err != nil {
		return RegisterResponse{}, fmt.Errorf("failed to decode user info")
	}

	return userInfo, nil
}
