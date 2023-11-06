package user

import (
	"encoding/json"
	"errors"

	"github.com/athunlal/serverless/pkg/validators"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
)

var (
	ErrorFaledToFetchRecord      = "failed to fetch record"
	ErrorFailedToUnmarshalRecord = "failed to unmarshal record"
	ErrorInvalidUserData         = "invalid user date"
	ErrorInvalidEmail            = "invalid email"
	ErrorCouldNotMarshallItem    = "could not marshal item"
	ErrorCouldNotDeleteItem      = "could not delete item"
	ErrorCouldNotDynamoPutItem   = "could not dynamo put item"
	ErrorUserAlreadyExists       = "user.User already exists"
	ErroorUserDoesNotExisst      = "user.User does not exist"
)

type User struct {
	Email     string `json:"email"`
	FirstName string `json:"firstname"`
	LastName  string `json:"lastname"`
}

func FetUser(email, tableName string, dynaClient dynamodbiface.DynamoDBAPI) (*User, error) {
	input := &dynamodb.GetItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"email": {
				S: aws.String(email),
			},
		},
		TableName: aws.String(tableName),
	}
	result, err := dynaClient.GetItem(input)
	if err != nil {
		return nil, errors.New(ErrorFaledToFetchRecord)
	}

	item := new(User)
	err = dynamodbattribute.UnmarshalMap(result.Item, item)
	if err != nil {
		return nil, errors.New(ErrorFailedToUnmarshalRecord)
	}
}

func FetchUsers(tableName string, dynaClient dynamodbiface.DynamoDBAPI) (*[]User, error) {
	input := &dynamodb.ScanInput{
		TableName: aws.String(tableName),
	}

	result, err := dynaClient.Scan(input)
	if err != nil {
		return nil, errors.New(ErrorFaledToFetchRecord)
	}
	item := new([]User)
	err = dynamodbattribute.UnmarshalMap(result.Items, item)
	return item, nil
}

func CreateUser(req events.APIGatewayProxyRequest, tableName string, dynaClinet dynamodbiface.DynamoDBAPI) (*User, error) {
	var u User
	if err := json.Unmarshal([]byte(req.Body), &u); err != nil {
		return nil, errors.New(ErrorInvalidUserData)
	}
	if !validators.IsEmailValid(u.Email) {
		return nil, errors.New(ErrorInvalidEmail)
	}

	currenctUser, _ := FetUser(u.Email, tableName, dynaClinet)
	if currenctUser != nil && len(currenctUser.Email) != 0 {
		return nil, errors.New(ErrorUserAlreadyExists)
	}

	av, err := dynamodbattribute.MarshalMap(u)
	if err != nil {
		return nil, errors.New(ErrorCouldNotMarshallItem)
	}
	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(tableName),
	}
	_, err = dynaClinet.PutItem(input)
	if err != nil {
		return nil, errors.New(ErrorCouldNotDynamoPutItem)
	}
	return &u, nil
}

func UpdateUser(req events.APIGatewayProxyRequest, tablenName string, dynaClient dynamodbiface.DynamoDBAPI) (*User, error) {
	var u User

	if err := json.Unmarshal([]byte(req.Body), &u); err != nil {
		return nil, errors.New(ErrorInvalidEmail)
	}

	currentUser, _ := FetchUsers(u.Email, tablenName, dynaClient)
	if currentUser != nil && len(currentUser.Email) == 0 {
		return nil, errors.New(ErroorUserDoesNotExisst)
	}

	av, err := dynamodbattribute.MarshalMap(u)
	if err != nil {
		return nil, errors.New(ErrorCouldNotMarshallItem)
	}

	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(tablenName),
	}
	_, err = dynaclient.PutItem(input)
	if err != nil {
		return nil, errors.New(ErrorCouldNotDynamoPutItem)
	}

	return &u, nil
}

func DeleteUser(req events.APIGatewayProxyRequest, tabelName string, dynaClient dynamodbiface.DynamoDBAPI) error {
	email := req.QueryStringParameters["email"]
	input := &dynamodb.DeleteItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"email": {
				S: aws.String(email),
			},
		},
		TableName: aws.String(tabelName),
	}
	_, err := dynaClient.DeleteItem(input)
	if err != nil {

		return errors.New(ErrorCouldNotDeleteItem)
	}
	return nil
}
