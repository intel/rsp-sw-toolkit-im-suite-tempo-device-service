package main

import (
	"io"
	"io/ioutil"
	"log"
)

// Drain reads r til EOF, then closes it.
func Drain(r io.ReadCloser) {
	LogIfErr(io.Copy(ioutil.Discard, r))
	LogIfErr(r.Close())
}

// IgnoreEOF ignores EOF errors.
var IgnoreEOF = ErrsFilter(io.EOF)

// IgnoreErr returns a decorator that filters out the given errors.
func ErrsFilter(errs ...error) func(args ...interface{}) error {
	if len(errs) == 0 {
		return func(args ...interface{}) error { return nil }
	}
	return func(args ...interface{}) error {
		if len(args) == 0 {
			return nil
		}
		thisErr, _ := args[len(args)-1].(error)
		if thisErr == nil {
			return nil
		}
		for _, err := range errs {
			if thisErr == err {
				return nil
			}
		}
		return thisErr
	}
}

// LogIfErr logs the last argument if it's a non-nil error.
func LogIfErr(args ...interface{}) {
	if len(args) == 0 {
		return
	}
	if err, _ := args[len(args)-1].(error); err != nil {
		log.Printf("error: %+v", err)
	}
}
