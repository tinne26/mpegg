package mpegg

import "io"
import "time"
import "sync"
import "unsafe"

import "github.com/gen2brain/mpeg"
import "github.com/hajimehoshi/ebiten/v2/audio"

// Implements both io.Reader for audio.Player and mpegController for mpegg.Player.
// Doing both things at once feels somewhat weird and confusing, but whatever.
type audioAdapter struct {
	mutex sync.Mutex
	audioPlayer *audio.Player
	mpg *mpeg.MPEG
	leftovers []byte

	// the position compensation is set during seeks to compensate
	// the audio player's time not being possible to reset
	positionCompensation time.Duration

	// a couple hacky variables to try to compensate ebitengine's 
	// low resolution on audio.Player.Current()
	latestAudioCurrentValue time.Duration
	latestAudioCurrentInstant time.Time
}

func newAudioAdapter(mpg *mpeg.MPEG) (*audioAdapter, error) {
	// if mpg.NumAudioStreams() == 0 { panic("internal code misuse") }

	ctx := audio.CurrentContext()
	if ctx == nil { return nil, ErrNilAudioContext }
	if mpg.Channels() > 2 { return nil, ErrTooManyChannels }
	if ctx.SampleRate() != mpg.Samplerate() {
		return nil, ErrBadSampleRate
	}

	mpg.SetAudioFormat(mpeg.AudioS16)
	adapter := &audioAdapter{ mpg: mpg, leftovers: make([]byte, 0) }
	audioPlayer, err := ctx.NewPlayer(adapter)
	if err != nil { return nil, err }
	audioPlayer.SetBufferSize(100*time.Millisecond)
	adapter.audioPlayer = audioPlayer

	return adapter, nil
}

func (self *audioAdapter) Read(buffer []byte) (int, error) {
	// helper function
	var eof = func(ended bool) error { if ended { return io.EOF } else { return nil } }

	// mutex
	self.mutex.Lock()
	defer self.mutex.Unlock()

	// get the audio decoder and ensure the stream hasn't ended
	var servedBytes int = 0
	audio := self.mpg.Audio()
	if len(buffer) == 0 { return 0, eof(audio.HasEnded()) }
	
	// if we had leftover bytes from the previous read, use that
	if len(self.leftovers) > 0 {
		copiedBytes := copy(buffer, self.leftovers)
		servedBytes += copiedBytes
		buffer = buffer[copiedBytes : ]
		if copiedBytes >= len(self.leftovers) {
			self.leftovers = self.leftovers[ : 0]
		} else {
			copy(self.leftovers, self.leftovers[copiedBytes : ])
		}
	}
	
	// stop if audio has already ended
	if audio.HasEnded() { return servedBytes, io.EOF }

	// decode audio and move it into the buffer
	for len(buffer) > 0 {
		// decode some samples
		samples := audio.Decode()
		
		// cast the samples int16 slice to a byte slice
		ptr := (*byte)(unsafe.Pointer(&samples.S16[0]))
		length := len(samples.S16)*2
		samplesAsBytes := unsafe.Slice(ptr, length)
		
		// copy into the output buffer
		copiedBytes := copy(buffer, samplesAsBytes)
		servedBytes += copiedBytes
		buffer = buffer[copiedBytes : ]
		
		// if some samples were left over, store them
		if copiedBytes < len(samplesAsBytes) {
			self.leftovers = append(self.leftovers, samplesAsBytes[copiedBytes : ]...)
			return servedBytes, eof(audio.HasEnded())
		}

		// if audio has ended, don't try to continue
		if audio.HasEnded() { return servedBytes, io.EOF }
	}

	return servedBytes, eof(audio.HasEnded())
}

func (self *audioAdapter) MPEG() (*mpeg.MPEG, *sync.Mutex) {
	return self.mpg, &self.mutex
}

func (self *audioAdapter) Play() {
	self.mutex.Lock()
	if self.audioPlayer.IsPlaying() { return }
	if self.mpg.HasEnded() { return }
	self.mutex.Unlock()
	self.audioPlayer.Play()
}

func (self *audioAdapter) IsPlaying() bool {
	return self.audioPlayer.IsPlaying()
}

func (self *audioAdapter) Pause() {
	self.audioPlayer.Pause()
}

func (self *audioAdapter) Rewind() {
	self.mutex.Lock()
	defer self.mutex.Unlock()
	self.mpg.Rewind()
}

func (self *audioAdapter) SeekFrame(t time.Duration, b bool) *mpeg.Frame {
	self.mutex.Lock()
	defer self.mutex.Unlock()
	frame := self.mpg.SeekFrame(t, b)
	if frame != nil {
		self.positionCompensation = self.audioPlayer.Current() - t
	}
	return frame
}

func (self *audioAdapter) Position() time.Duration {
	self.mutex.Lock()
	defer self.mutex.Unlock()

	current := self.audioPlayer.Current()

	// rectify ebitengine's time with some silly lousy messy hacks
	if current == self.latestAudioCurrentValue {
		offset := time.Now().Sub(self.latestAudioCurrentInstant)
		if offset > time.Millisecond*10 { offset = time.Millisecond*10 }
		current += offset
	} else {
		self.latestAudioCurrentInstant = time.Now()
		self.latestAudioCurrentValue = current
	}

	return current - self.positionCompensation
}
