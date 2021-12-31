package home

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func (h *Controller) Data(c *gin.Context) {

	conn, err := h.upgrade.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusBadRequest, map[string]string{"msg": "连接失败"})
		return
	}

	client, err := NewClient(conn)
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]string{"msg": "内部错误"})
		return
	}

	go client.ReadMessage()
	go client.WriteMessage()

	client.SendHello()

}
