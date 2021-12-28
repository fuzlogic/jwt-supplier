package main

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"jwt-supplier/auth"
	"jwt-supplier/check"
	"jwt-supplier/docs"
	"log"
)

// @title Auth Blueprint Swagger API
// @version 1.0
// @description Swagger API for Auth Blueprint.
// @termsOfService http://swagger.io/terms/
// @contact.name Roman Kasovsky
// @contact.email roman@kasovsky.ru
// @license.name Apache-2.0
// @license.url https://directory.fsf.org/wiki/License:Apache-2.0
// @BasePath /api/v1
func main() {

	auth.Dbpool = auth.OpenDB()
	defer auth.CloseDB(auth.Dbpool)

	router := gin.Default()
	docs.SwaggerInfo.BasePath = "/api/v1"
	apiV1 := router.Group("/api/v1")
	{
		authGroup := apiV1.Group("/auth")
		{
			authGroup.POST("/signup", auth.Signup)
			authGroup.POST("/signin", auth.Signin)
			authGroup.GET("/welcome", auth.Welcome)
			authGroup.POST("/refresh", auth.Refresh)
			authGroup.POST("/delete", auth.Delete)
		}
		otherGroup := apiV1.Group("/other")
		{
			otherGroup.GET("/healthcheck", check.HealthCheck)
		}
	}
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	log.Fatal(router.Run(":9000"))
}
