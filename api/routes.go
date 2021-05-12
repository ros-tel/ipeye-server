package api

import (
	"ipeye-server/api/balancer"

	"github.com/gin-gonic/gin"
)

func InitRoutes(r *gin.Engine) {
	r.GET("/balancer/server/:cloud_id", balancer.Server)
	r.GET("/balancer/uuid", balancer.UUID)
}
