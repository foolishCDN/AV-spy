**Note: Now only support FLV**

## Usage

### AV-spy
AV-Spy -- a simple interactive tool to analysis Media data
![screencast](asset/screencast.gif)
You can input the http-flv url in Terminal UI, or
```
AV-spy -i <url>
```

### simpleFlvParser
SimpleFlvParser is a simple tool to parse FLV stream

```
SimpleFlvParser is a simple tool to parse FLV stream

Usage:
  simpleFlvParser ...[flags] <file path of http url> ...[flags]

Flags:
      --diff_threshold int   when the diff between the real fps(using time) and the fps(using timestamp) is less than this threshold(percent), it is considered that all cache have been received (default 5)
  -f, --format string        output format (default "normal")
  -H, --header strings       http request header
  -h, --help                 help for simpleFlvParser
  -k, --insecure_tls         insecure TLS connection
  -L, --location             follow 302
  -n, --number n             show n packets (no limit if n<=0)
      --server_name string   server name for TLS handshake
      --show                 will show all message
      --show_extradata       will show codec extradata(sequence header)
      --show_header          will show flv file header
      --show_metadata        will show meta data
      --show_packets         will show packets info
  -t, --timeout int          timeout for http request(seconds) (default 10)
  -v, --verbose              verbose output
```

Output
```
$ simpleFlvParser -n 10 --show test.flv
---------- FLV Header ----------
Version: 1
HasVideo: true
HasAudio: true
HeaderSize: 9
------------------------------
---------- MetaData ----------
[
  "onMetaData",
  {
    "audiocodecid": 2,
    "audiodatarate": 0,
    "audiosamplerate": 44100,
    "audiosamplesize": 16,
    "compatible_brands": "mp42isom",
    "duration": 15.107,
    "encoder": "Lavf58.20.100",
    "filesize": 1524829,
    "framerate": 24.333333333333332,
    "height": 960,
    "major_brand": "mp42",
    "minor_version": "0",
    "stereo": true,
    "videocodecid": 7,
    "videodatarate": 0,
    "width": 544
  }
]
------------------------------
        StreamID     PTS     DTS    Size
SCRIPT       0       0       0     362
-- sequence header of video --
{
  ConfigurationVersion: 0x1,
  AVCProfileIndication: 0x64,
  ProfileCompatibility: 0x8,
  AVCLevelIndication: 0x1f,
  LengthSizeMinusOne: 0x3,
  SPS: [
    [
      0000 67  64  08  1f  ac  d9  40  88  1e  68  40  00  00  03  00  c0    'gd....@..h@.....'
      0016 00  00  24  83  c6  0c  65  80                                    '..$...e.'
    ]
  ],
  PPS: [
    [
      0000 68  eb  e3  cb  22  c0                                            'h...".'
    ]
  ],
  ChromaFormat: 0x0,
  BitDepthLumaMinus8: 0x0,
  BitDepthChromaMinus8: 0x0,
  NumOfSPSExt: 0x0,
  SPSExt: []
}
-- From SPS --
width x height: 96x16
fps: NaN (It's not mandatory)
------------------------------
   AVC       0       0       0      41 KeyFrame H264
 AUDIO       0       0       0     417 MP3 Stereo 16bit 44KHz
 VIDEO       0     107      25   10682 KeyFrame H264
 AUDIO       0      26      26     418 MP3 Stereo 16bit 44KHz
 AUDIO       0      52      52     418 MP3 Stereo 16bit 44KHz
 VIDEO       0     189      66    1703 InterFrame H264
 AUDIO       0      78      78     418 MP3 Stereo 16bit 44KHz
 AUDIO       0     104     104     418 MP3 Stereo 16bit 44KHz
 VIDEO       0     148     107    2903 InterFrame H264
 AUDIO       0     131     131     418 MP3 Stereo 16bit 44KHz

Summary:
video: 3/82/512.1µs, fps: 36.59, real fps: 5858.23, gap: 41, rewind: 0, duplicate: 0
audio: 6/131/512.1µs, pps: 45.80, real pps: 11716.46, gap: 27, rewind: 0, duplicate: 0
cache: 82(not yet over) was send within 512.1µs
```
## Install
```
go install github.com/foolishCDN/AV-spy/cmd/AV-spy@latest
```
or 
```
go install github.com/foolishCDN/AV-spy/cmd/simpleFlvParser@latest
```


We want to be a human spy to get the secrets of the media world.
You say that the media world is built by humans. 

What? ... (°ー°〃) 