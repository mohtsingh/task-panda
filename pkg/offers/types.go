package offers

type Offer struct {
	ID           int     `json:"id"`
	TaskID       int     `json:"task_id"`
	ProviderID   int     `json:"provider_id"`
	OfferedPrice float64 `json:"offered_price"`
	Message      string  `json:"message"`
	Status       string  `json:"status"`
	CreatedAt    string  `json:"created_at"`
	UpdatedAt    string  `json:"updated_at"`
	ProviderName string  `json:"provider_name,omitempty"`
}
