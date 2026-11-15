package auth

import (
	"encoding/json"
	"errors"
	"net/http"

	userController "github.com/Nithin0620/dynamodb-go-crud/internal/controllers/user"
	userEntity "github.com/Nithin0620/dynamodb-go-crud/internal/entities/user"
	"github.com/Nithin0620/dynamodb-go-crud/internal/repository/adapter"
	userRules "github.com/Nithin0620/dynamodb-go-crud/internal/rules/user"
	httpUtil "github.com/Nithin0620/dynamodb-go-crud/utils/http"
	"github.com/Nithin0620/dynamodb-go-crud/utils/logger"
)

type Handler struct {
	controller userController.Interface
	rules      *userRules.Rules
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func NewHandler(repository adapter.Interface) *Handler {
	return &Handler{
		controller: userController.NewController(repository),
		rules:      userRules.NewRules(),
	}
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var u userEntity.User
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		httpUtil.StatusBadRequest(w, r, errors.New("invalid request body"))
		return
	}

	if u.Role == "" {
		u.Role = "user"
	}

	if err := h.rules.Validate(&u); err != nil {
		httpUtil.StatusBadRequest(w, r, err)
		return
	}

	if len(u.Password) < 6 {
		httpUtil.StatusBadRequest(w, r, errors.New("password must be at least 6 characters"))
		return
	}

	createdUser, err := h.controller.Register(&u)
	if err != nil {
		logger.ERROR("Error during registration: ", err)
		httpUtil.StatusInternalServerError(w, r, err)
		return
	}

	httpUtil.StatusOK(w, r, createdUser)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpUtil.StatusBadRequest(w, r, errors.New("invalid request body"))
		return
	}

	token, u, err := h.controller.Login(req.Email, req.Password)
	if err != nil {
		httpUtil.StatusBadRequest(w, r, err)
		return
	}

	httpUtil.StatusOK(w, r, map[string]interface{}{
		"token": token,
		"user":  u,
	})
}
