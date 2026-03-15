package company

import (
	"context"
	"errors"
	"strings"
	"sync"

	"gorm.io/gorm"

	"autostack/internal/repository"
)

type gormCompanyRepository struct {
	db *gorm.DB
}

var companyCreateMu sync.Mutex

const maxCompanyIDCreateRetries = 3

// NewCompanyRepository 创建企业仓储实例
func NewCompanyRepository(db *gorm.DB) CompanyRepository {
	return &gormCompanyRepository{db: db}
}

func (r *gormCompanyRepository) getDB(ctx context.Context) *gorm.DB {
	return repository.GetDB(ctx, r.db)
}

func (r *gormCompanyRepository) FindByID(ctx context.Context, id uint) (*Company, error) {
	var company Company
	if err := r.getDB(ctx).First(&company, id).Error; err != nil {
		return nil, err
	}
	return &company, nil
}

func (r *gormCompanyRepository) Create(ctx context.Context, company *Company) error {
	if company == nil {
		return errors.New("company is nil")
	}

	db := r.getDB(ctx)
	companyCreateMu.Lock()
	defer companyCreateMu.Unlock()

	var lastErr error
	for attempt := 0; attempt < maxCompanyIDCreateRetries; attempt++ {
		nextID, err := r.nextCompanyID(db)
		if err != nil {
			lastErr = err
			continue
		}

		company.ID = nextID
		err = db.Create(company).Error
		if err == nil {
			return nil
		}

		if isDuplicateKeyError(err) {
			lastErr = err
			continue
		}
		return err
	}

	if lastErr != nil {
		return errors.New("创建 company_id 失败，已重试 3 次: " + lastErr.Error())
	}
	return errors.New("创建 company_id 失败，已重试 3 次")
}

func (r *gormCompanyRepository) Update(ctx context.Context, company *Company) error {
	return r.getDB(ctx).Save(company).Error
}

func (r *gormCompanyRepository) UpdateFields(ctx context.Context, id uint, fields map[string]interface{}) error {
	return r.getDB(ctx).Model(&Company{}).Where("id = ?", id).Updates(fields).Error
}

func (r *gormCompanyRepository) List(ctx context.Context) ([]Company, error) {
	var companies []Company
	if err := r.getDB(ctx).Find(&companies).Error; err != nil {
		return nil, err
	}
	return companies, nil
}

func isDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "Duplicate entry") || strings.Contains(msg, "UNIQUE constraint failed")
}

func (r *gormCompanyRepository) nextCompanyID(db *gorm.DB) (uint, error) {
	var maxID uint64
	if err := db.Model(&Company{}).
		Where("id >= ? AND id <= ?", CompanyIDMin, CompanyIDMax).
		Select("COALESCE(MAX(id), 0)").
		Scan(&maxID).Error; err != nil {
		return 0, err
	}

	if maxID < uint64(CompanyIDMin-1) {
		maxID = uint64(CompanyIDMin - 1)
	}

	nextID := maxID + 1
	if nextID > uint64(CompanyIDMax) {
		return 0, errors.New("company_id 已耗尽，请扩展编号规则")
	}

	return uint(nextID), nil
}
