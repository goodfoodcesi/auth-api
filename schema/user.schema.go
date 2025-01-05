package schema

type CreateUser struct {
	FirstName   string `json:"first_name" binding:"required,min=2,max=50"`
	LastName    string `json:"last_name" binding:"required,min=2,max=50"`
	Email       string `json:"email" binding:"required,email"`
	Password    string `json:"password" binding:"required,min=8"`
	PhoneNumber string `json:"phone_number" binding:"required,e164"`
}

type UpdateUser struct {
	FirstName   string `json:"first_name" binding:"omitempty,min=2,max=50"`
	LastName    string `json:"last_name" binding:"omitempty,min=2,max=50"`
	PhoneNumber string `json:"phone_number" binding:"omitempty,e164"`
}
