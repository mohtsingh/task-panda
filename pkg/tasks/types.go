package tasks

type Task struct {
	ID          int     `json:"id"`
	Category    string  `json:"category"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Budget      float64 `json:"budget"`
	Location    string  `json:"location"`
	Date        string  `json:"date"`
}
