package controller

import (
	"backend/internal/service"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// CreateOrganizationRequest 创建组织请求
type CreateOrganizationRequest struct {
	Code        string `json:"code" binding:"required"`
	Description string `json:"description"`
}

// UpdateOrganizationRequest 更新组织请求
type UpdateOrganizationRequest struct {
	Code        string `json:"code"`
	Description string `json:"description"`
}

// Organization 组织控制器
type Organization struct {
	orgService *service.OrganizationService
}

// NewOrganization creates a new Organization controller
func NewOrganization() *Organization {
	return &Organization{
		orgService: &service.OrganizationService{},
	}
}

// Create 创建组织
// @Summary      创建组织
// @Description  创建新的组织
// @Tags         organizations
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        request body CreateOrganizationRequest true "组织信息"
// @Success      201  {object}  model.Organization
// @Failure      400  {object}  response.ErrorResponse
// @Failure      401  {object}  response.ErrorResponse
// @Router       /organizations [post]
func (o *Organization) Create(c *gin.Context) {
	var req CreateOrganizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	org, err := o.orgService.Create(req.Code, req.Description)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, org)
}

// List 获取组织列表
// @Summary      获取组织列表
// @Description  获取所有组织的列表
// @Tags         organizations
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Success      200  {array}   model.Organization
// @Failure      401  {object}  response.ErrorResponse
// @Failure      500  {object}  response.ErrorResponse
// @Router       /organizations [get]
func (o *Organization) List(c *gin.Context) {
	orgs, err := o.orgService.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, orgs)
}

// Get 获取单个组织
// @Summary      获取组织详情
// @Description  根据ID获取组织详情
// @Tags         organizations
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        id   path      int  true  "组织ID"
// @Success      200  {object}  model.Organization
// @Failure      400  {object}  response.ErrorResponse
// @Failure      401  {object}  response.ErrorResponse
// @Failure      404  {object}  response.ErrorResponse
// @Router       /organizations/{id} [get]
func (o *Organization) Get(c *gin.Context) {
	id := c.Param("id")
	var orgID uint
	if _, err := fmt.Sscanf(id, "%d", &orgID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid organization id"})
		return
	}

	org, err := o.orgService.Get(orgID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, org)
}

// Update 更新组织
// @Summary      更新组织
// @Description  根据ID更新组织信息
// @Tags         organizations
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        id      path    int                       true  "组织ID"
// @Param        request body    UpdateOrganizationRequest true  "组织信息"
// @Success      200     {object} model.Organization
// @Failure      400     {object} response.ErrorResponse
// @Failure      401     {object} response.ErrorResponse
// @Failure      404     {object} response.ErrorResponse
// @Router       /organizations/{id} [put]
func (o *Organization) Update(c *gin.Context) {
	id := c.Param("id")
	var orgID uint
	if _, err := fmt.Sscanf(id, "%d", &orgID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid organization id"})
		return
	}

	var req UpdateOrganizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	org, err := o.orgService.Update(orgID, req.Code, req.Description)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, org)
}

// Delete 删除组织
// @Summary      删除组织
// @Description  根据ID删除组织
// @Tags         organizations
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        id   path      int  true  "组织ID"
// @Success      200  {object}  response.SuccessResponse
// @Failure      401  {object}  response.ErrorResponse
// @Failure      404  {object}  response.ErrorResponse
// @Router       /organizations/{id} [delete]
func (o *Organization) Delete(c *gin.Context) {
	id := c.Param("id")
	var orgID uint
	if _, err := fmt.Sscanf(id, "%d", &orgID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid organization id"})
		return
	}

	if err := o.orgService.Delete(orgID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Organization deleted successfully"})
}
