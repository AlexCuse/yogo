package signals

import (
	"github.com/BurntSushi/toml"
	"github.com/antonmedv/expr"
	"github.com/antonmedv/expr/vm"
	iex "github.com/goinvest/iexcloud/v2"
	"io/ioutil"
	"os"
)

type Target struct {
	Quote iex.PreviousDay
	Stats iex.KeyStats
}

type Signal struct {
	Name   string
	Source string
	check  *vm.Program
}

func (s Signal) Check(t Target) (bool, error) {
	res, err := expr.Run(s.check, t)

	if err != nil {
		return false, err
	}

	return res.(bool), err
}

func Load(filepath string) ([]*Signal, error) {

	var input map[string]string

	// Read config file
	f, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	buf, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}
	err = toml.Unmarshal(buf, &input)

	signals := make([]*Signal, 0)

	for name, prog := range input {
		s, err := NewSignal(name, prog)

		if err != nil {
			panic(err)
		}

		signals = append(signals, s)
	}

	return signals, nil
}

func NewSignal(name string, prog string) (*Signal, error) {
	p, err := expr.Compile(prog, expr.Env(Target{}))

	if err != nil {
		return nil, err
	}

	return &Signal{
		Name:  name,
		check: p,
	}, nil
}
