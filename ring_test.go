package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRingBufferAvg(t *testing.T) {
	tests := []struct {
		comment  string
		size     int
		values   []float64
		expected float64
	}{
		{
			comment:  "empty buffer",
			size:     3,
			values:   nil,
			expected: 0,
		},
		{
			comment:  "partially filled buffer",
			size:     5,
			values:   []float64{1, 2, 3},
			expected: (1 + 2 + 3 + 0 + 0) / 5.,
		},
		{
			comment:  "exactly full buffer",
			size:     3,
			values:   []float64{2, 4, 6},
			expected: (2 + 4 + 6) / 3.,
		},
		{
			comment:  "overwritten values",
			size:     3,
			values:   []float64{1, 2, 3, 4},
			expected: (2 + 3 + 4) / 3.,
		},
		{
			comment:  "multiple overwrites",
			size:     2,
			values:   []float64{10, 20, 30, 40},
			expected: (30 + 40) / 2.,
		},
	}

	for _, tc := range tests {
		t.Run(tc.comment, func(t *testing.T) {
			buf := NewRingBuffer(tc.size)
			for _, v := range tc.values {
				buf.Add(v)
			}

			require.Equal(t, tc.expected, buf.Avg(), "unexpected average")
		})
	}
}
