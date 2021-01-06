// cribbed from github.com/goinvest/iexcloud
package contracts

import (
	"encoding/json"
	"fmt"
	"time"
)

// Date models a report date
type Date time.Time

// String implements the stringer interface for the Date type.
func (d Date) String() string {
	return time.Time(d).Format("2006-01-02")
}

// UnmarshalJSON implements the Unmarshaler interface for Date.
func (d *Date) UnmarshalJSON(data []byte) error {
	var aux string
	err := json.Unmarshal(data, &aux)
	if err != nil {
		return fmt.Errorf("error unmarshaling date to string: %s", err)
	}
	if aux == "" {
		aux = "1929-10-24"
	}
	t, err := time.Parse("2006-01-02", aux)
	if err != nil {
		return fmt.Errorf("error converting %s string to date: %s", aux, err)
	}
	*d = Date(t)
	return nil
}

// MarshalJSON implements the Marshaler interface for Date.
func (d Date) MarshalJSON() ([]byte, error) {
	t := time.Time(d)
	return json.Marshal(t.Format("2006-01-02"))
}
