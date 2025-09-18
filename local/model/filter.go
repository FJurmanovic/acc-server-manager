package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BaseFilter struct {
	Page     int    `query:"page"`
	PageSize int    `query:"page_size"`
	SortBy   string `query:"sort_by"`
	SortDesc bool   `query:"sort_desc"`
}

type DateRangeFilter struct {
	StartDate time.Time `query:"start_date" time_format:"2006-01-02T15:04:05Z07:00"`
	EndDate   time.Time `query:"end_date" time_format:"2006-01-02T15:04:05Z07:00"`
}

type ServerBasedFilter struct {
	ServerID string `param:"id"`
}

type ConfigFilter struct {
	BaseFilter
	ServerBasedFilter
	ConfigFile string    `query:"config_file"`
	ChangedAt  time.Time `query:"changed_at" time_format:"2006-01-02T15:04:05Z07:00"`
}

type ServiceControlFilter struct {
	BaseFilter
	ServiceControl string `query:"serviceControl"`
}

type MembershipFilter struct {
	BaseFilter
	Username string `query:"username"`
	RoleName string `query:"role_name"`
	RoleID   string `query:"role_id"`
}

func (f *BaseFilter) Pagination() (offset, limit int) {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.PageSize < 1 {
		f.PageSize = 10
	}
	offset = (f.Page - 1) * f.PageSize
	limit = f.PageSize
	return
}

func (f *BaseFilter) GetSorting() (field string, desc bool) {
	if f.SortBy == "" {
		return "id", false
	}
	return f.SortBy, f.SortDesc
}

func (f *DateRangeFilter) IsDateRangeValid() bool {
	if f.StartDate.IsZero() || f.EndDate.IsZero() {
		return true
	}
	return f.StartDate.Before(f.EndDate)
}

func (f *MembershipFilter) ApplyFilter(query *gorm.DB) *gorm.DB {
	if f.Username != "" {
		query = query.Where("username LIKE ?", "%"+f.Username+"%")
	}
	if f.RoleName != "" {
		query = query.Joins("JOIN roles ON users.role_id = roles.id").Where("roles.name = ?", f.RoleName)
	}
	if f.RoleID != "" {
		if roleUUID, err := uuid.Parse(f.RoleID); err == nil {
			query = query.Where("role_id = ?", roleUUID)
		}
	}
	return query
}

func (f *MembershipFilter) Pagination() (offset, limit int) {
	return f.BaseFilter.Pagination()
}

func (f *MembershipFilter) GetSorting() (field string, desc bool) {
	return f.BaseFilter.GetSorting()
}
