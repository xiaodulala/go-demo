package user

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func (u *Controller) Get(c *gin.Context) {
	c.JSON(http.StatusOK, map[string]string{"name": "duyong"})
	return
}
