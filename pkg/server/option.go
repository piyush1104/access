package server

import (
	"github.com/100mslive/auth"
)

type Option func(s *optionsStruct)

type optionsStruct struct {
	auth           auth.Client
	tracingEnabled bool
}

func newOptions() *optionsStruct {
	return &optionsStruct{
		tracingEnabled: true,
	}
}

// WithAuth ...
func WithAuth(auth auth.Client) Option {
	return func(s *optionsStruct) {
		if auth != nil {
			s.auth = auth
		}
	}
}
