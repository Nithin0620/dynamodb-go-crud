package order

import (
	"encoding/json"
	"errors"
	"net/http"

	orderCtrl "github.com/Nithin0620/dynamodb-go-crud/internal/controllers/order"
	"github.com/Nithin0620/dynamodb-go-crud/internal/repository/adapter"
	"github.com/Nithin0620/dynamodb-go-crud/internal/routes/middlewares"
	httpUtil "github.com/Nithin0620/dynamodb-go-crud/utils/http"
	"github.com/go-chi/chi"
	"github.com/google/uuid"
)

type Handler struct {
	controller orderCtrl.Interface
}

func NewHandler(repository adapter.Interface) *Handler {
	return &Handler{
		controller: orderCtrl.NewController(repository),
	}
}

type ChallengeRequest struct {
	Items []orderCtrl.CheckoutItem `json:"items"`
}

type ConfirmOrderRequest struct {
	ChallengeID string                   `json:"challengeId"`
	Answer      string                   `json:"answer"`
	Items       []orderCtrl.CheckoutItem `json:"items"`
}

type UpdateStatusRequest struct {
	Status string `json:"status"`
}

// POST /orders/checkout-challenge (User)
func (h *Handler) RequestChallenge(w http.ResponseWriter, r *http.Request) {
	userIDStr := middlewares.GetAuthUserID(r)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		httpUtil.StatusBadRequest(w, r, errors.New("invalid user token"))
		return
	}

	var req ChallengeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpUtil.StatusBadRequest(w, r, errors.New("invalid request body"))
		return
	}

	res, err := h.controller.GenerateChallenge(userID, req.Items)
	if err != nil {
		httpUtil.StatusBadRequest(w, r, err)
		return
	}

	httpUtil.StatusOK(w, r, res)
}

// POST /orders/confirm (User)
func (h *Handler) ConfirmOrder(w http.ResponseWriter, r *http.Request) {
	userIDStr := middlewares.GetAuthUserID(r)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		httpUtil.StatusBadRequest(w, r, errors.New("invalid user token"))
		return
	}

	var req ConfirmOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpUtil.StatusBadRequest(w, r, errors.New("invalid request body"))
		return
	}

	ord, err := h.controller.ConfirmOrder(userID, req.ChallengeID, req.Answer, req.Items)
	if err != nil {
		httpUtil.StatusBadRequest(w, r, err)
		return
	}

	httpUtil.StatusOK(w, r, map[string]interface{}{
		"message": "Order successfully placed and paid!",
		"order":   ord,
	})
}

// GET /orders/my-orders (User)
func (h *Handler) GetUserOrders(w http.ResponseWriter, r *http.Request) {
	userIDStr := middlewares.GetAuthUserID(r)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		httpUtil.StatusBadRequest(w, r, errors.New("invalid user token"))
		return
	}

	orders, err := h.controller.ListUserOrders(userID)
	if err != nil {
		httpUtil.StatusInternalServerError(w, r, err)
		return
	}

	httpUtil.StatusOK(w, r, orders)
}

// GET /admin/orders (Store Owner Only - sees orders containing their products)
func (h *Handler) GetStoreOwnerOrders(w http.ResponseWriter, r *http.Request) {
	ownerIDStr := middlewares.GetAuthUserID(r)
	ownerID, err := uuid.Parse(ownerIDStr)
	if err != nil {
		httpUtil.StatusBadRequest(w, r, errors.New("invalid owner token"))
		return
	}

	orders, err := h.controller.ListStoreOwnerOrders(ownerID)
	if err != nil {
		httpUtil.StatusInternalServerError(w, r, err)
		return
	}

	httpUtil.StatusOK(w, r, orders)
}

// PUT /admin/orders/{id}/status (Store Owner)
func (h *Handler) UpdateOrderStatus(w http.ResponseWriter, r *http.Request) {
	orderID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpUtil.StatusBadRequest(w, r, errors.New("invalid order id"))
		return
	}

	var req UpdateStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpUtil.StatusBadRequest(w, r, errors.New("invalid request body"))
		return
	}

	err = h.controller.UpdateOrderStatus(orderID, req.Status)
	if err != nil {
		httpUtil.StatusInternalServerError(w, r, err)
		return
	}

	httpUtil.StatusOK(w, r, map[string]interface{}{
		"message": "order status updated",
	})
}
