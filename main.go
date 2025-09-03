package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/epicmet/dekamond-task/internal/otp"
	"github.com/gin-gonic/gin"
)

var OTP_LENGTH = 6
var stateManager = otp.NewMemStateManager(time.Minute * 2)
var otpProvider otp.OTPProvider = otp.NewConsoleOTP(
	stateManager,
	os.Stdout,
	OTP_LENGTH,
)

func main() {
	r := gin.Default()

	r.POST("/send-otp", func(c *gin.Context) {
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
	})

	r.POST("/verify-otp", func(c *gin.Context) {
		var req struct {
			Phone string `json:"phone"`
			OTP   string `json:"otp"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}

		storedOTP, err := stateManager.Get(req.Phone)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid or expired otp"})
			return
		}

		if storedOTP != req.OTP {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid otp"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "otp verified successfully"})
	})

	r.Run()
}

// package main
//
// import (
// 	"fmt"
// 	"io"
// 	"net/http"
// 	"os"
// 	"time"
//
// 	"github.com/epicmet/dekamond-task/internal/otp"
// 	"github.com/gin-gonic/gin"
// )
//
// var OTP_LENGTH = 6
//
// var otpProvider otp.OTPProvider = otp.NewConsoleOTP(
// 	&otp.MemStateManager{TTL: time.Minute * 2},
// 	os.Stdout,
// 	OTP_LENGTH,
// )
//
// func main() {
// 	r := gin.Default()
// 	// TODO: Routing and http stuff. This is just a place holder
//
// 	r.GET("/send-otp", func(c *gin.Context) {
// 		err := otpProvider.Send("09129377828")
//
// 		if err == nil {
// 			// TODO: Some sort of DTO?
// 			c.JSON(http.StatusOK, gin.H{
// 				"message": "otp has been sent",
// 			})
// 		} else {
// 			fmt.Printf("error while sending the otp: %v", err)
// 			c.JSON(http.StatusInternalServerError, gin.H{
// 				"message": "internal server error",
// 			})
// 		}
// 	})
//
// 	r.Run()
// }
