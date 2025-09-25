package user

import (
	"context"

	"belimang/internal/infrastructure/cache"
	"belimang/internal/infrastructure/database"
	"belimang/internal/pkg/jwt"
	logger "belimang/internal/pkg/logging"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	queries    *database.Queries
	jwtService *jwt.JWTService
	cache      *cache.RedisCache
}

func NewUserService(queries *database.Queries, jwtService *jwt.JWTService, cache *cache.RedisCache) *UserService {
	return &UserService{
		queries:    queries,
		jwtService: jwtService,
		cache:      cache,
	}
}

func (s *UserService) Register(req *RegisterRequest) (string, error) {
	ctx := context.Background()
	logger.InfoCtx(ctx, "Registering new user", "username", req.Username)

	// Check if username already exists
	usernameExists, err := s.queries.CheckUsernameExists(ctx, req.Username)
	if err != nil {
		logger.ErrorCtx(ctx, "Failed to check username existence", "error", err)
		return "", err
	}
	if usernameExists {
		logger.WarnCtx(ctx, "Username already exists", "username", req.Username)
		return "", ErrUsernameExists
	}

	// Check if email already exists for users
	emailExists, err := s.queries.CheckEmailExistsForRole(ctx, database.CheckEmailExistsForRoleParams{
		Email: req.Email,
		Role:  "user",
	})
	if err != nil {
		logger.ErrorCtx(ctx, "Failed to check email existence", "error", err)
		return "", err
	}
	if emailExists {
		logger.WarnCtx(ctx, "Email already exists for user", "email", req.Email)
		return "", ErrEmailExists
	}

	// Hash password
	logger.DebugCtx(ctx, "Hashing password", "username", req.Username)
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), 8)
	if err != nil {
		logger.ErrorCtx(ctx, "Failed to hash password", "error", err)
		return "", err
	}

	// Create user
	user := NewUser(req.Username, req.Email, string(hashedPassword))

	userUUID, err := uuid.Parse(user.ID)
	if err != nil {
		logger.ErrorCtx(ctx, "Failed to parse user ID as UUID", "error", err)
		return "", err
	}

	createdUser, err := s.queries.CreateUser(ctx, database.CreateUserParams{
		ID:           userUUID,
		Username:     user.Username,
		PasswordHash: user.Password,
		Email:        user.Email,
		Role:         database.UserRole(user.Role),
	})
	if err != nil {
		logger.ErrorCtx(ctx, "Failed to create user", "error", err, "username", req.Username)
		return "", err
	}

	// Generate JWT token
	token, err := s.jwtService.GenerateToken(createdUser.ID.String(), createdUser.Email, createdUser.Username)
	if err != nil {
		logger.ErrorCtx(ctx, "Failed to generate token", "error", err, "user_id", createdUser.ID)
		return "", err
	}

	logger.InfoCtx(ctx, "User registered successfully", "user_id", createdUser.ID, "username", createdUser.Username)
	return token, nil
}

func (s *UserService) Login(req *LoginRequest) (string, error) {
	ctx := context.Background()
	logger.InfoCtx(ctx, "Attempting user login", "username", req.Username)

	user, err := s.queries.GetUserByUsernameAndRole(ctx, database.GetUserByUsernameAndRoleParams{
		Username: req.Username,
		Role:     "user",
	})
	if err != nil {
		logger.WarnCtx(ctx, "User not found during login", "username", req.Username)
		return "", ErrInvalidCredentials
	}

	// Compare password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		logger.WarnCtx(ctx, "Invalid password during login", "username", req.Username)
		return "", ErrInvalidCredentials
	}

	// Generate JWT token
	token, err := s.jwtService.GenerateToken(user.ID.String(), user.Email, user.Username)
	if err != nil {
		logger.ErrorCtx(ctx, "Failed to generate token", "error", err, "user_id", user.ID)
		return "", err
	}

	logger.InfoCtx(ctx, "User logged in successfully", "user_id", user.ID, "username", user.Username)
	return token, nil
}
