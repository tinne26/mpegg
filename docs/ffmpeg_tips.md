# ffmpeg tips

Converting videos to mpeg and using ffmpeg can be tricky, so here's a list of useful commands and tips.

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

Set **framerate** explicitly (`-r 30` or `-filter:v fps=VALUE` (don't go below 24!)).
```
ffmpeg -i my_video.mp4 -c:v mpeg1video -c:a mp2 -f mpeg -r 24 my_video.mpg
```

Change the video **resolution** while converting (`-s 640x480`):
```
ffmpeg -i my_video.mp4 -c:v mpeg1video -c:a mp2 -f mpeg -s 640x480 my_video.mpg
```

Keep only video, **remove audio** (`-an`):
```
ffmpeg -i my_video.mp4 -c:v mpeg1video -an -f mpeg my_video.mpg
```

Two-pass encoding (can reduce artifacts, improving quality; `-b:v 1600k` is close to `-q 8`, `-b:v 2700k` is close to `-q 4`):
```
ffmpeg -i my_video.mp4 -c:v mpeg1video -c:a mp2 -f mpeg -b:v 1600k -pass 1 my_video_P1.mpg
ffmpeg -i my_video.mp4 -c:v mpeg1video -c:a mp2 -f mpeg -b:v 1600k -pass 2 my_video_P2.mpg
```

Show video info (`bitrate`, `fps` and `size` are probably the main values you care about):
```
ffprobe -i my_video.mpg -hide_banner
```
