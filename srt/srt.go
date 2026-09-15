// Package srt reads and writes SubRip (.srt) subtitle files one cue at a
// time, so a caller never has to hold a whole file in memory.
package srt

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

// Cue is a single subtitle entry: a time range and the lines of text shown
// during it.
type Cue struct {
	Index int
	Start time.Duration
	End   time.Duration
	Text  []string
}

// Reader pulls cues out of an SRT stream one block at a time. It only ever
// buffers the lines that belong to the cue currently being parsed, which is
// what lets Format handle multi-gigabyte subtitle dumps without loading them
// whole.
type Reader struct {
	br *bufio.Reader
}

// NewReader wraps r for cue-by-cue reading.
func NewReader(r io.Reader) *Reader {
	return &Reader{br: bufio.NewReader(r)}
}

// Next returns the next cue, or io.EOF once the stream is exhausted.
func (r *Reader) Next() (*Cue, error) {
	var lines []string
	for {
		line, atEOF, err := r.readLine()
		if err != nil {
			return nil, err
		}
		trimmed := strings.TrimRight(line, "\r\n")
		if trimmed != "" {
			lines = append(lines, trimmed)
		} else if len(lines) > 0 {
			break // blank line closes a block that already has content
		}
		if atEOF {
			break
		}
	}
	if len(lines) == 0 {
		return nil, io.EOF
	}
	return parseCueBlock(lines)
}

func (r *Reader) readLine() (line string, atEOF bool, err error) {
	line, err = r.br.ReadString('\n')
	if err != nil {
		if err == io.EOF {
			return line, true, nil
		}
		return "", false, err
	}
	return line, false, nil
}

func parseCueBlock(lines []string) (*Cue, error) {
	idx := 0
	i := 0
	if n, err := strconv.Atoi(strings.TrimSpace(lines[0])); err == nil {
		idx = n
		i = 1
	}
	if i >= len(lines) {
		return nil, fmt.Errorf("cue %d: missing timestamp line", idx)
	}
	start, end, err := parseTimestampLine(lines[i])
	if err != nil {
		return nil, fmt.Errorf("cue %d: %w", idx, err)
	}
	return &Cue{Index: idx, Start: start, End: end, Text: lines[i+1:]}, nil
}

func parseTimestampLine(line string) (start, end time.Duration, err error) {
	parts := strings.SplitN(line, "-->", 2)
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid timestamp line %q", line)
	}
	start, err = parseTimestamp(strings.TrimSpace(parts[0]))
	if err != nil {
		return 0, 0, err
	}
	// The end field may be followed by rendering hints (X1:.. Y1:..); only
	// the first token is the timestamp itself.
	fields := strings.Fields(parts[1])
	if len(fields) == 0 {
		return 0, 0, fmt.Errorf("invalid timestamp line %q", line)
	}
	end, err = parseTimestamp(fields[0])
	if err != nil {
		return 0, 0, err
	}
	return start, end, nil
}

// parseTimestamp accepts both the standard "HH:MM:SS,mmm" form and the
// WebVTT-style "HH:MM:SS.mmm" form, since messy files mix the two.
func parseTimestamp(s string) (time.Duration, error) {
	s = strings.ReplaceAll(s, ".", ",")
	main, frac, _ := strings.Cut(s, ",")

	hms := strings.Split(main, ":")
	if len(hms) != 3 {
		return 0, fmt.Errorf("invalid timestamp %q", s)
	}
	h, err1 := strconv.Atoi(hms[0])
	m, err2 := strconv.Atoi(hms[1])
	sec, err3 := strconv.Atoi(hms[2])
	if err1 != nil || err2 != nil || err3 != nil {
		return 0, fmt.Errorf("invalid timestamp %q", s)
	}

	frac = padOrTrim(frac, 3)
	ms, err := strconv.Atoi(frac)
	if err != nil {
		return 0, fmt.Errorf("invalid timestamp %q", s)
	}

	return time.Duration(h)*time.Hour +
		time.Duration(m)*time.Minute +
		time.Duration(sec)*time.Second +
		time.Duration(ms)*time.Millisecond, nil
}

func padOrTrim(s string, n int) string {
	if s == "" {
		return strings.Repeat("0", n)
	}
	if len(s) > n {
		return s[:n]
	}
	return s + strings.Repeat("0", n-len(s))
}

// Writer emits cues as a well-formed SRT stream, renumbering them
// sequentially from 1 regardless of what index they arrived with.
type Writer struct {
	w io.Writer
	n int
}

// NewWriter wraps w for cue-by-cue writing.
func NewWriter(w io.Writer) *Writer {
	return &Writer{w: w}
}

// WriteCue appends one cue block to the stream.
func (w *Writer) WriteCue(c *Cue) error {
	w.n++
	if _, err := fmt.Fprintf(w.w, "%d\n%s --> %s\n", w.n, formatTimestamp(c.Start), formatTimestamp(c.End)); err != nil {
		return err
	}
	for _, line := range c.Text {
		if _, err := fmt.Fprintln(w.w, line); err != nil {
			return err
		}
	}
	_, err := fmt.Fprintln(w.w)
	return err
}

func formatTimestamp(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second
	d -= s * time.Second
	ms := d / time.Millisecond
	return fmt.Sprintf("%02d:%02d:%02d,%03d", h, m, s, ms)
}
