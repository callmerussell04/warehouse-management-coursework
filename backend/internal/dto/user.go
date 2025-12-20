package dto

type CreateUserRequest struct {
	Username string `json:"username" binding:"required,min=1,max=255"`
	Email    string `json:"email" binding:"required,email,max=254"`
	FullName string `json:"full_name" binding:"required,min=1,max=255"`
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
	Username string `json:"username" binding:"required,max=255"`
	Password string `json:"password" binding:"required,max=72"`
}

type LoginResponse struct {
	AccessToken string       `json:"access_token"`
	User        UserResponse `json:"user"`
}

type SendOTPRequest struct {
	Email string `json:"email" binding:"required,email,max=254"`
}

type ForgotUsernameRequest struct {
	Email string `json:"email" binding:"required,email,max=254"`
}

type ResetPasswordRequest struct {
	Email       string `json:"email" binding:"required,email,max=254"`
	OTP         string `json:"otp" binding:"required,len=6"`
	NewPassword string `json:"new_password" binding:"required,min=12,max=72"`
}

type RefreshTokenResponse struct {
	AccessToken string `json:"access_token"`
}

type PagedUsers struct {
	Paging Paging         `json:"paging"`
	Items  []UserResponse `json:"items"`
}
