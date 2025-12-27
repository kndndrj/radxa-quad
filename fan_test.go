package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFan_Update(t *testing.T) {
	type tcase struct {
		comment  string
		window   int
		temps    []float64
		expected []float64
	}

	tests := []struct {
		comment string
		curve   []curvePoint
		cases   []tcase
	}{
		{
			comment: "rising linear curve",
			curve: []curvePoint{
				// Below 30, speed should be 20%.
				{Temperature: 30, Speed: 20},
				{Temperature: 50, Speed: 40},
				{Temperature: 70, Speed: 60},
				// Above 70, speed should be 60%.
			},
			cases: []tcase{
				{
					comment: "single sample",
					window:  3,
					temps:   []float64{40},
					// interpolated:   30
					expected: []float64{10},
				},
				{
					comment: "averaging ramps gradually",
					window:  3,
					temps:   []float64{40, 50, 90},
					// interpolated:   30, 40, 60
					expected: []float64{10, 23.33, 43.33},
				},
				{
					comment: "window fills then slides",
					window:  3,
					temps:   []float64{40, 50, 60, 60, 60, 55},
					// interpolated:   30, 40, 50, 50, 50, 45
					expected: []float64{10, 23.33, 40, 46.66, 50, 48.33},
				},
				{
					comment: "step change is delayed",
					window:  3,
					temps:   []float64{40, 40, 40, 40, 70, 70, 70},
					// interpolated:   30, 30, 30, 30, 60, 60, 60
					expected: []float64{10, 20, 30, 30, 40, 50, 60},
				},
			},
		},
		{
			comment: "up-down curve",
			curve: []curvePoint{
				// Below 30, speed should be 0%.
				{Temperature: 30, Speed: 10},
				{Temperature: 40, Speed: 0},
				{Temperature: 50, Speed: 80},
				{Temperature: 60, Speed: 100},
				{Temperature: 80, Speed: 70},
				// Above 80, speed should be 80%.
			},
			cases: []tcase{
				{
					comment: "gradual ramp",
					window:  3,
					temps:   []float64{35, 40, 50, 60, 60, 70, 80, 100},
					// interpolated:   5, 0, 80, 100, 100, 85, 70, 70
					expected: []float64{1.66, 1.66, 28.33, 60, 93.33, 95, 85, 75},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.comment, func(t *testing.T) {
			for _, tc := range tt.cases {
				t.Run(tc.comment, func(t *testing.T) {
					fan := &fan{
						curve:   tt.curve,
						history: NewRingBuffer(tc.window),
					}

					for i, temp := range tc.temps {
						require.InDelta(t, tc.expected[i], fan.update(temp), 0.01, "unexpected fan speed.")
					}
				})
			}
		})
	}
}
