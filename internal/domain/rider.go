package domain


type Rider struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Username       string    `json:"username"`
	CreatedAt      string    `json:"created_at"`
	Active         int       `json:"active"`
	WhatsappNumber string    `json:"whatsapp_number"`
}

type RiderResponse struct {
	Status string  `json:"status"`
	Data   []Rider `json:"data"`
}
