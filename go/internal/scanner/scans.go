package scanner

import (
	"time"

	"github.com/antonmedv/expr"
	"github.com/antonmedv/expr/vm"
	iex "github.com/goinvest/iexcloud/v2"
)

type Target struct {
	Quote iex.PreviousDay
	Stats iex.KeyStats
}

type Scan struct {
	Name   string `json:"name,omitempty"`
	Source string `json:"source,omitempty"`
	check  *vm.Program
}

type hit struct {
	RuleName  string
	Symbol    string
	QuoteDate time.Time
}

func (s *Scan) Check(t Target) (bool, error) {
	res, err := expr.Run(s.check, t)

	if err != nil {
		return false, err
	}

	return res.(bool), err
}

func (s *Scan) Compile() error {
	p, err := expr.Compile(s.Source, expr.Env(Target{}))

	if err != nil {
		return err
	}

	s.check = p

	return nil
}
