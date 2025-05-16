**Note: Now only support FLV**

This repo provides two command tool (**AV-spy** and **simpleFlvParser**) for analyzing media data.

## Usage

### AV-spy
AV-spy -- a simple interactive tool to analysis Media data
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
      --sei_format string    how to show SEI (default "hex")
      --server_name string   server name for TLS handshake
      --show                 will show all message
      --show_extradata       will show codec extradata(sequence header)
      --show_header          will show flv file header
      --show_metadata        will show meta data
      --show_packets         will show packets info
      --show_sei             will show SEI(Supplemental Enhancement Information)
  -t, --timeout int          timeout for http request(seconds) (default 10)
  -v, --verbose              verbose output
```

Output
```
$ simpleFlvParser -n 20 --show test.flv
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
resolution: 544x960
fps: 24.33 (It's not mandatory)
------------------------------
   AVC       0       0       0      41 KeyFrame H264 AVCC [7 8]
 AUDIO       0       0       0     417 MP3 Stereo 16bit 44KHz
------------- SEI ------------
"payload type: 5"
"payload size: 667"
[
  0000 dc  45  e9  bd  e6  d9  48  b7  96  2c  d8  20  d9  23  ee  ef    '.E....H..,. .#..'
  0016 78  32  36  34  20  2d  20  63  6f  72  65  20  31  35  37  20    'x264 - core 157 '
  0032 2d  20  48  2e  32  36  34  2f  4d  50  45  47  2d  34  20  41    '- H.264/MPEG-4 A'
  0048 56  43  20  63  6f  64  65  63  20  2d  20  43  6f  70  79  6c    'VC codec - Copyl'
  0064 65  66  74  20  32  30  30  33  2d  32  30  31  39  20  2d  20    'eft 2003-2019 - '
  0080 68  74  74  70  3a  2f  2f  77  77  77  2e  76  69  64  65  6f    'http://www.video'
  0096 6c  61  6e  2e  6f  72  67  2f  78  32  36  34  2e  68  74  6d    'lan.org/x264.htm'
  0112 6c  20  2d  20  6f  70  74  69  6f  6e  73  3a  20  63  61  62    'l - options: cab'
  0128 61  63  3d  31  20  72  65  66  3d  33  20  64  65  62  6c  6f    'ac=1 ref=3 deblo'
  0144 63  6b  3d  31  3a  30  3a  30  20  61  6e  61  6c  79  73  65    'ck=1:0:0 analyse'
  0160 3d  30  78  33  3a  30  78  31  31  33  20  6d  65  3d  68  65    '=0x3:0x113 me=he'
  0176 78  20  73  75  62  6d  65  3d  37  20  70  73  79  3d  31  20    'x subme=7 psy=1 '
  0192 70  73  79  5f  72  64  3d  31  2e  30  30  3a  30  2e  30  30    'psy_rd=1.00:0.00'
  0208 20  6d  69  78  65  64  5f  72  65  66  3d  31  20  6d  65  5f    ' mixed_ref=1 me_'
  0224 72  61  6e  67  65  3d  31  36  20  63  68  72  6f  6d  61  5f    'range=16 chroma_'
  0240 6d  65  3d  31  20  74  72  65  6c  6c  69  73  3d  31  20  38    'me=1 trellis=1 8'
  0256 78  38  64  63  74  3d  31  20  63  71  6d  3d  30  20  64  65    'x8dct=1 cqm=0 de'
  0272 61  64  7a  6f  6e  65  3d  32  31  2c  31  31  20  66  61  73    'adzone=21,11 fas'
  0288 74  5f  70  73  6b  69  70  3d  31  20  63  68  72  6f  6d  61    't_pskip=1 chroma'
  0304 5f  71  70  5f  6f  66  66  73  65  74  3d  2d  32  20  74  68    '_qp_offset=-2 th'
  0320 72  65  61  64  73  3d  31  32  20  6c  6f  6f  6b  61  68  65    'reads=12 lookahe'
  0336 61  64  5f  74  68  72  65  61  64  73  3d  32  20  73  6c  69    'ad_threads=2 sli'
  0352 63  65  64  5f  74  68  72  65  61  64  73  3d  30  20  6e  72    'ced_threads=0 nr'
  0368 3d  30  20  64  65  63  69  6d  61  74  65  3d  31  20  69  6e    '=0 decimate=1 in'
  0384 74  65  72  6c  61  63  65  64  3d  30  20  62  6c  75  72  61    'terlaced=0 blura'
  0400 79  5f  63  6f  6d  70  61  74  3d  30  20  63  6f  6e  73  74    'y_compat=0 const'
  0416 72  61  69  6e  65  64  5f  69  6e  74  72  61  3d  30  20  62    'rained_intra=0 b'
  0432 66  72  61  6d  65  73  3d  33  20  62  5f  70  79  72  61  6d    'frames=3 b_pyram'
  0448 69  64  3d  32  20  62  5f  61  64  61  70  74  3d  31  20  62    'id=2 b_adapt=1 b'
  0464 5f  62  69  61  73  3d  30  20  64  69  72  65  63  74  3d  31    '_bias=0 direct=1'
  0480 20  77  65  69  67  68  74  62  3d  31  20  6f  70  65  6e  5f    ' weightb=1 open_'
  0496 67  6f  70  3d  30  20  77  65  69  67  68  74  70  3d  32  20    'gop=0 weightp=2 '
  0512 6b  65  79  69  6e  74  3d  32  35  30  20  6b  65  79  69  6e    'keyint=250 keyin'
  0528 74  5f  6d  69  6e  3d  32  34  20  73  63  65  6e  65  63  75    't_min=24 scenecu'
  0544 74  3d  34  30  20  69  6e  74  72  61  5f  72  65  66  72  65    't=40 intra_refre'
  0560 73  68  3d  30  20  72  63  5f  6c  6f  6f  6b  61  68  65  61    'sh=0 rc_lookahea'
  0576 64  3d  34  30  20  72  63  3d  63  72  66  20  6d  62  74  72    'd=40 rc=crf mbtr'
  0592 65  65  3d  31  20  63  72  66  3d  32  33  2e  30  20  71  63    'ee=1 crf=23.0 qc'
  0608 6f  6d  70  3d  30  2e  36  30  20  71  70  6d  69  6e  3d  30    'omp=0.60 qpmin=0'
  0624 20  71  70  6d  61  78  3d  36  39  20  71  70  73  74  65  70    ' qpmax=69 qpstep'
  0640 3d  34  20  69  70  5f  72  61  74  69  6f  3d  31  2e  34  30    '=4 ip_ratio=1.40'
  0656 20  61  71  3d  31  3a  31  2e  30  30  00                        ' aq=1:1.00.'
]
------------------------------
 VIDEO       0     107      25   10682 KeyFrame H264 AVCC [6 5]
 AUDIO       0      26      26     418 MP3 Stereo 16bit 44KHz
 AUDIO       0      52      52     418 MP3 Stereo 16bit 44KHz
 VIDEO       0     189      66    1703 InterFrame H264 AVCC [1]
 AUDIO       0      78      78     418 MP3 Stereo 16bit 44KHz
 AUDIO       0     104     104     418 MP3 Stereo 16bit 44KHz
 VIDEO       0     148     107    2903 InterFrame H264 AVCC [1]
 AUDIO       0     131     131     418 MP3 Stereo 16bit 44KHz
 VIDEO       0     354     148    5249 InterFrame H264 AVCC [1]
 AUDIO       0     157     157     418 MP3 Stereo 16bit 44KHz
 AUDIO       0     183     183     418 MP3 Stereo 16bit 44KHz
 VIDEO       0     272     189     863 InterFrame H264 AVCC [1]
 AUDIO       0     209     209     418 MP3 Stereo 16bit 44KHz
 VIDEO       0     230     230     171 InterFrame H264 AVCC [1]
 AUDIO       0     235     235     418 MP3 Stereo 16bit 44KHz
 AUDIO       0     261     261     418 MP3 Stereo 16bit 44KHz
 VIDEO       0     313     272    2965 InterFrame H264 AVCC [1]
 AUDIO       0     287     287     418 MP3 Stereo 16bit 44KHz

Summary:
  Running time: 12.086ms
  video:
    resolution: 544x960, codec: avc, fps: 24.33 (from sps)
    count/timestamp: 7/247, fps: 28.34, real fps: 579.18, gap: 42, rewind: 0, duplicate: 0
    Estimated cache: 247(not yet over) was send within 12.086ms
  audio:
    count/timestamp: 12/287, pps: 41.81, real pps: 950.82, gap: 27, rewind: 0, duplicate: 0
    Estimated cache: 287(not yet over) was send within 12.6207ms
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
