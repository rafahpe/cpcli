package term

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/rafahpe/cpcli/model"
)

// Options stores pagination options
type Options struct {
	PageSize    int
	Paginate    bool
	SkipHeaders bool
	PrettyPrint bool
}

// Output the feed of replies, printing the given columns (if any)
func (options Options) Output(pages model.Reply, format []string) error {
	// If output is CSV-like, dump the header
	if format != nil && len(format) > 0 && !options.SkipHeaders {
		fmt.Println(strings.Join(format, ";"))
	}
	// Keep reading pages of data
	p := options.newPaginator()
	for pages.Next() {
		// Show next item
		page := pages.Get()
		output, err := options.serialize(page, format)
		if err != nil {
			return err
		}
		fmt.Println(output)
		// Check next page
		if ok, err := p.next(); !ok || err != nil {
			return err
		}
	}
	if err := pages.Error(); pages != nil {
		return err
	}
	return nil
}

// Serializes a reply according to the options
func (options Options) serialize(item model.RawReply, format []string) (string, error) {
	if format != nil && len(format) > 0 {
		return model.ToCSV(item, format), nil
	}
	var output []byte
	var err error
	if options.PrettyPrint {
		output, err = json.MarshalIndent(item, "", "  ")
	} else {
		output, err = json.Marshal(item)
	}
	if err != nil {
		return "", err
	}
	return string(output), nil
}
