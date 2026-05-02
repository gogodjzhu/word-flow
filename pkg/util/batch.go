package util

func BatchSegments(segments []string, threshold int) [][]string {
	if len(segments) == 0 {
		return nil
	}

	var result [][]string
	var current []string
	currentLen := 0

	for _, seg := range segments {
		if currentLen+len(seg) > threshold && len(current) > 0 {
			result = append(result, current)
			current = nil
			currentLen = 0
		}
		current = append(current, seg)
		currentLen += len(seg)
	}

	if len(current) > 0 {
		result = append(result, current)
	}

	return result
}