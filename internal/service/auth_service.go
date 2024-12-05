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
func (s *AuthService) Login(username string, password string, organizationCode string) (*model.User, string, error) {
	// 先查找组织
	var org model.Organization
	if err := database.DB.Where("code = ?", organizationCode).First(&org).Error; err != nil {
		return nil, "", errors.New("组织不存在")
	}

	var user model.User
	if err := database.DB.Preload("Organization").
		Where("username = ? AND organization_id = ? AND (role = ? OR role = ? OR role = ?)",
			username, org.ID, model.RoleOrgMember, model.RoleOrgAdmin, model.RoleSuperAdmin).
		First(&user).Error; err != nil {
		return nil, "", errors.New("错误的用户名/密码")
	}

	// 打印用户信息以帮助调试
	fmt.Printf("User found: %+v\n", user)

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, "", errors.New("错误的用户名/密码")
	}

	// 生成 token
	token, err := jwt.GenerateToken(user.ID, user.Username, user.Role, user.OrganizationID)
	if err != nil {
		return nil, "", errors.New("生成token失败")
	}

	return &user, token, nil
}

// Register 处理用户注册
func (s *AuthService) Register(username, password, email string, organizationID uint) (*model.User, error) {
	// 检查用户名是否已存在
	var count int64
	database.DB.Model(&model.User{}).Where("username = ? AND organization_id = ?", username, organizationID).Count(&count)
	if count > 0 {
		return nil, errors.New("用户名已存在于此组织")
	}

	// 创建用户
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("密码哈希失败")
	}

	user := model.User{
		Username:       username,
		Password:       string(hashedPassword),
		Email:          email,
		Role:           model.RoleOrgMember,
		OrganizationID: organizationID,
	}

	if err := database.DB.Create(&user).Error; err != nil {
		return nil, errors.New("创建用户失败")
	}

	return &user, nil
}

// ChangePassword 修改用户密码
func (s *AuthService) ChangePassword(userID uint, oldPassword, newPassword string) error {
	var user model.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		return errors.New("用户不存在")
	}

	// 验证旧密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword)); err != nil {
		return errors.New("旧密码错误")
	}

	// 检查新密码是否与旧密码相同
	if oldPassword == newPassword {
		return errors.New("新密码不能与旧密码相同")
	}

	// 生成新密码哈希
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("密码哈希失败")
	}

	// 更新密码
	user.Password = string(hashedPassword)
	if err := database.DB.Save(&user).Error; err != nil {
		return errors.New("密码更新失败")
	}

	return nil
}

// ResetPassword 重置用户密码（管理员功能）
func (s *AuthService) ResetPassword(userID uint, newPassword string) error {
	var user model.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		return errors.New("用户不存在")
	}

	// 生成新密码哈希
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("密码哈希失败")
	}

	// 更新密码
	user.Password = string(hashedPassword)
	if err := database.DB.Save(&user).Error; err != nil {
		return errors.New("密码更新失败")
	}

	return nil
}

// CreateAdmin 创建超级管理员
func (s *AuthService) CreateAdmin(username, password, email string) (*model.User, error) {
	// 检查用户名是否已存在
	var count int64
	database.DB.Model(&model.User{}).Where("username = ? AND role = ?", username, model.RoleSuperAdmin).Count(&count)
	if count > 0 {
		return nil, errors.New("管理员用户名已存在")
	}

	// 创建管理员用户
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("密码哈希失败")
	}

	admin := model.User{
		Username: username,
		Password: string(hashedPassword),
		Email:    email,
		Role:     model.RoleSuperAdmin,
	}

	if err := database.DB.Create(&admin).Error; err != nil {
		return nil, errors.New("管理员创建失败")
	}

	return &admin, nil
}
