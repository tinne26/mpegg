package mpegg

import "io"
import "time"
import "sync"
import "image/color"
import "errors"

import "github.com/hajimehoshi/ebiten/v2"
import "github.com/gen2brain/mpeg"

// NOTES:
// - check if mono audio is relevant (mpg.Channels() <- audio channels count)
// - Forward() Backward() ? or should we move in some other way?

// A collection of initialization errors defined by this package for [NewPlayer]().
// Other mpeg-specific errors are also possible.
var (
	ErrNoVideo         = errors.New("mpeg doesn't include any video stream")
	ErrMissingHeaders  = errors.New("one or more mpeg streams are missing headers")
	ErrNilAudioContext = errors.New("mpeg has audio stream but audio.Context is not initialized")
	ErrBadSampleRate   = errors.New("mpeg audio stream and audio context sample rates don't match")
	ErrTooManyChannels = errors.New("mpeg audio streams with more than 2 channels are not supported")
)

// Internal correction factor to try to compensate between video
// and audio delays inherent to Ebitengine's processing.
var syncCorrection time.Duration = 0*time.Millisecond

// A [Player] represents an mpeg video player, typically also including audio.
// 
// The player is a simple abstraction layer or wrapper around the lower level
// [gen2brain/mpeg] types, which implement the underlying decoders used to
// make playing video possible on Ebitengine.
//
// Usage is quite similar to Ebitengine audio players:
//  - Create a [NewPlayer]().
//  - Call [Player.Play()] to start the video.
//  - Audio will play automatically. Frames are obtained with [Player.CurrentFrame]().
//  - Use [Player.Pause]() or [Player.SeekFast]() to control the video.
// More methods are available, but that's the main idea.
//
// [gen2brain/mpeg]: https://github.com/gen2brain/mpeg
type Player struct {
	currentFrame *ebiten.Image
	controller mpegController
	onBlackFrame bool
}

// Creates a new mpeg video [Player]. The read-seeker source is usually
// a file opened with [os.Open]().
func NewPlayer(mpegVideoReader io.ReadSeeker) (*Player, error) {
	mpg, err := mpeg.New(mpegVideoReader)
	if err != nil { return nil, err }

	// make sure there's video stream and headers
	if mpg.NumVideoStreams() == 0 { return nil, ErrNoVideo }
	if !mpg.HasHeaders() { return nil, ErrMissingHeaders }
	
	// check if there's audio streams
	var controller mpegController
	if mpg.NumAudioStreams() > 0 {
		controller, err = newAudioAdapter(mpg)
		if err != nil { return nil, err }
	} else {
		controller = newVideoOnlyController(mpg)
	}

	// create mpeg player
	img := ebiten.NewImage(mpg.Width(), mpg.Height())
	img.Fill(color.Black)
	return &Player {
		currentFrame: img,
		controller: controller,
	}, nil
}

// Returns the image corresponding to the underlying mpeg's video frame at
// the current [Player.Position](). This means that as long as the mpeg is 
// playing, calling this method at different times will return different
// frames.
//
// The returned image is reused, so calling this method again is likely to
// overwrite its contents. This means you can use the image between calls, but
// you should not store it for later use expecting the image to remain the same.
func (self *Player) CurrentFrame() *ebiten.Image {
	// get mpeg and current position
	mpg, mutex := self.controller.MPEG()
	refSecPos := self.controller.Position().Seconds()
	mutex.Lock()
	defer mutex.Unlock()
	
	// compute frame duration and stop if we are already within the target
	videoDecoder := mpg.Video()
	if videoDecoder.HasEnded() {
		self.copyFrame(nil)
		return self.currentFrame
	}
	frameDuration := 1.0/mpg.Framerate()
	if videoDecoder.Time() + frameDuration >= refSecPos { return self.currentFrame }
	
	// keep decoding until we reach the target
	// TODO: if we have too-big gaps this may be extremely wasteful
	//       and direct seeks may be a better option. THIS IS THE 
	//       BIGGEST ISSUE IN THE PKG TO TEST AND MOST LIKELY CORRECT.
	//       in fact, if decoding takes too long, the controller's
	//       position may keep advancing.
	// TODO: actually looping is a bigger problem.
	
	var frame *mpeg.Frame
	for videoDecoder.Time() + frameDuration <= refSecPos {
		frame = videoDecoder.Decode()
		if videoDecoder.HasEnded() { break }
	}
	
	self.copyFrame(frame)
	return self.currentFrame
}

// Returns the width and height of the video.
func (self *Player) Resolution() (int, int) {
	mpg, mutex := self.controller.MPEG()
	mutex.Lock()
	defer mutex.Unlock()
	return mpg.Width(), mpg.Height()
}

// Returns the video's framerate (how many video frames are used per second).
func (self *Player) FrameRate() float64 {
	mpg, mutex := self.controller.MPEG()
	mutex.Lock()
	defer mutex.Unlock()
	return mpg.Framerate()
}

// Returns whether the video has audio.
func (self *Player) HasAudio() bool {
	_, isAudioAdapter := self.controller.(*audioAdapter)
	return isAudioAdapter
}

// Returns the video duration.
func (self *Player) Duration() time.Duration {
	mpg, mutex := self.controller.MPEG()
	mutex.Lock()
	defer mutex.Unlock()
	return mpg.Duration()
}

// Gets the video's volume. If the video has no audio, 0 will be returned.
func (self *Player) Volume() float64 {
	adapter, isAudioAdapter := self.controller.(*audioAdapter)
	if !isAudioAdapter { return 0 }
	return adapter.audioPlayer.Volume()
}

// Sets the volume of the video. If the video has no audio, this method will panic.
func (self *Player) SetVolume(volume float64) {
	adapter, isAudioAdapter := self.controller.(*audioAdapter)
	if isAudioAdapter {
		adapter.audioPlayer.SetVolume(volume)
	} else {
		panic("attempted to set volume on a video with no audio")
	}
}

func (self *Player) copyFrame(frame *mpeg.Frame) {
	if frame == nil {
		if self.onBlackFrame { return }
		self.currentFrame.Fill(color.Black)
		self.onBlackFrame = true
	} else {
		self.currentFrame.WritePixels(frame.RGBA().Pix)
		self.onBlackFrame = false
	}
}

// ---- methods derived from mpegController ----

// Play() activates the player's playback clock. If the player is already
// playing, it just keeps playing and nothing new happens.
//
// If the underlying mpeg contains any audio, the audio will also
// start or resume. Video frames need to be retrieved manually through
// [Player.CurrentFrame]() instead.
func (self *Player) Play() { self.controller.Play() }

// Returns whether the player's clock and audio are running or not.
// Notice that even when playing, video frames need to be retrieved manually
// through [Player.CurrentFrame]().
func (self *Player) IsPlaying() bool { return self.controller.IsPlaying() }

// Pauses the player's playback clock. If the player is already paused, it
// just stays paused and nothing new happens.
//
// If the underlying mpeg contains any audio, the audio will also be paused.
func (self *Player) Pause() { self.controller.Pause() }

// Rewinds the mpeg streams to the beginning. Behavior may differ from
// [Player.Seek](0) (TODO: figure out the differences in more detail).
func (self *Player) Rewind() { self.controller.Rewind() }

// Moves the player's playback position to the first [intra-frame] that can
// be found before the given target position.
//
// [Player.SeekPrecise]() is more precise but can also be significantly
// slower.
//
// TODO: better document the potential precision loss, common values in practical
// scenarios, or advice on format settings in order to make results more or less
// precise (if intra-frames can be easily configured). Also quantify how much slower
// "slower" means.
//
// [intra-frame]: https://en.wikipedia.org/wiki/Intra-frame_coding
func (self *Player) SeekFast(position time.Duration) {
	self.seek(position, false)
}

// Moves the player's playback position to the given one, relative to the start
// of the video.
//
// This method is more precise than [Player.SeekFast](), but it can also be
// significantly slower: a precise seek has to do an initial fast seek to the
// first [intra-frame] that can be found before the target position, but then
// it also has to keep decoding frames progressively until the actual target
// position is reached.
//
// [intra-frame]: https://en.wikipedia.org/wiki/Intra-frame_coding
func (self *Player) SeekPrecise(position time.Duration) {
	self.seek(position, true)
}

func (self *Player) seek(position time.Duration, sampleLevelPrecision bool) {
	frame := self.controller.SeekFrame(position, sampleLevelPrecision)
	self.copyFrame(frame)
}

// Returns the player's current playback position.
func (self *Player) Position() time.Duration { return self.controller.Position() }

// Returns the player's underlying mpeg and its associated mutex. The mutex
// is particularly critical when the mpeg also contains audio, as the Ebitengine
// audio process may be periodically reading from the mpeg, making concurrent
// access dangerous.
//
// Here's a list of common useful methods on the underlying mpeg:
//  - SetAudioStream(int) to select the audio stream [0 - 4].
//  - SetLoop(bool) to set mpeg looping mode.
//  - Loop() bool to determine if the mpeg is in looping mode.
func (self *Player) MPEG() (*mpeg.MPEG, *sync.Mutex) { return self.controller.MPEG() }
