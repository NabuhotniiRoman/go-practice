package config

import (
	"fmt"
	"time"

	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

// Duration is a custom type that can decode duration strings from HCL
type Duration time.Duration

// UnmarshalText implements encoding.TextUnmarshaler
func (d *Duration) UnmarshalText(text []byte) error {
	parsed, err := time.ParseDuration(string(text))
	if err != nil {
		return err
	}
	*d = Duration(parsed)
	return nil
}

// Duration returns the time.Duration value
func (d Duration) Duration() time.Duration {
	return time.Duration(d)
}

// String returns the string representation
func (d Duration) String() string {
	return time.Duration(d).String()
}

// DecodeCTY implements cty decoding for HCL
func (d *Duration) DecodeCTY(val cty.Value) error {
	if val.Type() != cty.String {
		return fmt.Errorf("duration must be a string")
	}

	var s string
	err := gocty.FromCtyValue(val, &s)
	if err != nil {
		return err
	}

	parsed, err := time.ParseDuration(s)
	if err != nil {
		return err
	}

	*d = Duration(parsed)
	return nil
}
