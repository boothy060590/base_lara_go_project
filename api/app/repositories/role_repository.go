package repositories

import (
	app_core "base_lara_go_project/app/core/go_core"
	"base_lara_go_project/app/models"
	"time"

	"gorm.io/gorm"
)

type RoleRepository struct {
	model *app_core.BaseModel[models.Role]
}

func NewRoleRepository(db *gorm.DB, cache app_core.Cache[models.Role]) *RoleRepository {
	config := app_core.ModelConfig{
		TableName: "roles",
		Traits: app_core.ModelTraits{
			Cacheable:   true,
			SoftDeletes: true,
			Timestamps:  true,
		},
		CacheTTL:    30 * time.Minute,
		CachePrefix: "role",
	}
	return &RoleRepository{
		model: app_core.NewBaseModel[models.Role](db, cache, config),
	}
}

func (r *RoleRepository) Find(id uint) (*models.Role, error) {
	return r.model.Find(id)
}

func (r *RoleRepository) FindByName(name string) (*models.Role, error) {
	roles, err := r.model.Where("name = ?", name).Get()
	if err != nil || len(roles) == 0 {
		return nil, err
	}
	return &roles[0], nil
}

func (r *RoleRepository) FindAll(page, perPage int) ([]models.Role, int64, error) {
	return r.model.Paginate(page, perPage)
}

func (r *RoleRepository) Create(role *models.Role) error {
	return r.model.Create(role)
}

func (r *RoleRepository) Update(role *models.Role) error {
	return r.model.Update(role)
}

func (r *RoleRepository) Delete(id uint) error {
	return r.model.Delete(id)
}
