package check

import (
	"fmt"
	"github.com/gin-gonic/gin"
)

// HealthCheck godoc
// @Summary HealthCheck Summary
// @Schemes
// @Description HealthCheck Description
// @Tags other
// @Produce json
// @Success 200 {object} string
// @Router /other/healthcheck [get]
func HealthCheck(context *gin.Context) {
	context.Writer.Write([]byte(fmt.Sprintf("%t", true)))
}

