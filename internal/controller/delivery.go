package controller

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"healthcheckProject/internal/service"
)

type DeliveryController struct {
	service *service.DeliveryService
}

func NewDeliveryController(service *service.DeliveryService) *DeliveryController {
	return &DeliveryController{
		service: service,
	}
}

type applyCourierRequest struct {
	OrderID uint64 `json:"order_id"`
}

type applyCourierResponse struct {
	CourierID  uint64 `json:"courier_id"`
	DeliveryID uint64 `json:"delivery_id"`
}

func (c *DeliveryController) ApplyCourierForOrder(ctx *gin.Context) {
	var req applyCourierRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	delivery, err := c.service.ApplyCourierForOrder(ctx, req.OrderID)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, applyCourierResponse{
		CourierID:  delivery.CourierID,
		DeliveryID: delivery.ID,
	})
}

type DeliveryItem struct {
	ID        uint64 `json:"id"`
	OrderID   uint64 `json:"order_id"`
	CourierID uint64 `json:"courier_id"`
	Status    string `json:"status"`
}

func (c *DeliveryController) GetDeliveryInfo(ctx *gin.Context) {
	deliveryID, err := strconv.ParseUint(ctx.Param("delivery_id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	info, err := c.service.GetDeliveryInfo(ctx, deliveryID)
	if err != nil {
		if errors.Is(err, service.ErrDeliveryNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, DeliveryItem{
		ID:        info.ID,
		OrderID:   info.OrderID,
		CourierID: info.CourierID,
		Status:    string(info.Status),
	})
}

type DeliveriesResponse struct {
	List []DeliveryItem `json:"list"`
}

func (c *DeliveryController) GetAllDeliveries(ctx *gin.Context) {
	deliveries, err := c.service.GetDeliveries(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

		return
	}

	deliveriesItems := make([]DeliveryItem, 0, len(deliveries))

	for _, delivery := range deliveries {
		deliveriesItems = append(deliveriesItems, DeliveryItem{
			ID:        delivery.ID,
			OrderID:   delivery.OrderID,
			CourierID: delivery.CourierID,
			Status:    string(delivery.Status),
		})
	}

	ctx.JSON(http.StatusOK, DeliveriesResponse{
		List: deliveriesItems,
	})
}
