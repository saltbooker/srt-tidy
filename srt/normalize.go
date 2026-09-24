package srt

import "strings"

// Normalize cleans up a single cue in place: it strips trailing whitespace
// from each line of text, drops blank lines at the start or end of the
// text block, and swaps a start/end pair that arrived reversed. It does not
// touch cue ordering or overlap between cues, since that requires looking
// across cues rather than at one in isolation.
func Normalize(c *Cue) {
	trimmed := make([]string, 0, len(c.Text))
	for _, line := range c.Text {
		trimmed = append(trimmed, strings.TrimRight(line, " \t"))
	}
	c.Text = trimEmptyEdges(trimmed)

	if c.End < c.Start {
		c.Start, c.End = c.End, c.Start
	}
}

// ResolveOverlaps walks cues in start-time order and clamps any cue whose
// end time runs past the next cue's start, so no two cues are ever shown
// at once. cues must already be sorted by Start. Checking only each cue
// against its immediate successor is enough: Start is non-decreasing
// across the slice, so if cue i no longer overlaps cue i+1 it can't
// overlap anything further along either.
func ResolveOverlaps(cues []*Cue) {
	for i := 0; i < len(cues)-1; i++ {
		if next := cues[i+1].Start; cues[i].End > next {
			cues[i].End = next
		}
	}
}

func trimEmptyEdges(lines []string) []string {
	start := 0
	for start < len(lines) && lines[start] == "" {
		start++
	}
	end := len(lines)
	for end > start && lines[end-1] == "" {
		end--
	}
	return lines[start:end]
}
