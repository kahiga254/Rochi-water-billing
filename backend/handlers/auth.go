package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"waterbilling/backend/models"
	"waterbilling/backend/services"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	userService *services.UserService
	jwtService  *services.JWTService
}

func NewAuthHandler(userService *services.UserService, jwtService *services.JWTService) *AuthHandler {
	return &AuthHandler{
		userService: userService,
		jwtService:  jwtService,
	}
}

// Login handles user authentication
// @Summary User login
// @Description Authenticate user and return JWT token
// @Tags Authentication
// @Accept json
// @Produce json
// @Param credentials body LoginRequest true "Login credentials"
// @Success 200 {object} Response "Login successful"
// @Failure 400 {object} Response "Invalid credentials"
// @Failure 401 {object} Response "Unauthorized"
// @Failure 500 {object} Response "Internal server error"
// @Router /auth/login [post]
// Login handles user authentication
// Login handles user authentication
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "Invalid login data", err)
		return
	}

	if req.Username == "" || req.Password == "" {
		BadRequest(c, "Username and password are required", nil)
		return
	}

	// Get user by username
	user, err := h.userService.GetUserByUsername(req.Username)
	if err != nil {
		// Check if it's a "not found" error
		if err.Error() == "mongo: no documents in result" ||
			err.Error() == "user not found" {
			Unauthorized(c, "Invalid credentials")
			return
		}
		InternalServerError(c, "Database error", err)
		return
	}

	// IMPORTANT: Check if user is nil
	if user == nil {
		Unauthorized(c, "Invalid credentials")
		return
	}

	// Check if user is active
	if !user.IsActive {
		Forbidden(c, "Account is deactivated")
		return
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		Unauthorized(c, "Invalid credentials")
		return
	}

	// Update last login
	now := time.Now()
	user.LastLogin = &now
	if err := h.userService.UpdateUser(user.ID.Hex(), map[string]interface{}{
		"last_login": now,
	}); err != nil {
		// Log error but continue with login
		fmt.Printf("Failed to update last login: %v\n", err)
	}

	// Generate JWT token
	token, err := h.jwtService.GenerateToken(user)
	if err != nil {
		InternalServerError(c, "Failed to generate token", err)
		return
	}

	// Return user info (excluding password) and token
	userResponse := UserResponse{
		ID:          user.ID.Hex(),
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		Email:       user.Email,
		Username:    user.Username,
		PhoneNumber: user.PhoneNumber,
		Role:        user.Role,
		EmployeeID:  user.EmployeeID,
		Department:  user.Department,
		Zone:        user.AssignedZone,
		IsActive:    user.IsActive,
		LastLogin:   user.LastLogin,
		CreatedAt:   user.CreatedAt,
	}

	response := gin.H{
		"user":  userResponse,
		"token": token,
	}

	SuccessResponse(c, "Login successful", response)
}

// Register handles new user registration
// @Summary Register new user
// @Description Register a new system user (admin only)
// @Tags Authentication
// @Accept json
// @Produce json
// @Param user body RegisterRequest true "User registration data"
// @Success 201 {object} Response "User registered successfully"
// @Failure 400 {object} Response "Invalid input"
// @Failure 403 {object} Response "Forbidden"
// @Failure 409 {object} Response "User already exists"
// @Failure 500 {object} Response "Internal server error"
// @Router /auth/register [post]
// Register handles new user registration
func (h *AuthHandler) Register(c *gin.Context) {
	// Check if user is admin (from JWT middleware)
	userRole, exists := c.Get("userRole")
	if !exists || userRole != "admin" {
		Forbidden(c, "Only administrators can register new users")
		return
	}

	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "Invalid registration data", err)
		return
	}

	// Validate required fields
	if req.FirstName == "" || req.LastName == "" {
		BadRequest(c, "First name and last name are required", nil)
		return
	}

	if req.Username == "" {
		BadRequest(c, "Username is required", nil)
		return
	}

	if req.Email == "" {
		BadRequest(c, "Email is required", nil)
		return
	}

	if req.Password == "" {
		BadRequest(c, "Password is required", nil)
		return
	}

	if req.Role == "" {
		req.Role = "reader" // Default role
	}

	// Validate role
	validRoles := map[string]bool{
		"admin":            true,
		"reader":           true,
		"cashier":          true,
		"manager":          true,
		"customer_service": true,
	}

	if !validRoles[req.Role] {
		BadRequest(c, "Invalid role", nil)
		return
	}

	// Create user model
	now := time.Now()
	user := &models.User{
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Email:        req.Email,
		Username:     req.Username,
		PhoneNumber:  req.PhoneNumber,
		Role:         req.Role,
		EmployeeID:   req.EmployeeID,
		Department:   req.Department,
		AssignedZone: req.Zone,
		IsActive:     true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	// Create user
	if err := h.userService.CreateUser(user, req.Password); err != nil {
		if err.Error() == "user with username "+req.Username+" already exists" {
			ErrorResponse(c, http.StatusConflict, "User already exists", err)
		} else if err.Error() == "user with email "+req.Email+" already exists" {
			ErrorResponse(c, http.StatusConflict, "Email already registered", err)
		} else {
			InternalServerError(c, "Failed to register user", err)
		}
		return
	}

	// ✅ FIXED: Return ALL user fields
	userResponse := UserResponse{
		ID:          user.ID.Hex(),
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		Email:       user.Email,
		Username:    user.Username,
		PhoneNumber: user.PhoneNumber,
		Role:        user.Role,
		EmployeeID:  user.EmployeeID,
		Department:  user.Department,
		Zone:        user.AssignedZone,
		IsActive:    user.IsActive,
		CreatedAt:   user.CreatedAt,
		LastLogin:   user.LastLogin,
	}

	CreatedResponse(c, "User registered successfully", userResponse)
}

// GetProfile gets current user profile
// @Summary Get user profile
// @Description Get current authenticated user's profile
// @Tags Authentication
// @Accept json
// @Produce json
// @Success 200 {object} Response "Profile retrieved"
// @Failure 401 {object} Response "Unauthorized"
// @Failure 500 {object} Response "Internal server error"
// @Router /auth/profile [get]
func (h *AuthHandler) GetProfile(c *gin.Context) {
	// Get user ID from context (set by JWT middleware)
	userID, exists := c.Get("userID")
	if !exists {
		Unauthorized(c, "Not authenticated")
		return
	}

	user, err := h.userService.GetUserByID(userID.(string))
	if err != nil {
		Unauthorized(c, "User not found")
		return
	}

	userResponse := UserResponse{
		ID:          user.ID.Hex(),
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		Email:       user.Email,
		Username:    user.Username,
		PhoneNumber: user.PhoneNumber,
		Role:        user.Role,
		EmployeeID:  user.EmployeeID,
		Department:  user.Department,
		Zone:        user.AssignedZone,
		IsActive:    user.IsActive,
		LastLogin:   user.LastLogin,
		CreatedAt:   user.CreatedAt,
	}

	SuccessResponse(c, "Profile retrieved", userResponse)
}

// UpdateProfile updates current user profile
// @Summary Update user profile
// @Description Update current authenticated user's profile
// @Tags Authentication
// @Accept json
// @Produce json
// @Param updates body UpdateProfileRequest true "Profile updates"
// @Success 200 {object} Response "Profile updated successfully"
// @Failure 400 {object} Response "Invalid input"
// @Failure 401 {object} Response "Unauthorized"
// @Failure 500 {object} Response "Internal server error"
// @Router /auth/profile [put]
func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		Unauthorized(c, "Not authenticated")
		return
	}

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "Invalid profile data", err)
		return
	}

	updates := make(map[string]interface{})

	if req.FirstName != "" {
		updates["first_name"] = req.FirstName
	}

	if req.LastName != "" {
		updates["last_name"] = req.LastName
	}

	if req.Email != "" {
		updates["email"] = req.Email
	}

	if req.PhoneNumber != "" {
		updates["phone_number"] = req.PhoneNumber
	}

	// Add updated_at
	updates["updated_at"] = time.Now()

	if err := h.userService.UpdateUser(userID.(string), updates); err != nil {
		if err.Error() == "user not found" {
			Unauthorized(c, "User not found")
		} else if strings.Contains(err.Error(), "already exists") {
			ErrorResponse(c, http.StatusConflict, "Email already registered", err)
		} else {
			InternalServerError(c, "Failed to update profile", err)
		}
		return
	}

	SuccessResponse(c, "Profile updated successfully", nil)
}

// ChangePassword changes user password
// @Summary Change password
// @Description Change current user's password
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body ChangePasswordRequest true "Password change request"
// @Success 200 {object} Response "Password changed successfully"
// @Failure 400 {object} Response "Invalid input"
// @Failure 401 {object} Response "Unauthorized"
// @Failure 500 {object} Response "Internal server error"
// @Router /auth/change-password [post]
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		Unauthorized(c, "Not authenticated")
		return
	}

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "Invalid password data", err)
		return
	}

	if req.CurrentPassword == "" || req.NewPassword == "" {
		BadRequest(c, "Current password and new password are required", nil)
		return
	}

	if req.NewPassword != req.ConfirmPassword {
		BadRequest(c, "New password and confirmation do not match", nil)
		return
	}

	if len(req.NewPassword) < 8 {
		BadRequest(c, "New password must be at least 8 characters", nil)
		return
	}

	// Verify current password
	user, err := h.userService.GetUserByID(userID.(string))
	if err != nil {
		Unauthorized(c, "User not found")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.CurrentPassword)); err != nil {
		BadRequest(c, "Current password is incorrect", nil)
		return
	}

	// Update password
	if err := h.userService.ChangePassword(userID.(string), req.NewPassword); err != nil {
		InternalServerError(c, "Failed to change password", err)
		return
	}

	SuccessResponse(c, "Password changed successfully", nil)
}

// RefreshToken refreshes JWT token
// @Summary Refresh JWT token
// @Description Refresh expired JWT token
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body RefreshTokenRequest true "Refresh token request"
// @Success 200 {object} Response "Token refreshed successfully"
// @Failure 400 {object} Response "Invalid token"
// @Failure 401 {object} Response "Unauthorized"
// @Router /auth/refresh-token [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "Invalid request", err)
		return
	}

	if req.RefreshToken == "" {
		BadRequest(c, "Refresh token is required", nil)
		return
	}

	// Validate and refresh token
	token, err := h.jwtService.RefreshToken(req.RefreshToken)
	if err != nil {
		Unauthorized(c, "Invalid or expired refresh token")
		return
	}

	response := gin.H{
		"token": token,
	}

	SuccessResponse(c, "Token refreshed successfully", response)
}

// Logout handles user logout
// @Summary User logout
// @Description Logout user (client should discard token)
// @Tags Authentication
// @Accept json
// @Produce json
// @Success 200 {object} Response "Logout successful"
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	// In JWT, logout is handled client-side by discarding the token
	// We could implement token blacklisting if needed
	SuccessResponse(c, "Logout successful", nil)
}

// GetUsers returns all users (admin only)
// GetUsers returns all users (admin only)
func (h *AuthHandler) GetUsers(c *gin.Context) {
	// Parse query parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	role := c.Query("role")

	// Build filter
	filter := bson.M{}
	if role != "" {
		filter["role"] = role
	}

	// Get users from service (returns 3 values)
	users, total, err := h.userService.ListUsers(filter, int64(page), int64(limit))
	if err != nil {
		InternalServerError(c, "Failed to fetch users", err)
		return
	}

	// Calculate total pages
	totalPages := (total + int64(limit) - 1) / int64(limit)

	SuccessResponse(c, "Users retrieved successfully", gin.H{
		"users":       users,
		"total":       total,
		"page":        page,
		"limit":       limit,
		"total_pages": totalPages,
	})
}

// Request/Response DTOs

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type RegisterRequest struct {
	FirstName   string `json:"first_name" binding:"required"`
	LastName    string `json:"last_name" binding:"required"`
	Email       string `json:"email" binding:"required"`
	Username    string `json:"username" binding:"required"`
	Password    string `json:"password" binding:"required"`
	PhoneNumber string `json:"phone_number,omitempty"`
	Role        string `json:"role"`
	EmployeeID  string `json:"employee_id,omitempty"`
	Department  string `json:"department,omitempty"`
	Zone        string `json:"zone,omitempty"`
}

type UpdateProfileRequest struct {
	FirstName   string `json:"first_name,omitempty"`
	LastName    string `json:"last_name,omitempty"`
	Email       string `json:"email,omitempty"`
	PhoneNumber string `json:"phone_number,omitempty"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required"`
	ConfirmPassword string `json:"confirm_password" binding:"required"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type UserResponse struct {
	ID          string     `json:"id"`
	FirstName   string     `json:"first_name"`
	LastName    string     `json:"last_name"`
	Email       string     `json:"email"`
	Username    string     `json:"username"`
	PhoneNumber string     `json:"phone_number,omitempty"`
	Role        string     `json:"role"`
	EmployeeID  string     `json:"employee_id,omitempty"`
	Department  string     `json:"department,omitempty"`
	Zone        string     `json:"zone,omitempty"`
	IsActive    bool       `json:"is_active"`
	LastLogin   *time.Time `json:"last_login,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}
