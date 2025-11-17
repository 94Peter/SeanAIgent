package locales

import (
	"embed"
	"io/fs"
	"strings"
)

//go:embed en
//go:embed zh-tw
//go:embed en-us

var Content embed.FS

func LocaleExist(name string) bool {
	// check if the locale exists
	_, err := Content.ReadDir(strings.ToLower(name))
	if err != nil {
		if _, ok := err.(*fs.PathError); ok {
			return false
		} else {
			panic(err)
		}
	}
	return true
}
