package filters

import (
	"fmt"
	"strings"
)

func filterDockerHistory(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", nil
	}
	if !looksLikeDockerHistoryOutput(trimmed) {
		return raw, nil
	}

	raw = trimmed
	lines := strings.Split(raw, "\n")
	if len(lines) < 2 {
		return raw, nil
	}

	header := lines[0]
	imageIdx := strings.Index(header, "IMAGE")
	createdIdx := strings.Index(header, "CREATED")
	createdByIdx := strings.Index(header, "CREATED BY")
	sizeIdx := strings.Index(header, "SIZE")

	if imageIdx == -1 || sizeIdx == -1 {
		return raw, nil
	}

	commentIdx := strings.Index(header, "COMMENT")
	sizeBound := commentIdx
	if sizeBound == -1 {
		sizeBound = len(header)
	}

	var out []string
	count := 0
	for _, line := range lines[1:] {
		if strings.TrimSpace(line) == "" {
			continue
		}

		size := extractColumn(line, sizeIdx, sizeBound)
		if size == "0B" {
			continue
		}

		count++
		image := extractColumn(line, imageIdx, createdIdx)
		if len(image) > 12 {
			image = image[:12]
		}

		createdBy := ""
		if createdByIdx != -1 {
			createdBy = extractColumn(line, createdByIdx, sizeIdx)
			if len(createdBy) > 60 {
				createdBy = createdBy[:57] + "..."
			}
		}

		entry := fmt.Sprintf("%s %s %s", image, size, createdBy)
		out = append(out, strings.TrimSpace(entry))
	}

	if count == 0 {
		return raw, nil
	}

	out = append(out, fmt.Sprintf("%d layers (non-zero)", count))
	result := strings.Join(out, "\n")
	return outputSanityCheck(raw, result), nil
}
