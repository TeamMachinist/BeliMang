package user

import (
	"context"
	"fmt"

	"belimang/internal/infrastructure/cache"
	logger "belimang/internal/pkg/logging"

	"golang.org/x/crypto/bcrypt"
)

// UserService handles user business logic
type UserService struct {
	repo  *UserRepository
	cache *cache.RedisCache
}

// NewUserService creates a new UserService
func NewUserService(repo *UserRepository, cache *cache.RedisCache) *UserService {
	return &UserService{repo: repo, cache: cache}
}

// Create creates a new user
func (s *UserService) Create(req *CreateUserRequest) (*UserResponse, error) {
	ctx := context.Background()
	logger.InfoCtx(ctx, "Creating new user", "email", req.Email)

	// Check if user with email already exists
	existingUser, _ := s.repo.GetByEmail(req.Email)
	if existingUser != nil {
		logger.WarnCtx(ctx, "User already exists", "email", req.Email)
		return nil, ErrUserAlreadyExists
	}

	// Hash password
	logger.DebugCtx(ctx, "Hashing password", "email", req.Email)
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		logger.ErrorCtx(ctx, "Failed to hash password", "error", err)
		return nil, err
	}

	// Create user entity
	user := &User{
		Email:    req.Email,
		Name:     req.Name,
		Password: string(hashedPassword),
	}

	// Save to repository
	if err := s.repo.Create(user); err != nil {
		logger.ErrorCtx(ctx, "Failed to create user", "error", err, "email", req.Email)
		return nil, err
	}

	logger.InfoCtx(ctx, "User created successfully", "user_id", user.ID, "email", user.Email)
	return user.ToResponse(), nil
}

// GetAll retrieves all users
func (s *UserService) GetAll(limit, offset int) (*ListUsersResponse, error) {
	ctx := context.Background()
	cacheKey := fmt.Sprintf("users:list:limit_%d:offset_%d", limit, offset)

	// Try to get from cache first
	var cachedResponse ListUsersResponse
	if s.cache != nil {
		err := s.cache.Get(ctx, cacheKey, &cachedResponse)
		if err == nil {
			logger.DebugCtx(ctx, "Cache hit for user list", "cache_key", cacheKey)
			return &cachedResponse, nil
		}
		logger.DebugCtx(ctx, "Cache miss for user list", "cache_key", cacheKey)
	}

	logger.InfoCtx(ctx, "Fetching users from database", "limit", limit, "offset", offset)
	users, total, err := s.repo.GetAll(limit, offset)
	if err != nil {
		logger.ErrorCtx(ctx, "Failed to fetch users", "error", err)
		return nil, err
	}

	userResponses := make([]*UserResponse, len(users))
	for i, user := range users {
		userResponses[i] = user.ToResponse()
	}

	response := &ListUsersResponse{
		Users:  userResponses,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}

	// Cache the response
	if s.cache != nil {
		err = s.cache.Set(ctx, cacheKey, response, cache.FileListTTL)
		if err != nil {
			logger.WarnCtx(ctx, "Failed to cache user list", "error", err, "cache_key", cacheKey)
		} else {
			logger.DebugCtx(ctx, "User list cached successfully", "cache_key", cacheKey, "ttl", cache.FileListTTL)
		}
	}

	return response, nil
}

// GetByID retrieves a user by ID
func (s *UserService) GetByID(id string) (*UserResponse, error) {
	ctx := context.Background()
	cacheKey := fmt.Sprintf(cache.UserProfileKey, id)

	// Try to get from cache first
	var cachedResponse UserResponse
	if s.cache != nil {
		err := s.cache.Get(ctx, cacheKey, &cachedResponse)
		if err == nil {
			logger.DebugCtx(ctx, "Cache hit for user profile", "cache_key", cacheKey, "user_id", id)
			return &cachedResponse, nil
		}
		logger.DebugCtx(ctx, "Cache miss for user profile", "cache_key", cacheKey, "user_id", id)
	}

	logger.InfoCtx(ctx, "Fetching user from database", "user_id", id)
	user, err := s.repo.GetByID(id)
	if err != nil {
		if err == ErrUserNotFound {
			logger.WarnCtx(ctx, "User not found", "user_id", id)
		} else {
			logger.ErrorCtx(ctx, "Failed to fetch user", "error", err, "user_id", id)
		}
		return nil, err
	}

	response := user.ToResponse()

	// Cache the response
	if s.cache != nil {
		err = s.cache.Set(ctx, cacheKey, response, cache.UserProfileTTL)
		if err != nil {
			logger.WarnCtx(ctx, "Failed to cache user profile", "error", err, "cache_key", cacheKey)
		} else {
			logger.DebugCtx(ctx, "User profile cached successfully", "cache_key", cacheKey, "ttl", cache.UserProfileTTL)
		}
	}

	return response, nil
}

// Update updates an existing user
func (s *UserService) Update(id string, req *UpdateUserRequest) (*UserResponse, error) {
	ctx := context.Background()
	logger.InfoCtx(ctx, "Updating user", "user_id", id)

	// Get existing user
	user, err := s.repo.GetByID(id)
	if err != nil {
		if err == ErrUserNotFound {
			logger.WarnCtx(ctx, "User not found for update", "user_id", id)
		} else {
			logger.ErrorCtx(ctx, "Failed to fetch user for update", "error", err, "user_id", id)
		}
		return nil, err
	}

	// Update fields if provided
	if req.Email != "" {
		user.Email = req.Email
	}
	if req.Name != "" {
		user.Name = req.Name
	}

	// Save to repository
	if err := s.repo.Update(user); err != nil {
		logger.ErrorCtx(ctx, "Failed to update user", "error", err, "user_id", id)
		return nil, err
	}

	response := user.ToResponse()

	// Invalidate cache
	if s.cache != nil {
		cacheKey := fmt.Sprintf(cache.UserProfileKey, id)
		err = s.cache.Delete(ctx, cacheKey)
		if err != nil {
			logger.WarnCtx(ctx, "Failed to invalidate user cache", "error", err, "cache_key", cacheKey)
		} else {
			logger.DebugCtx(ctx, "User cache invalidated", "cache_key", cacheKey)
		}

		// Also invalidate user list cache
		// Note: In a real application, you might want to be more selective about this
		// For now, we'll invalidate all user list caches
		// A better approach would be to use cache tags or a more sophisticated invalidation strategy
	}

	logger.InfoCtx(ctx, "User updated successfully", "user_id", user.ID)
	return response, nil
}

// Delete removes a user by ID
func (s *UserService) Delete(id string) error {
	ctx := context.Background()
	logger.InfoCtx(ctx, "Deleting user", "user_id", id)

	err := s.repo.Delete(id)
	if err != nil {
		if err == ErrUserNotFound {
			logger.WarnCtx(ctx, "User not found for deletion", "user_id", id)
		} else {
			logger.ErrorCtx(ctx, "Failed to delete user", "error", err, "user_id", id)
		}
		return err
	}

	// Invalidate cache
	if s.cache != nil {
		cacheKey := fmt.Sprintf(cache.UserProfileKey, id)
		err = s.cache.Delete(ctx, cacheKey)
		if err != nil {
			logger.WarnCtx(ctx, "Failed to invalidate user cache", "error", err, "cache_key", cacheKey)
		} else {
			logger.DebugCtx(ctx, "User cache invalidated", "cache_key", cacheKey)
		}
	}

	logger.InfoCtx(ctx, "User deleted successfully", "user_id", id)
	return nil
}

// Authenticate authenticates a user by email and password
func (s *UserService) Authenticate(email, password string) (*UserResponse, error) {
	ctx := context.Background()
	logger.InfoCtx(ctx, "Authenticating user", "email", email)

	// Get user by email
	user, err := s.repo.GetByEmail(email)
	if err != nil {
		if err == ErrUserNotFound {
			logger.WarnCtx(ctx, "User not found during authentication", "email", email)
		} else {
			logger.ErrorCtx(ctx, "Failed to fetch user during authentication", "error", err, "email", email)
		}
		return nil, err
	}

	// Compare password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		logger.WarnCtx(ctx, "Invalid password during authentication", "email", email)
		return nil, ErrInvalidPassword
	}

	response := user.ToResponse()
	logger.InfoCtx(ctx, "User authenticated successfully", "user_id", user.ID, "email", user.Email)
	return response, nil
}
