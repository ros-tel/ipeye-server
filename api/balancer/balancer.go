package balancer

import (
	"log"
	"net/http"
	"strconv"

	"ipeye-server/config"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

//Server получает адрес облака
func Server(c *gin.Context) {
	cnf := c.MustGet("cnf").(*config.Balancer)

	cloud_id := c.Param("cloud_id")

	// По умолчанию
	message, ok := getServer("default", cnf)
	if !ok {
		log.Println("[ERROR] Default cloud not set in config")
		c.JSON(
			http.StatusInternalServerError,
			gin.H{
				"code":    500,
				"message": message,
				"status":  1,
			},
		)
		return
	}

	if server, ok := getServer(cloud_id, cnf); ok {
		message = server
	}

	c.JSON(
		http.StatusOK,
		gin.H{
			"code":    200,
			"message": message,
			"status":  1,
		},
	)
}

func getServer(cloud_id string, cnf *config.Balancer) (string, bool) {
	if route, ok := cnf.Routes[cloud_id]; ok {
		if server, ok := cnf.Servers[route]; ok {
			return server.Host + "|" + strconv.Itoa(server.Port), true
		}
	}

	return "", false
}

//UUID получает новый id устройства
func UUID(c *gin.Context) {
	cloud_id := uuid.New()

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": cloud_id, "status": 1})
}
