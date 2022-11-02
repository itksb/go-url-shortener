package handler

import "fmt"

func createShortenURL(id string, baseURL string) string {
	return fmt.Sprintf("%s/%s", baseURL, id)
}
