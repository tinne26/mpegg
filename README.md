# mpegg
[![Go Reference](https://pkg.go.dev/badge/github.com/tinne26/mpegg.svg)](https://pkg.go.dev/github.com/tinne26/mpegg)

**NOTICE: Development is halted until some issues are resolved in gen2brain/mpeg. Don't recommend using this package yet.**

A simple abstraction layer over [gen2brain/mpeg](https://github.com/gen2brain/mpeg) to make it easier to play videos on [Ebitengine](https://ebitengine.org). It uses MPEG-1, which is an old video format but kinda does the job at the expense of x2-x3 video sizes.

## Code example

First, you need a video in MPEG-1 format. You can grab one [somewhere](https://github.com/gen2brain/mpeg/blob/main/testdata/test.mpg) or convert your own with [ffmpeg](https://ffmpeg.org/):
```
ffmpeg -i my_video.mp4 -c:v mpeg1video -c:a mp2 -f mpeg -q 8 my_video.mpg
```

Then you write a simple program to play it on Ebitengine:
```Golang
package main

import ( "os" ; "log" )
import "github.com/hajimehoshi/ebiten/v2"
import "github.com/hajimehoshi/ebiten/v2/audio"
import "github.com/tinne26/mpegg"

type VideoPlayer struct { player *mpegg.Player }
func (vp *VideoPlayer) Layout(w, h int) (int, int) { return w, h }
func (vp *VideoPlayer) Update() error { return nil }
func (vp *VideoPlayer) Draw(screen *ebiten.Image) {
	mpegg.Draw(screen, vp.player.CurrentFrame())
}

func main() {
	// initialize audio context (necessary if the video has audio)
	audio.NewContext(44100) // sample rate must match the video

	// open the video and create the mpeg player
	src, err := os.Open("my_video.mpg")
	if err != nil { log.Fatal(err) }
	mpeggPlayer, err := mpegg.NewPlayer(src)
	if err != nil { log.Fatal(err) }
	mpeggPlayer.Play() // activate the playback clock

	// run the video player!
	err = ebiten.RunGame(&VideoPlayer{ mpeggPlayer })
	if err != nil { log.Fatal(err) }
}
```

The `mpegg.Player` also provides easy access to the main mpeg properties and a few control functions:
- `Player.Play()`, `Player.Pause()` and `Player.IsPlaying()` to play or pause the video.
- `Player.SeekFast(time.Duration)`, and `Player.Position() time.Duration` to control the playback position.
- `Player.MPEG() (*mpeg.MPEG, *sync.Mutex)` to access the underlying mpeg.

Check the [package's documentation](https://pkg.go.dev/github.com/tinne26/mpegg) to explore these and a few more.

## Limitations

TODO: write about biggest sizes and framerates viable even for low-end laptop CPUs.

TODO: write about audio and video sync limitations on Ebitengine being ~10ms coarse.

## ffmpeg cheatsheet

Here's a small list of useful ffmpeg flags. See the [ffmpeg tips doc](https://github.com/tinne26/mpegg/blob/main/docs/ffmpeg_tips.md) for the full commands and further info:
- `-v quiet -stats`: prevent noisy command output but still show conversion progress.
- `-ss 3 -t 10`: convert only a fragment of the video (start at second 3, convert 10 seconds of video).
- `-q 4`: set the quality level explicitly. Lower is better.
- `-filter:v fps=VALUE`: set framerate explicitly (better don't go below 24!).
- `-s 640x480`: adjust video resolution.
- `-an` instead of `-c:a mp2`: remove the audio to keep only the video.
- `ffprobe -i my_video.mpg -hide_banner`: show video info if you have ffprobe.

## License

The code is licensed under the MIT license.

## Acknowledgements

This package is only a small wrapper, all credit goes to [gen2brain/mpeg](https://github.com/gen2brain/mpeg), based on the original [phoboslab/pl_mpeg](https://github.com/phoboslab/pl_mpeg) C library, and [hajimehoshi/ebiten](https://github.com/hajimehoshi/ebiten).


