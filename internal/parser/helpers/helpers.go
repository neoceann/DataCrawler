package helpers

import "regexp"

func ExtractIDFromURL(url string) string {
	re := regexp.MustCompile(`(\d{8,})`)
	return re.FindString(url)
}
