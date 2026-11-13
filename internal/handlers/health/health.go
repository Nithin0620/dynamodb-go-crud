package health

import(
	"errors"
	"net/http"
	"github.com/Nithin0620/dynamoDB-GO-CRUD/internal/handlers"
	"github.com/Nithin0620/dynamoDB-GO-CRUD/internal/adapter"
)


type Handler struct {
	handlers.Interface
	Repository adapter.Interface
}

func NewHandler(repository adapter.Interface) *Handler {
	return &Handler{
		Repository: repository,
	}
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	if !h.Repository.Health() {
		HttpStatus.StatusInternalServerError(w,r,errors.New("Relational database is not alive"))
		
		return	
	}
	HttpStatus.StatusOK(w,r,"Relational database is alive")
}
   
func (h *Handler) Post(w http.ResponseWriter, r *http.Request) {
	HttpStatus.StatusMethodNotAllowed(w,r,"Method not allowed")
}

func (h *Handler) Put(w http.ResponseWriter, r *http.Request) {
	HttpStatus.StatusMethodNotAllowed(w,r,"Method not allowed")
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	HttpStatus.StatusMethodNotAllowed(w,r,"Method not allowed")
}

func (h *Handler) Options(w http.ResponseWriter, r *http.Request) {
	HttpStatus.StatusNoContent(w,r)
}