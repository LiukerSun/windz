package service

import (
	"backend/internal/model"
	"backend/pkg/database"
	"backend/pkg/jwt"
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

type AuthService struct{}

// Login 处理用户登录
func (s *AuthService) Login(username string, password string, organizationID uint) (*model.User, string, error) {
	var user model.User
	if err := database.DB.Preload("Organization").
		Where("username = ? AND organization_id = ? AND (role = ? OR role = ? OR role = ?)",
			username, organizationID, model.RoleOrgMember, model.RoleOrgAdmin, model.RoleSuperAdmin).
		First(&user).Error; err != nil {
		return nil, "", errors.New("invalid username or password")
	}

	// 打印用户信息以帮助调试
	fmt.Printf("User found: %+v\n", user)

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, "", errors.New("invalid username or password")
	}

	// 生成 token
	token, err := jwt.GenerateToken(user.ID, user.Username, user.Role, user.OrganizationID)
	if err != nil {
		return nil, "", errors.New("failed to generate token")
	}

	return &user, token, nil
}

// Register 处理用户注册
func (s *AuthService) Register(username, password, email string, organizationID uint) (*model.User, error) {
	// 检查用户名是否已存在
	var count int64
	database.DB.Model(&model.User{}).Where("username = ? AND organization_id = ?", username, organizationID).Count(&count)
	if count > 0 {
		return nil, errors.New("username already exists in this organization")
	}

	// 创建用户
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	user := model.User{
		Username:       username,
		Password:       string(hashedPassword),
		Email:          email,
		Role:           model.RoleOrgMember,
		OrganizationID: organizationID,
	}

	if err := database.DB.Create(&user).Error; err != nil {
		return nil, errors.New("failed to create user")
	}

	return &user, nil
}

// ChangePassword 修改用户密码
func (s *AuthService) ChangePassword(userID uint, oldPassword, newPassword string) error {
	var user model.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		return errors.New("user not found")
	}

	// 验证旧密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword)); err != nil {
		return errors.New("invalid old password")
	}

	// 生成新密码哈希
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("failed to hash password")
	}

	// 更新密码
	user.Password = string(hashedPassword)
	if err := database.DB.Save(&user).Error; err != nil {
		return errors.New("failed to update password")
	}

	return nil
}

// ResetPassword 重置用户密码（管理员功能）
func (s *AuthService) ResetPassword(userID uint, newPassword string) error {
	var user model.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		return errors.New("user not found")
	}

	// 生成新密码哈希
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("failed to hash password")
	}

	// 更新密码
	user.Password = string(hashedPassword)
	if err := database.DB.Save(&user).Error; err != nil {
		return errors.New("failed to reset password")
	}

	return nil
}

// CreateAdmin 创建超级管理员
func (s *AuthService) CreateAdmin(username, password, email string) (*model.User, error) {
	// 检查用户名是否已存在
	var count int64
	database.DB.Model(&model.User{}).Where("username = ? AND role = ?", username, model.RoleSuperAdmin).Count(&count)
	if count > 0 {
		return nil, errors.New("admin username already exists")
	}

	// 创建管理员用户
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	admin := model.User{
		Username: username,
		Password: string(hashedPassword),
		Email:    email,
		Role:     model.RoleSuperAdmin,
	}

	if err := database.DB.Create(&admin).Error; err != nil {
		return nil, errors.New("failed to create admin")
	}

	return &admin, nil
}
