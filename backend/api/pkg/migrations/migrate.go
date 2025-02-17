package main

import (
	"BlessedApi/cmd/db"
	"BlessedApi/internal/models"
	"BlessedApi/internal/models/benefits"
	"BlessedApi/internal/models/benefits/benefit_progress"
	"BlessedApi/internal/models/exchange"
	"BlessedApi/internal/models/fortune_wheel"
	"BlessedApi/internal/models/requirements"
	"BlessedApi/internal/models/requirements/requirement_progress"
	"BlessedApi/internal/models/travepass"
	"BlessedApi/pkg/logger"
	"time"
)

func main() {
	// dropTables()
	// createTables()
	// seedTravepass()
	// seedFortuneWheelBenefits()

	logger.Info("Migrated.")
}

func dropTables() {
	db.DB.Migrator().DropTable(
		&models.User{},
		&models.Deposit{},
		&models.BinaryBet{},
		&models.UserReferral{},
		&models.RouletteX14Bet{},
		&models.RouletteX14GameResult{},
		&models.CrashGameBet{},
		&models.CrashGame{},
		&models.Withdrawal{},

		&exchange.ExchangeBalance{},

		&travepass.TravePassLevel{},
		&travepass.TravePassLevelRequirement{},
		&travepass.TravePassLevelBenefit{},

		&fortune_wheel.FortuneWheelSector{},

		&requirements.Requirement{},
		&requirements.RequirementBinaryOption{},
		&requirements.RequirementClicker{},
		&requirements.RequirementExchange{},
		&requirements.RequirementMiniGame{},
		&requirements.RequirementReplenishment{},
		&requirements.RequirementTurnover{},

		&requirement_progress.RequirementProgress{},
		&requirement_progress.RequirementProgressClicker{},
		&requirement_progress.RequirementProgressBinaryOption{},
		&requirement_progress.RequirementProgressExchange{},
		&requirement_progress.RequirementProgressMiniGame{},
		&requirement_progress.RequirementProgressReplenishment{},
		&requirement_progress.RequirementProgressTurnover{},

		&benefits.Benefit{},
		&benefits.BenefitFortuneWheel{},
		&benefits.BenefitBinaryOption{},
		&benefits.BenefitClicker{},
		&benefits.BenefitCredit{},
		&benefits.BenefitMiniGame{},
		&benefits.BenefitItem{},
		&benefits.BenefitReplenishment{},

		&benefit_progress.BenefitProgress{},
		&benefit_progress.BenefitProgressBinaryOption{},
		&benefit_progress.BenefitProgressClicker{},
		&benefit_progress.BenefitProgressFortuneWheel{},
		&benefit_progress.BenefitProgressMiniGame{},
		&benefit_progress.BenefitProgressReplenishment{},
	)
}

func createTables() {
	db.DB.AutoMigrate(
		&models.User{},
		&models.Deposit{},
		&models.BinaryBet{},
		&models.UserReferral{},
		&models.RouletteX14Bet{},
		&models.RouletteX14GameResult{},
		&models.CrashGameBet{},
		&models.CrashGame{},
		&models.Withdrawal{},

		&exchange.ExchangeBalance{},

		&travepass.TravePassLevel{},
		&travepass.TravePassLevelRequirement{},
		&travepass.TravePassLevelBenefit{},

		&fortune_wheel.FortuneWheelSector{},

		&requirements.Requirement{},
		&requirements.RequirementBinaryOption{},
		&requirements.RequirementClicker{},
		&requirements.RequirementExchange{},
		&requirements.RequirementMiniGame{},
		&requirements.RequirementReplenishment{},
		&requirements.RequirementTurnover{},

		&requirement_progress.RequirementProgress{},
		&requirement_progress.RequirementProgressClicker{},
		&requirement_progress.RequirementProgressBinaryOption{},
		&requirement_progress.RequirementProgressExchange{},
		&requirement_progress.RequirementProgressMiniGame{},
		&requirement_progress.RequirementProgressReplenishment{},
		&requirement_progress.RequirementProgressTurnover{},

		&benefits.Benefit{},
		&benefits.BenefitFortuneWheel{},
		&benefits.BenefitBinaryOption{},
		&benefits.BenefitClicker{},
		&benefits.BenefitCredit{},
		&benefits.BenefitMiniGame{},
		&benefits.BenefitItem{},
		&benefits.BenefitReplenishment{},

		&benefit_progress.BenefitProgress{},
		&benefit_progress.BenefitProgressBinaryOption{},
		&benefit_progress.BenefitProgressClicker{},
		&benefit_progress.BenefitProgressFortuneWheel{},
		&benefit_progress.BenefitProgressMiniGame{},
		&benefit_progress.BenefitProgressReplenishment{},
	)
}

func seedTravepass() {
	applyLevels()
	applyTravePassLevelRequirements()
	applyTravePassLevelBenefits()
}

func applyLevels() {
	db.DB.Exec("INSERT INTO trave_pass_levels (id) VALUES (?)", 0)
	for i := 1; i <= 60; i++ {
		db.DB.Create(&travepass.TravePassLevel{})
	}
}

var day int64 = int64((24 * time.Hour).Seconds())
var hour int64 = int64((time.Hour).Seconds())

func usdToRupee(usd float64) float64 {
	var rupee int
	rupee = int(84 * usd)
	rupee += rupee % 100
	return float64(rupee)
}

func applyTravePassLevelRequirements() {
	createBinaryOptionReq(requirements.RequirementBinaryOption{
		BetsAmount: 5, MinBetRupee: usdToRupee(1)}, 1)
	createReplenishmentReq(requirements.RequirementReplenishment{
		AmountRupee: usdToRupee(10)}, 2)
	createTurnoverReq(requirements.RequirementTurnover{
		AmountRupee: usdToRupee(20), TimeDuration: day}, 3)
	createClickerReq(requirements.RequirementClicker{
		ClicksAmount: 10000}, 4)
	createMiniGameReq(requirements.RequirementMiniGame{
		GameID: requirements.NvutiGameID, BetsAmount: 5, MinBetRupee: usdToRupee(2)}, 5)
	createReplenishmentReq(requirements.RequirementReplenishment{
		AmountRupee: usdToRupee(20)}, 6)
	createBinaryOptionReq(requirements.RequirementBinaryOption{
		WinsAmount: 5}, 7)
	createTurnoverReq(requirements.RequirementTurnover{
		AmountRupee: usdToRupee(75), TimeDuration: day * 2}, 8)
	createMiniGameReq(requirements.RequirementMiniGame{
		BetsAmount: 15}, 9)
	createClickerReq(requirements.RequirementClicker{
		ClicksAmount: 10000}, 10)
	createExchangeReq(requirements.RequirementExchange{
		BCoinsAmount: 1000}, 10) // same level
	// >10
	createBinaryOptionReq(requirements.RequirementBinaryOption{
		BetsAmount: 10, MinBetRupee: usdToRupee(30)}, 11)
	createReplenishmentReq(requirements.RequirementReplenishment{
		AmountRupee: usdToRupee(50)}, 12)
	createTurnoverReq(requirements.RequirementTurnover{
		AmountRupee: usdToRupee(200), TimeDuration: day}, 13)
	createClickerReq(requirements.RequirementClicker{
		ClicksAmount: 10000}, 14)
	createMiniGameReq(requirements.RequirementMiniGame{
		GameID: requirements.NvutiGameID, BetsAmount: 5}, 15)
	createBinaryOptionReq(requirements.RequirementBinaryOption{
		WinsAmount: 3}, 16)
	createReplenishmentReq(requirements.RequirementReplenishment{
		AmountRupee: usdToRupee(50)}, 17)
	createTurnoverReq(requirements.RequirementTurnover{
		AmountRupee: usdToRupee(300), TimeDuration: day * 2}, 18)
	createClickerReq(requirements.RequirementClicker{
		ClicksAmount: 10000}, 19)
	createMiniGameReq(requirements.RequirementMiniGame{
		BetsAmount: 20}, 20)
	// >20
	createBinaryOptionReq(requirements.RequirementBinaryOption{
		TotalWinningsRupee: usdToRupee(15)}, 21)
	createMiniGameReq(requirements.RequirementMiniGame{
		GameID: requirements.RouletteGameID, BetsAmount: 3}, 22)
	createClickerReq(requirements.RequirementClicker{
		HitLimit: true}, 23)
	createClickerReq(requirements.RequirementClicker{
		ClicksAmount: 20000}, 24)
	createMiniGameReq(requirements.RequirementMiniGame{
		GameID: requirements.NvutiGameID, BetsAmount: 5, MinBetRupee: usdToRupee(5)}, 25)
	createReplenishmentReq(requirements.RequirementReplenishment{
		AmountRupee: usdToRupee(40)}, 26)
	createBinaryOptionReq(requirements.RequirementBinaryOption{
		WinsAmount: 10}, 27)
	createTurnoverReq(requirements.RequirementTurnover{
		AmountRupee: usdToRupee(150), TimeDuration: day * 3}, 28)
	createMiniGameReq(requirements.RequirementMiniGame{
		BetsAmount: 30}, 29)
	createClickerReq(requirements.RequirementClicker{
		HitLimit: true}, 30)
	createExchangeReq(requirements.RequirementExchange{
		BCoinsAmount: 1000}, 30) // same level
	// >30
	createBinaryOptionReq(requirements.RequirementBinaryOption{
		BetsAmount: 25, MinBetRupee: usdToRupee(100)}, 31)
	createReplenishmentReq(requirements.RequirementReplenishment{
		AmountRupee: usdToRupee(150)}, 32)
	createTurnoverReq(requirements.RequirementTurnover{
		AmountRupee: usdToRupee(1000), TimeDuration: day * 3}, 33)
	createClickerReq(requirements.RequirementClicker{
		ClicksAmount: 30000}, 34)
	createMiniGameReq(requirements.RequirementMiniGame{
		GameID: requirements.NvutiGameID, BetsAmount: 10}, 35)
	createBinaryOptionReq(requirements.RequirementBinaryOption{
		WinsAmount: 5}, 36)
	createReplenishmentReq(requirements.RequirementReplenishment{
		AmountRupee: usdToRupee(100)}, 37)
	createTurnoverReq(requirements.RequirementTurnover{
		AmountRupee: usdToRupee(1500), TimeDuration: day * 4}, 38)
	createClickerReq(requirements.RequirementClicker{
		ClicksAmount: 30000}, 39)
	createMiniGameReq(requirements.RequirementMiniGame{
		BetsAmount: 30}, 40)
	// >40
	createBinaryOptionReq(requirements.RequirementBinaryOption{
		BetsAmount: 30, MinBetRupee: usdToRupee(150)}, 41)
	createReplenishmentReq(requirements.RequirementReplenishment{
		AmountRupee: usdToRupee(200)}, 42)
	createTurnoverReq(requirements.RequirementTurnover{
		AmountRupee: usdToRupee(2000), TimeDuration: day * 5}, 43)
	createClickerReq(requirements.RequirementClicker{
		ClicksAmount: 50000}, 44)
	createMiniGameReq(requirements.RequirementMiniGame{
		GameID: requirements.NvutiGameID, BetsAmount: 10, MinBetRupee: usdToRupee(10)}, 45)
	createReplenishmentReq(requirements.RequirementReplenishment{
		AmountRupee: usdToRupee(80)}, 46)
	createBinaryOptionReq(requirements.RequirementBinaryOption{
		WinsAmount: 15}, 47)
	createTurnoverReq(requirements.RequirementTurnover{
		AmountRupee: usdToRupee(2500), TimeDuration: day * 5}, 48)
	createMiniGameReq(requirements.RequirementMiniGame{
		BetsAmount: 40}, 49)
	createClickerReq(requirements.RequirementClicker{
		HitLimit: true}, 50)
	createExchangeReq(requirements.RequirementExchange{
		BCoinsAmount: 10000}, 50) // same level
	// >50
	createBinaryOptionReq(requirements.RequirementBinaryOption{
		BetsAmount: 40, MinBetRupee: usdToRupee(2000)}, 51)
	createReplenishmentReq(requirements.RequirementReplenishment{
		AmountRupee: usdToRupee(300)}, 52)
	createTurnoverReq(requirements.RequirementTurnover{
		AmountRupee: usdToRupee(3000), TimeDuration: day * 6}, 53)
	createClickerReq(requirements.RequirementClicker{
		HitLimit: true}, 54)
	createMiniGameReq(requirements.RequirementMiniGame{
		GameID: requirements.NvutiGameID, BetsAmount: 15}, 55)
	createBinaryOptionReq(requirements.RequirementBinaryOption{
		WinsAmount: 10}, 56)
	createReplenishmentReq(requirements.RequirementReplenishment{
		AmountRupee: usdToRupee(1000)}, 57)
	createTurnoverReq(requirements.RequirementTurnover{
		AmountRupee: usdToRupee(4000), TimeDuration: day}, 58)
	createClickerReq(requirements.RequirementClicker{
		HitLimit: true}, 59)
	createTurnoverReq(requirements.RequirementTurnover{
		AmountRupee: usdToRupee(100000), TimeDuration: day * 2}, 60)
}

func createBinaryOptionReq(req requirements.RequirementBinaryOption, level int64) {
	db.DB.Create(&req).Scan(&req)
	requirement := requirements.Requirement{
		PolymorphicRequirementID: req.ID, PolymorphicRequirementType: requirements.RequirementBinaryOptionType}
	db.DB.Create(&requirement).Scan(&requirement)
	db.DB.Create(&travepass.TravePassLevelRequirement{
		TravePassLevelID: level, RequirementID: requirement.ID})
}

func createMiniGameReq(req requirements.RequirementMiniGame, level int64) {
	db.DB.Create(&req).Scan(&req)
	requirement := requirements.Requirement{
		PolymorphicRequirementID: req.ID, PolymorphicRequirementType: requirements.RequirementMiniGameType}
	db.DB.Create(&requirement).Scan(&requirement)
	db.DB.Create(&travepass.TravePassLevelRequirement{
		TravePassLevelID: level, RequirementID: requirement.ID})
}

func createClickerReq(req requirements.RequirementClicker, level int64) {
	db.DB.Create(&req).Scan(&req)
	requirement := requirements.Requirement{
		PolymorphicRequirementID: req.ID, PolymorphicRequirementType: requirements.RequirementClickerType}
	db.DB.Create(&requirement).Scan(&requirement)
	db.DB.Create(&travepass.TravePassLevelRequirement{
		TravePassLevelID: level, RequirementID: requirement.ID})
}

func createExchangeReq(req requirements.RequirementExchange, level int64) {
	db.DB.Create(&req).Scan(&req)
	requirement := requirements.Requirement{
		PolymorphicRequirementID: req.ID, PolymorphicRequirementType: requirements.RequirementExchangeType}
	db.DB.Create(&requirement).Scan(&requirement)
	db.DB.Create(&travepass.TravePassLevelRequirement{
		TravePassLevelID: level, RequirementID: requirement.ID})
}

func createReplenishmentReq(req requirements.RequirementReplenishment, level int64) {
	db.DB.Create(&req).Scan(&req)
	requirement := requirements.Requirement{
		PolymorphicRequirementID: req.ID, PolymorphicRequirementType: requirements.RequirementReplenishmentType}
	db.DB.Create(&requirement).Scan(&requirement)
	db.DB.Create(&travepass.TravePassLevelRequirement{
		TravePassLevelID: level, RequirementID: requirement.ID})
}

func createTurnoverReq(req requirements.RequirementTurnover, level int64) {
	db.DB.Create(&req).Scan(&req)
	requirement := requirements.Requirement{
		PolymorphicRequirementID: req.ID, PolymorphicRequirementType: requirements.RequirementTurnoverType}
	db.DB.Create(&requirement).Scan(&requirement)
	db.DB.Create(&travepass.TravePassLevelRequirement{
		TravePassLevelID: level, RequirementID: requirement.ID})
}

func applyTravePassLevelBenefits() {
	createCreditBenefit(benefits.BenefitCredit{
		BCoinsAmount: 2000}, 1)
	createFortuneWheelBenefit(benefits.BenefitFortuneWheel{
		FreeSpinsAmount: 1}, 2)
	createClickerBenefit(benefits.BenefitClicker{
		BonusMultiplier: 2, TimeDuration: day}, 3)
	createCreditBenefit(benefits.BenefitCredit{
		BCoinsAmount: 2500}, 4)
	createMiniGameBenefit(benefits.BenefitMiniGame{
		GameID: requirements.DiceGameID, FreeBetsAmount: 5, FreeBetDepositRupee: usdToRupee(1)}, 5)
	createClickerBenefit(benefits.BenefitClicker{
		Reset: true}, 6)
	createCreditBenefit(benefits.BenefitCredit{
		BCoinsAmount: 5000}, 7)
	createReplenishmentBenefit(benefits.BenefitReplenishment{
		BonusMultiplier: 0.2, TimeDuration: day}, 8)
	createFortuneWheelBenefit(benefits.BenefitFortuneWheel{
		FreeSpinsAmount: 3}, 9)
	createMiniGameBenefit(benefits.BenefitMiniGame{
		GameID: requirements.NvutiGameID, FreeBetsAmount: 1, FreeBetDepositRupee: usdToRupee(8)}, 10)
	// >10
	createCreditBenefit(benefits.BenefitCredit{
		BCoinsAmount: 3000}, 11)
	createReplenishmentBenefit(benefits.BenefitReplenishment{
		BonusMultiplier: 0.1, TimeDuration: day}, 12)
	createClickerBenefit(benefits.BenefitClicker{
		BonusMultiplier: 2, TimeDuration: day}, 13)
	createCreditBenefit(benefits.BenefitCredit{
		RupeeAmount: usdToRupee(10)}, 14)
	createClickerBenefit(benefits.BenefitClicker{
		Reset: true}, 15)
	createCreditBenefit(benefits.BenefitCredit{
		BCoinsAmount: 3000}, 15) // same level
	createCreditBenefit(benefits.BenefitCredit{
		BCoinsAmount: 4000}, 16)
	createBinaryOptionBenefit(benefits.BenefitBinaryOption{
		FreeBetsAmount: 1, FreeBetDepositRupee: usdToRupee(10)}, 17)
	createFortuneWheelBenefit(benefits.BenefitFortuneWheel{
		FreeSpinsAmount: 5}, 18)
	createCreditBenefit(benefits.BenefitCredit{
		RupeeAmount: usdToRupee(15)}, 19)
	createCreditBenefit(benefits.BenefitCredit{
		BCoinsAmount: 10000}, 20)
	createClickerBenefit(benefits.BenefitClicker{
		BonusMultiplier: 2, TimeDuration: day * 3}, 20) // same level
	// >20
	createCreditBenefit(benefits.BenefitCredit{
		RupeeAmount: usdToRupee(15)}, 21)
	createFortuneWheelBenefit(benefits.BenefitFortuneWheel{
		FreeSpinsAmount: 3}, 22)
	createCreditBenefit(benefits.BenefitCredit{
		BCoinsAmount: 30000}, 23)
	createCreditBenefit(benefits.BenefitCredit{
		RupeeAmount: usdToRupee(20)}, 24)
	createMiniGameBenefit(benefits.BenefitMiniGame{
		GameID: requirements.DiceGameID, FreeBetsAmount: 7, FreeBetDepositRupee: usdToRupee(1.5)}, 25)
	createClickerBenefit(benefits.BenefitClicker{
		Reset: true}, 26)
	createCreditBenefit(benefits.BenefitCredit{
		BCoinsAmount: 12000}, 27)
	createReplenishmentBenefit(benefits.BenefitReplenishment{
		BonusMultiplier: 0.25, TimeDuration: day}, 28)
	createFortuneWheelBenefit(benefits.BenefitFortuneWheel{
		FreeSpinsAmount: 5}, 29)
	createMiniGameBenefit(benefits.BenefitMiniGame{
		GameID: requirements.DiceGameID, FreeBetsAmount: 1, FreeBetDepositRupee: usdToRupee(12)}, 30)
	// >30
	createCreditBenefit(benefits.BenefitCredit{
		BCoinsAmount: 15000}, 31)
	createReplenishmentBenefit(benefits.BenefitReplenishment{
		BonusMultiplier: 0.2, TimeDuration: day}, 32)
	createClickerBenefit(benefits.BenefitClicker{
		BonusMultiplier: 2, TimeDuration: day * 5}, 33)
	createCreditBenefit(benefits.BenefitCredit{
		RupeeAmount: usdToRupee(30)}, 34)
	createClickerBenefit(benefits.BenefitClicker{
		Reset: true}, 35)
	createCreditBenefit(benefits.BenefitCredit{
		BCoinsAmount: 7000}, 35) // same level
	createCreditBenefit(benefits.BenefitCredit{
		BCoinsAmount: 18000}, 36)
	createBinaryOptionBenefit(benefits.BenefitBinaryOption{
		FreeBetsAmount: 1, FreeBetDepositRupee: usdToRupee(20)}, 37)
	createFortuneWheelBenefit(benefits.BenefitFortuneWheel{
		FreeSpinsAmount: 7}, 38)
	createCreditBenefit(benefits.BenefitCredit{
		RupeeAmount: usdToRupee(25)}, 39)
	createCreditBenefit(benefits.BenefitCredit{
		BCoinsAmount: 20000}, 40)
	// >40
	createCreditBenefit(benefits.BenefitCredit{
		BCoinsAmount: 25000}, 41)
	createReplenishmentBenefit(benefits.BenefitReplenishment{
		BonusMultiplier: 0.25, TimeDuration: day}, 42)
	createClickerBenefit(benefits.BenefitClicker{
		BonusMultiplier: 2, TimeDuration: day * 7}, 43)
	createCreditBenefit(benefits.BenefitCredit{
		RupeeAmount: usdToRupee(50)}, 44)
	createMiniGameBenefit(benefits.BenefitMiniGame{
		GameID: requirements.DiceGameID, FreeBetsAmount: 10, FreeBetDepositRupee: usdToRupee(2)}, 45)
	createClickerBenefit(benefits.BenefitClicker{
		Reset: true}, 46)
	createCreditBenefit(benefits.BenefitCredit{
		BCoinsAmount: 40000}, 47)
	createReplenishmentBenefit(benefits.BenefitReplenishment{
		BonusMultiplier: 0.3, TimeDuration: day}, 48)
	createFortuneWheelBenefit(benefits.BenefitFortuneWheel{
		FreeSpinsAmount: 10}, 49)
	createMiniGameBenefit(benefits.BenefitMiniGame{
		GameID: requirements.DiceGameID, FreeBetsAmount: 1, FreeBetDepositRupee: usdToRupee(60)}, 50)
	// >50
	createCreditBenefit(benefits.BenefitCredit{
		BCoinsAmount: 35000}, 51)
	createReplenishmentBenefit(benefits.BenefitReplenishment{
		BonusMultiplier: 0.35, TimeDuration: day}, 52)
	createClickerBenefit(benefits.BenefitClicker{
		BonusMultiplier: 2, TimeDuration: day * 10}, 53)
	createCreditBenefit(benefits.BenefitCredit{
		RupeeAmount: usdToRupee(75)}, 54)
	createClickerBenefit(benefits.BenefitClicker{
		Reset: true}, 55)
	createCreditBenefit(benefits.BenefitCredit{
		BCoinsAmount: 90000}, 55) // same level
	createCreditBenefit(benefits.BenefitCredit{
		BCoinsAmount: 50000}, 56)
	createItemBenefit(benefits.BenefitItem{
		ItemName: "iPhone 15 256gb"}, 57)
	createFortuneWheelBenefit(benefits.BenefitFortuneWheel{
		FreeSpinsAmount: 15}, 58)
	createItemBenefit(benefits.BenefitItem{
		ItemName: "Unlock the chance to win a MacBook Air 13 M2 256GB"}, 59)
	createItemBenefit(benefits.BenefitItem{
		ItemName: "Chevrolet Camaro 2024, valued at 2,700,000 INR"}, 60)
	createCreditBenefit(benefits.BenefitCredit{
		BCoinsAmount: 1000000}, 60) // same level
}

func createFortuneWheelBenefit(ben benefits.BenefitFortuneWheel, level int64) {
	db.DB.Create(&ben).Scan(&ben)
	benefit := benefits.Benefit{
		PolymorphicBenefitID: ben.ID, PolymorphicBenefitType: benefits.BenefitFortuneWheelType}
	db.DB.Create(&benefit).Scan(&benefit)
	db.DB.Create(&travepass.TravePassLevelBenefit{
		TravePassLevelID: level, BenefitID: benefit.ID})
}

func createBinaryOptionBenefit(ben benefits.BenefitBinaryOption, level int64) {
	db.DB.Create(&ben).Scan(&ben)
	benefit := benefits.Benefit{
		PolymorphicBenefitID: ben.ID, PolymorphicBenefitType: benefits.BenefitBinaryOptionType}
	db.DB.Create(&benefit).Scan(&benefit)
	db.DB.Create(&travepass.TravePassLevelBenefit{
		TravePassLevelID: level, BenefitID: benefit.ID})
}

func createClickerBenefit(ben benefits.BenefitClicker, level int64) {
	db.DB.Create(&ben).Scan(&ben)
	benefit := benefits.Benefit{
		PolymorphicBenefitID: ben.ID, PolymorphicBenefitType: benefits.BenefitClickerType}
	db.DB.Create(&benefit).Scan(&benefit)
	db.DB.Create(&travepass.TravePassLevelBenefit{
		TravePassLevelID: level, BenefitID: benefit.ID})
}

func createCreditBenefit(ben benefits.BenefitCredit, level int64) {
	db.DB.Create(&ben).Scan(&ben)
	benefit := benefits.Benefit{
		PolymorphicBenefitID: ben.ID, PolymorphicBenefitType: benefits.BenefitCreditType}
	db.DB.Create(&benefit).Scan(&benefit)
	db.DB.Create(&travepass.TravePassLevelBenefit{
		TravePassLevelID: level, BenefitID: benefit.ID})
}

func createMiniGameBenefit(ben benefits.BenefitMiniGame, level int64) {
	db.DB.Create(&ben).Scan(&ben)
	benefit := benefits.Benefit{
		PolymorphicBenefitID: ben.ID, PolymorphicBenefitType: benefits.BenefitMiniGameType}
	db.DB.Create(&benefit).Scan(&benefit)
	db.DB.Create(&travepass.TravePassLevelBenefit{
		TravePassLevelID: level, BenefitID: benefit.ID})
}

func createItemBenefit(ben benefits.BenefitItem, level int64) {
	db.DB.Create(&ben).Scan(&ben)
	benefit := benefits.Benefit{
		PolymorphicBenefitID: ben.ID, PolymorphicBenefitType: benefits.BenefitItemType}
	db.DB.Create(&benefit).Scan(&benefit)
	db.DB.Create(&travepass.TravePassLevelBenefit{
		TravePassLevelID: level, BenefitID: benefit.ID})
}

func createReplenishmentBenefit(ben benefits.BenefitReplenishment, level int64) {
	db.DB.Create(&ben).Scan(&ben)
	benefit := benefits.Benefit{
		PolymorphicBenefitID: ben.ID, PolymorphicBenefitType: benefits.BenefitReplenishmentType}
	db.DB.Create(&benefit).Scan(&benefit)
	db.DB.Create(&travepass.TravePassLevelBenefit{
		TravePassLevelID: level, BenefitID: benefit.ID})
}

const (
	smallProb float64 = 0.1083
	midProb   float64 = 0.0625
	largeProb float64 = 0.05
)

func seedFortuneWheelBenefits() {
	createFWCreditBenefit(
		benefits.BenefitCredit{BCoinsAmount: 30000},
		midProb, "#800000")
	createFWFortuneWheelBenefit(
		benefits.BenefitFortuneWheel{FreeSpinsAmount: 1},
		smallProb, "#00FF00")
	createFWCreditBenefit(
		benefits.BenefitCredit{BCoinsAmount: 50000},
		largeProb, "#800080")
	createFWReplenishmentBenefit(
		benefits.BenefitReplenishment{
			BonusMultiplier: 0.05, TimeDuration: day},
		smallProb, "#0000FF")
	createFWClickerBenefit(
		benefits.BenefitClicker{
			BonusMultiplier: 2, TimeDuration: hour},
		smallProb, "#FFFF00")
	createFWFortuneWheelBenefit(
		benefits.BenefitFortuneWheel{FreeSpinsAmount: 2},
		midProb, "#008000")
	createFWCreditBenefit(
		benefits.BenefitCredit{BCoinsAmount: 10000},
		smallProb, "#FF0000")
	createFWFortuneWheelBenefit(
		benefits.BenefitFortuneWheel{FreeSpinsAmount: 3},
		largeProb, "#008080")
	createFWClickerBenefit(
		benefits.BenefitClicker{
			Reset: true},
		smallProb, "#FF00FF")
	createFWBinaryOptionBenefit(
		benefits.BenefitBinaryOption{
			FreeBetsAmount: 1, FreeBetDepositRupee: usdToRupee(3)},
		smallProb, "#00FFFF")
	createFWReplenishmentBenefit(
		benefits.BenefitReplenishment{
			BonusMultiplier: 0.1, TimeDuration: day},
		midProb, "#000080")
	// 12. End with a medium-high reward for a satisfying finish
	createFWClickerBenefit(
		benefits.BenefitClicker{
			BonusMultiplier: 2, TimeDuration: day},
		midProb, "#808000")
}

func createFWCreditBenefit(
	benefitCredit benefits.BenefitCredit,
	probability float64, colorHex string) {
	db.DB.Create(&benefitCredit).Scan(&benefitCredit)
	benefit := benefits.Benefit{
		PolymorphicBenefitID:   benefitCredit.ID,
		PolymorphicBenefitType: benefits.BenefitCreditType,
	}
	db.DB.Create(&benefit).Scan(&benefit)
	db.DB.Create(&fortune_wheel.FortuneWheelSector{
		Probability: probability,
		ColorHex:    colorHex,
		BenefitID:   benefit.ID,
	})
}

func createFWFortuneWheelBenefit(
	benefitFW benefits.BenefitFortuneWheel,
	probability float64, colorHex string) {
	db.DB.Create(&benefitFW).Scan(&benefitFW)
	benefit := benefits.Benefit{
		PolymorphicBenefitID:   benefitFW.ID,
		PolymorphicBenefitType: benefits.BenefitFortuneWheelType,
	}
	db.DB.Create(&benefit).Scan(&benefit)
	db.DB.Create(&fortune_wheel.FortuneWheelSector{
		Probability: probability,
		ColorHex:    colorHex,
		BenefitID:   benefit.ID,
	})
}

func createFWReplenishmentBenefit(
	benefitReplenishment benefits.BenefitReplenishment,
	probability float64, colorHex string) {
	db.DB.Create(&benefitReplenishment).Scan(&benefitReplenishment)
	benefit := benefits.Benefit{
		PolymorphicBenefitID:   benefitReplenishment.ID,
		PolymorphicBenefitType: benefits.BenefitReplenishmentType,
	}
	db.DB.Create(&benefit).Scan(&benefit)
	db.DB.Create(&fortune_wheel.FortuneWheelSector{
		Probability: probability,
		ColorHex:    colorHex,
		BenefitID:   benefit.ID,
	})
}

func createFWClickerBenefit(
	benefitClicker benefits.BenefitClicker,
	probability float64, colorHex string) {
	db.DB.Create(&benefitClicker).Scan(&benefitClicker)
	benefit := benefits.Benefit{
		PolymorphicBenefitID:   benefitClicker.ID,
		PolymorphicBenefitType: benefits.BenefitClickerType,
	}
	db.DB.Create(&benefit).Scan(&benefit)
	db.DB.Create(&fortune_wheel.FortuneWheelSector{
		Probability: probability,
		ColorHex:    colorHex,
		BenefitID:   benefit.ID,
	})
}

func createFWBinaryOptionBenefit(
	benefitBO benefits.BenefitBinaryOption,
	probability float64, colorHex string) {
	db.DB.Create(&benefitBO).Scan(&benefitBO)
	benefit := benefits.Benefit{
		PolymorphicBenefitID:   benefitBO.ID,
		PolymorphicBenefitType: benefits.BenefitBinaryOptionType,
	}
	db.DB.Create(&benefit).Scan(&benefit)
	db.DB.Create(&fortune_wheel.FortuneWheelSector{
		Probability: probability,
		ColorHex:    colorHex,
		BenefitID:   benefit.ID,
	})
}
