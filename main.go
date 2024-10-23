package main

import (
	"database/sql"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	_ "github.com/lib/pq"
)

type Activity struct {
    ID       	 int    	`json:"id"`
    Title 		 string 	`json:"title"`
	Category     string 	`json:"category"`
	Description  string    	`json:"description"`
	ActivityDate time.Time  `json:"activity_date"`
	Status       string 	`json:"status"`
	CreatedAt    time.Time  `json:"created_at"`
}

func initDB() (*sql.DB, error) {
	dns := "user=postgres.narqvhcfdhxovbnwfjky password=daviddmppdatabase15 host=aws-0-ap-southeast-1.pooler.supabase.com port=6543 dbname=postgres"
	db, err := sql.Open("postgres", dns)
	if err!= nil {
		return nil, err
	}

	err = db.Ping()
	if err!= nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

func main() {
	db, err := initDB()
	if err!= nil {
		panic(err)
	}
	defer db.Close()

	app := fiber.New()

	// Get /activities
	app.Get("/activities", func(c *fiber.Ctx) error {
        rows, err := db.Query("SELECT * FROM activities")
        if err!= nil {
            return c.Status(500).JSON(fiber.Map{"error": err.Error()})
        }
        defer rows.Close()

        var activities []Activity
        for rows.Next() {
            var activity Activity
            err := rows.Scan(&activity.ID, &activity.Title, &activity.Category, &activity.Description, &activity.ActivityDate, &activity.Status, &activity.CreatedAt)
            if err!= nil {
                return c.Status(500).JSON(fiber.Map{"error": err.Error()})
            }
            activities = append(activities, activity)
        }

        return c.JSON(activities)
    })

    // Get /activities/:id
    app.Get("/activities/:id", func(c *fiber.Ctx) error {
        id := c.Params("id")
		row := db.QueryRow("SELECT * FROM activities WHERE id=$1", id)
		var activity Activity
		err := row.Scan(&activity.ID, &activity.Title, &activity.Category, &activity.Description, &activity.ActivityDate, &activity.Status, &activity.CreatedAt)
		if err!= nil {
			return c.Status(404).JSON(fiber.Map{"error": err.Error()})
        }
		return c.JSON(activity)
	})

	// Post /activities
	app.Post("/activities", func(c *fiber.Ctx) error {
        var activity Activity
        if err := c.BodyParser(&activity); err!= nil {
            return c.Status(400).JSON(fiber.Map{"error": err.Error()})
        }

        // Create a new activity and return the inserted id
		err := db.QueryRow("INSERT INTO activities (title, category, description, activity_date, status, created_at) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id", activity.Title, activity.Category, activity.Description, activity.ActivityDate, activity.Status, time.Now()).Scan(&activity.ID)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		activity.CreatedAt = time.Now()
		return c.Status(201).JSON(activity)
		})

	// Update /activities/:id
    app.Put("/activities/:id", func(c *fiber.Ctx) error {
        id := c.Params("id")
        var activity Activity
        if err := c.BodyParser(&activity); err != nil {
            return c.Status(400).JSON(fiber.Map{"error": err.Error()})
        }

        // Convert id from string to int
        activityID, err := strconv.Atoi(id)
        if err != nil {
            return c.Status(400).JSON(fiber.Map{"error": "Invalid activity ID"})
        }

        // Update the activity in the database
        result, err := db.Exec("UPDATE activities SET title=$1, category=$2, description=$3, activity_date=$4, status=$5, created_at=$6 WHERE id=$7", activity.Title, activity.Category, activity.Description, activity.ActivityDate, activity.Status, time.Now(), activityID)
        if err != nil {
            return c.Status(500).JSON(fiber.Map{"error": err.Error()})
        }

        rowsAffected, err := result.RowsAffected()
        if err != nil {
            return c.Status(500).JSON(fiber.Map{"error": err.Error()})
        }

        if rowsAffected == 0 {
            return c.Status(404).JSON(fiber.Map{"error": "Activity not found"})
        }

        activity.ID = activityID // Set the correct ID
        activity.CreatedAt = time.Now() // Set the updated time

        return c.Status(200).JSON(activity)

	// delete /activities/:id
	})

    app.Delete("/activities/:id", func(c *fiber.Ctx) error {
        id := c.Params("id")
        _, err := db.Exec("DELETE FROM activities WHERE id=$1", id)
        if err != nil {
            return c.Status(500).JSON(fiber.Map{"error": err.Error()})
        }
        return c.Status(204).SendString("Activity deleted successfully")
    })



	app.Listen(":8000")
}