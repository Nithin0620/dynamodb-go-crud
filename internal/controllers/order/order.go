package order

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/Nithin0620/dynamodb-go-crud/internal/entities/order"
	prodEntity "github.com/Nithin0620/dynamodb-go-crud/internal/entities/product"
	"github.com/Nithin0620/dynamodb-go-crud/internal/repository/adapter"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"github.com/google/uuid"
)

type Challenge struct {
	ID        string
	Question  string
	Answer    string
	UserID    uuid.UUID
	ExpiresAt time.Time
}

var (
	challengesLock sync.Mutex
	challenges     = make(map[string]Challenge)
)

type Controller struct {
	repository adapter.Interface
}

type CheckoutItem struct {
	ProductID uuid.UUID `json:"productId"`
	Quantity  int       `json:"quantity"`
}

type ChallengeResponse struct {
	ChallengeID string            `json:"challengeId"`
	Question    string            `json:"question"`
	TotalAmount float64           `json:"totalAmount"`
	Items       []order.OrderItem `json:"items"`
}

type Interface interface {
	GenerateChallenge(userID uuid.UUID, items []CheckoutItem) (*ChallengeResponse, error)
	ConfirmOrder(userID uuid.UUID, challengeID string, answer string, items []CheckoutItem) (*order.Order, error)
	ListUserOrders(userID uuid.UUID) ([]order.Order, error)
	ListStoreOwnerOrders(ownerID uuid.UUID) ([]order.Order, error)
	UpdateOrderStatus(orderID uuid.UUID, status string) error
}

func NewController(repository adapter.Interface) Interface {
	return &Controller{
		repository: repository,
	}
}

func (c *Controller) getProduct(id uuid.UUID) (*prodEntity.Product, error) {
	var sample prodEntity.Product
	sample.ID = id
	res, err := c.repository.FindOne(sample.GetFilterId(), sample.TableName())
	if err != nil {
		return nil, err
	}
	if res == nil || res.Item == nil {
		return nil, fmt.Errorf("product %s not found", id.String())
	}
	p, err := prodEntity.ParseDynamoAttributeToStruct(res.Item)
	return &p, err
}

func (c *Controller) GenerateChallenge(userID uuid.UUID, checkoutItems []CheckoutItem) (*ChallengeResponse, error) {
	if len(checkoutItems) == 0 {
		return nil, fmt.Errorf("no items provided")
	}

	var totalAmount float64
	var orderItems []order.OrderItem

	for _, item := range checkoutItems {
		prod, err := c.getProduct(item.ProductID)
		if err != nil {
			return nil, err
		}
		if prod.Stock < item.Quantity {
			return nil, fmt.Errorf("insufficient stock for product: %s (available: %d)", prod.Name, prod.Stock)
		}

		lineTotal := prod.Price * float64(item.Quantity)
		totalAmount += lineTotal

		orderItems = append(orderItems, order.OrderItem{
			ProductID:   prod.ID,
			ProductName: prod.Name,
			Price:       prod.Price,
			Quantity:    item.Quantity,
			OwnerID:     prod.OwnerID,
		})
	}

	// Generate simple math puzzle
	rand.Seed(time.Now().UnixNano())
	a := rand.Intn(10) + 1
	b := rand.Intn(10) + 1
	ans := fmt.Sprintf("%d", a+b)
	challengeID := uuid.New().String()

	challenge := Challenge{
		ID:        challengeID,
		Question:  fmt.Sprintf("Solve to confirm order: %d + %d = ?", a, b),
		Answer:    ans,
		UserID:    userID,
		ExpiresAt: time.Now().Add(5 * time.Minute),
	}

	challengesLock.Lock()
	challenges[challengeID] = challenge
	challengesLock.Unlock()

	return &ChallengeResponse{
		ChallengeID: challengeID,
		Question:    challenge.Question,
		TotalAmount: totalAmount,
		Items:       orderItems,
	}, nil
}

func (c *Controller) ConfirmOrder(userID uuid.UUID, challengeID string, answer string, checkoutItems []CheckoutItem) (*order.Order, error) {
	challengesLock.Lock()
	ch, exists := challenges[challengeID]
	if exists {
		delete(challenges, challengeID)
	}
	challengesLock.Unlock()

	if !exists || ch.UserID != userID || time.Now().After(ch.ExpiresAt) {
		return nil, fmt.Errorf("invalid or expired puzzle challenge")
	}

	if ch.Answer != answer {
		return nil, fmt.Errorf("incorrect puzzle answer")
	}

	var totalAmount float64
	var orderItems []order.OrderItem

	// Decrement stock and assemble order
	for _, item := range checkoutItems {
		prod, err := c.getProduct(item.ProductID)
		if err != nil {
			return nil, err
		}
		if prod.Stock < item.Quantity {
			return nil, fmt.Errorf("stock changed, insufficient for: %s", prod.Name)
		}

		// Decrement inventory
		prod.Stock -= item.Quantity
		prod.UpdatedAt = time.Now()
		_, err = c.repository.CreateOrUpdate(prod.GetMap(), prod.TableName())
		if err != nil {
			return nil, fmt.Errorf("failed to update inventory: %w", err)
		}

		totalAmount += prod.Price * float64(item.Quantity)
		orderItems = append(orderItems, order.OrderItem{
			ProductID:   prod.ID,
			ProductName: prod.Name,
			Price:       prod.Price,
			Quantity:    item.Quantity,
			OwnerID:     prod.OwnerID,
		})
	}

	newOrder := &order.Order{
		UserID:      userID,
		Items:       orderItems,
		TotalAmount: totalAmount,
		Status:      "PAID",
	}
	newOrder.ID = uuid.New()
	newOrder.CreatedAt = time.Now()
	newOrder.UpdatedAt = time.Now()

	_, err := c.repository.CreateOrUpdate(newOrder.GetMap(), newOrder.TableName())
	if err != nil {
		return nil, err
	}

	return newOrder, nil
}

func (c *Controller) ListUserOrders(userID uuid.UUID) ([]order.Order, error) {
	var sample order.Order
	filter := expression.Name("userId").AttributeExists()
	condition, err := expression.NewBuilder().WithFilter(filter).Build()
	if err != nil {
		return nil, err
	}

	res, err := c.repository.FindAll(condition, sample.TableName())
	if err != nil {
		return nil, err
	}

	orders := []order.Order{}
	if res != nil {
		for _, item := range res.Items {
			ord, err := order.ParseDynamoAttributeToStruct(item)
			if err == nil && ord.UserID == userID {
				orders = append(orders, ord)
			}
		}
	}
	return orders, nil
}

func (c *Controller) ListStoreOwnerOrders(ownerID uuid.UUID) ([]order.Order, error) {
	var sample order.Order
	filter := expression.Name("status").AttributeExists()
	condition, err := expression.NewBuilder().WithFilter(filter).Build()
	if err != nil {
		return nil, err
	}

	res, err := c.repository.FindAll(condition, sample.TableName())
	if err != nil {
		return nil, err
	}

	var relevantOrders []order.Order
	if res != nil {
		for _, item := range res.Items {
			ord, err := order.ParseDynamoAttributeToStruct(item)
			if err != nil {
				continue
			}
			// Filter items belonging to this store owner
			var ownerItems []order.OrderItem
			for _, it := range ord.Items {
				if it.OwnerID == ownerID {
					ownerItems = append(ownerItems, it)
				}
			}
			if len(ownerItems) > 0 {
				ord.Items = ownerItems
				relevantOrders = append(relevantOrders, ord)
			}
		}
	}
	return relevantOrders, nil
}

func (c *Controller) UpdateOrderStatus(orderID uuid.UUID, status string) error {
	var sample order.Order
	sample.ID = orderID
	res, err := c.repository.FindOne(sample.GetFilterId(), sample.TableName())
	if err != nil {
		return err
	}
	if res == nil || res.Item == nil {
		return fmt.Errorf("order not found")
	}

	ord, err := order.ParseDynamoAttributeToStruct(res.Item)
	if err != nil {
		return err
	}

	ord.Status = status
	ord.UpdatedAt = time.Now()
	_, err = c.repository.CreateOrUpdate(ord.GetMap(), ord.TableName())
	return err
}
