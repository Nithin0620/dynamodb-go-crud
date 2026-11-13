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

func NewHandler() {

}

func Get() {

}

func Put() {

}

func Post() {

}

func Delete() {

}

func Options() {

}