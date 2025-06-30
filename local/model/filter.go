package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BaseFilter contains common filter fields that can be embedded in other filters
type BaseFilter struct {
	Page     int    `query:"page"`
	PageSize int    `query:"page_size"`
	SortBy   string `query:"sort_by"`
	SortDesc bool   `query:"sort_desc"`
}

// DateRangeFilter adds date range filtering capabilities
type DateRangeFilter struct {
	StartDate time.Time `query:"start_date" time_format:"2006-01-02T15:04:05Z07:00"`
	EndDate   time.Time `query:"end_date" time_format:"2006-01-02T15:04:05Z07:00"`
}

// ServerBasedFilter adds server ID filtering capability
type ServerBasedFilter struct {
	ServerID string `param:"id"`
}

// ConfigFilter defines filtering options for Config queries
type ConfigFilter struct {
	BaseFilter
	ServerBasedFilter
	ConfigFile string    `query:"config_file"`
	ChangedAt  time.Time `query:"changed_at" time_format:"2006-01-02T15:04:05Z07:00"`
}

// ApiFilter defines filtering options for Api queries
type ApiFilter struct {
	BaseFilter
	Api string `query:"api"`
}

// MembershipFilter defines filtering options for User queries
type MembershipFilter struct {
	BaseFilter
	Username string `query:"username"`
	RoleName string `query:"role_name"`
	RoleID   string `query:"role_id"`
}

// Pagination returns the offset and limit for database queries
func (f *BaseFilter) Pagination() (offset, limit int) {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.PageSize < 1 {
		f.PageSize = 10 // Default page size
	}
	offset = (f.Page - 1) * f.PageSize
	limit = f.PageSize
	return
}

// GetSorting returns the sort field and direction for database queries
func (f *BaseFilter) GetSorting() (field string, desc bool) {
	if f.SortBy == "" {
		return "id", false // Default sorting
	}
	return f.SortBy, f.SortDesc
}

// IsDateRangeValid checks if both dates are set and start date is before end date
func (f *DateRangeFilter) IsDateRangeValid() bool {
	if f.StartDate.IsZero() || f.EndDate.IsZero() {
		return true // If either date is not set, consider it valid
	}
	return f.StartDate.Before(f.EndDate)
}

// ApplyFilter applies the membership filter to a GORM query
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

// Pagination returns the offset and limit for database queries
func (f *MembershipFilter) Pagination() (offset, limit int) {
	return f.BaseFilter.Pagination()
}

// GetSorting returns the sort field and direction for database queries
func (f *MembershipFilter) GetSorting() (field string, desc bool) {
	return f.BaseFilter.GetSorting()
}
