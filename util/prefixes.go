package util

import (
	"fmt"
	"strings"
)

// PrefixRegistryToImageName prefixes the specified registry to the image.
func PrefixRegistryToImageName(registry string, img string) string {
	if registry == "" {
		return img
	}

	if !strings.HasPrefix(img, registry) && !strings.HasPrefix(img, "library/") {
		return fmt.Sprintf("%s/%s", registry, img)
	}

	return img
}

// PrefixTags prefixes tags in the specified command and returns the new command.
func PrefixTags(cmd string, registry string) (string, []string) {
	tags := []string{}
	fields := strings.Fields(cmd)
	for i := 1; i < len(fields); i++ {
		if fields[i-1] == "-t" || fields[i-1] == "--tag" {
			fields[i] = PrefixRegistryToImageName(registry, fields[i])
			tags = append(tags, fields[i])
		}
	}
	return strings.Join(fields, " "), tags
}
