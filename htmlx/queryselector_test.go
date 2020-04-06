package htmlx

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func ErrorIs(t *testing.T, target, err error) {
	t.Helper()
	if !errors.Is(err, target) {
		t.Fatalf("errors.Is failure target: %v, err: %v", target, err)
	}
}

func TestParseSelector(t *testing.T) {
	tests := []struct {
		input string
		want  *Selector
	}{
		{input: "DIV", want: &Selector{Tag: "div"}},
		{input: "div#id2", want: &Selector{Tag: "div", ID: "id2"}},
		{input: "div.big.red", want: &Selector{Tag: "div", Classes: []string{"big", "red"}}},
		{input: "#id1", want: &Selector{ID: "id1"}},
		{input: ".content", want: &Selector{Classes: []string{"content"}}},
		{input: "div#id5.highlighted", want: &Selector{Tag: "div", ID: "id5", Classes: []string{"highlighted"}}},
		{input: "#id5.highlighted", want: &Selector{ID: "id5", Classes: []string{"highlighted"}}},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.input, func(t *testing.T) {
			got, err := ParseSelector(tc.input)
			require.NoError(t, err)
			require.Equal(t, tc.want, got)
		})
	}
}

func TestValidateSelector(t *testing.T) {
	tests := map[string]struct {
		input *Selector
		want  error
	}{
		"empty_selector": {input: &Selector{}, want: ErrInvalidSelector},
		"invalid_tag":    {input: &Selector{Tag: "p100"}, want: ErrInvalidTag},
	}

	for name, tc := range tests {
		tc := tc
		t.Run(name, func(t *testing.T) {
			err := ValidateSelector(tc.input)
			ErrorIs(t, tc.want, err)
		})
	}
}

func TestParseSelectorErr(t *testing.T) {
	tests := map[string]struct {
		input string
		want  error
	}{
		"whitespace":    {input: "div .abc", want: ErrInvalidSelector},
		"duplicate_id":  {input: "#id1#id2.foo.bar", want: ErrDuplicateID},
		"invalid_class": {input: "..abc.x2", want: ErrInvalidSelector},
		"invalid_tag":   {input: "p100", want: ErrInvalidTag},
		"invalid_char":  {input: "div.x$2", want: ErrInvalidSelector},
	}

	for name, tc := range tests {
		tc := tc
		t.Run(name, func(t *testing.T) {
			_, err := ParseSelector(tc.input)
			ErrorIs(t, tc.want, err)
		})
	}
}
