package model

// Taco hold types of data stored in database
type Taco struct {
	Type        string `json:"type"`
	Description string `json:"description"`
}
