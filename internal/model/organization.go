package model

import (
	"gorm.io/gorm"
)

// Organization 组织模型
type Organization struct {
	gorm.Model
	Code        string `gorm:"size:32;not null;unique" json:"code"`    // 组织代码，用户自定义，不可重复
	Description string `gorm:"size:256" json:"description"`            // 组织描述
	Users       []User `gorm:"foreignKey:OrganizationID" json:"users"` // 组织下的用户
}
