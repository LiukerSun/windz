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
		return nil, errors.New("组织代码已存在")
	}

	// 创建组织
	org := model.Organization{
		Code:        code,
		Description: description,
	}

	if err := database.DB.Create(&org).Error; err != nil {
		return nil, errors.New("创建组织失败")
	}

	return &org, nil
}

// List 获取组织列表
func (s *OrganizationService) List() ([]model.Organization, error) {
	var orgs []model.Organization
	if err := database.DB.Preload("Users").Find(&orgs).Error; err != nil {
		return nil, errors.New("获取组织列表失败")
	}
	return orgs, nil
}

// Get 获取单个组织
func (s *OrganizationService) Get(id uint) (*model.Organization, error) {
	var org model.Organization
	if err := database.DB.First(&org, id).Error; err != nil {
		return nil, errors.New("组织不存在")
	}
	return &org, nil
}

// Update 更新组织信息
func (s *OrganizationService) Update(id uint, code, description string) (*model.Organization, error) {
	var org model.Organization
	if err := database.DB.First(&org, id).Error; err != nil {
		return nil, errors.New("组织不存在")
	}

	// 如果修改了组织代码，检查新代码是否已存在
	if org.Code != code {
		var count int64
		database.DB.Model(&model.Organization{}).Where("code = ? AND id != ?", code, id).Count(&count)
		if count > 0 {
			return nil, errors.New("组织代码已存在")
		}
	}

	// 更新组织信息
	org.Code = code
	org.Description = description

	if err := database.DB.Save(&org).Error; err != nil {
		return nil, errors.New("更新组织失败")
	}

	return &org, nil
}

// Delete 删除组织
func (s *OrganizationService) Delete(id uint) error {
	var org model.Organization
	if err := database.DB.First(&org, id).Error; err != nil {
		return errors.New("组织不存在")
	}

	// 检查组织是否有关联的用户
	var userCount int64
	if err := database.DB.Model(&model.User{}).Where("organization_id = ?", id).Count(&userCount).Error; err != nil {
		return errors.New("检查组织用户失败")
	}

	if userCount > 0 {
		return errors.New("组织下有用户，不能删除")
	}

	if err := database.DB.Delete(&org).Error; err != nil {
		return errors.New("删除组织失败")
	}

	return nil
}
