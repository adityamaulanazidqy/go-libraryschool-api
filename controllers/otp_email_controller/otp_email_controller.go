package otp_email_controller

import (
	context2 "context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"go-libraryschool/helpers"
	"go-libraryschool/models/request_models"
	"go-libraryschool/repository/otp_email_repository"
	"golang.org/x/net/context"
	"gopkg.in/gomail.v2"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"
)

type OtpEmailController struct {
	Db        *sql.DB
	logLogrus *logrus.Logger
	rdb       *redis.Client

	otpEmailRepo *otp_email_repository.OtpEmailRepository
}

var (
	smtpUser string
	smtpPass string
)

func NewOtpEmailController(db *sql.DB, logLogrus *logrus.Logger, rdb *redis.Client) *OtpEmailController {
	return &OtpEmailController{db, logLogrus, rdb, otp_email_repository.NewOtpEmailRepository(db, logLogrus)}
}

func SetOtpEmail() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	smtpUser = os.Getenv("SMTP_USER")
	smtpPass = os.Getenv("SMTP_PASSWORD")
}

func (controller *OtpEmailController) generateOtpEmail() string {
	rand.Seed(time.Now().UnixNano())
	return fmt.Sprintf("%06d", rand.Intn(1000000)) // OTP 6 digit
}

func (controller *OtpEmailController) sendEmail(to, otp string) (helpers.ApiResponse, int, error) {
	m := gomail.NewMessage()
	m.SetHeader("From", "adityamaullana234@gmail.com")
	m.SetHeader("To", to)
	m.SetHeader("Subject", "Kode OTP - Go-libraryschool")

	htmlBody := fmt.Sprintf(`
		<!DOCTYPE html>
		<html lang="id">
		<head>
			<meta charset="UTF-8">
			<title>Kode OTP Anda</title>
			<style>
				body {
					font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
					background-color: #f9fafb;
					margin: 0;
					padding: 0;
					color: #333;
				}
				.container {
					max-width: 600px;
					margin: 40px auto;
					background-color: #ffffff;
					padding: 30px;
					border-radius: 10px;
					box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
				}
				.header {
					text-align: center;
					border-bottom: 1px solid #eeeeee;
					padding-bottom: 20px;
					margin-bottom: 20px;
				}
				.header h1 {
					color: #2563eb;
					font-size: 24px;
					margin: 0;
				}
				.content p {
					font-size: 16px;
					line-height: 1.6;
				}
				.otp {
					font-size: 28px;
					font-weight: bold;
					color: #10b981;
					text-align: center;
					margin: 20px 0;
				}
				.footer {
					text-align: center;
					font-size: 12px;
					color: #999999;
					margin-top: 30px;
				}
			</style>
		</head>
		<body>
			<div class="container">
				<div class="header">
					<h1>Go-libraryschool</h1>
				</div>
				<div class="content">
					<p>Halo,</p>
					<p>Berikut adalah <strong>Kode OTP</strong> Anda untuk melanjutkan proses verifikasi akun di <strong>Go-libraryschool</strong>:</p>
					<div class="otp">%s</div>
					<p>Jangan berikan kode ini kepada siapa pun. Kode ini hanya berlaku untuk beberapa menit ke depan.</p>
				</div>
				<div class="footer">
					<p>Jika Anda tidak meminta kode ini, Anda bisa mengabaikan email ini.</p>
					<p>&copy; 2025 Go-libraryschool</p>
				</div>
			</div>
		</body>
		</html>
	`, otp)

	m.SetBody("text/html", htmlBody)

	d := gomail.NewDialer("smtp.gmail.com", 587, smtpUser, smtpPass)

	if err := d.DialAndSend(m); err != nil {
		fmt.Println("Error:", err)
		fmt.Println("OTP:", otp, "To:", to)
		return helpers.ApiResponse{Message: "Failed to send email", Data: nil}, http.StatusInternalServerError, err
	}

	return helpers.ApiResponse{Message: "Successfully sent email"}, http.StatusOK, nil
}

// OtpEmail godoc
// @Summary Otp Email
// @Description Send otp to email
// @Tags Otp
// @Accept json
// @Produce json
// @Param request body request_models.RequestOtpEmail true "User Email"
// @Success 200 {object} helpers.ApiResponse
// @Success 400 {object} helpers.ApiResponse
// @Success 409 {object} helpers.ApiResponse
// @Failure 500 {object} helpers.ApiResponse
// @Router /otp/send-otp [post]
func (controller *OtpEmailController) OtpEmail(w http.ResponseWriter, r *http.Request) {
	var request request_models.RequestOtpEmail

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		controller.logLogrus.WithFields(logrus.Fields{
			"error":   err,
			"message": "Failed to parse request body",
		}).Error("Failed to parse request body")

		helpers.SendJson(w, http.StatusBadRequest, helpers.ApiResponse{
			Message: "Failed to parse request body",
			Data:    nil,
		})
		return
	}

	responseRepo, code, err := controller.otpEmailRepo.VerificationEmail(request.Email)
	if err != nil {
		helpers.SendJson(w, code, responseRepo)
		return
	}

	var otp = controller.generateOtpEmail()
	otpJson, err := json.Marshal(otp)
	if err != nil {
		controller.logLogrus.WithFields(logrus.Fields{
			"error":   err,
			"message": "Failed to marshal otp email",
		}).Error("Failed to marshal otp email")

		helpers.SendJson(w, http.StatusInternalServerError, helpers.ApiResponse{
			Message: "Failed to marshal otp email",
			Data:    nil,
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()

	err = controller.rdb.Set(ctx, fmt.Sprintf("Otp %s: ", request.Email), otpJson, 2*time.Minute).Err()
	if err != nil {
		controller.logLogrus.WithFields(logrus.Fields{
			"error":   err,
			"message": "Failed to set otp store",
		}).Error("Failed to set otp store")

		helpers.SendJson(w, http.StatusInternalServerError, helpers.ApiResponse{
			Message: "Failed to set otp store",
			Data:    nil,
		})
		return
	}

	responseRepo, code, err = controller.sendEmail(request.Email, otp)
	if err != nil {
		helpers.SendJson(w, code, responseRepo)
		return
	}

	helpers.SendJson(w, code, responseRepo)
}

// VerifyOtp godoc
// @Summary Verify Otp
// @Description Verify otp email
// @Tags Otp
// @Accept json
// @Produce json
// @Param request body request_models.VerificationOtpEmail true "User Email and Otp"
// @Success 200 {object} helpers.ApiResponse
// @Success 400 {object} helpers.ApiResponse
// @Success 404 {object} helpers.ApiResponse
// @Failure 500 {object} helpers.ApiResponse
// @Router /otp/verify-otp [post]
func (controller *OtpEmailController) VerifyOtp(w http.ResponseWriter, r *http.Request) {
	var request request_models.VerificationOtpEmail

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		controller.logLogrus.WithFields(logrus.Fields{
			"error":   err,
			"message": "Failed to parse request body",
		}).Error("Failed to parse request body")

		helpers.SendJson(w, http.StatusBadRequest, helpers.ApiResponse{
			Message: "Failed to parse request body",
			Data:    nil,
		})
		return
	}

	ctx, cancel := context2.WithTimeout(context2.Background(), 4*time.Second)
	defer cancel()

	otpJson, err := controller.rdb.Get(ctx, fmt.Sprintf("Otp %s: ", request.Email)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			controller.logLogrus.WithFields(logrus.Fields{
				"error":   err,
				"message": fmt.Sprintf("Otp %s does not exist", request.Email),
			}).Error("Otp does not exist")

			helpers.SendJson(w, http.StatusNotFound, helpers.ApiResponse{
				Message: "Otp does not exist",
				Data:    nil,
			})
			return
		}

		controller.logLogrus.WithFields(logrus.Fields{
			"error":   err,
			"message": "Failed to get otp store",
		}).Error("Failed to get otp store")

		helpers.SendJson(w, http.StatusInternalServerError, helpers.ApiResponse{
			Message: "Failed to get otp store",
			Data:    nil,
		})
		return
	}

	var otp string
	err = json.Unmarshal([]byte(otpJson), &otp)
	if err != nil {
		controller.logLogrus.WithFields(logrus.Fields{
			"error":   err,
			"message": "Failed to unmarshal otp store",
		}).Error("Failed to unmarshal otp store")

		helpers.SendJson(w, http.StatusInternalServerError, helpers.ApiResponse{
			Message: "Failed to unmarshal otp store",
			Data:    nil,
		})
		return
	}

	if otp != request.Otp {
		controller.logLogrus.WithFields(logrus.Fields{
			"error":   err,
			"message": "Invalid otp store",
		}).Error("Invalid otp store")

		helpers.SendJson(w, http.StatusInternalServerError, helpers.ApiResponse{
			Message: "Invalid otp store",
			Data:    nil,
		})
		return
	}

	var result = struct {
		Email string `json:"email"`
		Otp   string `json:"otp"`
	}{
		Email: request.Email,
		Otp:   otp,
	}

	helpers.SendJson(w, http.StatusOK, helpers.ApiResponse{Message: "Successfully verified otp store", Data: result})
}
