package controller

import (
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"healthcheckProject/internal/service"
)

type AuthController struct {
	as *service.AuthService
}

func NewAuthController(authService *service.AuthService) *AuthController {
	return &AuthController{
		as: authService,
	}
}

type LoginArgs struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	AccessToken string `json:"access_token"`
}

type RegisterArgs struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type RegisterResponse struct {
	PasswordHash string `json:"password_hash"`
}

func (ac *AuthController) Register(ctx *gin.Context) {
	var registerArgs RegisterArgs

	err := ctx.ShouldBindJSON(&registerArgs)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	register, err := ac.as.Register(ctx, registerArgs.Login, registerArgs.Password)
	if err != nil {
		if errors.Is(err, service.ErrUserAlreadyExists) {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "User already exists"})
			return
		}

		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, RegisterResponse{
		PasswordHash: register.PasswordHash,
	})
}

func (ac *AuthController) Login(ctx *gin.Context) {
	log.Println("start")
	var loginArgs LoginArgs

	err := ctx.ShouldBindJSON(&loginArgs)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	authResult, err := ac.as.Authorize(ctx, loginArgs.Username, loginArgs.Password)
	if err != nil {
		log.Printf("failed to authorize %s\n", err)
		if errors.Is(err, service.ErrUserNotFound) {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "User not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, LoginResponse{AccessToken: authResult.AccessToken})
}

func (ac *AuthController) AuthCheck(ctx *gin.Context) {
	authHeader := ctx.GetHeader("Authorization")

	if authHeader == "" {
		log.Println("Authorization header is empty")
		ctx.Status(http.StatusUnauthorized)
		return
	}

	if !strings.HasPrefix(authHeader, "Bearer ") {
		log.Println("Authorization header is not bearer token")
		ctx.Status(http.StatusUnauthorized)
		return
	}

	result, err := ac.as.ValidateAuthorization(ctx, strings.TrimPrefix(authHeader, "Bearer "))

	if err != nil {
		log.Println("ValidateAuthorization err:", err)
		ctx.Status(http.StatusUnauthorized)
		return
	}

	ctx.Status(http.StatusOK)
	ctx.Header("X-Login", result.Login)
	ctx.Header("X-User-Id", strconv.Itoa(result.ID))
}
