package order

import (
	"encoding/json"

	"github.com/Nithin0620/dynamodb-go-crud/internal/entities"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/google/uuid"
)

type OrderItem struct {
	ProductID   uuid.UUID `json:"productId"`
	ProductName string    `json:"productName"`
	Price       float64   `json:"price"`
	Quantity    int       `json:"quantity"`
	OwnerID     uuid.UUID `json:"ownerId"`
}

type Order struct {
	entities.Base
	UserID      uuid.UUID   `json:"userId"`
	Items       []OrderItem `json:"items"`
	TotalAmount float64     `json:"totalAmount"`
	Status      string      `json:"status"` // "PAID", "SHIPPED", "DELIVERED", "CANCELLED"
}

func InterfaceToModel(data interface{}) (instance *Order, err error) {
	bytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(bytes, &instance)
	return instance, err
}

func (o *Order) GetFilterId() map[string]interface{} {
	return map[string]interface{}{
		"_id": o.ID.String(),
	}
}

func (o *Order) TableName() string {
	return "orders"
}

func (o *Order) GetMap() map[string]interface{} {
	serializedItems := make([]map[string]interface{}, len(o.Items))
	for i, item := range o.Items {
		serializedItems[i] = map[string]interface{}{
			"productId":   item.ProductID.String(),
			"productName": item.ProductName,
			"price":       item.Price,
			"quantity":    item.Quantity,
			"ownerId":     item.OwnerID.String(),
		}
	}

	return map[string]interface{}{
		"_id":         o.ID.String(),
		"userId":      o.UserID.String(),
		"items":       serializedItems,
		"totalAmount": o.TotalAmount,
		"status":      o.Status,
		"createdAt":   o.CreatedAt.Format(entities.GetTimeFormat()),
		"updatedAt":   o.UpdatedAt.Format(entities.GetTimeFormat()),
	}
}

func ParseDynamoAttributeToStruct(response map[string]*dynamodb.AttributeValue) (o Order, err error) {
	if response == nil {
		return o, nil
	}
	err = entities.UnmarshalItem(response, &o)
	return o, err
}
