package user

import (
	"encoding/json"

	"github.com/Nithin0620/dynamodb-go-crud/internal/entities"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

type User struct {
	entities.Base
	Email    string `json:"email"`
	Password string `json:"password,omitempty"`
	Name     string `json:"name"`
	Role     string `json:"role"` // "user" or "admin"
}

func InterfaceToModel(data interface{}) (instance *User, err error) {
	bytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(bytes, &instance)
	return instance, err
}

func (u *User) GetFilterId() map[string]interface{} {
	return map[string]interface{}{
		"_id": u.ID.String(),
	}
}

func (u *User) TableName() string {
	return "users"
}

func (u *User) GetMap() map[string]interface{} {
	return map[string]interface{}{
		"_id":       u.ID.String(),
		"email":     u.Email,
		"password":  u.Password,
		"name":      u.Name,
		"role":      u.Role,
		"createdAt": u.CreatedAt.Format(entities.GetTimeFormat()),
		"updatedAt": u.UpdatedAt.Format(entities.GetTimeFormat()),
	}
}

func ParseDynamoAttributeToStruct(response map[string]*dynamodb.AttributeValue) (u User, err error) {
	if response == nil {
		return u, nil
	}
	err = entities.UnmarshalItem(response, &u)
	return u, err
}
