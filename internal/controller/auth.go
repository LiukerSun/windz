package controller

import (
	"backend/internal/model"
	"backend/internal/service"
	"backend/pkg/database"
	"backend/pkg/jwt"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// LoginRequest 普通用户登录请求
type LoginRequest struct {
	Username       string `json:"username" binding:"required"`
	Password       string `json:"password" binding:"required"`
	OrganizationID uint   `json:"organization_id" binding:"required"`
}

// AdminLoginRequest 管理员登录请求
type AdminLoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username       string `json:"username" binding:"required,min=3,max=32"`
	Password       string `json:"password" binding:"required,min=6,max=32"`
	Email          string `json:"email" binding:"required,email"`
	OrganizationID uint   `json:"organization_id" binding:"required"` // 设为必填项
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token        string `json:"token"`
	UserID       uint   `json:"user_id"`
	Username     string `json:"username"`
	Role         string `json:"role"`
	Organization string `json:"organization"`
}

// CreateAdminRequest 创建超级管理员请求
type CreateAdminRequest struct {
	Username string `json:"username" binding:"required,min=3,max=32"`
	Password string `json:"password" binding:"required,min=6,max=32"`
	Email    string `json:"email" binding:"required,email"`
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6,max=32"`
}

// ResetPasswordRequest 重置密码请求（管理员使用）
type ResetPasswordRequest struct {
	UserID      uint   `json:"user_id" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6,max=32"`
}

// Auth 认证控制器
type Auth struct {
	authService *service.AuthService
}

// NewAuth creates a new Auth controller
func NewAuth() *Auth {
	return &Auth{
		authService: &service.AuthService{},
	}
}

// Login 普通用户登录
// @Summary      用户登录
// @Description  普通用户登录接口
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body LoginRequest true "登录信息"
// @Success      200  {object}  LoginResponse
// @Failure      400  {object}  response.ErrorResponse
// @Failure      401  {object}  response.ErrorResponse
// @Router       /auth/login [post]
func (a *Auth) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, token, err := a.authService.Login(req.Username, req.Password, req.OrganizationID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, LoginResponse{
		Token:        token,
		UserID:       user.ID,
		Username:     user.Username,
		Role:         user.Role,
		Organization: user.Organization.Code,
	})
}

// AdminLogin 管理员登录
func (a *Auth) AdminLogin(c *gin.Context) {
	var req AdminLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 查找管理员用户
	var user model.User
	var systemOrg model.Organization
	if err := database.DB.Where("code = ?", "system").First(&systemOrg).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "system organization not found"})
		return
	}

	if err := database.DB.Preload("Organization").
		Where("username = ? AND role = ? AND organization_id = ?",
			req.Username, model.RoleSuperAdmin, systemOrg.ID).
		First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid username or password"})
		return
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid username or password"})
		return
	}

	// 生成 token
	token, err := jwt.GenerateToken(user.ID, user.Username, user.Role, user.OrganizationID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	// 返回登录信息
	c.JSON(http.StatusOK, LoginResponse{
		Token:        token,
		UserID:       user.ID,
		Username:     user.Username,
		Role:         user.Role,
		Organization: user.Organization.Code,
	})
}

// Register 用户注册
// @Summary      用户注册
// @Description  注册新用户
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body RegisterRequest true "注册信息"
// @Success      201  {object}  model.User
// @Failure      400  {object}  response.ErrorResponse
// @Router       /auth/register [post]
func (a *Auth) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := a.authService.Register(req.Username, req.Password, req.Email, req.OrganizationID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, user)
}

// GetCurrentUser 获取当前用户信息
func (a *Auth) GetCurrentUser(c *gin.Context) {
	user, exists := c.Get("currentUser")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	currentUser := user.(*model.User)
	var org model.Organization
	if err := database.DB.First(&org, currentUser.OrganizationID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch organization"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id":      currentUser.ID,
		"username":     currentUser.Username,
		"email":        currentUser.Email,
		"role":         currentUser.Role,
		"organization": org.Code,
	})
}

// ChangePassword 用户修改密码
// @Summary      修改密码
// @Description  用户修改自己的密码
// @Tags         auth
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        request body ChangePasswordRequest true "密码修改信息"
// @Success      200  {object}  response.SuccessResponse
// @Failure      400  {object}  response.ErrorResponse
// @Failure      401  {object}  response.ErrorResponse
// @Router       /auth/change-password [post]
func (a *Auth) ChangePassword(c *gin.Context) {
	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetUint("user_id")
	if err := a.authService.ChangePassword(userID, req.OldPassword, req.NewPassword); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "password changed successfully"})
}

// ResetPassword 重置用户密码（需要管理员权限）
// @Summary      重置密码
// @Description  管理员重置用户密码
// @Tags         auth
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        request body ResetPasswordRequest true "密码重置信息"
// @Success      200  {object}  response.SuccessResponse
// @Failure      400  {object}  response.ErrorResponse
// @Failure      401  {object}  response.ErrorResponse
// @Failure      403  {object}  response.ErrorResponse
// @Router       /auth/reset-password [post]
func (a *Auth) ResetPassword(c *gin.Context) {
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := a.authService.ResetPassword(req.UserID, req.NewPassword); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "password reset successfully"})
}

// CreateAdmin 创建超级管理员（需要超级管理员权限）
// @Summary      创建管理员
// @Description  创建新的超级管理员（需要超级管理员权限）
// @Tags         auth
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        request body CreateAdminRequest true "管理员创建信息"
// @Success      201  {object}  model.User
// @Failure      400  {object}  response.ErrorResponse
// @Failure      401  {object}  response.ErrorResponse
// @Failure      403  {object}  response.ErrorResponse
// @Router       /auth/create-admin [post]
func (a *Auth) CreateAdmin(c *gin.Context) {
	var req CreateAdminRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	admin, err := a.authService.CreateAdmin(req.Username, req.Password, req.Email)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, admin)
}
