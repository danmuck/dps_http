package admin

// AdminService implements the service interface
// route handlers are defined in their corresponding files
// (note: there are none for this template service)
// //
// var service *AdminService

// type AdminService struct {
// 	endpoint string // endpoint for the service; canonically /api/v_/<service_name>
// 	version  string // api version; the version this template resides in

// 	// service specific structures
// 	userDB  string
// 	storage *mongo_client.MongoClient

// 	datagen chan any
// }

// assign routes for the service and initialize any resources
// routes are structured `api/v1/<service_name>/<your_endpoints>`
// func (svc *AdminService) Up(root *gin.RouterGroup) {

// 	rg := root.Group("/")
// 	// rg.Use(middleware.JWTMiddleware(), middleware.AuthorizeByRoles("admin"), middleware.RateLimitMiddleware())

// 	ug := rg.Group("/users")
// 	ug.POST("/:id", CreateUser())
// 	ug.DELETE("/:id", DeleteUser()) // Delete user by ID

// 	admin := rg.Group("/admin")
// 	admin.POST("/gcx", CreateUsersX())
// 	admin.POST("/dcx", DeleteUsersX()) // Delete user by ID

// 	logs.Info("[AdminService] up at %s and %s", root.BasePath(), ug.BasePath())
// }

// bring the service down gracefully and release all resources
// func (svc *AdminService) Down() error {
// 	logs.Info("auth service Down() is partially implemented")
// 	svc.datagen <- nil
// 	return fmt.Errorf("error not yet implemented")
// }

// // returns the API version this depends on
// func (svc *AdminService) Version() string {
// 	return svc.version
// }

// // any other services this depends on
// func (svc *AdminService) DependsOn() []string {
// 	return nil
// }

// returns a pointer to the server instance
// expects its state to be initialized and ready for Up()
// func NewAdminService(endpoint string) *AdminService {
// 	// m, err := mongo_client.NewMongoStore(cfg.DB.MongoURI, cfg.DB.Name)
// 	// if err != nil {
// 	// 	logs.Log("failed to create mongo store: %v", err)
// 	// 	return nil
// 	// }
// 	version := "v1"
// 	service = &AdminService{
// 		endpoint: endpoint,
// 		version:  version,
// 		userDB:   "users" + version,
// 		storage:  nil,
// 		datagen:  newDataGenerator(),
// 	}
// 	return service
// }

// for development purposes
// func dummyString(length int, postfix string) string {
// 	const letters = "abcdefghijklmnopqrstuvwxyz01"
// 	b := make([]byte, length)
// 	for i := range b {
// 		b[i] = letters[rand.Intn(len(letters))]
// 	}
// 	return fmt.Sprintf("%s%s", string(b), postfix)
// }

// func createDummyUser(username string) error {
// 	email := dummyString(8, "@dirtranch.io")
// 	password := dummyString(4, "crypt")
// 	if username == "" || username == "undefined" {
// 		// logs.Log("[DEV]> No username provided, generating a random one")
// 		username = dummyString(4, "dps")
// 	}

// 	if _, found := service.storage.Lookup(service.userDB, bson.M{"username": username}); found {
// 		logs.Warn("[DEV]> User %s already exists, generating a new one", username)
// 		username = dummyString(4, "dps")
// 	}

// 	// logs.Log("[DEV]> Creating user: %s", username)

// 	hash, err := auth.HashPassword(password)
// 	if err != nil {
// 		logs.Err("[DEV]> Hashing error: %v", err)
// 		return err
// 	}
// 	// t := sha512.Sum512([]byte(username))
// 	// logs.Dev("token: %s", t)

// 	// logs.Log("[DEV]> Hashed password: %s", hash)
// 	user := api.User{
// 		ID:           primitive.NewObjectID(),
// 		Username:     username,
// 		Email:        email,
// 		PasswordHash: hash,
// 		Roles:        []string{"dummy"},
// 		Token:        "NEW_USER_TOKEN",
// 		Bio:          "the dev. the creator.. ... ..i'm special",
// 		AvatarURL:    "banner.svg",
// 		CreatedAt:    primitive.NewDateTimeFromTime(time.Now()),
// 		UpdatedAt:    primitive.NewDateTimeFromTime(time.Now()),
// 	}

// 	logs.Debug("[DEV]> User object: %s", user.String())

// 	// Store the user in the database
// 	if err := service.storage.Store(service.userDB, user.ID.Hex(), user); err != nil {
// 		return err
// 	}
// 	return nil
// }
