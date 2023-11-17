package httpadapter

import (
	"context"
	"errors"
	"iconrepo/internal/app/security/authr"
	"iconrepo/internal/app/services"
	"net/http"

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

func subscriptionHandler(getUserInfo func(c *gin.Context) authr.UserInfo, ns *services.Notification, loadBalancerAddress string) gin.HandlerFunc {
	return func(g *gin.Context) {
		logger := zerolog.Ctx(g.Request.Context()).With().Str("function", "subscriptionHandler").Logger()

		wsConn, subsErr := websocket.Accept(g.Writer, g.Request, &websocket.AcceptOptions{
			OriginPatterns: []string{loadBalancerAddress},
		})
		if subsErr != nil {
			logger.Error().Err(subsErr).Msg("Failed to accept WS connection request")
			g.Error(subsErr)
			g.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		defer wsConn.Close(websocket.StatusInternalError, "")

		userInfo := getUserInfo(g)

		curriedContext := wsConn.CloseRead(g.Request.Context())                                    // Clients wan't write to the WS.(?)
		subscriptionError := ns.Subscribe(curriedContext, &socketAdapter{wsConn}, userInfo.UserId) // we block here until Error or Done

		if errors.Is(subscriptionError, context.Canceled) {
			logger.Error().Err(subsErr).Msg("subscription terminated due to context-cancelation")
			return // Done
		}

		if websocket.CloseStatus(subscriptionError) == websocket.StatusNormalClosure ||
			websocket.CloseStatus(subscriptionError) == websocket.StatusGoingAway {
			logger.Error().Err(subscriptionError).Msg("subscription terminated due to StatusNormalClosure or StatusGoingAway")
			return
		}
		if subscriptionError != nil {
			logger.Error().Err(subscriptionError).Msg("subscription terminated without error")
			return
		}
	}
}
