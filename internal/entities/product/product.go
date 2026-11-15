package product

import (
	"encoding/json"
	"time"

	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/google/uuid"

	"github.com/Nithin0620/dynamodb-go-crud/internal/entities"
)

type Product struct {
	entities.Base
	Name string `json:"name" dynamodbav:"name"`
}

func InterfaceToModel(data interface{}) (instance *Product, err error) {
	bytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	instance = &Product{}
	if err := json.Unmarshal(bytes, instance); err != nil {
		return nil, err
	}

	return instance, nil
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
		"_id":       p.ID.String(),
		"name":      p.Name,
		"createdAt": p.CreatedAt,
		"updatedAt": p.UpdatedAt,
	}
}

func ParseDynamoAttributeToStruct(item map[string]*dynamodb.AttributeValue) (Product, error) {
	entity := Product{}
	if len(item) == 0 {
		return entity, nil
	}

	var raw struct {
		ID        string    `dynamodbav:"_id"`
		Name      string    `dynamodbav:"name"`
		CreatedAt time.Time `dynamodbav:"createdAt"`
		UpdatedAt time.Time `dynamodbav:"updatedAt"`
	}
	if err := dynamodbattribute.UnmarshalMap(item, &raw); err != nil {
		return entity, err
	}

	id, err := uuid.Parse(raw.ID)
	if err != nil {
		return entity, err
	}

	entity.ID = id
	entity.Name = raw.Name
	entity.CreatedAt = raw.CreatedAt
	entity.UpdatedAt = raw.UpdatedAt

	return entity, nil
}
