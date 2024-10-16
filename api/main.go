package api

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"qa/model"
)

// Index 主页
func Index(c *gin.Context) {
	c.String(http.StatusOK, "======= https://github.com/Hui4401/qa =======")
}

// CurrentUser 获取当前用户
func CurrentUser(c *gin.Context) *model.User {
	if userID, _ := c.Get("user_id"); userID != nil {
		if user, err := model.GetUser(userID); err == nil {
			return &user
		}
	}
	return nil
}
