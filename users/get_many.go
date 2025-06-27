package users

import (
	"net/http"
	"strconv"

	"github.com/danmuck/dps_http/mongo_client"
	"github.com/danmuck/dps_lib/logs"
	"github.com/gin-gonic/gin"
)

/*
*    Expects:
*    - rows: []users
*    - totalCount: int
 */
func ListUsers(client *mongo_client.MongoClient) gin.HandlerFunc {
	logs.Dev("ListUsers from storage: %s", client.Name())
	return func(c *gin.Context) {
		pageStr := c.Query("page")
		pageSizeStr := c.Query("pageSize")
		logs.Dev("ListUsers: page=%s, pageSize=%s", pageStr, pageSizeStr)
		page, err := strconv.Atoi(pageStr)
		if err != nil {
			page = 0
		}
		pageSize, err := strconv.Atoi(pageSizeStr)
		if err != nil || pageSize <= 0 {
			pageSize = 10
		}

		users, err := ListUsersT()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch users"})
			return
		}

		start := page * pageSize
		end := start + pageSize
		if start > len(users) {
			start = len(users)
		}
		if end > len(users) {
			end = len(users)
		}
		// temp debugging
		// func(users []*users.User) {
		// 	for _, user := range users {
		// 		logs.Debug("List of user: %+v", user)
		// 	}
		// }(users)

		c.JSON(http.StatusOK, gin.H{
			"rows":       users[start:end],
			"totalCount": len(users),
		})
	}
}
