package formatting

func FormatWindowHeading(word string) string {
	const maxWordLen = 48
	const totalWidth = 54

	if len(word) > maxWordLen {
		word = word[:maxWordLen]
	}

	dashCount := (totalWidth - len(word)) / 2

	heading := make([]byte, 0, totalWidth)

	for range dashCount {
		heading = append(heading, '-')
	}
	heading = append(heading, ' ')
	heading = append(heading, word...)
	heading = append(heading, ' ')
	for range dashCount {
		heading = append(heading, '-')
	}

	return string(heading)
}
