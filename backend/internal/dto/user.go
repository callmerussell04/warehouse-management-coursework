package dto

type CreateUserRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	FullName string `json:"full_name" binding:"required"`
	Role     string `json:"role" binding:"required,oneof=admin worker"`
}

type UserResponse struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	FullName string `json:"full_name"`
	Role     string `json:"role"`
	IsActive bool   `json:"is_active"`
}

// type UpdateUserRequest struct {
// 	FullName *string `json:"full_name"`
// 	Email    *string `json:"email" binding:"omitempty,email"`
// 	Role     *string `json:"role" binding:"omitempty,oneof=admin worker"`
// }

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	AccessToken string       `json:"access_token"`
	User        UserResponse `json:"user"`
}

type SendOTPRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type ForgotUsernameRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type ResetPasswordRequest struct {
	Email       string `json:"email" binding:"required,email"`
	OTP         string `json:"otp" binding:"required,len=6"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

type RefreshTokenResponse struct {
	AccessToken string `json:"access_token"`
}

type PagedUsers struct {
	Paging Paging         `json:"paging"`
	Items  []UserResponse `json:"items"`
}
