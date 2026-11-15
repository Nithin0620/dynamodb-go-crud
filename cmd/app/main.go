package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/Nithin0620/dynamodb-go-crud/config"
	"github.com/Nithin0620/dynamodb-go-crud/internal/repository/adapter"
	"github.com/Nithin0620/dynamodb-go-crud/internal/repository/instance"
	"github.com/Nithin0620/dynamodb-go-crud/internal/routes"
	"github.com/Nithin0620/dynamodb-go-crud/internal/rules"
	orderRules "github.com/Nithin0620/dynamodb-go-crud/internal/rules/order"
	productRules "github.com/Nithin0620/dynamodb-go-crud/internal/rules/product"
	userRules "github.com/Nithin0620/dynamodb-go-crud/internal/rules/user"
	"github.com/Nithin0620/dynamodb-go-crud/utils/logger"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

func main() {
	configs := config.GetConfig()
	connection := instance.GetConnection()
	repository := adapter.NewAdapter(connection)

	logger.INFO("Starting service on port: ", configs.Port)

	if err := waitForDynamoDB(connection); err != nil {
		logger.PANIC("DynamoDB is not reachable: ", err)
	}

	errors := Migrate(connection)
	if len(errors) > 0 {
		for _, err := range errors {
			logger.ERROR("Database migration warning: ", err)
		}
	}

	if err := checkTables(connection); err != nil {
		logger.ERROR("Warning checking tables: ", err)
	}

	port := fmt.Sprintf(":%d", configs.Port)
	router := routes.NewRouter().SetRouters(repository)
	logger.INFO("Server is running on port: ", port)

	server := http.ListenAndServe(port, router)
	log.Fatal(server)
}

func waitForDynamoDB(connection *dynamodb.DynamoDB) error {
	var err error
	for attempt := 1; attempt <= 60; attempt++ {
		_, err = connection.ListTables(&dynamodb.ListTablesInput{})
		if err == nil {
			return nil
		}

		logger.INFO("Waiting for DynamoDB to become available", err)
		time.Sleep(time.Second)
	}

	return err
}

func Migrate(connection *dynamodb.DynamoDB) []error {
	var errors []error
	callMigrateAndAppendError(&errors, connection, productRules.NewRules())
	callMigrateAndAppendError(&errors, connection, userRules.NewRules())
	callMigrateAndAppendError(&errors, connection, orderRules.NewRules())
	return errors
}

func callMigrateAndAppendError(errors *[]error, connection *dynamodb.DynamoDB, r rules.Interface) {
	err := r.Migrate(connection)
	if err != nil {
		*errors = append(*errors, err)
	}
}

func checkTables(connection *dynamodb.DynamoDB) error {
	response, err := connection.ListTables(&dynamodb.ListTablesInput{})
	if err != nil {
		return fmt.Errorf("error while checking tables: %v", err)
	}

	if response == nil || len(response.TableNames) == 0 {
		return fmt.Errorf("no tables found in database")
	}

	for _, tableName := range response.TableNames {
		logger.INFO("Table found: ", *tableName)
	}
	return nil
}
