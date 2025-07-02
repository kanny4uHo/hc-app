package controller

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"healthcheckProject/internal/entity"
	"healthcheckProject/internal/service"
)

type UserController struct {
	userService service.UserService
}

func CreateUserController(userService service.UserService) UserController {
	return UserController{
		userService: userService,
	}
}

type CreateUserArgs struct {
	Login     string `json:"login" binding:"required,min=4,max=15"`
	Password  string `json:"password" binding:"required,min=8,max=32"`
	FirstName string `json:"first_name" binding:"min=2,max=32,alpha"`
	LastName  string `json:"last_name" binding:"min=2,max=32,alpha"`
	Email     string `json:"email" binding:"required,email"`
}

type NameResponse struct {
	First string `json:"first"`
	Last  string `json:"last"`
}

type UserResponse struct {
	ID    int64        `json:"id"`
	Login string       `json:"login"`
	Email string       `json:"email"`
	Name  NameResponse `json:"name"`
}

func (c UserController) CreateUser(ctx *gin.Context) {
	var createUserArgs CreateUserArgs

	err := ctx.ShouldBindJSON(&createUserArgs)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	createUser, err := c.userService.CreateUser(ctx, entity.AddUserArgs{
		Login:    createUserArgs.Login,
		Password: createUserArgs.Password,
		Meta: entity.UserMeta{
			Name: entity.UserName{
				First: createUserArgs.FirstName,
				Last:  createUserArgs.LastName,
			},
			Email: createUserArgs.Email,
		},
	})

	if err != nil {
		var invalidArgumentError service.InvalidArgumentError
		if errors.As(err, &invalidArgumentError) {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, UserResponse{
		ID:    createUser.ID,
		Login: createUser.Login,
		Email: createUser.Meta.Email,
		Name: NameResponse{
			First: createUser.Meta.Name.First,
			Last:  createUser.Meta.Name.Last,
		},
	})

}

func (c UserController) GetUser(ctx *gin.Context) {
	userIDParam := ctx.Param("user_id")

	userID, err := strconv.ParseInt(userIDParam, 10, 64)

	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := c.userService.GetUser(ctx, userID)

	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "no user found"})
			return
		}

		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, UserResponse{
		ID:    user.ID,
		Login: user.Login,
		Email: user.Meta.Email,
		Name: NameResponse{
			First: user.Meta.Name.First,
			Last:  user.Meta.Name.Last,
		},
	})
}

func (c UserController) DeleteUser(ctx *gin.Context) {
	userIDParam := ctx.Param("user_id")

	userID, err := strconv.ParseInt(userIDParam, 10, 64)

	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := c.userService.DeleteUser(ctx, userID)

	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "no user found"})
			return
		}

		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, UserResponse{
		ID:    user.ID,
		Login: user.Login,
		Email: user.Meta.Email,
		Name: NameResponse{
			First: user.Meta.Name.First,
			Last:  user.Meta.Name.Last,
		},
	})
}

type UpdateUserArgs struct {
	FirstName string `json:"first_name" binding:"min=4,max=32,alpha"`
	LastName  string `json:"last_name" binding:"min=2,max=32,alpha"`
	Email     string `json:"email" binding:"email"`
}

func (c UserController) UpdateUser(ctx *gin.Context) {
	userIDParam := ctx.Param("user_id")
	userID, err := strconv.ParseInt(userIDParam, 10, 64)

	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updateArgs := UpdateUserArgs{}
	err = ctx.ShouldBindJSON(&updateArgs)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if updateArgs.FirstName == "" && updateArgs.LastName == "" && updateArgs.Email == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "must provide either first_name, last_name, email"})
	}

	updatedUser, err := c.userService.UpdateUser(ctx, userID, service.UpdateUserArgs{
		FirstName: updateArgs.FirstName,
		LastName:  updateArgs.LastName,
		Email:     updateArgs.Email,
	})

	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "no user found"})
			return
		}

		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, UserResponse{
		ID:    updatedUser.ID,
		Login: updatedUser.Login,
		Email: updatedUser.Meta.Email,
		Name: NameResponse{
			First: updatedUser.Meta.Name.First,
			Last:  updatedUser.Meta.Name.Last,
		},
	})
}
