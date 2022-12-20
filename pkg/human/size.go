package human

import (
	"errors"
	"strconv"
	"strings"
	"unicode"
)

const (
	b = 1 << (10 * iota)
	kilo
	mega
	giga
	tera
	peta
	exa
)

var errInvalidUnits = errors.New("value must be positive and must contains size valid units like m, Mb, GiB, p, EiB or gb")

// To human readable byte string like 20M, 18.7K, etc.
// Valid units:
//
//	e: Exabyte
//	p: Petabyte
//	t: Terabyte
//	g: Gigabyte
//	m: Megabyte
//	k: Kilobyte
//	b: Byte
//
// The tesults falls to smallest number >= 1 of the unit.
func FormatSizeToHuman(bytes uint64) string {
	unit := ""
	val := float64(bytes)

	switch {
	case bytes >= exa:
		unit = "eb"
		val /= exa
	case bytes >= peta:
		unit = "pb"
		val /= peta
	case bytes >= tera:
		unit = "tb"
		val /= tera
	case bytes >= giga:
		unit = "gb"
		val /= giga
	case bytes >= mega:
		unit = "mb"
		val /= mega
	case bytes >= kilo:
		unit = "kb"
		val /= kilo
	case bytes >= b:
		unit = "b"
	case bytes == 0:
		return "0b"
	}

	return strings.TrimSuffix(strconv.FormatFloat(val, 'f', 1, 64), ".0") + unit
}

// ToBytes parses a string formatted by ByteSize as bytes. Note binary-prefixed and SI prefixed units both mean a base-2 units
// SI       IEC
// kb - k = KiB	= 1024
// mb - m - MiB - 1024 * kilo
// gb - g - GiB - 1024 * mega
// tb - t - TiB - 1024 * giga
// pb - p - PiB - 1024 * peta
// eb - e - EiB - 1024 * tera

func ParseSizeFromHuman(s string) (uint64, error) {
	s = strings.TrimSpace(s)
	s = strings.ToLower(s)

	i := strings.IndexFunc(s, unicode.IsLetter)

	if i == -1 {
		return 0, errInvalidUnits
	}

	vStr, multiple := s[:i], s[i:]
	v, err := strconv.ParseFloat(vStr, 64)
	if err != nil || v < 0 {
		return 0, errInvalidUnits
	}

	switch multiple {
	case "e", "eb", "eib":
		return uint64(v * exa), nil
	case "p", "pb", "pib":
		return uint64(v * peta), nil
	case "t", "tb", "tib":
		return uint64(v * tera), nil
	case "g", "gb", "gib":
		return uint64(v * giga), nil
	case "m", "mb", "mib":
		return uint64(v * mega), nil
	case "k", "kb", "kib":
		return uint64(v * kilo), nil
	case "b":
		return uint64(v), nil
	default:
		return 0, errInvalidUnits
	}
}

type Size uint64

func (s Size) String() string {
	return FormatSizeToHuman(uint64(s))
}

func (s *Size) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var vStr string
	var v uint64
	err := unmarshal(&vStr)
	if err != nil {
		err = unmarshal(&v)
		if err != nil {
			return err
		}
	} else {
		if vStr == "" {
			v = 0
		} else {
			v, err = ParseSizeFromHuman(vStr)
			if err != nil {
				return err
			}
		}
	}

	*s = Size(v)
	return nil
}

func (s Size) MarshalYAML() (interface{}, error) {
	return FormatSizeToHuman(uint64(s)), nil
}

func (s *Size) Value() int64 {
	return int64(*s)
}

func (s *Size) IsZero() bool {
	return int(*s) == 0
}

func MustParseSize(s string) Size {
	v, err := ParseSizeFromHuman(s)
	if err != nil {
		panic(err)
	}
	return Size(v)
}
