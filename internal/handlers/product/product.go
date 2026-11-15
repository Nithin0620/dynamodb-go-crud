package product

import (
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	product "github.com/Nithin0620/dynamodb-go-crud/internal/controllers"
	EntityProduct "github.com/Nithin0620/dynamodb-go-crud/internal/entities/product"
	"github.com/Nithin0620/dynamodb-go-crud/internal/handlers"
	"github.com/Nithin0620/dynamodb-go-crud/internal/repository/adapter"
	Rules "github.com/Nithin0620/dynamodb-go-crud/internal/rules"
	RulesProduct "github.com/Nithin0620/dynamodb-go-crud/internal/rules/product"
	HttpStatus "github.com/Nithin0620/dynamodb-go-crud/utils/http"
	"github.com/Nithin0620/dynamodb-go-crud/utils/logger"
)

type Handler struct {
	handlers.Interface
	Controller product.Interface
	Rules      Rules.Interface
}

func NewHandler(repository adapter.Interface) handlers.Interface {
	return &Handler{
		Controller: product.NewController(repository),
		Rules:      RulesProduct.NewRules(),
	}
}

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
		logger.ERROR("Error while parsing the ID: ", err)
		HttpStatus.StatusBadRequest(w, r, errors.New("ID is not a valid uuid"))
		return
	}

	response, err := h.Controller.ListOne(ID)
	if err != nil {
		logger.ERROR("Error while getting the product: ", err)
		if errors.Is(err, product.ErrNotFound) {
			HttpStatus.StatusNotFound(w, r, err)
			return
		}
		HttpStatus.StatusInternalServerError(w, r, err)
		return
	}

	HttpStatus.StatusOK(w, r, response)
}

func (h *Handler) getAll(w http.ResponseWriter, r *http.Request) {
	response, err := h.Controller.ListAll()
	if err != nil {
		logger.ERROR("Error while getting the products: ", err)
		HttpStatus.StatusInternalServerError(w, r, err)
		return
	}

	HttpStatus.StatusOK(w, r, response)
}

func (h *Handler) Post(w http.ResponseWriter, r *http.Request) {
	productBody, err := h.getBodyAndValidate(r, uuid.Nil)
	if err != nil {
		logger.ERROR("Error while getting the body and validating: ", err)
		HttpStatus.StatusBadRequest(w, r, err)
		return
	}

	ID, err := h.Controller.Create(productBody)
	if err != nil {
		logger.ERROR("Error while creating the product: ", err)
		HttpStatus.StatusInternalServerError(w, r, err)
		return
	}

	HttpStatus.StatusOK(w, r, map[string]interface{}{"ID": ID.String()})
}

func (h *Handler) Put(w http.ResponseWriter, r *http.Request) {
	ID, err := uuid.Parse(chi.URLParam(r, "ID"))
	if err != nil {
		logger.ERROR("Error while parsing the ID: ", err)
		HttpStatus.StatusBadRequest(w, r, errors.New("ID is not a valid uuid"))
		return
	}

	productBody, err := h.getBodyAndValidate(r, ID)
	if err != nil {
		logger.ERROR("Error while getting the body and validating: ", err)
		HttpStatus.StatusBadRequest(w, r, err)
		return
	}

	err = h.Controller.Update(ID, productBody)
	if err != nil {
		logger.ERROR("Error while updating the product: ", err)
		if errors.Is(err, product.ErrNotFound) {
			HttpStatus.StatusNotFound(w, r, err)
			return
		}
		HttpStatus.StatusInternalServerError(w, r, err)
		return
	}

	HttpStatus.StatusNoContent(w, r)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	ID, err := uuid.Parse(chi.URLParam(r, "ID"))
	if err != nil {
		logger.ERROR("Error while parsing the ID: ", err)
		HttpStatus.StatusBadRequest(w, r, errors.New("ID is not a valid uuid"))
		return
	}

	err = h.Controller.Remove(ID)
	if err != nil {
		logger.ERROR("Error while deleting the product: ", err)
		if errors.Is(err, product.ErrNotFound) {
			HttpStatus.StatusNotFound(w, r, err)
			return
		}
		HttpStatus.StatusInternalServerError(w, r, errors.New("Error while deleting the product"))
		return
	}

	HttpStatus.StatusNoContent(w, r)
}

func (h *Handler) Options(w http.ResponseWriter, r *http.Request) {
	HttpStatus.StatusNoContent(w, r)
}

func (h *Handler) getBodyAndValidate(r *http.Request, ID uuid.UUID) (*EntityProduct.Product, error) {
	productBody := &EntityProduct.Product{}
	body, err := h.Rules.ConvertIoReaderToStruct(r.Body, productBody)
	if err != nil {
		return &EntityProduct.Product{}, errors.New("Body is required and must be a valid JSON")
	}

	productParsed, err := EntityProduct.InterfaceToModel(body)
	if err != nil {
		return &EntityProduct.Product{}, errors.New("Error while parsing the product")
	}

	setDefaultValues(productParsed, ID)
	return productParsed, h.Rules.Validate(productParsed)
}

func setDefaultValues(product *EntityProduct.Product, ID uuid.UUID) {
	product.UpdatedAt = time.Now()
	if ID == uuid.Nil {
		product.ID = uuid.New()
		product.CreatedAt = time.Now()
	} else {
		product.ID = ID
	}
}
