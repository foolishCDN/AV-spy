## SimpleFlvParser

### Usage
```
Usage:
        simpleFlvParser [options] input.flv
        simpleFlvParser [options] http://path/to/input.flv

Options:
  -show_extradata
        will show codec extradata(sequence header)
  -show_metadata
        will show meta data (default true)
  -show_packets
        will show packets info
```
### output

```
---------- FLV Header ----------
Version: 1
HasVideo: true
HasAudio: true
HeaderSize: 9
------------------------------
---------- MetaData ----------
[ "onMetaData", { "videocodecid": 7, "audiosamplesize": 16, "videodatarate": 0, "audiodatarate": 0, "audiosamplerate": 44100, "major_brand": "mp42", "file
size": 1524829, "duration": 15.107, "height": 960, "framerate": 24.333333333333332, "encoder": "Lavf58.20.100", "audiocodecid": 2, "minor_version": "0", "
compatible_brands": "mp42isom", "width": 544, "stereo": true } ]
------------------------------
{SCRIPT} 0 0
-- sequence header of video --
{ ConfigurationVersion: 1, AVCProfileIndication: 100, ProfileCompatibility: 8, AVCLevelIndication: 31, LengthSizeMinusOne: 3, NumOfSPS: 1, LenOfSPS: 24, S
PS: [ 103, 100, 8, 31, 172, 217, 64, 136, 30, 104, 64, 0, 0, 3, 0, 192, 0, 0, 36, 131, 198, 12, 101, 128 ], NumOfPPS: 1, LenOfPPS: 6, PPS: [ 104, 235, 227
, 203, 34, 192 ] }
------------------------------
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
{ AUDIO} 0 157 AudioMP3 AudioStereo Audio16bit Audio44KHz
...
```