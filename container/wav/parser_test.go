package wav

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/davecgh/go-spew/spew"

	"github.com/sikasjc/AV-spy/encoding/riff"

	"github.com/sikasjc/AV-spy/utils"

	"github.com/hajimehoshi/oto"
)

func TestParser(t *testing.T) {
	//path := "forest.wav"
	path := "../../encoding/riff/test.wav"
	absPath, err := filepath.Abs(path)
	if err != nil {
		t.Fatal(err)
	}
	f, err := os.Open(absPath)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			t.Error(err)
		}
	}()
	player := new(Player)
	player.Close()

	parser := NewParser(player)

	defer utils.Elapsed("play")()
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
}

type Player struct {
	context *oto.Context
	player  *oto.Player
}

func (p *Player) OnFormat(f *Format) error {
	fmt.Printf("%+v\n", f)
	context, err := oto.NewContext(
		int(f.SampleRate),
		int(f.NumOfChannels),
		int(f.BitPerSample/8),
		int(f.SampleRate/8),
	)
	if err != nil {
		return fmt.Errorf("player: new context err %v", err)
	}
	p.context = context
	p.player = context.NewPlayer()
	return nil
}

func (p *Player) OnPCM(data []byte) error {
	if p.player == nil {
		return errors.New("player: no player")
	}
	if _, err := p.player.Write(data); err != nil {
		return fmt.Errorf("player: write to player err %v", err)
	}
	return nil
}

func (p *Player) OnMIDISample(sample *MIDISample) error {
	spew.Dump(sample)
	return nil
}

func (p *Player) OnUnknownChunk(chunk *riff.Chunk) error {
	fmt.Printf("player: unknown chunk id: %q size: %d length %d\n",
		chunk.ID, chunk.Size, len(chunk.Data))
	return nil
}

func (p *Player) Close() {
	if p.player != nil {
		if err := p.player.Close(); err != nil {
			log.Println(err)
		}
	}
	if p.context != nil {
		if err := p.context.Close(); err != nil {
			log.Println(err)
		}
	}
}
