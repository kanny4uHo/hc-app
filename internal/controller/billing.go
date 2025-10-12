package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/segmentio/kafka-go"
	"healthcheckProject/internal/entity"
	"healthcheckProject/internal/gateway"
	"healthcheckProject/internal/service"
	"log"
	"net/http"
	"strconv"
)

type Controller struct {
	service *service.BillingService
}

func NewBillingController(service *service.BillingService) *Controller {
	return &Controller{
		service: service,
	}
}

func (c *Controller) Consume(ctx context.Context, reader *kafka.Reader) error {
	err := consumeTopic(ctx, reader, func(message kafka.Message) error {
		newUserEvent := &gateway.NewUserEvent{}
		err := json.Unmarshal(message.Value, newUserEvent)

		if err != nil {
			return fmt.Errorf("failed to unmarshal event: %s", err)
		}

		err = c.service.OnNewUserRegistered(ctx, newUserEvent.ID)
		if err != nil {
			return fmt.Errorf("failed to register user: %s", err)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to consume: %s", err)
	}

	return nil
}

func (c *Controller) ConsumeNewOrders(ctx context.Context, reader *kafka.Reader) error {
	err := consumeTopic(ctx, reader, func(message kafka.Message) error {
		newOrderEvent := &gateway.OrderEvent{}
		err := json.Unmarshal(message.Value, newOrderEvent)

		if err != nil {
			return fmt.Errorf("failed to unmarshal event: %s", err)
		}

		if newOrderEvent.Type != gateway.OrderCreated {
			log.Printf("wrong order type, skip: %s", newOrderEvent.Type)
			return nil
		}

		err = c.service.OnOrderCreated(ctx, newOrderEvent.ID)
		if err != nil {
			return fmt.Errorf("failed to process order: %s", err)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to consume: %s", err)
	}

	return nil
}

type CreditMoneyRequest struct {
	UserID uint64 `json:"user_id"`
	Amount int    `json:"amount"`
}

type CreditMoneyResponse struct {
	UserID  uint64 `json:"user_id"`
	Balance int64  `json:"amount"`
}

func (c *Controller) CreditMoney(ctx *gin.Context) {
	request := &CreditMoneyRequest{}
	err := ctx.BindJSON(request)
	if err != nil {
		log.Printf("failed to unmarshal request: %s", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

		return
	}

	var userAccount entity.UserAccount
	userAccount, err = c.service.CreditMoney(ctx, request.UserID, uint64(request.Amount))
	if err != nil {
		log.Printf("failed to credit money: %s", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

		return
	}

	ctx.JSON(http.StatusOK, CreditMoneyResponse{
		UserID:  userAccount.UserID,
		Balance: userAccount.Balance,
	})
}

type WithdrawMoneyRequest struct {
	UserID uint64 `json:"user_id"`
	Amount int    `json:"amount"`
}

type WithdrawMoneyResponse struct {
	UserID  uint64 `json:"user_id"`
	Balance int64  `json:"balance"`
}

func (c *Controller) WithdrawMoney(ctx *gin.Context) {
	request := &WithdrawMoneyRequest{}
	err := ctx.BindJSON(request)
	if err != nil {
		log.Printf("failed to unmarshal request: %s", err)
		return
	}

	var userAccount entity.UserAccount
	userAccount, err = c.service.WithdrawMoney(ctx, request.UserID, uint64(request.Amount))
	if err != nil {
		log.Printf("failed to credit money: %s", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

		return
	}

	ctx.JSON(http.StatusOK, WithdrawMoneyResponse{
		UserID:  userAccount.UserID,
		Balance: userAccount.Balance,
	})
}

type GetUserAccountByIDResponse struct {
	ID      uint64 `json:"id"`
	UserID  uint64 `json:"user_id"`
	Balance int64  `json:"balance"`
}

func (c *Controller) GetUserAccountInfo(ctx *gin.Context) {
	userIDParam := ctx.Param("user_id")
	userID, err := strconv.ParseUint(userIDParam, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userAccount, err := c.service.GetUserAccountByID(ctx, userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, GetUserAccountByIDResponse{
		ID:      userAccount.ID,
		UserID:  userAccount.UserID,
		Balance: userAccount.Balance,
	})
}
