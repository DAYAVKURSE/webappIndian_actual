package app

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	_ "BlessedApi/cmd/db"
	"BlessedApi/internal/middleware"
	"BlessedApi/internal/service"
	"BlessedApi/pkg/binance"
	"BlessedApi/pkg/logger"
	"BlessedApi/pkg/redis"
)

const apiPrefix = "api/"

func Start() {
	gin.DisableConsoleColor()

	router := gin.Default()
	router.Use(middleware.CORSMiddleware())
	router.Use(middleware.BlockBadActorsMiddleware())
	fromTelegram := router.Group("/", middleware.ValidateTelegramInitDataMiddleware())
	authorized := fromTelegram.Group("/", middleware.AuthMiddleware())

	// Initialize Redis and Binance WebSocket services
	redisService := redis.NewRedisService("redis:6379", "")
	logger.Info("Starting Binance WebSocket service...")
	binanceWS := binance.NewBinanceWebsocketService(redisService)
	binanceWS.Start()

	// Binary options WebSocket routes
	apiWebsocketService := service.NewAPIWebsocketServiceBinaryOptions(redisService, binanceWS)
	// fromTelegram
	{
		fromTelegram.GET(apiPrefix+"ws/kline", apiWebsocketService.WebsocketHandler)
		fromTelegram.GET(apiPrefix+"ws/kline/latest", apiWebsocketService.LatestKlineWebsocketHandler)
	}

	// Start the Roulette X14 game loop in a separate goroutine
	go service.SuperviseRouletteX14Game()

	// Start the Crash Game game loop in a separate goroutine
	go service.SuperviseCrashGame()

	// Fortune Wheel WebSocket routes
	service.InitFortuneWheelService(redisService)
	fortuneWheelWebsocketService := service.NewFortuneWheelWebsocketService(redisService)

	// router
	{
		// payment system
		router.POST(apiPrefix+"payments/postback", service.PaymentSystemPostback)
	}

	// fromTelegram
	{
		fromTelegram.GET(apiPrefix+"ws/fortunewheel/live", fortuneWheelWebsocketService.LiveWinsWebsocketHandler)

		// Roulette X14 WebSocket routes
		//fromTelegram.GET(apiPrefix+"ws/roulettex14/live", service.RouletteWebsocketService.LiveRouletteX14WebsocketHandler)

		// Crash Game WebSocket routes
		fromTelegram.GET(apiPrefix+"ws/crashgame/live", service.CrashGameWS.LiveCrashGameWebsocketHandler)

		// auth
		fromTelegram.GET(apiPrefix+"users/auth", service.Auth)
		fromTelegram.POST(apiPrefix+"users/auth/signup", service.SignUp)
	}

	// authorized
	{
		// payment system
		authorized.POST(apiPrefix+"/payments/withdrawal", service.CreateWithdrawal)

		// trave pass
		authorized.GET(apiPrefix+"travepass", service.GetAllTravePassLevelsWithRequirementsAndBenefits)
		authorized.GET(apiPrefix+"travepass/requirements", service.GetNextLevelRequirements)

		// requirements
		authorized.GET(apiPrefix+"requirements/progress", service.GetUserRequirementsProgress)

		// exchange
		authorized.GET(apiPrefix+"users/exchange", service.GetUserExchangeBalance)
		authorized.POST(apiPrefix+"users/exchange", service.ExchangeBcoinsToRupee)

		// benefits
		authorized.GET(apiPrefix+"benefits/progress", service.GetUserBenefitsProgress)

		// users
		authorized.GET(apiPrefix+"users", service.GetUser)

		// referrals
		authorized.GET(apiPrefix+"users/referrals", service.GetUserReferrals)

		// deposits
		authorized.POST(apiPrefix+"payments/create", service.CreatePaymentPageHandler)
		authorized.GET(apiPrefix+"users/deposits", service.GetUserDeposits)

		// clicker
		authorized.POST(apiPrefix+"clicker", service.AddClicks)
		authorized.GET(apiPrefix+"clicker", service.GetUserCurrentBiPerClickCost)

		// fortune wheel
		authorized.GET(apiPrefix+"games/fortunewheel/info", service.GetFortuneWheelInfo)
		authorized.POST(apiPrefix+"games/fortunewheel/spin", service.SpinFortuneWheel)
		authorized.GET(apiPrefix+"games/fortunewheel/spins", service.GetUserFortuneWheelAvailableSpins)
		// authorized.POST(apiPrefix+"games/fortunewheel/add", service.AddSpins)
		authorized.GET(apiPrefix+"games/fortunewheel/wins", fortuneWheelWebsocketService.GetRecentWins)

		// games
		authorized.GET(apiPrefix+"games/benefits", service.GetUserFreeMiniGameBets)

		// nvuti
		authorized.POST(apiPrefix+"games/nvuti/place", service.NvutiPlaceBet)

		// dice
		authorized.POST(apiPrefix+"games/dice/place", service.DicePlaceBet)

		// binary
		authorized.POST(apiPrefix+"games/binary/place", func(c *gin.Context) {
			service.PlaceBinaryBet(c, redisService)
		})
		authorized.GET(apiPrefix+"games/binary/outcome", service.GetUserBetOutcome)
		authorized.GET(apiPrefix+"games/binary/benefits",
			service.GetUserFreeBinaryOptionBets)

		// Roulette X14
		authorized.POST(apiPrefix+"games/roulettex14/place", service.PlaceRouletteX14Bet)
		authorized.GET(apiPrefix+"games/roulettex14/info", service.GetRouletteX14Info)
		authorized.GET(apiPrefix+"games/roulettex14/history", service.GetRouletteX14History)

		// Crash Game
		authorized.POST(apiPrefix+"games/crashgame/place", service.PlaceCrashGameBet)
		authorized.POST(apiPrefix+"games/crashgame/cashout", service.ManualCashout)
		authorized.GET(apiPrefix+"games/crashgame/history", service.CrashGameWS.GetLast50CrashGames)

	}

	srv := &http.Server{
		Addr:    ":8080",
		Handler: router.Handler(),
	}

	go func() {
		// service connections
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("listen: %s\n", err)
		}
	}()

	// HTTPS server configuration
	tlsSrv := &http.Server{
		Addr:    ":8443", // This is the HTTPS port
		Handler: router.Handler(),
	}

	// Paths to SSL certificate and key
	certFile := "./ssl/certificate.crt"
	keyFile := "./ssl/private.key"

	// service HTTPS connections
	if err := tlsSrv.ListenAndServeTLS(certFile, keyFile); err != nil && err != http.ErrServerClosed {
		logger.Fatal("listen: %s\n", err)
	}

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal, 1)
	// kill (no param) default send syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall. SIGKILL but can"t be catch, so don't need add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutdown Server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server Shutdown: %v", err)
	}
	// catching ctx.Done(). timeout of 5 seconds.

	<-ctx.Done()
	logger.Info("timeout of 5 seconds.")
	logger.Info("Server exiting")
}
