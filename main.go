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
	ID        primitive.ObjectID `json:"_id, omitempty" bson:"_id,omitempty"`
	Completed bool               `json:"completed"`
	Body      string             `json:"body"`
}

var collection *mongo.Collection

func main() {

	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading env")
	}
	MONGO_URL := os.Getenv("MONGO_URL")
	clientOptions := options.Client().ApplyURI(MONGO_URL)
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(context.Background())

	err = client.Ping(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Connected to mongodb Atlas")
	collection = client.Database("go-todo").Collection("todos")

	app := fiber.New()
	app.Use(cors.New(cors.Config{
		AllowOrigins: "http://localhost:3000",
		AllowHeaders: "Origin,Content-type,Accept",
	}))
	app.Get("/api/todos", getTodos)
	app.Post("/api/todos", createTodo)
	app.Patch("/api/todos/:id", updateTodo)
	app.Delete("/api/todos/:id", deleteTodo)

	PORT := os.Getenv("PORT")
	if PORT == "" {
		PORT = "5000"
	}
	log.Fatal(app.Listen("0.0.0.0:" + PORT))
}
func getTodos(c *fiber.Ctx) error {
	var todos []Todo
	cursor, err := collection.Find(context.Background(), bson.M{})
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
	return c.Status(200).JSON(fiber.Map{"data": todos})
}

func createTodo(c *fiber.Ctx) error {
	todo := new(Todo)
	if err := c.BodyParser(todo); err != nil {
		return err
	}
	if todo.Body == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Body is required"})
	}
	insertResult, err := collection.InsertOne(context.Background(), todo)
	if err != nil {
		log.Fatal(err)
	}
	todo.ID = insertResult.InsertedID.(primitive.ObjectID)
	return c.Status(201).JSON(fiber.Map{"data": todo, "message": "success"})
}

func updateTodo(c *fiber.Ctx) error {
	id := c.Params("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid id"})
	}

	type UpdateTodo struct {
		Completed *bool   `json:"completed,omitempty"`
		Body      *string `json:"body,omitempty"`
	}

	var updateData UpdateTodo
	if err := c.BodyParser(&updateData); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Cannot parse JSON"})
	}

	update := bson.M{}
	if updateData.Completed != nil {
		update["completed"] = *updateData.Completed
	}
	if updateData.Body != nil {
		update["body"] = *updateData.Body
	}

	if len(update) == 0 {
		return c.Status(400).JSON(fiber.Map{"error": "No valid fields to update"})
	}

	updateCommand := bson.M{"$set": update}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)

	var updatedTodo Todo
	err = collection.FindOneAndUpdate(
		context.Background(),
		bson.M{"_id": objectID},
		updateCommand,
		opts,
	).Decode(&updatedTodo)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to update todo"})
	}

	return c.Status(200).JSON(fiber.Map{"data": updatedTodo, "message": "Todo updated successfully"})
}

func deleteTodo(c *fiber.Ctx) error {
	id := c.Params("id")
	objectID, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid todo ID"})
	}

	filter := bson.M{"_id": objectID}
	_, err = collection.DeleteOne(context.Background(), filter)

	if err != nil {
		return err
	}

	return c.Status(200).JSON(fiber.Map{"success": true})
}
