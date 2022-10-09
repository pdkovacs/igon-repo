package httpadapter

import (
	"context"
	"errors"
	"igo-repo/internal/app/security/authr"
	"igo-repo/internal/app/services"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"nhooyr.io/websocket"
)

type socketAdapter struct {
	wsConn *websocket.Conn
}

func (sa *socketAdapter) CloseRead(ctx context.Context) context.Context {
	return sa.wsConn.CloseRead(ctx)
}

func (sa *socketAdapter) Close() error {
	return sa.wsConn.Close(websocket.StatusPolicyViolation, "connection too slow to keep up with messages")
}

func (sa *socketAdapter) Write(ctx context.Context, msg string) error {
	return sa.wsConn.Write(ctx, websocket.MessageText, []byte(msg))
}

func subscriptionHandler(getUserInfo func(c *gin.Context) authr.UserInfo, ns *services.Notification, logger zerolog.Logger) gin.HandlerFunc {
	return func(g *gin.Context) {
		wsConn, subsErr := websocket.Accept(g.Writer, g.Request, nil)
		if subsErr != nil {
			logger.Error().Msgf("Failed to accept WS connection request: %v", subsErr)
			g.Error(subsErr)
			g.AbortWithStatus(500)
			return
		}

		defer wsConn.Close(websocket.StatusInternalError, "")

		userInfo := getUserInfo(g)

		curriedContext := wsConn.CloseRead(g.Request.Context())                                    // Clients wan't write to the WS.(?)
		subscriptionError := ns.Subscribe(curriedContext, &socketAdapter{wsConn}, userInfo.UserId) // we block here until Error or Done

		if errors.Is(subscriptionError, context.Canceled) {
			return // Done
		}

		if websocket.CloseStatus(subscriptionError) == websocket.StatusNormalClosure ||
			websocket.CloseStatus(subscriptionError) == websocket.StatusGoingAway {
			return
		}
		if subscriptionError != nil {
			logger.Error().Msgf("%v", subscriptionError)
			return
		}
	}
}
