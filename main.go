package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	docs "github.com/epicmet/dekamond-task/docs"
	"github.com/epicmet/dekamond-task/internal/otp"
	ratelimit "github.com/epicmet/dekamond-task/internal/rate-limit"
	"github.com/epicmet/dekamond-task/internal/users"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// TODO:
// Readme & deployment

var OTP_LENGTH = 6
var stateManager = otp.NewMemStateManager(time.Minute * 2)
var otpProvider otp.OTPProvider = otp.NewConsoleOTP(
	stateManager,
	os.Stdout,
	OTP_LENGTH,
)

var usersRepo users.UserRepository

var jwtSecretKey = os.Getenv("JWT_SECRET_KEY")

func generateJWT(phoneNumber string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"phone_number": phoneNumber,
		"exp":          time.Now().Add(time.Hour * 24).Unix(),
	})
	return token.SignedString([]byte(jwtSecretKey))
}

// @Summary		Send OTP
// @Description	Send OTP to phone number
// @Tags			OTP
// @Accept			json
// @Produce		json
// @Param			request	body		object{phone=string}	true	"Phone number"
// @Success		200		{object}	object{message=string}
// @Failure		400		{object}	object{error=string}
// @Failure		500		{object}	object{error=string}
// @Router			/send-otp [post]
func sendOtp(c *gin.Context) {
	var req struct {
		Phone string `json:"phone"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	err := otpProvider.Send(req.Phone)
	if err != nil {
		fmt.Printf("error while sending the otp: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to send otp"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "otp has been sent"})
}

// @Summary		Verify OTP
// @Description	Verify OTP for phone number
// @Tags			OTP
// @Accept			json
// @Produce		json
// @Param			request	body		object{phone=string,otp=string}	true	"Phone and OTP"
// @Success		200		{object}	object{message=string}
// @Failure		400		{object}	object{error=string}
// @Router			/verify-otp [post]
func verifyOtp(c *gin.Context) {
	var req struct {
		Phone string `json:"phone"`
		OTP   string `json:"otp"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	isOtpCorrect := otpProvider.Check(req.Phone, req.OTP)

	if !isOtpCorrect {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid otp"})
		return
	}

	user, err := usersRepo.Upsert(req.Phone)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch data from db"})
		return
	}

	token, err := generateJWT(user.PhoneNumber)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
	}

	c.JSON(http.StatusOK, gin.H{"message": "otp verified successfully", "token": token})
}

// @Summary		Get user by ID
// @Description	Retrieve single user details by ID
// @Tags			Users
// @Accept			json
// @Produce		json
// @Param			id	path		string	true	"User ID"
// @Success		200	{object}	users.User
// @Failure		400	{object}	object{error=string}
// @Failure		404	{object}	object{error=string}
// @Failure		500	{object}	object{error=string}
// @Router			/users/{id} [get]
func getUserByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user id is required"})
		return
	}

	user, err := usersRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, users.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		fmt.Printf("error while fetching user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch user"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// @Summary		Get all users
// @Description	Retrieve list of users with pagination
// @Tags			Users
// @Accept			json
// @Produce		json
// @Param			page		query		int	false	"Page number"	default(1)
// @Param			page_size	query		int	false	"Page size"		default(10)
// @Success		200			{object}	users.PaginatedUsers
// @Failure		400			{object}	object{error=string}
// @Failure		500			{object}	object{error=string}
// @Router			/users [get]
func getUsers(c *gin.Context) {
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid page parameter"})
		return
	}

	pageSize, err := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if err != nil || pageSize < 1 || pageSize > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid page_size parameter (1-100)"})
		return
	}

	result, err := usersRepo.GetAll(page, pageSize)
	if err != nil {
		fmt.Printf("error while fetching users: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch users"})
		return
	}

	c.JSON(http.StatusOK, result)
}

// @Summary		Search users by phone
// @Description	Search users by phone number prefix
// @Tags			Users
// @Accept			json
// @Produce		json
// @Param			phone		query		string	true	"Phone number prefix to search"
// @Param			page		query		int		false	"Page number"	default(1)
// @Param			page_size	query		int		false	"Page size"		default(10)
// @Success		200			{object}	users.PaginatedUsers
// @Failure		400			{object}	object{error=string}
// @Failure		500			{object}	object{error=string}
// @Router			/users/search [get]
func searchUsers(c *gin.Context) {
	phonePrefix := c.Query("phone")
	if phonePrefix == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "phone parameter is required"})
		return
	}

	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid page parameter"})
		return
	}

	pageSize, err := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if err != nil || pageSize < 1 || pageSize > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid page_size parameter (1-100)"})
		return
	}

	result, err := usersRepo.SearchByPhone(phonePrefix, page, pageSize)
	if err != nil {
		fmt.Printf("error while searching users: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to search users"})
		return
	}

	c.JSON(http.StatusOK, result)
}

func main() {
	var err error
	usersRepo, err = users.NewMongoUserRepository("mongodb://localhost:27017", "dekamond-task")
	if err != nil {
		log.Fatal(err.Error())
	}

	r := gin.Default()

	docs.SwaggerInfo.Title = "Dekamond Task"
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	tb := ratelimit.NewTokenBucket("send-otp", 3, time.Minute*10, ratelimit.NewInMemoryStateManager())
	r.POST("/send-otp", tb.GinMiddleware(), sendOtp)
	r.POST("/verify-otp", verifyOtp)

	r.GET("/users/:id", getUserByID)
	r.GET("/users", getUsers)
	r.GET("/users/search", searchUsers)

	r.Run()
}
