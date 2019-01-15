package term

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/peterh/liner"
)

// Input object, "sort of" iterable object that allows iteration.
// you can iterate over this with a loop like:
// for r.Next() {
//   current := r.Get()
// }
// if r.Error() != nil {
//   iteration ended with error
// }
type Input interface {
	Next() bool           // Next returns true if there is a reply to Get
	Get() json.RawMessage // Get the current reply
	Error() error         // Error returns the non-nil error if the iteration broke
}

// inputReader implements a line-oriented stdin reader
type reader struct {
	scanner *bufio.Scanner
	current json.RawMessage
	err     error
}

// once implements a one-shot Input
type once struct {
	value json.RawMessage
	done  bool
}

// Once creates an Input that returns a single value
func Once(msg json.RawMessage) Input {
	return &once{value: msg}
}

// Stdin returns a Input stream if stdin is not a tty, nil otherwise
func Stdin() (Input, error) {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return nil, fmt.Errorf("Error stating os.Stdin: %s", err)
	}
	if (stat.Mode() & os.ModeCharDevice) != 0 {
		return nil, nil
	}
	return &reader{
		scanner: bufio.NewScanner(os.Stdin),
	}, nil
}

// Readline reads a single line of input
func Readline(prompt string, password bool) (string, error) {
	line := liner.NewLiner()
	defer line.Close()
	line.SetCtrlCAborts(true)
	var result string
	var err error
	if !password {
		result, err = line.Prompt(prompt)
	} else {
		result, err = line.PasswordPrompt(prompt)
	}
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(result), nil
}

// Next implements model.Reply
func (i *reader) Next() bool {
	if i.scanner == nil || i.err != nil {
		return false
	}
	if !i.scanner.Scan() {
		i.scanner = nil
		return false
	}
	if err := json.Unmarshal(i.scanner.Bytes(), &(i.current)); err != nil {
		i.scanner = nil
		i.err = err
		return false
	}
	return true
}

// Get implements Input
func (i *reader) Get() json.RawMessage {
	return i.current
}

// Error implements Imput
func (i *reader) Error() error {
	return i.err
}

// Next implements Input
func (o *once) Next() bool {
	done := o.done
	o.done = true
	return !done
}

// Get implements Input
func (o *once) Get() json.RawMessage {
	return o.value
}

// Error implements Input
func (o *once) Error() error {
	return nil
}
