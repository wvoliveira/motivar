package data

// Phrase phrases struct
type Phrase struct {
	Author   string `json:"author" csv:"author"`
	Phrase   string `json:"phrase" csv:"phrase"`
	Language string `json:"language" csv:"language"`
}
