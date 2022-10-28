package main

import "os"
import "fmt"
import "log"

import "github.com/hajimehoshi/ebiten/v2"
import "github.com/hajimehoshi/ebiten/v2/audio"
import "github.com/tinne26/mpegg"

type VideoPlayer struct { player *mpegg.Player }
func (vp *VideoPlayer) Layout(w, h int) (int, int) { return w, h }
func (vp *VideoPlayer) Update() error { return nil }
func (vp *VideoPlayer) Draw(screen *ebiten.Image) {
	mpegg.Draw(screen, vp.player.CurrentFrame())
	ebiten.SetWindowTitle(fmt.Sprintf("ebitengine + mpeg | %.2f FPS", ebiten.ActualFPS()))
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
	ebiten.SetWindowResizable(true)
	ebiten.SetRunnableOnUnfocused(true)
	err = ebiten.RunGame(&VideoPlayer{ mpeggPlayer })
	if err != nil { log.Fatal(err) }
}
