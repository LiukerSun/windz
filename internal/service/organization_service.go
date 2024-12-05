package service

import (
	"backend/internal/model"
	"backend/pkg/database"
	"errors"
)

type OrganizationService struct{}

// Create 创建新组织
func (s *OrganizationService) Create(code, description string) (*model.Organization, error) {
	// 检查组织代码是否已存在
	var count int64
	database.DB.Model(&model.Organization{}).Where("code = ?", code).Count(&count)
	if count > 0 {
		return nil, errors.New("organization code already exists")
	}

	// 创建组织
	org := model.Organization{
		Code:        code,
		Description: description,
	}

	if err := database.DB.Create(&org).Error; err != nil {
		return nil, errors.New("failed to create organization")
	}

	return &org, nil
}

// List 获取组织列表
func (s *OrganizationService) List() ([]model.Organization, error) {
	var orgs []model.Organization
	if err := database.DB.Preload("Users").Find(&orgs).Error; err != nil {
		return nil, errors.New("failed to fetch organizations")
	}
	return orgs, nil
}

// Get 获取单个组织
func (s *OrganizationService) Get(id uint) (*model.Organization, error) {
	var org model.Organization
	if err := database.DB.First(&org, id).Error; err != nil {
		return nil, errors.New("organization not found")
	}
	return &org, nil
}

// Update 更新组织信息
func (s *OrganizationService) Update(id uint, code, description string) (*model.Organization, error) {
	var org model.Organization
	if err := database.DB.First(&org, id).Error; err != nil {
		return nil, errors.New("organization not found")
	}

	// 如果修改了组织代码，检查新代码是否已存在
	if org.Code != code {
		var count int64
		database.DB.Model(&model.Organization{}).Where("code = ? AND id != ?", code, id).Count(&count)
		if count > 0 {
			return nil, errors.New("organization code already exists")
		}
	}

	// 更新组织信息
	org.Code = code
	org.Description = description

	if err := database.DB.Save(&org).Error; err != nil {
		return nil, errors.New("failed to update organization")
	}

	return &org, nil
}

// Delete 删除组织
func (s *OrganizationService) Delete(id uint) error {
	var org model.Organization
	if err := database.DB.First(&org, id).Error; err != nil {
		return errors.New("organization not found")
	}

	// 检查组织是否有关联的用户
	var userCount int64
	if err := database.DB.Model(&model.User{}).Where("organization_id = ?", id).Count(&userCount).Error; err != nil {
		return errors.New("failed to check organization users")
	}

	if userCount > 0 {
		return errors.New("cannot delete organization with existing users")
	}

	if err := database.DB.Delete(&org).Error; err != nil {
		return errors.New("failed to delete organization")
	}

	return nil
}
