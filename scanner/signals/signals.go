package signals

import (
	"github.com/antonmedv/expr"
	"github.com/antonmedv/expr/vm"
	iex "github.com/goinvest/iexcloud/v2"
	"time"
)

type Target struct {
	Quote iex.PreviousDay
	Stats iex.KeyStats
}

type Signal struct {
	Name   string `gorm:"primaryKey;autoIncrement:false"`
	Source string
}

type Scan struct {
	Signal
	check *vm.Program
}

type hit struct {
	RuleName  string
	Symbol    string
	QuoteDate time.Time
}

func (s Scan) Check(t Target) (bool, error) {
	res, err := expr.Run(s.check, t)

	if err != nil {
		return false, err
	}

	return res.(bool), err
}

func NewScan(sig Signal) (*Scan, error) {
	p, err := expr.Compile(sig.Source, expr.Env(Target{}))

	if err != nil {
		return nil, err
	}

	return &Scan{
		Signal: sig,
		check:  p,
	}, nil
}
