package scanner

type Token interface{}

type Scanner interface {
	Scan() ([]Token, error)
}

type scanner struct {
	input string
}

// Scan implements Scanner.
func (s *scanner) Scan() ([]Token, error) {
	panic("unimplemented scanner.Scan()")
}

// NewScanner returns a new Scanner.
func NewScanner(input string) Scanner {
	return &scanner{input: input}
}

var _ Scanner = (*scanner)(nil)
