package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/epicmet/dekamond-task/internal/otp"
	"github.com/epicmet/dekamond-task/internal/users"
	"github.com/gin-gonic/gin"
)

var OTP_LENGTH = 6
var stateManager = otp.NewMemStateManager(time.Minute * 2)
var otpProvider otp.OTPProvider = otp.NewConsoleOTP(
	stateManager,
	os.Stdout,
	OTP_LENGTH,
)

var usersRepo users.UserRepository

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

	// TODO: JWT
	fmt.Printf("user: %v\n", user)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch data from db"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "otp verified successfully"})
}

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

func searchUsers(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "search query is required"})
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

	result, err := usersRepo.Search(query, page, pageSize)
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

	r.POST("/send-otp", sendOtp)
	r.POST("/verify-otp", verifyOtp)

	r.GET("/users/:id", getUserByID)
	r.GET("/users", getUsers)
	r.GET("/users/search", searchUsers)

	r.Run()
}
