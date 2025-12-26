package repositories

import (
	"Pay2Go/entities"
	"gorm.io/gorm"
)

// repo = data access layer ที่ติดต่อกับ database โดยตรง
type TransactionRepository interface {
	CreateTransaction(tx *entities.Transaction) error
	GetTransactionByID(id uint) (*entities.Transaction, error)
	UpdateTransaction(tx *entities.Transaction) error
}

// ใช้ struct เพื่อ implement interface ในการติดต่อกับ database
type GormTransactionRepository struct {
	db *gorm.DB
}

// สร้าง instance ของ GormTransactionRepository เพื่อใช้ในที่อื่น
func NewGormTransactionRepository(db *gorm.DB) *GormTransactionRepository {
	return &GormTransactionRepository{db: db}
}

func (r *GormTransactionRepository) CreateTransaction(tx *entities.Transaction) error {
	return r.db.Create(tx).Error
}

func (r *GormTransactionRepository) GetTransactionByID(id uint) (*entities.Transaction, error) {
	var tx entities.Transaction
	if err := r.db.First(&tx, id).Error; err != nil {
		return nil, err
	}
	return &tx, nil
}

func (r *GormTransactionRepository) UpdateTransaction(tx *entities.Transaction) error {
	return r.db.Save(tx).Error
}



