package user

import (
	"fmt"
	"time"

	"github.com/Nithin0620/dynamodb-go-crud/internal/entities/user"
	"github.com/Nithin0620/dynamodb-go-crud/internal/repository/adapter"
	authUtil "github.com/Nithin0620/dynamodb-go-crud/utils/auth"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"github.com/google/uuid"
)

type Controller struct {
	repository adapter.Interface
}

type Interface interface {
	Register(u *user.User) (*user.User, error)
	Login(email, password string) (token string, u *user.User, err error)
	FindByID(id uuid.UUID) (*user.User, error)
}

func NewController(repository adapter.Interface) Interface {
	return &Controller{
		repository: repository,
	}
}

func (c *Controller) FindByEmail(email string) (*user.User, error) {
	var sample user.User
	filter := expression.Name("email").Equal(expression.Value(email))
	condition, err := expression.NewBuilder().WithFilter(filter).Build()
	if err != nil {
		return nil, err
	}

	response, err := c.repository.FindAll(condition, sample.TableName())
	if err != nil {
		return nil, err
	}
	if response != nil && len(response.Items) > 0 {
		u, err := user.ParseDynamoAttributeToStruct(response.Items[0])
		return &u, err
	}
	return nil, nil
}

func (c *Controller) Register(u *user.User) (*user.User, error) {
	existing, err := c.FindByEmail(u.Email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, fmt.Errorf("email already in use")
	}

	hashedPassword, err := authUtil.HashPassword(u.Password)
	if err != nil {
		return nil, err
	}
	u.Password = hashedPassword
	u.ID = uuid.New()
	u.CreatedAt = time.Now()
	u.UpdatedAt = time.Now()

	_, err = c.repository.CreateOrUpdate(u.GetMap(), u.TableName())
	if err != nil {
		return nil, err
	}
	u.Password = ""
	return u, nil
}

func (c *Controller) Login(email, password string) (string, *user.User, error) {
	u, err := c.FindByEmail(email)
	if err != nil {
		return "", nil, err
	}
	if u == nil {
		return "", nil, fmt.Errorf("invalid email or password")
	}

	if !authUtil.CheckPassword(password, u.Password) {
		return "", nil, fmt.Errorf("invalid email or password")
	}

	token, err := authUtil.GenerateToken(u.ID.String(), u.Email, u.Role)
	if err != nil {
		return "", nil, err
	}
	u.Password = ""
	return token, u, nil
}

func (c *Controller) FindByID(id uuid.UUID) (*user.User, error) {
	var sample user.User
	sample.ID = id
	res, err := c.repository.FindOne(sample.GetFilterId(), sample.TableName())
	if err != nil {
		return nil, err
	}
	if res == nil || res.Item == nil {
		return nil, fmt.Errorf("user not found")
	}
	u, err := user.ParseDynamoAttributeToStruct(res.Item)
	if err != nil {
		return nil, err
	}
	u.Password = ""
	return &u, nil
}
