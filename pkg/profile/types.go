package profile

type Profile struct {
	ID          int    `json:"id"`
	FullName    string `json:"full_name"`
	Email       string `json:"email"`
	Address     string `json:"address"`
	PhoneNumber string `json:"phone_number"`
	Bio         string `json:"bio"`
	Role        string `json:"role"` // CUSTOMER or SERVICE_PROVIDER
	Photo       []byte `json:"-"`
}
