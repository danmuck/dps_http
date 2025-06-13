package api

import "github.com/gin-gonic/gin"

// Services must register their own routes with a gin router
type ServiceReg interface {
	Register(r *gin.Engine)
}
