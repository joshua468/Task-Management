package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

var db *gorm.DB

func main() {
	var err error
	db, err := gorm.Open("sqlite3", "testdb")
	if err != nil {
		panic("failed to connect to database")
	}
	defer db.Close()

	db.AutoMigrate(&Task{})

	r := gin.Default()
	r.Use(errorHandler)
	r.POST("/login", loginHandler)

	authMiddleware := authMiddleware()

	authenticated := r.Group("/api")
	authenticated.Use(authMiddleware)
	{
		authenticated.POST("/tasks", createTaskHandler)
		authenticated.GET("/tasks/:id", getTaskHandler)
		authenticated.PUT("/tasks/:id", updateTaskHandler)
		authenticated.DELETE("/tasks/:id", deleteTaskHandler)
	}
	r.Run(":8080")
}

func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Implement your authentication logic here
		// For example, you can check the presence of a token in the request headers
		token := c.GetHeader("Authorization")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}
		// If authentication succeeds, you can proceed to the next middleware
		c.Next()
	}
}

type Task struct {
	ID     uint   `json:"id" gorm:"primary_key"`
	Title  string `json:"title"`
	Status string `json:"status"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func createTaskHandler(c *gin.Context) {
	var task Task
	if err := c.ShouldBindJSON(&task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	db.Create(&task)
	c.JSON(http.StatusCreated, task)
}

func getTaskHandler(c *gin.Context) {
	id := c.Param("id")
	var task Task
	if err := db.First(&task, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "task Not found"})
		return
	}
	c.JSON(http.StatusOK, task)
}

func updateTaskHandler(c *gin.Context) {
	id := c.Param("id")
	var task Task
	if err := db.First(&task, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "task Not Found"})
		return
	}
	var updatedTask Task
	if err := c.ShouldBindJSON(&updatedTask); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	db.Model(&task).Updates(&updatedTask)
	c.JSON(http.StatusOK, task)
}

func deleteTaskHandler(c *gin.Context) {
	id := c.Param("id")
	var task Task
	if err := db.First(&task, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "task Not Found"})
		return
	}
	db.Delete(&task)
	c.JSON(http.StatusOK, gin.H{"message": "task deleted successfully"})
}

func loginHandler(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Username == "user" && req.Password == "password" {
		token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3MTI2ODA2NjYsInVzZXJuYW1lIjoiZXhhbXBsZVVzZXIifQ.QyEq2PgO7fyixGcCqdZ07nFj8AV4YIvnNQL6wYuqf5k"
		c.JSON(http.StatusOK, gin.H{"token": token})
		return
	}
	c.JSON(http.StatusOK, gin.H{"error": "invalid credentials"})
}

func errorHandler(c *gin.Context) {
	c.Next()
	if len(c.Errors) == 0 {
		return
	}
	for _, e := range c.Errors {
		fmt.Println("Error", e.Error())
	}
}
