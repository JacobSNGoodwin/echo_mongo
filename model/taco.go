package model

// Taco hold types of data stored in database
type Taco struct {
	Meat        string `json:"type"`
	Description string `json:"description"`
}
