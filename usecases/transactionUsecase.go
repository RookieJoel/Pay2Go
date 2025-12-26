package usecases

import (
	"Pay2Go/entities"
)
//usecase = service ที่ใช้ติดต่อกับ repository 	อีกที

type TransactionUseCase interface {
	CreateTransaction(tx *entities.Transaction) error
	GetTransactionByID(id uint) (*entities.Transaction, error)
	UpdateTransaction(tx *entities.Transaction) error
}

//ใช้ struct เพื่อ implement interface ในการสร้าง service เพื่อเรียกใช้ repository
type TransactionService struct {
	repo TransactionUseCase
}

// สร้าง instance ของ TransactionService เพื่อใช้ในที่อื่น
func NewTransactionService(repo TransactionUseCase) *TransactionService {
	return &TransactionService{repo: repo}
}

func (s *TransactionService) CreateTransaction(tx *entities.Transaction) error {
	return s.repo.CreateTransaction(tx)
}

func (s *TransactionService) GetTransactionByID(id uint) (*entities.Transaction, error) {
	return s.repo.GetTransactionByID(id)
}

func (s *TransactionService) UpdateTransaction(tx *entities.Transaction) error {
	return s.repo.UpdateTransaction(tx)
}


