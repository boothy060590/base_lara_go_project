package repositories

import (
	app_core "base_lara_go_project/app/core/go_core"
	"base_lara_go_project/app/models"
	"time"

	"gorm.io/gorm"
)

type PermissionRepository struct {
	model *app_core.BaseModel[models.Permission]
}

func NewPermissionRepository(db *gorm.DB, cache app_core.Cache[models.Permission], wsp any, ca any, pgo any) *PermissionRepository {
	config := app_core.ModelConfig{
		TableName: "permissions",
		Traits: app_core.ModelTraits{
			Cacheable:   true,
			SoftDeletes: true,
			Timestamps:  true,
		},
		CacheTTL:    30 * time.Minute,
		CachePrefix: "permission",
	}

	return &PermissionRepository{
		model: app_core.NewBaseModel[models.Permission](db, cache, config, wsp, ca, pgo),
	}
}

func (r *PermissionRepository) Find(id uint) (*models.Permission, error) {
	return r.model.Find(id)
}

func (r *PermissionRepository) FindAll(page, perPage int) ([]models.Permission, int64, error) {
	return r.model.Paginate(page, perPage)
}

func (r *PermissionRepository) Create(permission *models.Permission) error {
	return r.model.Create(permission)
}

func (r *PermissionRepository) Update(permission *models.Permission) error {
	return r.model.Update(permission)
}

func (r *PermissionRepository) Delete(id uint) error {
	return r.model.Delete(id)
}

// Add more CRUD methods as needed...
