package util

// SplitWorker split string by separatorChars, and keep the separator
func SplitWorker(str string, separatorChars []string) []string {
	if str == "" {
		return []string{}
	}

	var list []string
	start := 0
	length := len(str)

	for i := 0; i < length; i++ {
		for _, sep := range separatorChars {
			sepLen := len(sep)

			if i+sepLen <= length && str[i:i+sepLen] == sep {
				if i > start {
					list = append(list, str[start:i]) // add token
				}
				list = append(list, str[i:i+sepLen]) // add separator
				start = i + sepLen
				i += sepLen - 1
				break
			}
		}
	}

	if start < length {
		list = append(list, str[start:]) // add last token
	}

	return list
}

func Contains(arr []string, str string) bool {
	for _, val := range arr {
		if val == str {
			return true
		}
	}
	return false
}
