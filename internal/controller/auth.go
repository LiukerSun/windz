package controller

import (
	"backend/internal/model"
	"backend/pkg/database"
	"backend/pkg/jwt"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
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

// Auth 认证控制器
type Auth struct{}

// Login 普通用户登录
func (a *Auth) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 查找用户
	var user model.User
	if err := database.DB.Preload("Organization").
		Where("username = ? AND organization_id = ? AND role = ?",
			req.Username, req.OrganizationID, model.RoleOrgMember).
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
func (a *Auth) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 禁止注册到系统组织（ID为1）
	if req.OrganizationID == 1 {
		c.JSON(http.StatusForbidden, gin.H{"error": "registration to system organization is not allowed"})
		return
	}

	// 验证组织是否存在
	var org model.Organization
	if err := database.DB.First(&org, req.OrganizationID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "organization not found"})
		return
	}

	// 检查用户名在组织内是否唯一
	var existingUser model.User
	if err := database.DB.Where("username = ? AND organization_id = ?", req.Username, req.OrganizationID).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username already exists in this organization"})
		return
	}

	// 对密码进行加密
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}

	// 创建用户
	user := model.User{
		Username:       req.Username,
		Password:       string(hashedPassword),
		Email:          req.Email,
		OrganizationID: req.OrganizationID,
		Role:           model.RoleOrgMember, // 默认为组织成员
	}

	if err := database.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 生成 token
	token, err := jwt.GenerateToken(user.ID, user.Username, user.Role, user.OrganizationID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	// 返回用户信息和 token
	c.JSON(http.StatusCreated, LoginResponse{
		Token:        token,
		UserID:       user.ID,
		Username:     user.Username,
		Role:         user.Role,
		Organization: org.Code,
	})
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
func (a *Auth) ChangePassword(c *gin.Context) {
	// 获取当前用户
	currentUser, exists := c.Get("currentUser")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	user := currentUser.(*model.User)

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 验证旧密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.OldPassword)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid old password"})
		return
	}

	// 对新密码进行加密
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}

	// 更新密码
	if err := database.DB.Model(user).Update("password", string(hashedPassword)).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "password changed successfully",
	})
}

// ResetPasswordRequest 重置密码请求（管理员使用）
type ResetPasswordRequest struct {
	UserID      uint   `json:"user_id" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6,max=32"`
}

// ResetPassword 重置用户密码（需要管理员权限）
func (a *Auth) ResetPassword(c *gin.Context) {
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 获取当前用户
	currentUser, exists := c.Get("currentUser")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	admin := currentUser.(*model.User)

	// 获取目标用户
	var user model.User
	if err := database.DB.First(&user, req.UserID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch user"})
		}
		return
	}

	// 检查权限：
	// 1. 超级管理员可以重置任何用户的密码
	// 2. 组织管理员只能重置同组织内的普通成员密码
	if admin.Role != model.RoleSuperAdmin {
		if admin.Role != model.RoleOrgAdmin ||
			admin.OrganizationID != user.OrganizationID ||
			user.Role == model.RoleOrgAdmin {
			c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
			return
		}
	}

	// 加密新密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}

	// 更新密码
	if err := database.DB.Model(&user).Update("password", string(hashedPassword)).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "password reset successfully",
	})
}

// CreateAdmin 创建超级管理员（需要超级管理员权限）
func (a *Auth) CreateAdmin(c *gin.Context) {
	// 获取当前用户
	currentUser, exists := c.Get("currentUser")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// 验证当前用户是否为超级管理员
	user := currentUser.(*model.User)
	if user.Role != model.RoleSuperAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "only super admin can create new admin users"})
		return
	}

	var req CreateAdminRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 检查用户名在系统组织内是否已存在
	var existingUser model.User
	if err := database.DB.Where("username = ? AND organization_id = ?", req.Username, 1).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username already exists in system organization"})
		return
	}

	// 对密码进行加密
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}

	// 创建超级管理员用户
	newAdmin := model.User{
		Username:       req.Username,
		Password:       string(hashedPassword),
		Email:          req.Email,
		Role:           model.RoleSuperAdmin,
		OrganizationID: 1, // 系统组织ID
	}

	if err := database.DB.Create(&newAdmin).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create admin user"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "admin user created successfully",
		"user": gin.H{
			"id":       newAdmin.ID,
			"username": newAdmin.Username,
			"email":    newAdmin.Email,
			"role":     newAdmin.Role,
		},
	})
}
