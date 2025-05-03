package controllers

import (
	"go-libraryschool/helpers"
	"go-libraryschool/middlewares"
	"go-libraryschool/models/jwt_models"
	"net/http"
)

func GetProfile(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value(middlewares.UserContextKey).(*jwt_models.JWTClaims)

	response := map[string]interface{}{
		"userID": claims.UserID,
		"email":  claims.Email,
		"role":   claims.Roles,
	}

	helpers.SendJson(w, http.StatusOK, helpers.ApiResponse{
		Message: "Your profile",
		Data:    response,
	})
}
