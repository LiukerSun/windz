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

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username       string `json:"username" binding:"required,min=3,max=32"`
	Password       string `json:"password" binding:"required,min=6,max=32"`
	Email          string `json:"email" binding:"required,email"`
	OrganizationID uint   `json:"organization_id"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token        string `json:"token"`
	UserID       uint   `json:"user_id"`
	Username     string `json:"username"`
	Role         string `json:"role"`
	Organization string `json:"organization"`
}

// Auth 认证控制器
type Auth struct{}

// Login 用户登录
func (a *Auth) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 查找用户
	var user model.User
	if err := database.DB.Preload("Organization").Where("username = ?", req.Username).First(&user).Error; err != nil {
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

	// 如果不是超级管理员，验证组织是否存在
	if req.OrganizationID != 0 {
		var org model.Organization
		if err := database.DB.First(&org, req.OrganizationID).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "organization not found"})
			return
		}
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
		Organization: "",
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

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6,max=32"`
}

// ChangePassword 修改密码
func (a *Auth) ChangePassword(c *gin.Context) {
	var req ChangePasswordRequest
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
	user := currentUser.(*model.User)

	// 验证旧密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.OldPassword)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "incorrect old password"})
		return
	}

	// 检查新密码是否与旧密码相同
	if req.OldPassword == req.NewPassword {
		c.JSON(http.StatusBadRequest, gin.H{"error": "new password must be different from old password"})
		return
	}

	// 加密新密码
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

	// 生成新的 token
	token, err := jwt.GenerateToken(user.ID, user.Username, user.Role, user.OrganizationID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "password changed successfully",
		"token":   token, // 返回新的 token
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