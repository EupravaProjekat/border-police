package Repo

import (
	"context"
	"errors"
	"fmt"
	"github.com/EupravaProjekat/border-police/Models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"log"
	"os"
	"time"
)

type Repo struct {
	logger *log.Logger
	cli    *mongo.Client
}

// New NoSQL: Constructor which reads db configuration from environment
func New(ctx context.Context, logger *log.Logger) (*Repo, error) {
	dburi := os.Getenv("MONGO_DB_URI")

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(dburi))
	if err != nil {
		return nil, err
	}

	return &Repo{
		cli:    client,
		logger: logger,
	}, nil
}

// Disconnect from database
func (ar *Repo) Disconnect(ctx context.Context) error {
	err := ar.cli.Disconnect(ctx)
	if err != nil {
		return err
	}
	return nil
}

// Ping Check database connection
func (ar *Repo) Ping() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Check connection -> if no error, connection is established
	err := ar.cli.Ping(ctx, readpref.Primary())
	if err != nil {
		ar.logger.Println(err)
	}
	// Print available databases
	databases, err := ar.cli.ListDatabaseNames(ctx, bson.M{})
	if err != nil {
		ar.logger.Println(err)
	}
	fmt.Println(databases)
}
func (ar *Repo) GetAll() ([]*Models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	Collection := ar.getCollection()
	var accommodationsSlice []*Models.User

	accommodationCursor, err := Collection.Find(ctx, bson.M{})
	if err != nil {
		ar.logger.Println(err)
		return nil, err
	}
	defer func(accommodationCursor *mongo.Cursor, ctx context.Context) {
		err := accommodationCursor.Close(ctx)
		if err != nil {
			ar.logger.Println(err)
		}
	}(accommodationCursor, ctx)

	for accommodationCursor.Next(ctx) {
		var user Models.User
		if err := accommodationCursor.Decode(&user); err != nil {
			ar.logger.Println(err)
			return nil, err
		}
		accommodationsSlice = append(accommodationsSlice, &user)
	}

	if err := accommodationCursor.Err(); err != nil {
		ar.logger.Println(err)
		return nil, err
	}

	return accommodationsSlice, nil
}

func (ar *Repo) UpdateRequest(uuid string) (*Models.Request, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	accCollection := ar.getCollection()
	var acc Models.Request

	err := accCollection.FindOne(ctx, bson.M{"uuid": uuid}).Decode(&acc)
	if err != nil {
		ar.logger.Println(err)
		return nil, err
	}

	return &acc, nil
}
func (ar *Repo) GetRequest(uuid string) (*Models.Request, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	accCollection := ar.getCollection()
	var acc Models.Request

	err := accCollection.FindOne(ctx, bson.M{"requests.uuid": uuid}).Decode(&acc)
	if err != nil {
		ar.logger.Println(err)
		return nil, err
	}

	return &acc, nil
}
func (ar *Repo) GetAllRequest() ([]*Models.Request, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var allRequests []*Models.Request

	Collection := ar.getCollection()
	cursor, err := Collection.Find(ctx, bson.M{})
	if err != nil {
		log.Fatal(err)
	}
	defer cursor.Close(ctx)

	// Iterate over the cursor
	for cursor.Next(ctx) {
		var user Models.User
		err := cursor.Decode(&user)
		if err != nil {
			log.Fatal(err)
		}
		// Convert each Request struct to a pointer to Request and append to allRequests
		for _, req := range user.Requests {
			allRequests = append(allRequests, &req)
		}
	}

	// Check for errors during cursor iteration
	if err := cursor.Err(); err != nil {
		log.Fatal(err)
	}

	// Print all requests
	for _, req := range allRequests {
		fmt.Printf("%+v\n", *req) // Dereference the pointer to print the request
	}

	return allRequests, nil
}
func (ar *Repo) GetAllNews() ([]*Models.News, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var allRequests []*Models.News

	collection := ar.getCollectionVehicles()
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err // Avoid using log.Fatal in library code; return the error instead
	}
	defer cursor.Close(ctx)

	// Iterate over the cursor
	for cursor.Next(ctx) {
		var request Models.News
		if err := cursor.Decode(&request); err != nil {
			return nil, err
		}
		allRequests = append(allRequests, &request)
	}
	if err := cursor.Err(); err != nil {
		return nil, err
	}

	// Reverse the slice
	reverse(allRequests)

	return allRequests, nil
}
func reverse(news []*Models.News) {
	for i, j := 0, len(news)-1; i < j; i, j = i+1, j-1 {
		news[i], news[j] = news[j], news[i]
	}
}
func (ar *Repo) GetAllCausings() ([]*Models.VehicleCausing, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var allRequests []*Models.VehicleCausing

	Collection := ar.getCollectionVehicles()
	cursor, err := Collection.Find(ctx, bson.M{})
	if err != nil {
		log.Fatal(err)
	}
	defer cursor.Close(ctx)

	// Iterate over the cursor
	for cursor.Next(ctx) {
		var request Models.VehicleCausing
		err := cursor.Decode(&request)
		if err != nil {
			log.Fatal(err)
		}
		allRequests = append(allRequests, &request)
	}
	if err := cursor.Err(); err != nil {
		log.Println(err)
	}

	return allRequests, nil
}
func (ar *Repo) GetByEmail(email string) (*Models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	accCollection := ar.getCollection()
	var acc Models.User

	err := accCollection.FindOne(ctx, bson.M{"email": email}).Decode(&acc)
	if err != nil {
		ar.logger.Println(err)
		return nil, err
	}
	if acc.Email == "" || len(acc.Email) < 3 {
		return nil, errors.New("invalid email")
	}
	return &acc, nil
}
func (ar *Repo) NewUser(Request *Models.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	accCollection := ar.getCollection()

	result, err := accCollection.InsertOne(ctx, &Request)
	if err != nil {
		return err
	}
	ar.logger.Printf("Documents ID: %v\n", result.InsertedID)
	return nil
}
func (ar *Repo) NewCausing(Request *Models.VehicleCausing) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	accCollection := ar.getCollectionVehicles()

	result, err := accCollection.InsertOne(ctx, &Request)
	if err != nil {
		return err
	}
	ar.logger.Printf("Documents ID: %v\n", result.InsertedID)
	return nil
}
func (ar *Repo) Update(User *Models.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	accCollection := ar.getCollection()
	filter := bson.M{"uuid": User.Uuid}
	update := bson.M{
		"$set": bson.M{
			"requests": User.Requests,
		},
	}

	// Perform the update operation
	updateResult, err := accCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		log.Fatal(err)
	}

	// Check the number of documents updated
	fmt.Printf("Updated %v document\n", updateResult.ModifiedCount)
	return nil
}

func (ar *Repo) Create(user *Models.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	accommodationCollection := ar.getCollection()
	result, err := accommodationCollection.InsertOne(ctx, &user)
	if err != nil {
		ar.logger.Println(err)
		return err
	}
	ar.logger.Printf("Documents ID: %v\n", result.InsertedID)
	return nil
}
func (ar *Repo) CreateNews(user *Models.News) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	accommodationCollection := ar.getCollection()
	result, err := accommodationCollection.InsertOne(ctx, &user)
	if err != nil {
		ar.logger.Println(err)
		return err
	}
	ar.logger.Printf("Documents ID: %v\n", result.InsertedID)
	return nil
}

func (ar *Repo) DeleteByEmail(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	accommodationCollection := ar.getCollection()

	filter := bson.M{"email": id}

	result, err := accommodationCollection.DeleteOne(ctx, filter)
	if err != nil {
		ar.logger.Println(err)
		return err
	}

	ar.logger.Printf("Documents deleted: %v\n", result.DeletedCount)

	return nil
}

func (ar *Repo) getCollection() *mongo.Collection {
	accommodationDatabase := ar.cli.Database("mongoBorder")
	accommodationCollection := accommodationDatabase.Collection("border")
	return accommodationCollection
}
func (ar *Repo) getCollectionNews() *mongo.Collection {
	accommodationDatabase := ar.cli.Database("mongoBorder")
	accommodationCollection := accommodationDatabase.Collection("border")
	return accommodationCollection
}
func (ar *Repo) getCollectionVehicles() *mongo.Collection {
	accommodationDatabase := ar.cli.Database("mongoBorder")
	accommodationCollection := accommodationDatabase.Collection("border-vehicles")
	return accommodationCollection
}
