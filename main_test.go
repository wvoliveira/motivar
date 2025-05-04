package main

import (
	"testing"
)

func TestCheckLanguages(t *testing.T) {
	langError := "language not supported. Use 'br' or 'us'"
	languages := []struct {
		lang   string
		result string
	}{
		{"br", ""},
		{"us", ""},
		{"fr", langError},
		{"ko", langError},
		{"la", langError},
		{"zh", langError},
	}

	for _, l := range languages {
		err := CheckLanguages(l.lang)
		if err != nil {
			if err.Error() != l.result {
				t.Errorf("Language %v was not supported, got: %v, want: %v.", l.lang, err, l.result)
			}
		}
	}
}
