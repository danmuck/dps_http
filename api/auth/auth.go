package auth

import (
	"github.com/danmuck/dps_http/configs"
	"github.com/danmuck/dps_http/lib/logs"
	"github.com/danmuck/dps_http/lib/middleware"
	"github.com/danmuck/dps_http/storage/mongo"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

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
	if check != nil {
		return false
	}
	return true
}

// interface impl
// //
var service *AuthService

type AuthService struct {
	endpoint, version string
	secret            string
	userDB            string
	storage           *mongo.MongoClient
}

func (as *AuthService) Up(rg *gin.RouterGroup) {
	middleware.InitAuthMiddleware(service.secret)
	rg.POST("/login", LoginHandler())
	rg.POST("/logout", LogoutHandler())
	rg.POST("/register", RegisterHandler())
}

func (as *AuthService) Down() error {
	return nil
}

func NewAuthService(endpoint, version string) *AuthService {
	cfg, err := configs.LoadConfig()
	if err != nil {
		logs.Fatal(err.Error())
	}
	m, err := mongo.NewMongoStore(cfg.DB.MongoURI, cfg.DB.Name)
	if err != nil {
		logs.Log("failed to create mongo store: %v", err)
		return nil
	}
	service = &AuthService{
		endpoint: endpoint,
		version:  version,
		userDB:   "users" + version, // need to fix this
		secret:   cfg.Auth.JWTSecret,
		storage:  m,
	}
	return service
}
