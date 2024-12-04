package model

import (
	"fmt"

	"gorm.io/gorm"
)

const (
	RoleSuperAdmin = "super_admin" // 超级管理员
	RoleOrgAdmin   = "org_admin"   // 组织管理员
	RoleOrgMember  = "org_member"  // 组织成员
)

// User 用户模型
type User struct {
	gorm.Model
	Username       string       `gorm:"size:32;not null" json:"username"`   // 用户名
	Password       string       `gorm:"size:128;not null" json:"-"`         // 密码
	Email          string       `gorm:"size:128" json:"email"`              // 邮箱
	Role           string       `gorm:"size:32;not null" json:"role"`       // 角色
	OrganizationID uint         `gorm:"default:0" json:"organization_id"`   // 组织ID，超级管理员为0
	Organization   Organization `gorm:"foreignKey:OrganizationID" json:"-"` // 所属组织，在json中忽略
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}

// BeforeCreate 创建前的钩子
func (u *User) BeforeCreate(tx *gorm.DB) error {
	// 超级管理员不需要组织ID
	if u.Role == RoleSuperAdmin {
		u.OrganizationID = 0
		return nil
	}

	// 非超级管理员必须属于一个组织
	if u.OrganizationID == 0 {
		return fmt.Errorf("organization_id is required for non-super-admin users")
	}

	// 检查同一组织下用户名和邮箱是否唯一
	var count int64
	tx.Model(&User{}).Where("(username = ? OR email = ?) AND organization_id = ?",
		u.Username, u.Email, u.OrganizationID).Count(&count)
	if count > 0 {
		return fmt.Errorf("username or email already exists in the organization")
	}
	return nil
}
