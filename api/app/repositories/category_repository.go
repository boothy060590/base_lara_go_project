package repositories

import (
	app_core "base_lara_go_project/app/core/go_core"
	"base_lara_go_project/app/models"
	"time"

	"gorm.io/gorm"
)

type CategoryRepository struct {
	model *app_core.BaseModel[models.Category]
}

func NewCategoryRepository(db *gorm.DB, cache app_core.Cache[models.Category]) *CategoryRepository {
	config := app_core.ModelConfig{
		TableName: "categories",
		Traits: app_core.ModelTraits{
			Cacheable:   true,
			SoftDeletes: true,
			Timestamps:  true,
		},
		CacheTTL:    30 * time.Minute,
		CachePrefix: "category",
	}
	return &CategoryRepository{
		model: app_core.NewBaseModel[models.Category](db, cache, config),
	}
}

func (r *CategoryRepository) Find(id uint) (*models.Category, error) {
	return r.model.Find(id)
}

func (r *CategoryRepository) FindAll(page, perPage int) ([]models.Category, int64, error) {
	return r.model.Paginate(page, perPage)
}

func (r *CategoryRepository) Create(category *models.Category) error {
	return r.model.Create(category)
}

func (r *CategoryRepository) Update(category *models.Category) error {
	return r.model.Update(category)
}

func (r *CategoryRepository) Delete(id uint) error {
	return r.model.Delete(id)
}

// Add more CRUD methods as needed...
