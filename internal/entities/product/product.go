package product

import (
	"encoding/json"

	"github.com/Nithin0620/dynamodb-go-crud/internal/entities"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/google/uuid"
)

type Product struct {
	entities.Base
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
	Stock       int       `json:"stock"`
	OwnerID     uuid.UUID `json:"ownerId"`
}

func InterfaceToModel(data interface{}) (instance *Product, err error) {
	bytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(bytes, &instance)
	return instance, err
}

func (p *Product) GetFilterId() map[string]interface{} {
	return map[string]interface{}{
		"_id": p.ID.String(),
	}
}

func (p *Product) TableName() string {
	return "products"
}

func (p *Product) Bytes() ([]byte, error) {
	return json.Marshal(p)
}

func (p *Product) GetMap() map[string]interface{} {
	return map[string]interface{}{
		"_id":         p.ID.String(),
		"name":        p.Name,
		"description": p.Description,
		"price":       p.Price,
		"stock":       p.Stock,
		"ownerId":     p.OwnerID.String(),
		"createdAt":   p.CreatedAt.Format(entities.GetTimeFormat()),
		"updatedAt":   p.UpdatedAt.Format(entities.GetTimeFormat()),
	}
}

func ParseDynamoAttributeToStruct(response map[string]*dynamodb.AttributeValue) (p Product, err error) {
	if response == nil {
		return p, nil
	}
	err = entities.UnmarshalItem(response, &p)
	return p, err
}
