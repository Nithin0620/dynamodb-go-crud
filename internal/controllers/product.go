package product

import (
	"fmt"
	"time"

	"github.com/Nithin0620/dynamodb-go-crud/internal/entities/product"
	"github.com/Nithin0620/dynamodb-go-crud/internal/repository/adapter"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"github.com/google/uuid"
)

type Controller struct {
	repository adapter.Interface
}

type Interface interface {
	ListOne(ID uuid.UUID) (entity product.Product, err error)
	ListAll() (entities []product.Product, err error)
	ListByOwner(ownerID uuid.UUID) (entities []product.Product, err error)
	Create(entity *product.Product) (uuid.UUID, error)
	Update(id uuid.UUID, entity *product.Product, ownerID uuid.UUID) error
	Remove(id uuid.UUID, ownerID uuid.UUID) error
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
	if response == nil || response.Item == nil {
		return entity, fmt.Errorf("product not found")
	}

	return product.ParseDynamoAttributeToStruct(response.Item)
}

func (c *Controller) ListAll() (entities []product.Product, err error) {
	entities = []product.Product{}
	var sample product.Product

	filter := expression.Name("name").AttributeExists()
	condition, err := expression.NewBuilder().WithFilter(filter).Build()
	if err != nil {
		return entities, err
	}

	response, err := c.repository.FindAll(condition, sample.TableName())
	if err != nil {
		return entities, err
	}

	if response != nil {
		for _, value := range response.Items {
			p, err := product.ParseDynamoAttributeToStruct(value)
			if err != nil {
				return entities, err
			}
			entities = append(entities, p)
		}
	}

	return entities, nil
}

func (c *Controller) ListByOwner(ownerID uuid.UUID) (entities []product.Product, err error) {
	all, err := c.ListAll()
	if err != nil {
		return nil, err
	}
	ownerProducts := make([]product.Product, 0)
	for _, p := range all {
		if p.OwnerID == ownerID {
			ownerProducts = append(ownerProducts, p)
		}
	}
	return ownerProducts, nil
}

func (c *Controller) Create(entity *product.Product) (uuid.UUID, error) {
	entity.CreatedAt = time.Now()
	entity.UpdatedAt = time.Now()
	_, err := c.repository.CreateOrUpdate(entity.GetMap(), entity.TableName())
	return entity.ID, err
}

func (c *Controller) Update(id uuid.UUID, entity *product.Product, ownerID uuid.UUID) error {
	found, err := c.ListOne(id)
	if err != nil {
		return err
	}
	// Verify owner
	if found.OwnerID != ownerID {
		return fmt.Errorf("unauthorized: not the product owner")
	}

	found.Name = entity.Name
	found.Description = entity.Description
	found.Price = entity.Price
	found.Stock = entity.Stock
	found.UpdatedAt = time.Now()

	_, err = c.repository.CreateOrUpdate(found.GetMap(), found.TableName())
	return err
}

func (c *Controller) Remove(id uuid.UUID, ownerID uuid.UUID) error {
	entity, err := c.ListOne(id)
	if err != nil {
		return err
	}
	if entity.OwnerID != ownerID {
		return fmt.Errorf("unauthorized: not the product owner")
	}

	_, err = c.repository.Delete(entity.GetFilterId(), entity.TableName())
	return err
}
