package product

import (
	"errors"
	"time"

	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"github.com/google/uuid"

	"github.com/Nithin0620/dynamodb-go-crud/internal/entities/product"
	"github.com/Nithin0620/dynamodb-go-crud/internal/repository/adapter"
)

var ErrNotFound = errors.New("product not found")

type Controller struct {
	repository adapter.Interface
}

type Interface interface {
	ListOne(ID uuid.UUID) (entity product.Product, err error)
	ListAll() (entities []product.Product, err error)
	Create(entity *product.Product) (uuid.UUID, error)
	Update(id uuid.UUID, entity *product.Product) error
	Remove(id uuid.UUID) error
}

func NewController(repository adapter.Interface) Interface {
	return &Controller{
		repository: repository,
	}
}

func (c *Controller) ListOne(id uuid.UUID) (entity product.Product, err error) {
	entity.ID = id
	response, err := c.repository.FindOne(entity.GetFilterId(), entity.TableName())
	if err != nil {
		return entity, err
	}

	if response == nil || len(response.Item) == 0 {
		return entity, ErrNotFound
	}

	return product.ParseDynamoAttributeToStruct(response.Item)
}

func (c *Controller) ListAll() (entities []product.Product, err error) {
	entities = []product.Product{}
	table := product.Product{}

	filter := expression.Name("name").NotEqual(expression.Value(""))
	condition, err := expression.NewBuilder().WithFilter(filter).Build()
	if err != nil {
		return entities, err
	}

	response, err := c.repository.FindAll(condition, table.TableName())
	if err != nil {
		return entities, err
	}

	if response != nil {
		for _, value := range response.Items {
			entity, err := product.ParseDynamoAttributeToStruct(value)
			if err != nil {
				return entities, err
			}
			entities = append(entities, entity)
		}
	}

	return entities, nil
}

func (c *Controller) Create(entity *product.Product) (uuid.UUID, error) {
	if entity.ID == uuid.Nil {
		entity.ID = uuid.New()
	}
	if entity.CreatedAt.IsZero() {
		entity.CreatedAt = time.Now()
	}
	entity.UpdatedAt = time.Now()

	_, err := c.repository.CreateOrUpdate(entity.GetMap(), entity.TableName())
	return entity.ID, err
}

func (c *Controller) Update(id uuid.UUID, entity *product.Product) error {
	found, err := c.ListOne(id)
	if err != nil {
		return err
	}

	found.Name = entity.Name
	found.UpdatedAt = time.Now()

	_, err = c.repository.CreateOrUpdate(found.GetMap(), found.TableName())
	return err
}

func (c *Controller) Remove(id uuid.UUID) error {
	entity, err := c.ListOne(id)
	if err != nil {
		return err
	}

	_, err = c.repository.Delete(entity.GetFilterId(), entity.TableName())

	return err
}
