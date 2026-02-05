package prowl

type Datum struct {
	Position int     `json:"position"`
	Title    string  `json:"title"`
	Link     string  `json:"link"`
	Source   string  `json:"source"`
	Snippet  string  `json:"snippet"`
	Price    float64 `json:"price,omitempty"`
}
