## flv muxer, demuxer/parser
### Usage
Please refer to [flv_test.go](https://github.com/foolishCDN/AV-spy/blob/master/container/flv/flv_test.go) and [simpleFlvParser](https://github.com/foolishCDN/AV-spy/tree/master/cmd/simpleFlvParser) for usage.
#### muxer
```Go
...
muxer := new(Muxer)
if err := muxer.WriteHeader(w, header.HasAudio, header.HasVideo); err != nil {
    log.Fatal(err)
}

for {
    ...
    if err := muxer.WriteTag(w, tag); err != nil {
        log.Fatal(err)
    }
    ...
}
...
```
#### demuxer
```Go
...
demuxer := new(Demuxer)
header, err := demuxer.ReadHeader(f)
if err != nil {
    log.Fatalf("read header err, %v", err)
}

for {
    tag, err := demuxer.ReadTag(f)
    if err != nil {
        if err != io.EOF {
            log.Fatalf("read tag err, %v", err)
        } else {
            break
        }
    }
    ...
}
...
```
#### parser
```Go
...
parser := NewParser(func(tag TagI) error {
    fmt.Println(tag.Info())
    return nil
})

for {
    var b = make([]byte, 1024)
    n, err := r.Read(b)
    if err != nil {
        if err != io.EOF {
            log.Fatalf("read tag err, %v", err)
        } else {
            break
        }
    }
    if err := parser.Input(b[:n]); err != nil {
        log.Fatal(err)
    }
}
...
```