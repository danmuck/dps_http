package users

import (
	"net/http"

	logs "github.com/danmuck/dps_lib/logs"
	"github.com/gin-gonic/gin"
)

func UpdateUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		logs.Init("UpdateUser: updating user %s", id)

		// bind only the updatable fields
		var patch struct {
			Email     string   `json:"email,omitempty"`
			Bio       string   `json:"bio,omitempty"`
			AvatarURL string   `json:"avatarURL,omitempty"`
			Roles     []string `json:"roles,omitempty"` // if you want to allow role changes
		}
		if err := c.ShouldBindJSON(&patch); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
			return
		}

		// build a map of just the changed fields
		updates := map[string]any{}
		if patch.Email != "" {
			logs.Log("UpdateUser: patching email to %s", patch.Email)
			updates["email"] = patch.Email
		}
		if patch.Bio != "" {
			logs.Log("UpdateUser: patching bio to %s", patch.Bio)
			updates["bio"] = patch.Bio
		}
		if patch.AvatarURL != "" {
			logs.Log("UpdateUser: patching avatarURL to %s", patch.AvatarURL)
			updates["avatarURL"] = patch.AvatarURL
		}
		if len(patch.Roles) > 0 {
			logs.Log("UpdateUser: patching roles to %v", patch.Roles)
			updates["roles"] = patch.Roles
		}
		if len(updates) == 0 {
			logs.Log("UpdateUser: nothing to update for user %s", id)
			c.JSON(http.StatusBadRequest, gin.H{"error": "nothing to update"})
			return
		}

		// apply the patch
		if err := service.storage.Patch(service.userDB, id, updates); err != nil {
			logs.Log("UpdateUser: failed to update user %s: %v", id, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update"})
			return
		}

		// return the updated document (you can re‐fetch with Retrieve)
		logs.Log("UpdateUser: retreiving updated user %s", id)
		updated, _ := service.storage.Retrieve(service.userDB, id)
		logs.Log("UpdateUser: updated user %s: %v", id, updated)
		c.JSON(http.StatusOK, updated)
	}
}
