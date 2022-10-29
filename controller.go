package mpegg

import "sync"
import "time"

import "github.com/gen2brain/mpeg"

// A common interface that helps us control the timing and position
// of the mpeg.
type mpegController interface {
	Play()
	IsPlaying() bool
	Pause()
	Rewind()
	SeekFrame(time.Duration, bool) *mpeg.Frame
	Position() time.Duration
	MPEG() (*mpeg.MPEG, *sync.Mutex)
}
