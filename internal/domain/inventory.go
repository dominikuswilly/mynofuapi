package domain

type Category struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Slug     string `json:"slug"`
	IconURL  *string `json:"icon_url"`
	IsActive bool   `json:"is_active"`
}

type CategoryResponse struct {
	Status string     `json:"status"`
	Data   []Category `json:"data"`
}
