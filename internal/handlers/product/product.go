package product

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	prodCtrl "github.com/Nithin0620/dynamodb-go-crud/internal/controllers"
	EntityProduct "github.com/Nithin0620/dynamodb-go-crud/internal/entities/product"
	"github.com/Nithin0620/dynamodb-go-crud/internal/repository/adapter"
	"github.com/Nithin0620/dynamodb-go-crud/internal/routes/middlewares"
	Rules "github.com/Nithin0620/dynamodb-go-crud/internal/rules"
	RulesProduct "github.com/Nithin0620/dynamodb-go-crud/internal/rules/product"
	HttpStatus "github.com/Nithin0620/dynamodb-go-crud/utils/http"
	"github.com/Nithin0620/dynamodb-go-crud/utils/logger"
	"github.com/go-chi/chi"
	"github.com/google/uuid"
)

type Handler struct {
	Controller prodCtrl.Interface
	Rules      Rules.Interface
}

func NewHandler(repository adapter.Interface) *Handler {
	return &Handler{
		Controller: prodCtrl.NewController(repository),
		Rules:      RulesProduct.NewRules(),
	}
}

// GET /products or GET /products/{ID} (Public)
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	if chi.URLParam(r, "ID") != "" {
		h.getOne(w, r)
		return
	}
	h.getAll(w, r)
}

func (h *Handler) getOne(w http.ResponseWriter, r *http.Request) {
	ID, err := uuid.Parse(chi.URLParam(r, "ID"))
	if err != nil {
		HttpStatus.StatusBadRequest(w, r, errors.New("invalid product UUID"))
		return
	}

	response, err := h.Controller.ListOne(ID)
	if err != nil {
		HttpStatus.StatusNotFound(w, r, err)
		return
	}

	HttpStatus.StatusOK(w, r, response)
}

func (h *Handler) getAll(w http.ResponseWriter, r *http.Request) {
	response, err := h.Controller.ListAll()
	if err != nil {
		logger.ERROR("Error while getting products: ", err)
		HttpStatus.StatusInternalServerError(w, r, err)
		return
	}

	HttpStatus.StatusOK(w, r, response)
}

// GET /admin/products (Store Owner Only - views their own inventory)
func (h *Handler) GetMyProducts(w http.ResponseWriter, r *http.Request) {
	ownerIDStr := middlewares.GetAuthUserID(r)
	ownerID, err := uuid.Parse(ownerIDStr)
	if err != nil {
		HttpStatus.StatusBadRequest(w, r, errors.New("invalid owner token"))
		return
	}

	products, err := h.Controller.ListByOwner(ownerID)
	if err != nil {
		HttpStatus.StatusInternalServerError(w, r, err)
		return
	}
	HttpStatus.StatusOK(w, r, products)
}

// POST /admin/products (Store Owner Only)
func (h *Handler) Post(w http.ResponseWriter, r *http.Request) {
	ownerIDStr := middlewares.GetAuthUserID(r)
	ownerID, err := uuid.Parse(ownerIDStr)
	if err != nil {
		HttpStatus.StatusBadRequest(w, r, errors.New("invalid owner token"))
		return
	}

	var p EntityProduct.Product
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		HttpStatus.StatusBadRequest(w, r, errors.New("invalid json payload"))
		return
	}

	p.ID = uuid.New()
	p.OwnerID = ownerID
	p.CreatedAt = time.Now()
	p.UpdatedAt = time.Now()

	if err := h.Rules.Validate(&p); err != nil {
		HttpStatus.StatusBadRequest(w, r, err)
		return
	}

	id, err := h.Controller.Create(&p)
	if err != nil {
		logger.ERROR("Error creating product: ", err)
		HttpStatus.StatusInternalServerError(w, r, err)
		return
	}

	HttpStatus.StatusOK(w, r, map[string]interface{}{"id": id.String()})
}

// PUT /admin/products/{ID} (Store Owner Only)
func (h *Handler) Put(w http.ResponseWriter, r *http.Request) {
	ownerIDStr := middlewares.GetAuthUserID(r)
	ownerID, err := uuid.Parse(ownerIDStr)
	if err != nil {
		HttpStatus.StatusBadRequest(w, r, errors.New("invalid owner token"))
		return
	}

	ID, err := uuid.Parse(chi.URLParam(r, "ID"))
	if err != nil {
		HttpStatus.StatusBadRequest(w, r, errors.New("invalid product UUID"))
		return
	}

	var p EntityProduct.Product
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		HttpStatus.StatusBadRequest(w, r, errors.New("invalid json payload"))
		return
	}
	p.ID = ID
	p.OwnerID = ownerID

	err = h.Controller.Update(ID, &p, ownerID)
	if err != nil {
		HttpStatus.StatusInternalServerError(w, r, err)
		return
	}

	HttpStatus.StatusNoContent(w, r)
}

// DELETE /admin/products/{ID} (Store Owner Only)
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	ownerIDStr := middlewares.GetAuthUserID(r)
	ownerID, err := uuid.Parse(ownerIDStr)
	if err != nil {
		HttpStatus.StatusBadRequest(w, r, errors.New("invalid owner token"))
		return
	}

	ID, err := uuid.Parse(chi.URLParam(r, "ID"))
	if err != nil {
		HttpStatus.StatusBadRequest(w, r, errors.New("invalid product UUID"))
		return
	}

	err = h.Controller.Remove(ID, ownerID)
	if err != nil {
		HttpStatus.StatusInternalServerError(w, r, err)
		return
	}

	HttpStatus.StatusNoContent(w, r)
}

func (h *Handler) Options(w http.ResponseWriter, r *http.Request) {
	HttpStatus.StatusNoContent(w, r)
}
