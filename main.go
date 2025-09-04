package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
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

func main() {
	var err error
	usersRepo, err = users.NewMongoUserRepository("mongodb://localhost:27017", "dekamond-task")
	if err != nil {
		log.Fatal(err.Error())
	}

	r := gin.Default()

	r.POST("/send-otp", sendOtp)
	r.POST("/verify-otp", verifyOtp)

	r.Run()
}
