package term

import (
	"bufio"
	"log"
	"os"
	"strings"
)

// paginator Keeps count of output lines, prompts for continue if page size exceeded
type paginator struct {
	reader           *bufio.Reader
	logger           *log.Logger
	lineno, pageSize int
}

// newPaginator creates a paginator from the given options
func (options Options) newPaginator() paginator {
	if !options.Paginate {
		return paginator{}
	}
	return paginator{
		reader:   bufio.NewReader(os.Stdin),
		logger:   log.New(os.Stderr, "", 0),
		pageSize: options.PageSize,
	}
}

// next checks if the user wants to advance to following page
func (p *paginator) next() (bool, error) {
	if p.reader == nil {
		// If no pagination, just keep going
		return true, nil
	}
	p.lineno++
	if p.lineno >= p.pageSize {
		p.logger.Println("Press q to quit, enter to continue")
		r, _, err := p.reader.ReadRune()
		if err != nil {
			return false, err
		}
		if strings.ToLower(string(r)) == "q" {
			return false, nil
		}
		p.lineno = 0
	}
	return true, nil
}
