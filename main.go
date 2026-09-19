// Command srt-tidy reads an SRT subtitle file and writes a cleaned-up
// version: sequential numbering, consistent timestamp formatting, and no
// stray whitespace.
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/saltbooker/srt-tidy/srt"
)

func main() {
	output := flag.String("o", "", "output file (default: stdout)")
	flag.Parse()

	in := io.Reader(os.Stdin)
	if args := flag.Args(); len(args) > 0 {
		f, err := os.Open(args[0])
		if err != nil {
			fatal(err)
		}
		defer f.Close()
		in = f
	}

	out := io.Writer(os.Stdout)
	if *output != "" {
		f, err := os.Create(*output)
		if err != nil {
			fatal(err)
		}
		defer f.Close()
		out = f
	}

	if err := format(in, out); err != nil {
		fatal(err)
	}
}

// format reads all cues from in, normalizes and sorts them by start time,
// and writes them back out. Sorting means every cue has to be held in
// memory for the duration of the run, unlike the parse and write steps
// themselves, which each only ever touch one cue at a time.
func format(in io.Reader, out io.Writer) error {
	r := srt.NewReader(in)
	var cues []*srt.Cue
	for {
		cue, err := r.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		srt.Normalize(cue)
		cues = append(cues, cue)
	}

	// Stable so cues that share a start time keep their original relative
	// order instead of shuffling on every run.
	sort.SliceStable(cues, func(i, j int) bool {
		return cues[i].Start < cues[j].Start
	})

	w := srt.NewWriter(out)
	for _, cue := range cues {
		if err := w.WriteCue(cue); err != nil {
			return err
		}
	}
	return nil
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "srt-tidy:", err)
	os.Exit(1)
}
