package services

import "strings"

var likeReplacer = strings.NewReplacer(
	"\\", "\\\\",
	"%", "\\%",
	"_", "\\_",
)

func sanitizeLikeInput(s string) string {
	return likeReplacer.Replace(s)
}
