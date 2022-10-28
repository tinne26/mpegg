# mpegg
[![Go Reference](https://pkg.go.dev/badge/github.com/tinne26/mpegg.svg)](https://pkg.go.dev/github.com/tinne26/mpegg)

**NOTICE: work in progress, I'll improve it in the next days. only videos with audio work**

A simple abstraction layer over [gen2brain/mpeg](https://github.com/gen2brain/mpeg) to make it easier to play videos on [Ebitengine](https://ebitengine.org).

First, you need a video in mpeg format. You can convert your own with [ffmpeg](https://ffmpeg.org/), e.g.:
```
ffmpeg -i my_video.mp4 -c:v mpeg1video -c:a mp2 -f mpeg -q 8 my_video.mpg
```

Then you write a simple Ebitengine program to play it:
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

You will probably combine a few of these while converting your videos!

Silence **noisy** ffmpeg (`-v quiet -stats`):
```
ffmpeg -i my_video.mp4 -v quiet -stats -c:v mpeg1video -c:a mp2 -f mpeg my_video.mpg
```

Convert only a **fragment** of the video (start second `-ss`, duration `-t`):
```
ffmpeg -i my_video.mp4 -c:v mpeg1video -c:a mp2 -f mpeg -ss 3 -t 10 my_video.mpg
```

Set **quality** level explicitly (`-q`, lower is better, I find 4 good, 6 kinda ok, 8 at the limit, but all these still produce very large video sizes, so you may also have to compromise on resolution or framerate):
```
ffmpeg -i my_video.mp4 -c:v mpeg1video -c:a mp2 -f mpeg -q 6 my_video.mpg
```

Set **framerate** explicitly (`-filter:v fps=VALUE`, but going below 24 is almost never a good idea).
```
ffmpeg -i my_video.mp4 -c:v mpeg1video -c:a mp2 -f mpeg -r 24 my_video.mpg
```

Change the video **resolution** while converting (`-s WIDTHxHEIGHT`, where "s" stands for "size"):
```
ffmpeg -i my_video.mp4 -c:v mpeg1video -c:a mp2 -f mpeg -s 640x480 my_video.mpg
```

Keep only video, **remove audio** (`-an`):
```
ffmpeg -i my_video.mp4 -c:v mpeg1video -c:a mp2 -f mpeg -q 6 my_video.mpg
```

Two-pass encoding (can reduce artifacts, improving quality; `-b:v 1600k` is close to `-q 8`, `-b:v 2700k` is close to `-q 4`):
```
ffmpeg -i my_video.mp4 -c:v mpeg1video -c:a mp2 -f mpeg -b:v 1600k -pass 1 my_video_P1.mpg
ffmpeg -i my_video.mp4 -c:v mpeg1video -c:a mp2 -f mpeg -b:v 1600k -pass 2 my_video_P2.mpg
```

Show video info (bitrate, fps and size are probably the only values you will care about):
```
ffprobe -i my_video.mpg -hide_banner
```

## License

The code is licensed under the MIT license.

## Acknowledgements

This package is only a small wrapper, all credit goes to [gen2brain/mpeg](https://github.com/gen2brain/mpeg), based on the original [phoboslab/pl_mpeg](https://github.com/phoboslab/pl_mpeg) C library, and [hajimehoshi/ebiten](https://github.com/hajimehoshi/ebiten).


