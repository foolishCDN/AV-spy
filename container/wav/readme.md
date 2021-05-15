## WAV (WAVE) parser
### Usage

For more info, please see [parser_test.go](https://github.com/foolishCDN/AV-spy/blob/master/container/wav/parser_test.go), which also uses [oto](https://github.com/hajimehoshi/oto) to play wav file.

```Go
    parser := NewParser(player)

    buf := make([]byte, 1024)
    for {
        n, err := f.Read(buf)
        if err != nil {
            if err == io.EOF {
                break
            }
            t.Fatal(err)
        }
        if err := parser.Input(buf[:n]); err != nil {
            t.Fatal(err)
        }
    }
```

### TODO
- Implement Writer
- Support more types of chunk
