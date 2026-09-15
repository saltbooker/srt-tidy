// Command srt-tidy reads an SRT subtitle file and writes a cleaned-up
// version: sequential numbering, consistent timestamp formatting, and no
// stray whitespace.
package main

import (
	"flag"
	"fmt"
	"io"
	"os"

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

// format streams cues from in to out, normalizing each one as it passes
// through. Only one cue is ever held in memory at a time.
func format(in io.Reader, out io.Writer) error {
	r := srt.NewReader(in)
	w := srt.NewWriter(out)
	for {
		cue, err := r.Next()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		srt.Normalize(cue)
		if err := w.WriteCue(cue); err != nil {
			return err
		}
	}
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "srt-tidy:", err)
	os.Exit(1)
}
