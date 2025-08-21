package proto

import (
	"bufio"
	"errors"
	"io"
	"strings"
)

var ErrEmpty = errors.New("empty")

// ReadCommand reads one line and tokenizes it by spaces.
// Returns uppercased command name and args (unmodified case).
func ReadCommand(r *bufio.Reader) (cmd string, args []string, err error) {
	line, err := r.ReadString('\n')
	if err != nil {
		return "", nil, err
	}
	line = strings.TrimSpace(line)
	if line == "" {
		return "", nil, ErrEmpty
	}
	fields := strings.Fields(line)
	if len(fields) == 0 {
		return "", nil, ErrEmpty
	}
	cmd = strings.ToUpper(fields[0])
	args = fields[1:]
	return cmd, args, nil
}

func WriteString(w io.Writer, s string) error {
	_, err := io.WriteString(w, s)
	return err
}
