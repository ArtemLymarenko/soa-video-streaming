package cookie

import (
	"time"

	"github.com/gin-gonic/gin"
)

const AccessToken = "access_token"

func GetAccessToken(gc *gin.Context) (string, error) {
	token, err := gc.Cookie(AccessToken)
	if err != nil {
		return "", err
	}

	return token, nil
}

func SetAccessToken(gc *gin.Context, token string, ttl time.Duration) {
	gc.SetCookie(AccessToken, token, int(ttl.Seconds()), "/", "", false, true)
}
