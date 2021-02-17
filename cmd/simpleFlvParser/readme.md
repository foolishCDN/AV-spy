## SimpleFlvParser

### Usage
```
Usage:
        simpleFlvParser [options] input.flv
        simpleFlvParser [options] http://path/to/input.flv

Options:
  -n n
        show n tags, default: no limit
  -show
        will show all message
  -show_extradata
        will show codec extradata(sequence header)
  -show_header
        will show flv file header
  -show_metadata
        will show meta data
  -show_packets
        will show packets info
```
### output

```
> simpleFlvParser -n 10 -show .\test.flv

---------- FLV Header ----------
Version: 1
HasVideo: true
HasAudio: true
HeaderSize: 9
------------------------------
---------- MetaData ----------
[ "onMetaData", { "width": 544, "videodatarate": 0, "minor_version": "0", "filesize": 1524829, "audiodatarate": 0, "audiosamplesize": 16, "encoder": "Lavf58.20.100", "audiocodecid": 2,
 "compatible_brands": "mp42isom", "duration": 15.107, "height": 960, "framerate": 24.333333333333332, "videocodecid": 7, "audiosamplerate": 44100, "stereo": true, "major_brand": "mp42" } ]
------------------------------
        StreamID     PTS     DTS
{SCRIPT}       0       0
-- sequence header of video --
{ ConfigurationVersion: 1, AVCProfileIndication: 100, ProfileCompatibility: 8, AVCLevelIndication: 31, LengthSizeMinusOne: 3, NumOfSPS: 1, SPS: [ [ 103, 100, 8, 31, 172, 217, 64, 136, 30, 104, 64, 0, 0, 3, 0, 192, 0, 0, 36, 131, 198, 12, 101, 128 ] ], NumOfPPS: 1, PPS: [ [ 104, 235, 227, 203, 34, 192 ] ] }
------------------------------
{   AVC}       0       0       0 KeyFrame H264
{ AUDIO}       0       0       0 MP3 Stereo 16bit 44KHz
{ VIDEO}       0     107      25 KeyFrame H264
{ AUDIO}       0      26      26 MP3 Stereo 16bit 44KHz
{ AUDIO}       0      52      52 MP3 Stereo 16bit 44KHz
{ VIDEO}       0     189      66 InterFrame H264
{ AUDIO}       0      78      78 MP3 Stereo 16bit 44KHz
{ AUDIO}       0     104     104 MP3 Stereo 16bit 44KHz
{ VIDEO}       0     148     107 InterFrame H264
{ AUDIO}       0     131     131 MP3 Stereo 16bit 44KHz
...
```