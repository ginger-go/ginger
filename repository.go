package ginger

import (
	"github.com/ginger-go/sql"
	"gorm.io/gorm"
)

type BaseRepository[T any] struct{}

func (r *BaseRepository[T]) Save(tx *gorm.DB, entity *T) (*T, error) {
	return sql.Save(tx, entity)
}

func (r *BaseRepository[T]) SaveAll(tx *gorm.DB, entities []T) ([]T, error) {
	return sql.SaveAll(tx, entities)
}

func (r *BaseRepository[T]) Delete(tx *gorm.DB, entity *T) error {
	return sql.Delete(tx, entity)
}

func (r *BaseRepository[T]) DeleteAll(tx *gorm.DB, entities []T) error {
	return sql.DeleteAll(tx, entities)
}

func (r *BaseRepository[T]) DeleteBy(tx *gorm.DB, clause *sql.Clause) error {
	return sql.DeleteAllByClause[T](tx, clause)
}

func (r *BaseRepository[T]) FindOne(tx *gorm.DB, clause *sql.Clause) (*T, error) {
	return sql.FindOne[T](tx, clause)
}

func (r *BaseRepository[T]) FindAll(tx *gorm.DB, clause *sql.Clause) ([]T, error) {
	return sql.FindAll[T](tx, clause)
}

func (r *BaseRepository[T]) FindAllComplex(tx *gorm.DB, clause *sql.Clause, sort *sql.Sort, page *sql.Pagination) ([]T, *sql.Pagination, error) {
	return sql.FindAllComplex[T](tx, clause, sort, page)
}

func (r *BaseRepository[T]) Count(tx *gorm.DB, clause *sql.Clause) (int64, error) {
	return sql.Count[T](tx, clause)
}

func (r *BaseRepository[T]) FindByID(tx *gorm.DB, id uint) (*T, error) {
	return sql.FindOne[T](tx, sql.Eq("id", id))
}
