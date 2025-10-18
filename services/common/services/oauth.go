package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/seentics/seentics/services/common/config"
	"github.com/seentics/seentics/services/common/database"
	"github.com/seentics/seentics/services/common/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

type GoogleTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

type GoogleUserInfo struct {
	ID      string `json:"id"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
}

type GitHubTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope"`
}

type GitHubUserInfo struct {
	ID        int    `json:"id"`
	Login     string `json:"login"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
}

type GitHubEmail struct {
	Email   string `json:"email"`
	Primary bool   `json:"primary"`
}

// Google OAuth
func GoogleAuth(c *gin.Context) {
	var req struct {
		Code string `json:"code" validate:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cfg := config.Load()
	if cfg.GoogleClientID == "" || cfg.GoogleClientSecret == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Google OAuth not configured"})
		return
	}

	// Exchange code for access token
	tokenData := url.Values{}
	tokenData.Set("client_id", cfg.GoogleClientID)
	tokenData.Set("client_secret", cfg.GoogleClientSecret)
	tokenData.Set("code", req.Code)
	tokenData.Set("grant_type", "authorization_code")
	tokenData.Set("redirect_uri", fmt.Sprintf("%s/auth/google/callback", cfg.FrontendURL))

	tokenResp, err := http.Post("https://oauth2.googleapis.com/token", "application/x-www-form-urlencoded", bytes.NewBufferString(tokenData.Encode()))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to exchange code for token"})
		return
	}
	defer tokenResp.Body.Close()

	var tokenResponse GoogleTokenResponse
	if err := json.NewDecoder(tokenResp.Body).Decode(&tokenResponse); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse token response"})
		return
	}

	// Get user info
	userReq, _ := http.NewRequest("GET", "https://www.googleapis.com/oauth2/v2/userinfo", nil)
	userReq.Header.Set("Authorization", "Bearer "+tokenResponse.AccessToken)

	client := &http.Client{Timeout: 10 * time.Second}
	userResp, err := client.Do(userReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user info"})
		return
	}
	defer userResp.Body.Close()

	var googleUser GoogleUserInfo
	if err := json.NewDecoder(userResp.Body).Decode(&googleUser); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse user info"})
		return
	}

	// Find or create user
	user, err := findOrCreateOAuthUser("google", googleUser.ID, googleUser.Email, googleUser.Name, googleUser.Picture)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	// Generate JWT token
	token, err := generateJWTToken(user.ID.Hex(), user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// Remove password from response
	user.Password = ""

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Google OAuth successful",
		"data": gin.H{
			"user": user,
			"tokens": gin.H{
				"accessToken": token,
			},
		},
	})
}

// GitHub OAuth
func GitHubAuth(c *gin.Context) {
	var req struct {
		Code string `json:"code" validate:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cfg := config.Load()
	if cfg.GitHubClientID == "" || cfg.GitHubClientSecret == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "GitHub OAuth not configured"})
		return
	}

	// Exchange code for access token
	tokenData := map[string]string{
		"client_id":     cfg.GitHubClientID,
		"client_secret": cfg.GitHubClientSecret,
		"code":          req.Code,
	}

	tokenJSON, _ := json.Marshal(tokenData)
	tokenReq, _ := http.NewRequest("POST", "https://github.com/login/oauth/access_token", bytes.NewBuffer(tokenJSON))
	tokenReq.Header.Set("Accept", "application/json")
	tokenReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	tokenResp, err := client.Do(tokenReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to exchange code for token"})
		return
	}
	defer tokenResp.Body.Close()

	var tokenResponse GitHubTokenResponse
	if err := json.NewDecoder(tokenResp.Body).Decode(&tokenResponse); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse token response"})
		return
	}

	// Get user info
	userReq, _ := http.NewRequest("GET", "https://api.github.com/user", nil)
	userReq.Header.Set("Authorization", "Bearer "+tokenResponse.AccessToken)
	userReq.Header.Set("Accept", "application/vnd.github.v3+json")

	userResp, err := client.Do(userReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user info"})
		return
	}
	defer userResp.Body.Close()

	var githubUser GitHubUserInfo
	if err := json.NewDecoder(userResp.Body).Decode(&githubUser); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse user info"})
		return
	}

	// Get user email if not provided
	if githubUser.Email == "" {
		emailReq, _ := http.NewRequest("GET", "https://api.github.com/user/emails", nil)
		emailReq.Header.Set("Authorization", "Bearer "+tokenResponse.AccessToken)
		emailReq.Header.Set("Accept", "application/vnd.github.v3+json")

		emailResp, err := client.Do(emailReq)
		if err == nil {
			defer emailResp.Body.Close()
			var emails []GitHubEmail
			if json.NewDecoder(emailResp.Body).Decode(&emails) == nil {
				for _, email := range emails {
					if email.Primary {
						githubUser.Email = email.Email
						break
					}
				}
				if githubUser.Email == "" && len(emails) > 0 {
					githubUser.Email = emails[0].Email
				}
			}
		}
	}

	if githubUser.Email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No email address found for GitHub user"})
		return
	}

	// Find or create user
	user, err := findOrCreateOAuthUser("github", fmt.Sprintf("%d", githubUser.ID), githubUser.Email, githubUser.Name, githubUser.AvatarURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	// Generate JWT token
	token, err := generateJWTToken(user.ID.Hex(), user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// Remove password from response
	user.Password = ""

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "GitHub OAuth successful",
		"data": gin.H{
			"user": user,
			"tokens": gin.H{
				"accessToken": token,
			},
		},
	})
}

// OAuth Health Check
func OAuthHealthCheck(c *gin.Context) {
	cfg := config.Load()

	health := gin.H{
		"google": gin.H{
			"configured":   cfg.GoogleClientID != "" && cfg.GoogleClientSecret != "",
			"clientId":     ternary(cfg.GoogleClientID != "", "configured", "missing"),
			"clientSecret": ternary(cfg.GoogleClientSecret != "", "configured", "missing"),
		},
		"github": gin.H{
			"configured":   cfg.GitHubClientID != "" && cfg.GitHubClientSecret != "",
			"clientId":     ternary(cfg.GitHubClientID != "", "configured", "missing"),
			"clientSecret": ternary(cfg.GitHubClientSecret != "", "configured", "missing"),
		},
		"frontend": gin.H{
			"url": cfg.FrontendURL,
			"redirectUris": gin.H{
				"google": fmt.Sprintf("%s/auth/google/callback", cfg.FrontendURL),
				"github": fmt.Sprintf("%s/auth/github/callback", cfg.FrontendURL),
			},
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "OAuth health check",
		"data":    health,
	})
}

func findOrCreateOAuthUser(provider, providerID, email, name, avatar string) (*models.User, error) {
	collection := database.GetUserDB().Collection("users")

	// Try to find existing user by email first
	var user models.User
	err := collection.FindOne(context.Background(), bson.M{"email": email}).Decode(&user)
	
	if err == nil {
		// User exists, update OAuth info
		updateDoc := bson.M{
			"updatedAt":    time.Now(),
			"lastLogin":    time.Now(),
			"isEmailVerified": true,
		}

		if provider == "google" {
			updateDoc["googleId"] = providerID
		} else if provider == "github" {
			updateDoc["githubId"] = providerID
		}

		if user.ProfilePicture == "" && avatar != "" {
			updateDoc["profilePicture"] = avatar
		}

		_, err = collection.UpdateOne(
			context.Background(),
			bson.M{"_id": user.ID},
			bson.M{"$set": updateDoc},
		)
		if err != nil {
			return nil, err
		}

		// Fetch updated user
		err = collection.FindOne(context.Background(), bson.M{"_id": user.ID}).Decode(&user)
		return &user, err
	}

	// Create new user
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("oauth-user"), bcrypt.DefaultCost)
	
	newUser := models.User{
		ID:                primitive.NewObjectID(),
		Email:             email,
		Password:          string(hashedPassword),
		FirstName:         name,
		LastName:          "",
		IsEmailVerified:   true,
		Role:              "user",
		IsActive:          true,
		LastLogin:         time.Now(),
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
		ProfilePicture:    avatar,
		SubscriptionPlan:  "free",
		SubscriptionStatus: "active",
		Timezone:          "UTC",
		Language:          "en",
	}

	if provider == "google" {
		newUser.GoogleID = providerID
	} else if provider == "github" {
		newUser.GitHubID = providerID
	}

	_, err = collection.InsertOne(context.Background(), newUser)
	if err != nil {
		return nil, err
	}

	return &newUser, nil
}

func ternary(condition bool, trueVal, falseVal string) string {
	if condition {
		return trueVal
	}
	return falseVal
}
