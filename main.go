package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
)

type Todo struct {
	ID        int    `json: "id"`
	Completed bool   `json: "completed"`
	Body      string `json: "body"`
}

func main() {
	app := fiber.New()

	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loadin env")
	}
	PORT := os.Getenv("PORT")
	todos := []Todo{}

	app.Get("/api/todos", func(c *fiber.Ctx) error {
		return c.Status(200).JSON(todos)
	})

	// Create a Todo
	app.Post("/api/todos", func(c *fiber.Ctx) error {
		todo := &Todo{} //the & gets the memorry address of Todos
		if err := c.BodyParser(todo); err != nil {
			return err
		}
		if todo.Body == "" {
			return c.Status(400).JSON(fiber.Map{"error": "Body is required"})
		}
		todo.ID = len(todos) + 1
		todos = append(todos, *todo) //add * to get the memory address value
		return c.Status(201).JSON(todos)
	})

	//Update a Todo
	app.Patch("/api/todos/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")

		type UpdateTodo struct {
			Completed *bool   `json:"completed"`
			Body      *string `json:"body"`
		}

		var updateData UpdateTodo
		// Parse the incoming JSON request body
		if err := c.BodyParser(&updateData); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Cannot parse JSON"})
		}

		for i, todo := range todos {
			if fmt.Sprint(todo.ID) == id {
				// Only update fields if they are present in the request
				if updateData.Completed != nil {
					todos[i].Completed = *updateData.Completed
				}
				if updateData.Body != nil {
					todos[i].Body = *updateData.Body
				}
				return c.Status(200).JSON(todos[i])
			}
		}
		return c.Status(404).JSON(fiber.Map{"Error": "Todo not found"})
	})

	//Dekete a Todo

	app.Delete("/api/todos/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")

		for i, todo := range todos {
			// Only update fields if they are present in the request
			if fmt.Sprint(todo.ID) == id {
				todos = append(todos[:i], todos[i+1:]...)
				return c.Status(200).JSON(fiber.Map{"data": todos, "message": "todo deleted"})

			}

		}
		return c.Status(404).JSON(fiber.Map{"Error": "Todo not found"})
	})

	log.Fatal(app.Listen(":" + PORT))
}
