package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Todo struct {
	ID        primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Completed bool               `json:"completed"`
	Body      string             `json:"body"`
}

var coll *mongo.Collection

func main() {
	fmt.Println("hello world")

	if os.Getenv("ENV") != "production" {
		// Load the .env file
		if err := godotenv.Load(); err != nil {
			log.Fatal("Error loading .env file")
		}
	}

	// Get the MongoDB URI from the environment variable
	MONGODB_URI := os.Getenv("MONGODB_URI")
	if MONGODB_URI == "" {
		log.Fatal("MONGODB_URI environment variable not set")
	}

	// Create a MongoDB client
	clientOptions := options.Client().ApplyURI(MONGODB_URI)

	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	defer client.Disconnect(context.Background())

	err = client.Ping(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Connected to MongoDB!")

	coll = client.Database("golang_db").Collection("todos")

	app := fiber.New()

	app.Use(cors.New(cors.Config{
		AllowOrigins: "http://localhost:5173",
		AllowHeaders: "Origin,Content-Type,Accepted",
	}))

	app.Get("/api/todos", getTodos)
	app.Post("/api/todos", createTodos)
	app.Patch("/api/todos/:id", updateTodos)
	app.Delete("/api/todos/:id", deleteTodos)

	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}

	if os.Getenv("ENV") == "production" {
		app.Static("/", "./client/dist")
	}

	log.Fatal(app.Listen("0.0.0.0:" + port))

}

func getTodos(c *fiber.Ctx) error {
	var todos []Todo

	cursor, err := coll.Find(context.Background(), bson.M{})
	if err != nil {
		return err
	}

	defer cursor.Close(context.Background())
	for cursor.Next(context.Background()) {
		var todo Todo
		if err := cursor.Decode(&todo); err != nil {
			return err
		}

		todos = append(todos, todo)
	}
	return c.JSON(todos)
}

func createTodos(c *fiber.Ctx) error {
	todo := new(Todo)

	if err := c.BodyParser(todo); err != nil {
		return err
	}

	if todo.Body == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Todo body cannot be empty"})
	}

	insertedResult, err := coll.InsertOne(context.Background(), todo)
	if err != nil {
		return err
	}

	todo.ID = insertedResult.InsertedID.(primitive.ObjectID)
	return c.Status(201).JSON(todo)

}

func updateTodos(c *fiber.Ctx) error {
	id := c.Params("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid todo id"})
	}

	filter := bson.M{"_id": objectID}
	update := bson.M{"$set": bson.M{"completed": true}}
	_, err = coll.UpdateOne(context.Background(), filter, update)

	if err != nil {
		return err
	}

	return c.Status(200).JSON(fiber.Map{"success": true})
}

func deleteTodos(c *fiber.Ctx) error {
	id := c.Params("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid todo id"})
	}

	filter := bson.M{"_id": objectID}

	_, err = coll.DeleteOne(context.Background(), filter)
	if err != nil {
		return err
	}

	return c.Status(200).JSON(fiber.Map{"success": true})

}
