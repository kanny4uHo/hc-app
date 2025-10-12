package controller

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"healthcheckProject/internal/entity"
	"healthcheckProject/internal/service"
)

type InventoryController struct {
	service *service.InventoryService
}

func NewInventoryController(service *service.InventoryService) *InventoryController {
	return &InventoryController{
		service: service,
	}
}

type addItemRequest struct {
	Item Item `json:"item"`
}

type addItemResponse struct {
	Item Item `json:"item"`
}

type Item struct {
	ItemID string `json:"item_id"`
	Amount uint64 `json:"amount"`
}

func (c *InventoryController) AddItems(ctx *gin.Context) {
	req := &addItemRequest{}
	err := ctx.BindJSON(req)
	if err != nil {
		return
	}

	if req.Item.ItemID == "" || req.Item.Amount == 0 {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "invalid request",
		})
	}

	item, err := c.service.AddItem(ctx, entity.Item{
		ItemID: req.Item.ItemID,
		Amount: req.Item.Amount,
	})

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, err)

		return
	}

	ctx.JSON(http.StatusOK, addItemResponse{
		Item: Item{
			ItemID: item.ItemID,
			Amount: item.Amount,
		},
	})
}

type reserveItemRequest struct {
	OrderID uint64 `json:"order_id"`
	Item    Item   `json:"item"`
}

type reserveItemResponse struct {
	ReservationID uint64 `json:"reservation_id"`
}

func (c *InventoryController) ReserveItems(ctx *gin.Context) {
	reserveItemsRequest := &reserveItemRequest{}
	err := ctx.BindJSON(reserveItemsRequest)
	if err != nil {
		log.Printf("failed to unmarshal request: %s", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

		return
	}

	reservation, err := c.service.ReserveItem(ctx, reserveItemsRequest.OrderID, entity.Item{
		ItemID: reserveItemsRequest.Item.ItemID,
		Amount: reserveItemsRequest.Item.Amount,
	})

	if err != nil {
		if errors.Is(err, service.ErrNoItemsInInventory) {
			log.Printf("no items enough in inventory: %s", err)
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

			return
		}

		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

		return
	}

	ctx.JSON(http.StatusOK, reserveItemResponse{
		ReservationID: reservation.ID,
	})
}

type reservationInfo struct {
	ID      uint64 `json:"id"`
	OrderID uint64 `json:"order_id"`
	ItemID  string `json:"item_id"`
	Amount  uint64 `json:"amount"`
}

func (c *InventoryController) GetReservationInfo(ctx *gin.Context) {
	param := ctx.Param("reservation_id")
	if param == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "no reservation id"})

		return
	}

	reservationID, err := strconv.ParseUint(param, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

		return
	}

	info, err := c.service.GetReservationInfo(ctx, reservationID)
	if err != nil {
		if errors.Is(err, service.ErrNoReservation) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})

			return
		}

		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

		return
	}

	ctx.JSON(http.StatusOK, reservationInfo{
		ID:      info.ID,
		OrderID: info.OrderID,
		ItemID:  info.Item.ItemID,
		Amount:  info.Item.Amount,
	})
}

func (c *InventoryController) CancelReservation(ctx *gin.Context) {
	param := ctx.Param("reservation_id")
	if param == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "no reservation id"})

		return
	}

	reservationID, err := strconv.ParseUint(param, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

		return
	}

	err = c.service.CancelReservation(ctx, reservationID)

	if err != nil {
		if errors.Is(err, service.ErrNoItemsInInventory) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.Status(http.StatusOK)
}
