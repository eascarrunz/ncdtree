package fasta

import (
	"bufio"
	"errors"
	"fmt"
	"strings"
)

// The id is the string after '>' up to the first whitespace.
// Sequences are concatenated with newlines removed.
func ReadFasta(scanner bufio.Scanner) (*[]string, *[][]byte, error) {
	nameList := make([]string, 0)
	fastaStrings := make([][]byte, 0)
	// fastaStrings := make([]string, 0)
	nameSet := make(map[string]int) // Set for checking duplicates of identifiers, with dummy int values
	var curID string
	var sb strings.Builder

	flush := func() error {
		if curID != "" {
			fastaStrings = append(fastaStrings, []byte(sb.String()))
			nameList = append(nameList, curID)
			_, ok := nameSet[curID]
			if ok {
				return errors.New("duplicated identifier in Fasta file: " + curID)
			}
			nameSet[curID] = 0
			sb.Reset()
		}
		return nil
	}

	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 {
			// skip empty lines
			continue
		}
		if line[0] == '>' {
			// new record: flush previous
			if err := flush(); err != nil {
				return nil, nil, err
			}
			rest := line[1:]
			parts := strings.Fields(rest)
			if len(parts) == 0 {
				return nil, nil, fmt.Errorf("empty Fasta descriptor in line: %q", line)
			}
			curID = parts[0]
		} else {
			sb.WriteString(strings.TrimSpace(line))
		}
	}
	if err := flush(); err != nil {
		return nil, nil, err
	}

	if err := scanner.Err(); err != nil {
		return nil, nil, err
	}

	return &nameList, &fastaStrings, nil
}
