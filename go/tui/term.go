package tui

import (
	"os"

	"golang.org/x/term"
)

// State wraps the term.State
type State struct {
	state *term.State
}

func enableRawMode(f *os.File) (*State, error) {
	oldState, err := term.MakeRaw(int(f.Fd()))
	if err != nil {
		return nil, err
	}
	return &State{state: oldState}, nil
}

func disableRawMode(f *os.File, s *State) error {
	if s == nil || s.state == nil {
		return nil
	}
	return term.Restore(int(f.Fd()), s.state)
}
