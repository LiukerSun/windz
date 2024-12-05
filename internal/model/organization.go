package model

// Organization 组织模型
type Organization struct {
	BaseModel
	Code        string `gorm:"size:32;unique;not null" json:"code" example:"company_a"`     // 组织代码
	Description string `gorm:"size:256" json:"description" example:"A sample organization"` // 组织描述
	Users       []User `gorm:"foreignKey:OrganizationID" json:"users,omitempty"`            // 组织成员
}

// TableName 指定表名
func (Organization) TableName() string {
	return "organizations"
}
