package chunk

import (
	"strings"
)

func SplitText(text string, maxChunkSize int, overlap int) []string {
	paragraphs := strings.Split(text, "\n\n")
	var chunks []string
	current := ""

	for _, paragraph := range paragraphs {
		paragraph = strings.TrimSpace(paragraph)
		if paragraph == "" {
			continue
		}

		// if a single paragraph alone exceeds maxChunkSize, split it directly
		for len(paragraph) > maxChunkSize {
			chunks = append(chunks, paragraph[:maxChunkSize])
			paragraph = paragraph[maxChunkSize-overlap:]
		}

		// would adding this paragraph overflow the current chunk?
		if len(current)+len(paragraph) > maxChunkSize {
			chunks = append(chunks, current)

			overlapText := ""
			if len(current) > overlap {
				overlapText = current[len(current)-overlap:]
			} else {
				overlapText = current
			}

			current = overlapText + "\n\n" + paragraph
		} else {
			if current != "" {
				current += "\n\n"
			}
			current += paragraph
		}
	}

	if current != "" {
		chunks = append(chunks, current)
	}

	return chunks
}