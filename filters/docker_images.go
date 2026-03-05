package filters

import (
	"fmt"
	"strings"
)

func filterDockerImages(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", nil
	}
	if !looksLikeDockerImagesOutput(trimmed) {
		return raw, nil
	}

	raw = trimmed
	lines := strings.Split(raw, "\n")
	if len(lines) == 0 {
		return "", nil
	}

	// Skip warning lines at the top
	startIdx := 0
	for startIdx < len(lines) && strings.HasPrefix(strings.ToUpper(strings.TrimSpace(lines[startIdx])), "WARNING") {
		startIdx++
	}
	if startIdx >= len(lines) {
		return raw, nil
	}

	header := lines[startIdx]

	// New format: IMAGE, ID, DISK USAGE, CONTENT SIZE, EXTRA
	if strings.Contains(header, "DISK USAGE") {
		return filterDockerImagesNewFormat(raw, lines[startIdx:])
	}

	// Classic format: REPOSITORY, TAG, IMAGE ID, CREATED, SIZE
	repoIdx := strings.Index(header, "REPOSITORY")
	tagIdx := strings.Index(header, "TAG")
	sizeIdx := strings.Index(header, "SIZE")

	if repoIdx == -1 || tagIdx == -1 || sizeIdx == -1 {
		return raw, nil
	}

	imageIDIdx := strings.Index(header, "IMAGE ID")
	if imageIDIdx == -1 {
		imageIDIdx = sizeIdx
	}

	var tagged []string
	var noneEntries []string

	for _, line := range lines[startIdx+1:] {
		if strings.TrimSpace(line) == "" {
			continue
		}

		repo := extractColumn(line, repoIdx, tagIdx)
		tag := extractColumn(line, tagIdx, imageIDIdx)
		size := extractColumn(line, sizeIdx, len(line))

		if repo == "<none>" || tag == "<none>" {
			noneEntries = append(noneEntries, fmt.Sprintf("%s:%s %s", repo, tag, size))
			continue
		}

		tagged = append(tagged, fmt.Sprintf("%s:%s %s", repo, tag, size))
	}

	var result []string
	if len(tagged) > 0 {
		result = tagged
	} else {
		result = noneEntries
	}

	total := len(tagged) + len(noneEntries)
	if total == 0 {
		return "", nil
	}

	result = append(result, fmt.Sprintf("%d images total", total))
	out := strings.Join(result, "\n")
	return outputSanityCheck(raw, out), nil
}

// filterDockerImagesNewFormat handles the newer docker images format with
// IMAGE, ID, DISK USAGE, CONTENT SIZE columns.
func filterDockerImagesNewFormat(raw string, lines []string) (string, error) {
	header := lines[0]
	imageIdx := strings.Index(header, "IMAGE")
	idIdx := strings.Index(header, "ID")
	diskIdx := strings.Index(header, "DISK USAGE")
	contentIdx := strings.Index(header, "CONTENT SIZE")

	if imageIdx == -1 || diskIdx == -1 {
		return raw, nil
	}

	// Use ID column to bound image name, fallback to diskIdx
	nameEnd := idIdx
	if nameEnd == -1 || nameEnd <= imageIdx {
		nameEnd = diskIdx
	}

	// Bound disk usage column end
	diskEnd := len(header)
	if contentIdx > diskIdx {
		diskEnd = contentIdx
	}

	var images []string
	for _, line := range lines[1:] {
		if strings.TrimSpace(line) == "" {
			continue
		}
		name := strings.TrimSpace(extractColumn(line, imageIdx, nameEnd))
		disk := strings.TrimSpace(extractColumn(line, diskIdx, diskEnd))
		images = append(images, fmt.Sprintf("%s %s", name, disk))
	}

	if len(images) == 0 {
		return raw, nil
	}

	images = append(images, fmt.Sprintf("%d images total", len(images)))
	out := strings.Join(images, "\n")
	return outputSanityCheck(raw, out), nil
}
