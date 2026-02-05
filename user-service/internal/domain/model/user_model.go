package model

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// UserStatus represents user account status
type UserStatus string

const (
	UserStatusActive    UserStatus = "ACTIVE"
	UserStatusInactive  UserStatus = "INACTIVE"
	UserStatusSuspended UserStatus = "SUSPENDED"
)

// User represents user domain model for business logic
type User struct {
	ID           string     `json:"id" db:"id"`
	Email        string     `json:"email" db:"email"`
	Username     string     `json:"username" db:"username"`
	FullName     string     `json:"full_name" db:"full_name"`
	PhoneNumber  string     `json:"phone_number" db:"phone_number"`
	PasswordHash string     `json:"-" db:"password_hash"` // Hidden from JSON
	Status       UserStatus `json:"status" db:"status"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
}

// CreateUserRequest represents request for creating user (Service Layer DTO)
type CreateUserRequest struct {
	Email       string `json:"email" validate:"required,email"`
	Username    string `json:"username" validate:"required,min=3,max=50"`
	Password    string `json:"password" validate:"required,min=6"`
	FullName    string `json:"full_name" validate:"required,min=2,max=100"`
	PhoneNumber string `json:"phone_number" validate:"omitempty,phone"`
}

// UpdateUserRequest represents request for updating user (Service Layer DTO)
type UpdateUserRequest struct {
	ID          string     `json:"id" validate:"required"`
	Email       string     `json:"email" validate:"required,email"`
	Username    string     `json:"username" validate:"required,min=3,max=50"`
	FullName    string     `json:"full_name" validate:"required,min=2,max=100"`
	PhoneNumber string     `json:"phone_number" validate:"omitempty,phone"`
	Status      UserStatus `json:"status" validate:"omitempty,oneof=ACTIVE INACTIVE SUSPENDED"`
}

// ListUsersRequest represents request for listing users with filters
type ListUsersRequest struct {
	Page   int        `json:"page" validate:"min=1"`
	Limit  int        `json:"limit" validate:"min=1,max=100"`
	Search string     `json:"search" validate:"omitempty,max=100"`
	Status UserStatus `json:"status" validate:"omitempty,oneof=ACTIVE INACTIVE SUSPENDED"`
}

// UserResponse represents user response (without sensitive data)
type UserResponse struct {
	ID          string     `json:"id"`
	Email       string     `json:"email"`
	Username    string     `json:"username"`
	FullName    string     `json:"full_name"`
	PhoneNumber string     `json:"phone_number"`
	Status      UserStatus `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// ListUsersResponse represents paginated list response
type ListUsersResponse struct {
	Users []*UserResponse `json:"users"`
	Total int             `json:"total"`
	Page  int             `json:"page"`
	Limit int             `json:"limit"`
}

// Business Logic Methods for User

// IsActive checks if user is active
func (u *User) IsActive() bool {
	return u.Status == UserStatusActive
}

// CanLogin checks if user can login
func (u *User) CanLogin() bool {
	return u.Status == UserStatusActive
}

// IsSuspended checks if user is suspended
func (u *User) IsSuspended() bool {
	return u.Status == UserStatusSuspended
}

// ToResponse converts User to UserResponse (removes sensitive data)
func (u *User) ToResponse() *UserResponse {
	return &UserResponse{
		ID:          u.ID,
		Email:       u.Email,
		Username:    u.Username,
		FullName:    u.FullName,
		PhoneNumber: u.PhoneNumber,
		Status:      u.Status,
		CreatedAt:   u.CreatedAt,
		UpdatedAt:   u.UpdatedAt,
	}
}

// Validation Methods

// Validate validates user data
func (u *User) Validate() error {
	if err := u.validateEmail(); err != nil {
		return err
	}
	if err := u.validateUsername(); err != nil {
		return err
	}
	if err := u.validateFullName(); err != nil {
		return err
	}
	if err := u.validatePhoneNumber(); err != nil {
		return err
	}
	return nil
}

// validateEmail validates email format
func (u *User) validateEmail() error {
	if u.Email == "" {
		return ErrEmailRequired
	}

	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(u.Email) {
		return ErrInvalidEmailFormat
	}

	if len(u.Email) > 255 {
		return ErrEmailTooLong
	}

	return nil
}

// validateUsername validates username format
func (u *User) validateUsername() error {
	if u.Username == "" {
		return ErrUsernameRequired
	}

	if len(u.Username) < 3 {
		return ErrUsernameTooShort
	}

	if len(u.Username) > 50 {
		return ErrUsernameTooLong
	}

	// Username hanya boleh alphanumeric dan underscore
	usernameRegex := regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
	if !usernameRegex.MatchString(u.Username) {
		return ErrInvalidUsernameFormat
	}

	return nil
}

// validateFullName validates full name
func (u *User) validateFullName() error {
	if u.FullName == "" {
		return ErrFullNameRequired
	}

	if len(strings.TrimSpace(u.FullName)) < 2 {
		return ErrFullNameTooShort
	}

	if len(u.FullName) > 100 {
		return ErrFullNameTooLong
	}

	return nil
}

// validatePhoneNumber validates phone number format
func (u *User) validatePhoneNumber() error {
	if u.PhoneNumber == "" {
		return nil // Phone number is optional
	}

	// Remove spaces and dashes
	phone := strings.ReplaceAll(u.PhoneNumber, " ", "")
	phone = strings.ReplaceAll(phone, "-", "")

	// Indonesian phone number format
	phoneRegex := regexp.MustCompile(`^(\+62|62|0)[0-9]{8,12}$`)
	if !phoneRegex.MatchString(phone) {
		return ErrInvalidPhoneFormat
	}

	return nil
}

// Validate methods for request DTOs

// Validate validates CreateUserRequest
func (r *CreateUserRequest) Validate() error {
	if r.Email == "" {
		return ErrEmailRequired
	}
	if r.Username == "" {
		return ErrUsernameRequired
	}
	if r.Password == "" {
		return ErrPasswordRequired
	}
	if len(r.Password) < 6 {
		return ErrPasswordTooShort
	}
	if r.FullName == "" {
		return ErrFullNameRequired
	}

	// Validate email format
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(r.Email) {
		return ErrInvalidEmailFormat
	}

	return nil
}

// Validate validates UpdateUserRequest
func (r *UpdateUserRequest) Validate() error {
	if r.ID == "" {
		return ErrUserIDRequired
	}
	if r.Email == "" {
		return ErrEmailRequired
	}
	if r.Username == "" {
		return ErrUsernameRequired
	}
	if r.FullName == "" {
		return ErrFullNameRequired
	}

	// Validate email format
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(r.Email) {
		return ErrInvalidEmailFormat
	}

	return nil
}

// Validate validates ListUsersRequest
func (r *ListUsersRequest) Validate() error {
	if r.Page <= 0 {
		r.Page = 1 // Set default
	}
	if r.Limit <= 0 {
		r.Limit = 10 // Set default
	}
	if r.Limit > 100 {
		r.Limit = 100 // Set max limit
	}

	return nil
}

// ToUser converts CreateUserRequest to User model
func (r *CreateUserRequest) ToUser() *User {
	return &User{
		Email:       strings.ToLower(strings.TrimSpace(r.Email)),
		Username:    strings.ToLower(strings.TrimSpace(r.Username)),
		FullName:    strings.TrimSpace(r.FullName),
		PhoneNumber: strings.TrimSpace(r.PhoneNumber),
		Status:      UserStatusActive, // Default status
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// ApplyToUser applies UpdateUserRequest to existing User
func (r *UpdateUserRequest) ApplyToUser(user *User) {
	user.Email = strings.ToLower(strings.TrimSpace(r.Email))
	user.Username = strings.ToLower(strings.TrimSpace(r.Username))
	user.FullName = strings.TrimSpace(r.FullName)
	user.PhoneNumber = strings.TrimSpace(r.PhoneNumber)
	if r.Status != "" {
		user.Status = r.Status
	}
	user.UpdatedAt = time.Now()
}

// Domain Errors - Business Logic Errors
var (
	// User validation errors
	ErrUserIDRequired   = NewDomainError("USER_ID_REQUIRED", "User ID is required")
	ErrEmailRequired    = NewDomainError("EMAIL_REQUIRED", "Email is required")
	ErrUsernameRequired = NewDomainError("USERNAME_REQUIRED", "Username is required")
	ErrPasswordRequired = NewDomainError("PASSWORD_REQUIRED", "Password is required")
	ErrFullNameRequired = NewDomainError("FULL_NAME_REQUIRED", "Full name is required")

	// Format validation errors
	ErrInvalidEmailFormat    = NewDomainError("INVALID_EMAIL_FORMAT", "Invalid email format")
	ErrInvalidUsernameFormat = NewDomainError("INVALID_USERNAME_FORMAT", "Username can only contain letters, numbers, and underscores")
	ErrInvalidPhoneFormat    = NewDomainError("INVALID_PHONE_FORMAT", "Invalid phone number format")

	// Length validation errors
	ErrEmailTooLong     = NewDomainError("EMAIL_TOO_LONG", "Email is too long (max 255 characters)")
	ErrUsernameTooShort = NewDomainError("USERNAME_TOO_SHORT", "Username is too short (min 3 characters)")
	ErrUsernameTooLong  = NewDomainError("USERNAME_TOO_LONG", "Username is too long (max 50 characters)")
	ErrPasswordTooShort = NewDomainError("PASSWORD_TOO_SHORT", "Password is too short (min 6 characters)")
	ErrFullNameTooShort = NewDomainError("FULL_NAME_TOO_SHORT", "Full name is too short (min 2 characters)")
	ErrFullNameTooLong  = NewDomainError("FULL_NAME_TOO_LONG", "Full name is too long (max 100 characters)")

	// Business logic errors
	ErrUserNotFound       = NewDomainError("USER_NOT_FOUND", "User not found")
	ErrEmailExists        = NewDomainError("EMAIL_EXISTS", "Email already exists")
	ErrUsernameExists     = NewDomainError("USERNAME_EXISTS", "Username already exists")
	ErrUserInactive       = NewDomainError("USER_INACTIVE", "User account is inactive")
	ErrUserSuspended      = NewDomainError("USER_SUSPENDED", "User account is suspended")
	ErrInvalidCredentials = NewDomainError("INVALID_CREDENTIALS", "Invalid email or password")
)

// DomainError represents business logic error with code and message
type DomainError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *DomainError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// NewDomainError creates new domain error
func NewDomainError(code, message string) *DomainError {
	return &DomainError{
		Code:    code,
		Message: message,
	}
}

// IsUserNotFoundError checks if error is user not found
func IsUserNotFoundError(err error) bool {
	if domainErr, ok := err.(*DomainError); ok {
		return domainErr.Code == "USER_NOT_FOUND"
	}
	return false
}

// IsValidationError checks if error is validation error
func IsValidationError(err error) bool {
	if domainErr, ok := err.(*DomainError); ok {
		return strings.HasSuffix(domainErr.Code, "_REQUIRED") ||
			strings.HasSuffix(domainErr.Code, "_FORMAT") ||
			strings.HasSuffix(domainErr.Code, "_SHORT") ||
			strings.HasSuffix(domainErr.Code, "_LONG")
	}
	return false
}

// IsConflictError checks if error is conflict error (duplicate)
func IsConflictError(err error) bool {
	if domainErr, ok := err.(*DomainError); ok {
		return domainErr.Code == "EMAIL_EXISTS" || domainErr.Code == "USERNAME_EXISTS"
	}
	return false
}
