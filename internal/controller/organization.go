package controller

import (
	"backend/internal/model"
	"backend/pkg/database"
	"net/http"

	"github.com/gin-gonic/gin"
)

// CreateOrganizationRequest 创建组织请求
type CreateOrganizationRequest struct {
	Code        string `json:"code" binding:"required"`
	Description string `json:"description"`
}

// Organization 组织控制器
type Organization struct{}

// Create 创建组织
func (o *Organization) Create(c *gin.Context) {
	var req CreateOrganizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 检查组织代码是否已存在
	var count int64
	database.DB.Model(&model.Organization{}).Where("code = ?", req.Code).Count(&count)
	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "organization code already exists"})
		return
	}

	// 创建组织
	org := model.Organization{
		Code:        req.Code,
		Description: req.Description,
	}

	if err := database.DB.Create(&org).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create organization"})
		return
	}

	c.JSON(http.StatusCreated, org)
}

// List 获取组织列表
func (o *Organization) List(c *gin.Context) {
	var orgs []model.Organization
	if err := database.DB.Preload("Users").Find(&orgs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch organizations"})
		return
	}

	c.JSON(http.StatusOK, orgs)
}

// Get 获取单个组织
func (o *Organization) Get(c *gin.Context) {
	id := c.Param("id")
	var org model.Organization

	if err := database.DB.First(&org, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "organization not found"})
		return
	}

	c.JSON(http.StatusOK, org)
}

// Update 更新组织
func (o *Organization) Update(c *gin.Context) {
	id := c.Param("id")
	var req CreateOrganizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 检查组织是否存在
	var org model.Organization
	if err := database.DB.First(&org, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "organization not found"})
		return
	}

	// 如果修改了组织代码，检查新代码是否已存在
	if org.Code != req.Code {
		var count int64
		database.DB.Model(&model.Organization{}).Where("code = ? AND id != ?", req.Code, id).Count(&count)
		if count > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "organization code already exists"})
			return
		}
	}

	// 更新组织
	org.Code = req.Code
	org.Description = req.Description
	if err := database.DB.Save(&org).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update organization"})
		return
	}

	c.JSON(http.StatusOK, org)
}

// Delete 删除组织
func (o *Organization) Delete(c *gin.Context) {
	id := c.Param("id")

	// 检查组织是否存在
	var org model.Organization
	if err := database.DB.First(&org, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "organization not found"})
		return
	}

	// 检查组织下是否还有用户
	var count int64
	database.DB.Model(&model.User{}).Where("organization_id = ?", id).Count(&count)
	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot delete organization with existing users"})
		return
	}

	// 删除组织
	if err := database.DB.Delete(&org).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete organization"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "organization deleted successfully"})
}
