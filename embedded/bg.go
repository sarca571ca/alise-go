package embedded

import "embed"

//go:embed assets/*.png
var AssetsFS embed.FS

func GetHNMMapBytes(id string) ([]byte, error) {
	return AssetsFS.ReadFile("assets/" + id + "_map.png")
}

func GetHNMThumbnailBytes(id string) ([]byte, error) {
	return AssetsFS.ReadFile("assets/" + id + ".png")
}
