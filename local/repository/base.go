package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

// BaseRepository provides generic CRUD operations for any model
type BaseRepository[T any, F any] struct {
	db        *gorm.DB
	modelType T
}

// NewBaseRepository creates a new base repository for the given model type
func NewBaseRepository[T any, F any](db *gorm.DB, model T) *BaseRepository[T, F] {
	return &BaseRepository[T, F]{
		db:        db,
		modelType: model,
	}
}

// GetAll retrieves all records based on the filter
func (r *BaseRepository[T, F]) GetAll(ctx context.Context, filter *F) (*[]T, error) {
	result := new([]T)
	query := r.db.WithContext(ctx).Model(&r.modelType)

	// Apply filter conditions if filter implements Filterable
	if filterable, ok := any(filter).(Filterable); ok {
		query = filterable.ApplyFilter(query)
	}

	// Apply pagination if filter implements Pageable
	if pageable, ok := any(filter).(Pageable); ok {
		offset, limit := pageable.Pagination()
		query = query.Offset(offset).Limit(limit)
	}

	// Apply sorting if filter implements Sortable
	if sortable, ok := any(filter).(Sortable); ok {
		field, desc := sortable.GetSorting()
		if desc {
			query = query.Order(field + " DESC")
		} else {
			query = query.Order(field)
		}
	}

	if err := query.Find(result).Error; err != nil {
		return nil, fmt.Errorf("error getting records: %w", err)
	}

	return result, nil
}

// GetByID retrieves a single record by ID
func (r *BaseRepository[T, F]) GetByID(ctx context.Context, id interface{}) (*T, error) {
	result := new(T)
	if err := r.db.WithContext(ctx).First(result, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("error getting record by ID: %w", err)
	}
	return result, nil
}

// Insert creates a new record
func (r *BaseRepository[T, F]) Insert(ctx context.Context, model *T) error {
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return fmt.Errorf("error creating record: %w", err)
	}
	return nil
}

// Update modifies an existing record
func (r *BaseRepository[T, F]) Update(ctx context.Context, model *T) error {
	if err := r.db.WithContext(ctx).Save(model).Error; err != nil {
		return fmt.Errorf("error updating record: %w", err)
	}
	return nil
}

// Delete removes a record by ID
func (r *BaseRepository[T, F]) Delete(ctx context.Context, id interface{}) error {
	if err := r.db.WithContext(ctx).Delete(new(T), id).Error; err != nil {
		return fmt.Errorf("error deleting record: %w", err)
	}
	return nil
}

// Count returns the total number of records matching the filter
func (r *BaseRepository[T, F]) Count(ctx context.Context, filter *F) (int64, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&r.modelType)

	if filterable, ok := any(filter).(Filterable); ok {
		query = filterable.ApplyFilter(query)
	}

	if err := query.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("error counting records: %w", err)
	}

	return count, nil
}

// Interfaces for filter capabilities

type Filterable interface {
	ApplyFilter(*gorm.DB) *gorm.DB
}

type Pageable interface {
	Pagination() (offset, limit int)
}

type Sortable interface {
	GetSorting() (field string, desc bool)
} 