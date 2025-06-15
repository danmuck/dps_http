package api

import "github.com/gin-gonic/gin"

type Service struct {
	Name     string
	Version  string
	Endpoint string
	Bucket   string
	Running  bool
}

// Services must register their own routes with a gin router
type ServiceReg interface {
	Register(r *gin.Engine)
}
