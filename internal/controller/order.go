package controller

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/segmentio/kafka-go"

	"healthcheckProject/internal/entity"
	"healthcheckProject/internal/gateway"
	"healthcheckProject/internal/service"
)

type OrderController struct {
	service *service.OrderService
}

type CreateOrderArgs struct {
	Item   string `json:"item"`
	Price  uint64 `json:"price"`
	Amount uint64 `json:"amount"`
}

type CreateOrderResponse struct {
	ID     uint64 `json:"id"`
	Price  uint64 `json:"price"`
	Status string `json:"status"`
	Item   string `json:"item"`

	ReservationID uint64 `json:"reservation_id"`
	DeliveryID    uint64 `json:"delivery_id"`
}

func NewOrderController(service *service.OrderService) *OrderController {
	return &OrderController{
		service: service,
	}
}

func (c *OrderController) CreateOrder(ctx *gin.Context) {
	createOrderArgs := &CreateOrderArgs{}
	err := ctx.BindJSON(createOrderArgs)
	if err != nil {
		return
	}

	//createdOrder, err := c.service.CreateOrder(ctx, entity.CreateOrderArgs{
	//	Item:  createOrderArgs.Item,
	//	Price: createOrderArgs.Price,
	//})

	createdOrder, err := c.service.CreateOrderSaga(ctx, entity.CreateOrderArgs{
		Item:  createOrderArgs.Item,
		Price: createOrderArgs.Price,
	})

	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

		return
	}

	ctx.JSON(http.StatusCreated, CreateOrderResponse{
		ID:     createdOrder.ID,
		Price:  createdOrder.Price,
		Item:   createdOrder.Item,
		Status: string(createdOrder.Status),

		ReservationID: createdOrder.ReservationID,
		DeliveryID:    createdOrder.DeliveryID,
	})
}

type OrderResponse struct {
	ID     uint64 `json:"id"`
	Price  uint64 `json:"price"`
	Status string `json:"status"`
	Item   string `json:"item"`
	UserID uint64 `json:"user_id"`
}

func (c *OrderController) GetOrder(ctx *gin.Context) {
	orderIDstr := ctx.Param("order_id")
	orderID, err := strconv.ParseUint(orderIDstr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

		return
	}

	order, err := c.service.GetOrder(ctx, orderID)
	if err != nil {
		log.Printf("failed to get order: %v", err)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

		return
	}

	ctx.JSON(http.StatusOK, OrderResponse{
		ID:     order.ID,
		Price:  order.Price,
		Status: string(order.Status),
		Item:   order.Item,
		UserID: order.Owner.ID,
	})
}

func (c *OrderController) ConsumeOrderIsPaid(ctx context.Context, reader *kafka.Reader) error {
	err := consumeTopic(ctx, reader, func(message kafka.Message) error {
		var event gateway.OrderIsPaidEvent
		err := json.Unmarshal(message.Value, &event)
		if err != nil {
			return fmt.Errorf("unmarshal event err: %w", err)
		}

		err = c.service.OnOrderPaid(ctx, event.OrderID)
		if err != nil {
			return fmt.Errorf("onOrderPaid err: %w", err)
		}

		return nil
	})

	if err != nil {
		log.Printf("failed to consume order is paid: %v", err)
	}

	return nil
}

func consumeTopic(ctx context.Context, reader *kafka.Reader, consumer func(message kafka.Message) error) error {
	for {
		log.Printf("Attempting to fetch message...")
		message, err := reader.FetchMessage(ctx)
		if err != nil {
			if errors.Is(ctx.Err(), context.Canceled) {
				break
			}

			log.Printf("Fetch error: %v (type: %T)", err, err)

			//log.Printf("failed to fetch message: %s", err)
			time.Sleep(3 * time.Second)
			continue
		}

		log.Printf("message received: %s", message)

		err = consumer(message)
		if err != nil {
			log.Printf("failed to process message: %s", err)
			continue
		}

		err = reader.CommitMessages(ctx, message)
		if err != nil {
			log.Printf("failed to commit message: %s", err)
		}

		log.Printf("Committed message: %v", message)
	}

	return nil
}
