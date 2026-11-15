package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/aws/aws-sdk-go/service/dynamodb"

	"github.com/Nithin0620/dynamodb-go-crud/config"
	"github.com/Nithin0620/dynamodb-go-crud/internal/repository/adapter"
	"github.com/Nithin0620/dynamodb-go-crud/internal/repository/instance"
	"github.com/Nithin0620/dynamodb-go-crud/internal/routes"
	"github.com/Nithin0620/dynamodb-go-crud/internal/rules"
	RulesProduct "github.com/Nithin0620/dynamodb-go-crud/internal/rules/product"
	"github.com/Nithin0620/dynamodb-go-crud/utils/logger"
)

func main() {
	configs := config.GetConfig()
	connection := instance.GetConnection()
	repository := adapter.NewAdapter(connection)

	logger.INFO("Waiting for the service to start on port: ", configs.Port)

	migrationErrors := Migrate(connection)
	if len(migrationErrors) > 0 {
		for _, err := range migrationErrors {
			logger.PANIC("Error while migrating the database: ", err)
		}
	}

	if err := checkTables(connection); err != nil {
		logger.PANIC("Error while checking tables: ", err)
	}

	port := fmt.Sprintf(":%d", configs.Port)
	router := routes.NewRouter().SetRouters(repository)
	logger.INFO("Server is running on port: ", port)

	server := http.ListenAndServe(port, router)
	log.Fatal(server)
}

func Migrate(connection *dynamodb.DynamoDB) []error {
	var errors []error
	callMigrateAndAppendError(&errors, connection, &RulesProduct.Rules{})
	return errors
}

func callMigrateAndAppendError(errors *[]error, connection *dynamodb.DynamoDB, rules rules.Interface) {
	err := rules.Migrate(connection)
	if err != nil {
		*errors = append(*errors, err)
	}
}

func checkTables(connection *dynamodb.DynamoDB) error {
	response, err := connection.ListTables(&dynamodb.ListTablesInput{})
	if err != nil {
		return fmt.Errorf("Error while checking tables: %v", err)
	}

	if response == nil || len(response.TableNames) == 0 {
		return fmt.Errorf("No tables found in the database")
	}

	for _, tableName := range response.TableNames {
		logger.INFO("Table found: ", *tableName)
	}

	return nil
}
