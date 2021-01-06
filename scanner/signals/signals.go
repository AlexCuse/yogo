package signals

import (
	"github.com/BurntSushi/toml"
	"github.com/alexcuse/yogo/common/contracts"
	"github.com/antonmedv/expr"
	"github.com/antonmedv/expr/vm"
	iex "github.com/goinvest/iexcloud/v2"
	"io/ioutil"
	"os"
)

type Signal struct {
	Name  string
	check *vm.Program
}

func (s Signal) Check(q contracts.Movement) bool {
	res, err := expr.Run(s.check, q)

	if err != nil {
		panic(err)
	}

	return res.(bool)
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
	p, err := expr.Compile(prog, expr.Env(iex.PreviousDay{}))

	if err != nil {
		return nil, err
	}

	return &Signal{
		Name:  name,
		check: p,
	}, nil
}
