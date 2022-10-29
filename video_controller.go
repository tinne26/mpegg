package mpegg

import "time"
import "sync"

import "github.com/gen2brain/mpeg"

// The default implementation of the controller interface for video-only
// mpegs (videos with no audio streams).
type videoOnlyController struct {
	mutex sync.Mutex
	mpg *mpeg.MPEG
	referenceTime time.Time
	referencePosition time.Duration
	playing bool
}

func newVideoOnlyController(mpg *mpeg.MPEG) *videoOnlyController {
	return &videoOnlyController { mpg: mpg, referenceTime: time.Now() }
}

func (self *videoOnlyController) Play() {
	self.mutex.Lock()
	defer self.mutex.Unlock()
	if !self.playing {
		self.referenceTime = time.Now()
		self.playing = true
	}
}

func (self *videoOnlyController) IsPlaying() bool {
	self.mutex.Lock()
	defer self.mutex.Unlock()
	return self.playing
}

func (self *videoOnlyController) Pause() {
	self.mutex.Lock()
	defer self.mutex.Unlock()
	if self.playing {
		position := self.noLockPosition()
		self.playing = false
		self.referencePosition = position
		self.referenceTime = time.Now()
	}
}

func (self *videoOnlyController) Rewind() {
	self.mutex.Lock()
	defer self.mutex.Unlock()
	self.mpg.Rewind()
	self.referenceTime = time.Now()
	self.referencePosition = 0
}

func (self *videoOnlyController) SeekFrame(t time.Duration, b bool) *mpeg.Frame {
	self.mutex.Lock()
	defer self.mutex.Unlock()
	frame := self.mpg.SeekFrame(t, b)
	if frame != nil {
		self.referencePosition = t
		self.referenceTime = time.Now()
	}
	return frame
}

func (self *videoOnlyController) Position() time.Duration {
	self.mutex.Lock()
	defer self.mutex.Unlock()
	return self.noLockPosition()
}

func (self *videoOnlyController) noLockPosition() time.Duration {
	if self.playing {
		return self.referencePosition + time.Now().Sub(self.referenceTime)
	} else {
		return self.referencePosition
	}
}

func (self *videoOnlyController) MPEG() (*mpeg.MPEG, *sync.Mutex) {
	return self.mpg, &self.mutex
}
