package main

import "bytes"

// splitCRLF is a split function for a Scanner that splits on '\r\n'.
func splitCRLF(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}

	// Find the index of '\r\n'.
	if i := bytes.Index(data, []byte("\r\n")); i >= 0 {
		// We have a full '\r\n'-terminated line.
		return i + 2, data[0:i], nil
	}

	// If at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		return len(data), data, nil
	}

	// Request more data.
	return 0, nil, nil
}
