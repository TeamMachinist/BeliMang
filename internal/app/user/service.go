package user

import (
	"context"
	"time"

	"belimang/internal/infrastructure/cache"
	"belimang/internal/infrastructure/database"
	"belimang/internal/pkg/jwt"
	logger "belimang/internal/pkg/logging"
	"belimang/internal/pkg/utils"

	"github.com/google/uuid"
)

type UserService struct {
	queries         *database.Queries
	cache           *cache.RedisCache
	jwtService      *jwt.JWTService
	passwordService *utils.PasswordService
}

func NewUserService(queries *database.Queries, cache *cache.RedisCache, jwtService *jwt.JWTService, passwordService *utils.PasswordService) *UserService {
	return &UserService{
		queries:         queries,
		cache:           cache,
		jwtService:      jwtService,
		passwordService: passwordService,
	}
}

// Register creates a new user or admin account
func (s *UserService) Register(req *RegisterRequest, role UserRole) (string, error) {
	ctx := context.Background()
	logger.InfoCtx(ctx, "Registering new user", "username", req.Username, "role", role)

	// Check if username already exists (across all user types)
	usernameExists, err := s.queries.CheckUsernameExists(ctx, req.Username)
	if err != nil {
		logger.ErrorCtx(ctx, "Failed to check username existence", "error", err)
		return "", err
	}
	if usernameExists {
		logger.WarnCtx(ctx, "Username already exists", "username", req.Username)
		return "", ErrUsernameExists
	}

	// Check email uniqueness based on role
	emailExists, err := s.queries.CheckEmailExistsForRole(ctx, database.CheckEmailExistsForRoleParams{
		Email: req.Email,
		Role:  database.UserRole(role),
	})
	if err != nil {
		logger.ErrorCtx(ctx, "Failed to check email existence", "error", err)
		return "", err
	}
	if emailExists {
		logger.WarnCtx(ctx, "Email already exists for role", "email", req.Email, "role", role)
		return "", ErrEmailExists
	}

	// Hash password
	logger.DebugCtx(ctx, "Hashing password", "username", req.Username)
	hashedPassword, err := s.passwordService.HashPassword(req.Password)
	if err != nil {
		logger.ErrorCtx(ctx, "Failed to hash password", "error", err)
		return "", err
	}

	// Create user
	user := s.newUser(req.Username, req.Email, string(hashedPassword), role)

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
	token, err := s.jwtService.GenerateToken(createdUser.ID.String(), string(role))
	if err != nil {
		logger.ErrorCtx(ctx, "Failed to generate token", "error", err, "user_id", createdUser.ID)
		return "", err
	}

	logger.InfoCtx(ctx, "User registered successfully", "user_id", createdUser.ID, "username", createdUser.Username, "role", role)
	return token, nil
}

// Login authenticates a user or admin and returns JWT token
func (s *UserService) Login(req *LoginRequest, role UserRole) (string, error) {
	ctx := context.Background()
	logger.InfoCtx(ctx, "Attempting login", "username", req.Username, "role", role)

	user, err := s.queries.GetUserByUsernameAndRole(ctx, database.GetUserByUsernameAndRoleParams{
		Username: req.Username,
		Role:     database.UserRole(role),
	})
	if err != nil {
		logger.WarnCtx(ctx, "User not found during login", "username", req.Username, "role", role)
		return "", ErrInvalidCredentials
	}

	// Compare password
	isVerified := s.passwordService.VerifyPassword(req.Password, user.PasswordHash)
	if !isVerified {
		logger.WarnCtx(ctx, "Invalid password during login", "username", req.Username)
		return "", ErrInvalidCredentials
	}

	// Generate JWT token
	token, err := s.jwtService.GenerateToken(user.ID.String(), string(role))
	if err != nil {
		logger.ErrorCtx(ctx, "Failed to generate token", "error", err, "user_id", user.ID)
		return "", err
	}

	logger.InfoCtx(ctx, "Login successful", "user_id", user.ID, "username", user.Username, "role", role)
	return token, nil
}

func (s *UserService) newUser(username, email, hashedPassword string, role UserRole) *User {
	return &User{
		ID:        uuid.New().String(),
		Username:  username,
		Password:  hashedPassword,
		Email:     email,
		Role:      role,
		CreatedAt: time.Now(),
	}
}
