package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID                primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Email             string             `json:"email" bson:"email" validate:"required,email"`
	Password          string             `json:"-" bson:"password" validate:"required,min=6"`
	FirstName         string             `json:"firstName" bson:"firstName" validate:"required"`
	LastName          string             `json:"lastName" bson:"lastName" validate:"required"`
	IsEmailVerified   bool               `json:"isEmailVerified" bson:"isEmailVerified"`
	EmailVerifyToken  string             `json:"-" bson:"emailVerifyToken"`
	ResetPasswordToken string            `json:"-" bson:"resetPasswordToken"`
	ResetPasswordExpires time.Time       `json:"-" bson:"resetPasswordExpires"`
	Role              string             `json:"role" bson:"role" validate:"required"`
	IsActive          bool               `json:"isActive" bson:"isActive"`
	LastLogin         time.Time          `json:"lastLogin" bson:"lastLogin"`
	CreatedAt         time.Time          `json:"createdAt" bson:"createdAt"`
	UpdatedAt         time.Time          `json:"updatedAt" bson:"updatedAt"`
	
	// Profile information
	ProfilePicture    string             `json:"profilePicture" bson:"profilePicture"`
	PhoneNumber       string             `json:"phoneNumber" bson:"phoneNumber"`
	Company           string             `json:"company" bson:"company"`
	JobTitle          string             `json:"jobTitle" bson:"jobTitle"`
	
	// Preferences
	Timezone          string             `json:"timezone" bson:"timezone"`
	Language          string             `json:"language" bson:"language"`
	
	// OAuth
	GoogleID          string             `json:"-" bson:"googleId"`
	GitHubID          string             `json:"-" bson:"githubId"`
	
	// Billing
	StripeCustomerID  string             `json:"-" bson:"stripeCustomerId"`
	SubscriptionPlan  string             `json:"subscriptionPlan" bson:"subscriptionPlan"`
	SubscriptionStatus string            `json:"subscriptionStatus" bson:"subscriptionStatus"`
}

type Website struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	UserID      primitive.ObjectID `json:"userId" bson:"userId" validate:"required"`
	Name        string             `json:"name" bson:"name" validate:"required"`
	URL         string             `json:"url" bson:"url" validate:"required,url"`
	Domain      string             `json:"domain" bson:"domain" validate:"required"`
	TrackingID  string             `json:"trackingId" bson:"trackingId"`
	IsActive    bool               `json:"isActive" bson:"isActive"`
	CreatedAt   time.Time          `json:"createdAt" bson:"createdAt"`
	UpdatedAt   time.Time          `json:"updatedAt" bson:"updatedAt"`
	
	// Settings
	AllowedOrigins []string           `json:"allowedOrigins" bson:"allowedOrigins"`
	DataRetention  int                `json:"dataRetention" bson:"dataRetention"` // days
	
	// Privacy settings
	PrivacySettings PrivacySettings   `json:"privacySettings" bson:"privacySettings"`
}

type PrivacySettings struct {
	CookieConsent     bool `json:"cookieConsent" bson:"cookieConsent"`
	AnonymizeIPs      bool `json:"anonymizeIPs" bson:"anonymizeIPs"`
	DataRetentionDays int  `json:"dataRetentionDays" bson:"dataRetentionDays"`
	GDPRCompliant     bool `json:"gdprCompliant" bson:"gdprCompliant"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type RegisterRequest struct {
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=6"`
	FirstName string `json:"firstName" validate:"required"`
	LastName  string `json:"lastName" validate:"required"`
}

type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}
