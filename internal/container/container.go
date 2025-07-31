package container

import (
	"khoomi-api-io/api/pkg/controllers"
	"khoomi-api-io/api/pkg/services"
)

type ServiceContainer struct {
	ReviewService       services.ReviewService
	CartService         services.CartService
	ShopService         services.ShopService
	UserService         services.UserService
	UserAddressService  services.UserAddressService
	UserFavoriteService services.UserFavoriteService
	VerificationService services.VerificationService
	NotificationService services.NotificationService
	EmailService        services.EmailService
	ShippingService     services.ShippingService
	PaymentService      services.PaymentService
	ListingService      services.ListingService

	ReviewController       *controllers.ReviewController
	CartController         *controllers.CartController
	ShopController         *controllers.ShopController
	UserController         *controllers.UserController
	UserAddressController  *controllers.UserAddressController
	UserFavoriteController *controllers.UserFavoriteController
	VerificationController *controllers.VerificationController
	ShippingController     *controllers.ShippingController
	PaymentController      *controllers.PaymentController
	ListingController      *controllers.ListingController
}

func NewServiceContainer() *ServiceContainer {
	reviewService := services.NewReviewService()
	cartService := services.NewCartService()
	shopService := services.NewShopService()
	userService := services.NewUserService()
	userAddressService := services.NewUserAddressService()
	userFavoriteService := services.NewUserFavoriteService()
	verificationService := services.NewVerificationService()
	notificationService := services.NewNotificationService()
	emailService := services.NewEmailService()
	shippingService := services.NewShippingService()
	paymentService := services.NewPaymentService()
	listingService := services.NewListingService()

	reviewController := controllers.InitReviewController(reviewService, notificationService)
	cartController := controllers.InitCartController(cartService, notificationService)
	shopController := controllers.InitShopController(shopService, notificationService, emailService)
	userController := controllers.InitUserController(userService, notificationService)
	userAddressController := controllers.InitUserAddressController(userAddressService, notificationService)
	userFavoriteController := controllers.InitUserFavoriteController(userFavoriteService, notificationService)
	verificationController := controllers.InitVerificationController(verificationService, notificationService)
	shippingController := controllers.InitShippingController(shippingService, shopService, notificationService)
	paymentController := controllers.InitPaymentController(paymentService, userService, notificationService)
	listingController := controllers.InitListingController(listingService, shopService, notificationService, emailService)

	return &ServiceContainer{
		ReviewService:       reviewService,
		CartService:         cartService,
		ShopService:         shopService,
		UserService:         userService,
		UserAddressService:  userAddressService,
		UserFavoriteService: userFavoriteService,
		VerificationService: verificationService,
		NotificationService: notificationService,
		EmailService:        emailService,
		ShippingService:     shippingService,
		PaymentService:      paymentService,
		ListingService:      listingService,

		ReviewController:       reviewController,
		CartController:         cartController,
		ShopController:         shopController,
		UserController:         userController,
		UserAddressController:  userAddressController,
		UserFavoriteController: userFavoriteController,
		VerificationController: verificationController,
		ShippingController:     shippingController,
		PaymentController:      paymentController,
		ListingController:      listingController,
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

// GetEmailService returns the email service instance
func (sc *ServiceContainer) GetEmailService() services.EmailService {
	return sc.EmailService
}

// GetShopController returns the shop controller instance
func (sc *ServiceContainer) GetShopController() *controllers.ShopController {
	return sc.ShopController
}

// GetUserAddressController returns the user address controller instance
func (sc *ServiceContainer) GetUserAddressController() *controllers.UserAddressController {
	return sc.UserAddressController
}

// GetUserFavoriteController returns the user favorite controller instance
func (sc *ServiceContainer) GetUserFavoriteController() *controllers.UserFavoriteController {
	return sc.UserFavoriteController
}

// GetVerificationController returns the verification controller instance
func (sc *ServiceContainer) GetVerificationController() *controllers.VerificationController {
	return sc.VerificationController
}

// GetUserController returns the user controller instance
func (sc *ServiceContainer) GetUserController() *controllers.UserController {
	return sc.UserController
}

// GetShippingController returns the shipping controller instance
func (sc *ServiceContainer) GetShippingController() *controllers.ShippingController {
	return sc.ShippingController
}

// GetPaymentController returns the payment controller instance
func (sc *ServiceContainer) GetPaymentController() *controllers.PaymentController {
	return sc.PaymentController
}

// GetListingController returns the listing controller instance
func (sc *ServiceContainer) GetListingController() *controllers.ListingController {
	return sc.ListingController
}
