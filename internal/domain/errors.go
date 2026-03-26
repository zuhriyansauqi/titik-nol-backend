package domain

import "errors"

var (
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrInternalServer     = errors.New("internal server error")
	ErrNotFound           = errors.New("not found")

	// Account errors
	ErrAccountNotFound    = errors.New("account not found")
	ErrInvalidAccountType = errors.New("invalid account type")

	// Transaction errors
	ErrTransactionNotFound = errors.New("transaction not found")
	ErrInvalidTxType       = errors.New("invalid transaction type")

	// Category errors
	ErrCategoryNotFound    = errors.New("category not found")
	ErrInvalidCategoryType = errors.New("invalid category type")

	// Authorization errors
	ErrForbidden = errors.New("forbidden: resource belongs to another user")

	// Auth errors
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrPasswordNotSet     = errors.New("password not set for this account")
	ErrInvalidTokenClaims = errors.New("token missing required claims")

	// Validation errors
	ErrNegativeBalance  = errors.New("initial balance cannot be negative")
	ErrEmptyBulkRequest = errors.New("bulk request cannot be empty")
	ErrAlreadyDeleted   = errors.New("resource already deleted")
	ErrValidationFailed = errors.New("validation failed")
)
