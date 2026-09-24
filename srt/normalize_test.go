package srt

import (
	"testing"
	"time"
)

func TestResolveOverlaps(t *testing.T) {
	sec := time.Second

	cases := []struct {
		name string
		in   []*Cue
		want []time.Duration // expected End for each cue, in order
	}{
		{
			name: "no overlap",
			in: []*Cue{
				{Start: 0, End: 2 * sec},
				{Start: 3 * sec, End: 5 * sec},
			},
			want: []time.Duration{2 * sec, 5 * sec},
		},
		{
			name: "simple overlap clamped to next start",
			in: []*Cue{
				{Start: 0, End: 4 * sec},
				{Start: 2 * sec, End: 6 * sec},
			},
			want: []time.Duration{2 * sec, 6 * sec},
		},
		{
			name: "cascading overlap across three cues",
			in: []*Cue{
				{Start: 0, End: 10 * sec},
				{Start: 1 * sec, End: 9 * sec},
				{Start: 2 * sec, End: 8 * sec},
			},
			want: []time.Duration{1 * sec, 2 * sec, 8 * sec},
		},
		{
			name: "identical start times collapse to zero length",
			in: []*Cue{
				{Start: 5 * sec, End: 9 * sec},
				{Start: 5 * sec, End: 12 * sec},
			},
			want: []time.Duration{5 * sec, 12 * sec},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ResolveOverlaps(tc.in)
			for i, cue := range tc.in {
				if cue.End != tc.want[i] {
					t.Errorf("cue %d: End = %v, want %v", i, cue.End, tc.want[i])
				}
			}
		})
	}
}
