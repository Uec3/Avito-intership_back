package main

import (
	"avito_intern_dev/handlers"
	"avito_intern_dev/models"
	"log"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	dsn := "host = db user=postgres password=pass dbname=prdb port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Could not connect to DB", err)
	}
	db.AutoMigrate(&models.Team{}, &models.User{}, &models.PullRequest{})

	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
	r.POST("/team/add", handlers.AddTeam(db))
	r.GET("/team/get", handlers.GetTeam(db))
	r.POST("/users/setIsActive", handlers.SetIsActive(db))
	r.POST("/pullRequest/create", handlers.CreatePR(db))
	r.Run(":8080")
}
