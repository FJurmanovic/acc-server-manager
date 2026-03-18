package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

type BaseRepository[T any, F any] struct {
	db        *gorm.DB
	modelType T
}

func NewBaseRepository[T any, F any](db *gorm.DB, model T) *BaseRepository[T, F] {
	return &BaseRepository[T, F]{
		db:        db,
		modelType: model,
	}
}

func (r *BaseRepository[T, F]) GetAll(ctx context.Context, filter *F) (*[]T, error) {
	result := new([]T)
	query := r.db.WithContext(ctx).Model(&r.modelType)

	if filterable, ok := any(filter).(Filterable); ok {
		query = filterable.ApplyFilter(query)
	}

	if pageable, ok := any(filter).(Pageable); ok {
		offset, limit := pageable.Pagination()
		query = query.Offset(offset).Limit(limit)
	}

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

func (r *BaseRepository[T, F]) GetByID(ctx context.Context, id interface{}) (*T, error) {
	result := new(T)
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(result).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("error getting record by ID: %w", err)
	}
	return result, nil
}

func (r *BaseRepository[T, F]) Insert(ctx context.Context, model *T) error {
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return fmt.Errorf("error creating record: %w", err)
	}
	return nil
}

func (r *BaseRepository[T, F]) Update(ctx context.Context, model *T) error {
	if err := r.db.WithContext(ctx).Save(model).Error; err != nil {
		return fmt.Errorf("error updating record: %w", err)
	}
	return nil
}

func (r *BaseRepository[T, F]) Delete(ctx context.Context, id interface{}) error {
	if err := r.db.WithContext(ctx).Delete(new(T), id).Error; err != nil {
		return fmt.Errorf("error deleting record: %w", err)
	}
	return nil
}

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

type Filterable interface {
	ApplyFilter(*gorm.DB) *gorm.DB
}

type Pageable interface {
	Pagination() (offset, limit int)
}

type Sortable interface {
	GetSorting() (field string, desc bool)
}
