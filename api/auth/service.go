package auth

import (
	"fmt"

	"github.com/danmuck/dps_http/configs"
	"github.com/danmuck/dps_http/lib/storage/mongo"
	"github.com/danmuck/dps_http/middleware"
	logs "github.com/danmuck/dps_lib/logs"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// interface impl
// //
var service *AuthService

type AuthService struct {
	endpoint, version string
	secret            string
	userDB            string
	storage           *mongo.MongoClient
}

func (svc *AuthService) Up(rg *gin.RouterGroup) {

	middleware.InitAuthMiddleware(service.secret)
	ag := rg.Group("/auth")
	ag.POST("/login", LoginHandler())
	ag.POST("/logout", LogoutHandler())
	ag.POST("/register", RegisterHandler())
}

func (svc *AuthService) Down() error {
	logs.Dev("auth service Down() not yet implemented")
	return fmt.Errorf("not yet implemented")
}

func (svc *AuthService) Version() string {
	return svc.version
}

func (svc *AuthService) DependsOn() []string {
	return nil
}

func NewAuthService(endpoint string) *AuthService {
	cfg, err := configs.LoadConfig()
	if err != nil {
		logs.Fatal(err.Error())
	}
	m, err := mongo.NewMongoStore(cfg.DB.MongoURI, cfg.DB.Name)
	if err != nil {
		logs.Err("failed to create mongo store: %v", err)
		return nil
	}
	version := "v1"
	service = &AuthService{
		endpoint: endpoint,
		version:  version,
		userDB:   "users" + version, // need to fix this
		secret:   cfg.Auth.JWTSecret,
		storage:  m,
	}
	return service
}

// registerPayload defines the input for user registration.
type registerPayload struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Confirm  string `json:"confirm"  binding:"required,eqfield=Password"` // confirm password must match
}

// loginPayload represents the input for user login.
type loginPayload struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

func VerifyPassword(hashed, password string) bool {
	check := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(password))
	return check == nil
}
