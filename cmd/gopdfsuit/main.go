package main

import (
	"github.com/chinmay-sawant/gopdfsuit/internal/handlers"
	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()
	handlers.RegisterRoutes(router)
	router.Run()
}
