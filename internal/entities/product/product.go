package product

import (
	"encoding/json"
	"errors"
	"time"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/google/uuid"
)

type Product struct {
	entities.Base
	Name		string  `json:"name"`
}

func InterfaceToModal(data interface{}) (instance *Product,err error) {
	bytes,err := json.Marshal(data)
	if err != nil {
		return nil,err
	}

	return instance,json.Unmarshal(bytes,&instance)
}

func (p *Product) GetFilterId() map[string]interface{} {
	return map[string]interface{}{
		"_id": p.ID,
	}
}

func (p *Product) TableName() string {
	return "products"
}

func (p *Product) Bytes() ([]byte,error) {
	return json.Marshal(p)
}

func (p *Product) GetMap() map[string]interface{} {

}

func ParseDynamoAttributeToStruct()() {

}
