package main

import "testing"

func TestSniffAudio(t *testing.T) {
	for _, c := range []struct {
		data []byte
		ext  string
		ok   bool
	}{
		{append([]byte("RIFF\x24\x00\x00\x00WAVEfmt "), make([]byte, 16)...), "wav", true},
		{append([]byte("OggS\x00\x02"), make([]byte, 20)...), "ogg", true},
		{append([]byte("ID3\x04\x00"), make([]byte, 20)...), "mp3", true},
		{append([]byte{0xFF, 0xFB, 0x90, 0x00}, make([]byte, 20)...), "mp3", true},
		{[]byte("GIF89a...."), "", false},
		{[]byte{0xFF, 0xE0}, "", false},
		{[]byte("RIFF\x24\x00\x00\x00AVI "), "", false},
	} {
		ext, ok := sniffAudio(c.data)
		if ok != c.ok || ext != c.ext {
			t.Fatalf("%q: got %s %v", c.data[:4], ext, ok)
		}
	}
}
