package checksum

import (
	"errors"
	"fmt"
	"io"
)

type Scanner struct {
	r        io.Reader
	buf      []byte
	Err      error
	Checksum Sum
}

func NewScanner(r io.Reader) (s *Scanner) {
	s = &Scanner{r: r}
	s.buf = make([]byte, len(s.Checksum))
	return
}

func (s *Scanner) Scan() bool {
	err := s.scan()
	if err != nil {
		if err == io.EOF {
			err = nil
		}
		s.Err = err
		return false
	}
	return true
}

func (s *Scanner) scan() error {
	n, err := fmt.Fscanf(s.r, "%x\n", &s.buf)
	if err != nil {
		return err
	}
	if n != 1 {
		return fmt.Errorf("failed to read checksum hex")
	}
	n = copy(s.Checksum[:], s.buf)
	if n != len(s.Checksum) {
		return errors.New("invalid checksum length")
	}
	return nil
}
