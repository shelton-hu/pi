package websocket

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type Upgrader struct {
	websocket.Upgrader
}

func NewUpgrader() *Upgrader {
	return &Upgrader{
		websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
}

func (u *Upgrader) UpgradeGin(c *gin.Context) (*websocket.Conn, error) {
	return u.Upgrader.Upgrade(c.Writer, c.Request, nil)
}
