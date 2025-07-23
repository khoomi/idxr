package container

import (
	"khoomi-api-io/api/pkg/controllers"
	"khoomi-api-io/api/pkg/services"
)

type ServiceContainer struct {
	ReviewService       services.ReviewService
	CartService         services.CartService
	NotificationService services.NotificationService

	ReviewController *controllers.ReviewController
	CartController   *controllers.CartController
}

func NewServiceContainer() *ServiceContainer {
	reviewService := services.NewReviewService()
	cartService := services.NewCartService()
	notificationService := services.NewNotificationService()

	reviewController := controllers.InitReviewController(reviewService, notificationService)
	cartController := controllers.InitCartController(cartService, notificationService)

	return &ServiceContainer{
		ReviewService:       reviewService,
		CartService:         cartService,
		NotificationService: notificationService,

		ReviewController: reviewController,
		CartController:   cartController,
	}
}

// GetReviewController returns the review controller instance
func (sc *ServiceContainer) GetReviewController() *controllers.ReviewController {
	return sc.ReviewController
}

// GetCartController returns the cart controller instance
func (sc *ServiceContainer) GetCartController() *controllers.CartController {
	return sc.CartController
}

// GetShopController returns the shop controller instance
// func (sc *ServiceContainer) GetShopController() *controllers.ShopController {
// 	return sc.ShopController
// }

// GetUserController returns the user controller instance
// func (sc *ServiceContainer) GetUserController() *controllers.UserController {
// 	return sc.UserController
// }
