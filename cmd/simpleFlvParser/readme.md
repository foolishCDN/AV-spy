## SimpleFlvParser

### Usage
```
usage:
simpleFlvParser input.flv
simpleFlvParser http://path/to/input.flv
```
### output

```
---------- FLV Header ----------
Version: 1
HasVideo: true
HasAudio: true
HeaderSize: 9
------------------------------
{SCRIPT} 0 0
---------- MetaData ----------
([]interface {}) (len=2 cap=2) {
 (string) (len=10) "onMetaData",
 (amf.ECMAArray) (len=16) {
  (string) (len=15) "audiosamplesize": (float64) 16,
  (string) (len=7) "encoder": (string) (len=13) "Lavf58.20.100",
  (string) (len=8) "filesize": (float64) 1.524829e+06,
  (string) (len=12) "videocodecid": (float64) 7,
  (string) (len=6) "stereo": (bool) true,
  (string) (len=12) "audiocodecid": (float64) 2,
  (string) (len=13) "minor_version": (string) (len=1) "0",
  (string) (len=8) "duration": (float64) 15.107,
  (string) (len=5) "width": (float64) 544,
  (string) (len=13) "audiodatarate": (float64) 0,
  (string) (len=15) "audiosamplerate": (float64) 44100,
  (string) (len=17) "compatible_brands": (string) (len=8) "mp42isom",
  (string) (len=6) "height": (float64) 960,
  (string) (len=13) "videodatarate": (float64) 0,
  (string) (len=9) "framerate": (float64) 24.333333333333332,
  (string) (len=11) "major_brand": (string) (len=4) "mp42"
 }
}
------------------------------
{ VIDEO} 0 0 VideoKeyFrame VideoH264
{ AUDIO} 0 0 AudioMP3 AudioStereo Audio16bit Audio44KHz
{ VIDEO} 0 107 VideoKeyFrame VideoH264
{ AUDIO} 0 26 AudioMP3 AudioStereo Audio16bit Audio44KHz
{ AUDIO} 0 52 AudioMP3 AudioStereo Audio16bit Audio44KHz
{ VIDEO} 0 189 VideoInterFrame VideoH264
{ AUDIO} 0 78 AudioMP3 AudioStereo Audio16bit Audio44KHz
{ AUDIO} 0 104 AudioMP3 AudioStereo Audio16bit Audio44KHz
{ VIDEO} 0 148 VideoInterFrame VideoH264
{ AUDIO} 0 131 AudioMP3 AudioStereo Audio16bit Audio44KHz
{ VIDEO} 0 354 VideoInterFrame VideoH264
...
```