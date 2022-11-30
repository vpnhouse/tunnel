package human

import (
	"fmt"
	"time"
)

type Interval time.Duration

func (s *Interval) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var vStr string
	var v time.Duration
	err := unmarshal(&vStr)
	if err != nil {
		var val int
		err = unmarshal(&val)
		if err != nil {
			return err
		}
		// Convert to seconds
		v = time.Second * time.Duration(val)
	} else {
		v, err = time.ParseDuration(vStr)
		if err != nil {
			return err
		}
	}

	*s = Interval(v)
	return nil
}

func (s Interval) MarshalYAML() (interface{}, error) {
	return fmt.Sprint(time.Duration(s)), nil
}

func (s *Interval) Value() time.Duration {
	return time.Duration(*s)
}

func MustParseInterval(s string) Interval {
	v, err := time.ParseDuration(s)
	if err != nil {
		panic(err)
	}
	return Interval(v)
}

func (s Interval) String() string {
	return fmt.Sprint(time.Duration(s))
}
