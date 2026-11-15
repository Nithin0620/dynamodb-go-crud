package order

import (
	"encoding/json"
	"errors"
	"io"
	"strings"
	"time"

	"github.com/Nithin0620/dynamodb-go-crud/internal/entities/order"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	Validation "github.com/go-ozzo/ozzo-validation/v4"
)

type Rules struct{}

func NewRules() *Rules {
	return &Rules{}
}

func (r *Rules) ConvertIoReaderToStruct(data io.Reader, model interface{}) (interface{}, error) {
	if data == nil {
		return nil, errors.New("data is nil")
	}
	return model, json.NewDecoder(data).Decode(model)
}

func (r *Rules) Migrate(connection *dynamodb.DynamoDB) error {
	return r.CreateTable(connection)
}

func (r *Rules) GetMock() interface{} {
	return order.Order{
		Status: "PAID",
	}
}

func (r *Rules) Validate(model interface{}) error {
	orderModel, err := order.InterfaceToModel(model)
	if err != nil {
		return err
	}

	return Validation.ValidateStruct(orderModel,
		Validation.Field(&orderModel.Status, Validation.Required),
	)
}

func (r *Rules) CreateTable(connection *dynamodb.DynamoDB) error {
	table := &order.Order{}
	input := &dynamodb.CreateTableInput{
		AttributeDefinitions: []*dynamodb.AttributeDefinition{
			{
				AttributeName: aws.String("_id"),
				AttributeType: aws.String("S"),
			},
		},
		KeySchema: []*dynamodb.KeySchemaElement{
			{
				AttributeName: aws.String("_id"),
				KeyType:       aws.String("HASH"),
			},
		},
		ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(5),
			WriteCapacityUnits: aws.Int64(5),
		},
		TableName: aws.String(table.TableName()),
	}

	response, err := connection.CreateTable(input)
	if err != nil && (strings.Contains(err.Error(), "ResourceInUseException") || strings.Contains(err.Error(), "Table already exists")) {
		return nil
	}
	if response != nil && strings.Contains(response.GoString(), "TableStatus: \"CREATING\"") {
		time.Sleep(3 * time.Second)
	}
	return err
}
