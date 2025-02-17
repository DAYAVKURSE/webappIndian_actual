package models

type UserReferral struct {
	ID                     int64 `gorm:"primaryKey,autoIncrement"`
	ReferrerID             int64 `gorm:"index"`
	ReferredID             int64 `gorm:"index"`
	ReferredNickname       string
	ReferredFirstDepositID *int64   `gorm:"index"`
	ReferredFirstDeposit   *Deposit `gorm:"foreignKey:ReferredFirstDepositID"`
	EarnedAmount           float64
}
