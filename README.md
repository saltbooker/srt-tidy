# srt-tidy

A command-line tool that cleans up messy SubRip (`.srt`) subtitle files.

Subtitles pulled from different rippers, editors, and streaming sites tend to
drift out of shape in small, annoying ways: cue numbers that don't run in
order, timestamps that mix commas and periods for the millisecond field,
trailing whitespace on text lines, blank lines left over from an editor. None
of it breaks playback outright, but it makes the files unpleasant to diff,
edit, or feed into another tool. `srt-tidy` reads a file and writes it back
out with consistent numbering and formatting.

The other thing it's built around: subtitle files for long recordings (a
multi-hour lecture, a full season muxed into one track) can get big, and
there's no reason a formatter needs the whole thing in memory to fix it. The
reader and writer in `srt/` each handle one cue block at a time. `srt-tidy`
itself does hold every cue in memory for the length of a run, because
putting cues back in start-time order means looking at all of them before
any of them can be written.

## Usage

```
srt-tidy input.srt > output.srt
srt-tidy -o output.srt input.srt
cat input.srt | srt-tidy > output.srt
```

With no file argument it reads from stdin, so it drops into a pipeline
alongside anything else that produces or consumes subtitle text.

## Example

Input with a mixed-up decimal separator and stray whitespace:

```
2
00:00:05.200 --> 00:00:07,000
Second line   

1
00:00:01,000 --> 00:00:04,000
Hello there.
```

Output:

```
1
00:00:01,000 --> 00:00:04,000
Hello there.

2
00:00:05,200 --> 00:00:07,000
Second line
```

## Status

Early. What's there: streaming SRT parsing and writing, sorting cues by
start time, sequential renumbering, timestamp normalization, whitespace
cleanup within a cue.

Not there yet: detecting or fixing overlapping cue timespans, WebVTT
input/output, a `--check` mode that reports problems without rewriting the
file.

## License

MIT, see [LICENSE](LICENSE).
