package product

import (
	"errors"
	"net/http"
	"time"
	"github.com/go-chi/chi"
	"github.com/google/uuid"

	"github.com/Nithin0620/dynamoDB-GO-CRUD/internal/adapter"
	"github.com/Nithin0620/dynamoDB-GO-CRUD/controllers/product"
	EntityProduct "github.com/Nithin0620/dynamoDB-GO-CRUD/internal/entities/product"
	"github.com/Nithin0620/dynamoDB-GO-CRUD/internal/handlers"
	Rules "github.com/Nithin0620/dynamoDB-GO-CRUD/internal/rules"
	RulesProduct "github.com/Nithin0620/dynamoDB-GO-CRUD/internal/rules/product"
	HttpStatus "github.com/Nithin0620/dynamoDB-GO-CRUD/utils/http"
)

type Handler struct {
	handlers.Interface
	Controller product.Interface
	Rules Rules.interface
}

func NewHandler(repository adapter.Interface) handlers.Interface {
	return &Handler{
		Controller: product.NewController(repository),
		Rules: RulesProduct.NewRules(),
	}


}

func (h * Handler) Get(w http.ResponseWriter, r *http.Request) {
	if chi.URLParam(r,"ID") != "" {
		h.getOne(w,r)
		return
	}
	h.getAll(w,r)
}


func (h *Handler) getOne(w http.ResponseWriter, r *http.Request){
	ID,err := uuid.Parse(chi.URLParam(r,"ID"))
	if err != nil {
		logger.ERROR("Error while parsing the ID: ", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		HttpStatus.StatusBadRequest(w,r,err.New("ID is not uuid valid"))
		return
	}

	response,err := h.Controller.ListOne(ID)
	if err != nil {
		logger.ERROR("Error while getting the product: ", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		HttpStatus.StatusInternalServerError(w,r,err.Error())
		return
	}

	if response == nil {
		logger.ERROR("Product not found")
		http.Error(w, "Product not found", http.StatusNotFound)
		HttpStatus.StatusNotFound(w,r,"Product not found")
		return
	}

	HttpStatus.StatusOK(w,r,response)
}

func (h *Handler) getAll(w http.ResponseWriter, r *http.Request) {
	response,err := h.Controller.ListAll()
	if err != nil {
		logger.ERROR("Error while getting the products: ", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		HttpStatus.StatusInternalServerError(w,r,err.Error())
		return
	}

	if len(response) == 0 {
		logger.ERROR("No products found")
		http.Error(w, "No products found", http.StatusNotFound)
		HttpStatus.StatusNotFound(w,r,"No products found")
		return
	}

	HttpStatus.StatusOK(w,r,response)
}

func (h * Handler) Post(w http.ResponseWriter, r *http.Request) {
	productBody,err := h.getBodyAndValidate(r,uuid.Nil)
	
	if err != nil {
		logger.ERROR("Error while getting the body and validating: ", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		HttpStatus.StatusBadRequest(w,r,err.Error())
		return
	}
	
	ID,err := h.Controller.Create(productBody)
	
	if err != nil {
		logger.ERROR("Error while creating the product: ", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		HttpStatus.StatusInternalServerError(w,r,err.Error())
		return
	}
	
	HttpStatus.StatusOK(w,r,map[string]interface{}{"ID": ID.Stribng()})
}

func (h * Handler) Put(w http.ResponseWriter, r *http.Request) {
	ID, err := uuid.Parse(chi.URLParam(r,"ID"))
	if err != nil {
		logger.ERROR("Error while parsing the ID: ", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		HttpStatus.StatusBadRequest(w,r,err.New("ID is not uuid valid"))
		return
	}

	productBody,err := h.getBodyAndValidate(r,ID)
	
	if err != nil {
		logger.ERROR("Error while getting the body and validating: ", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		HttpStatus.StatusBadRequest(w,r,err.Error())
		return
	}
	
	err = h.Controller.Update(ID,productBody)
	
	if err != nil {
		logger.ERROR("Error while updating the product: ", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		HttpStatus.StatusInternalServerError(w,r,err.Error())
		return
	}
	
	HttpStatus.StatusNoContent(w,r)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	ID, err := uuid.Parse(chi.URLParam(r,"ID"))
	if err != nil {
		logger.ERROR("Error while parsing the ID: ", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		HttpStatus.StatusBadRequest(w,r,errors.New("ID is not uuid valid"))
		return
	}

	err = h.Controller.Remove(ID)
	if err != nil {
		logger.ERROR("Error while deleting the product: ", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		HttpStatus.StatusInternalServerError(w,r,errors.New("Error while deleting the product"))
		return
	}

	HttpStatus.StatusNoContent(w,r)
}

func (h *Handler) Options(w http.ResponseWriter, r *http.Request) {
	HttpStatus.StatusNoContent(w,r)
}

func (h *Handler) getBodyAndValidate(r *http.Request,ID uuid.UUID)(*EntityProduct.Product,error) {
	productBody := &EntityProduct.Product{}
	body,err	:=	h.Rules.ConvertIoReaderToStruct(r.Body,productBody)
	if err != nil {
		return &EntityProduct.Product{},errors.New("Body is required and must be a valid JSON")
	}
	productParsed,err := EntityProduct.InterfaceToModel(body)
	if err != nil {
		return &EntityProduct.Product{},errors.New("Error while parsing the product")
	}
	setDefaultValues(productParsed,ID)
	return productParsed,h.Rule.Validate(productParsed)

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