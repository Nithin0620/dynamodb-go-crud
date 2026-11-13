package main

import(
	"fmt"
	"log"
	"net/http"
	
	"github.com/Nithin0620/dynamoDB-CRUD/internal/repository/adapter"
	"github.com/Nithin0620/dynamoDB-CRUD/internal/repository/instance"
	"github.com/Nithin0620/dynamoDB-CRUD/internal/routes"
	"github.com/Nithin0620/dynamoDB-CRUD/internal/rules"
	"github.com/Nithin0620/dynamoDB-CRUD/internal/rules/product"
	"github.com/Nithin0620/dynamoDB-CRUD/utils/logger"
	
	"github.com/aws/aws-sdk-go/service/dynamodb"
	
)

func main() {
	configs := config.GetConfig()
	connection := instance.GetConnection()
	repository := adapter.NewAdapter(connection)

	logger.INFO("Waiting for the service to start on port: ", configs.Port)

	errors := Migrate(connection)
	if len(errors) > 0{
		for _,err := range errors{
			logger.PANIC("Error while migrating the database: ", err)
		}
	}

	logger.PANIC("",checkTables(connection))

	port := fmt.Sprintf(":%d", configs.Port) 
	router := routes.NewRouter().SetRouters(repository)
	logger.INFO("Server is running on port: ", port)

	server := http.ListenAndServe(port,router)

	log.Fatal(server)
}	



func Migrate(connection *dynamodb.DynamoDB) []error {
 	var errors []error
	callMigrateAndAppendError(&errors, connection, &RulesProduct.Rules{})	
	return errors
}

func callMigrateAndAppendError(errors *[]error, connection *dynamodb.DynamoDB,rules rules.Interface) {
	// Implementation for calling migration and appending errors
	err := rules.Migrate(connection)
	if err != nil {
		*errors = append(*errors, err)
	}

}

func checkTables(connection *dynamodb.DynamoDB) error {
	response,err := connection.ListTables(&dynamodb.ListTables(&dynamodb.ListtablesInput{}))
	if err != nil {
		logger.PANIC("Error while checking tables: ", err)
		return fmt.Errorf("Error while checking tables: %v", err)
	}

	if response == nil || len(response.TableNames) == 0 {
		logger.PANIC("No tables found in the database",nil)
		return fmt.Errorf("No tables found in the database")
	}

	for _,tableName := range response.TableNames {
		logger.INFO("Table found: ", *tableName)
	}
	return err
}