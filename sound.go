package main

import (
	"bytes"
	"fmt"
	"io"
	"os"

	alsa "github.com/cocoonlife/goalsa"
	"github.com/cryptix/wav"
)

func playSound(path string, isEmbedded bool) error {
	var reader io.ReadSeeker
	var size int64

	if isEmbedded {
		// Read embedded file into memory
		data, err := soundFiles.ReadFile(path)
		if err != nil {
			return fmt.Errorf("reading embedded file %s: %v", path, err)
		}
		reader = bytes.NewReader(data)
		size = int64(len(data))
	} else {
		// Open local file
		file, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("opening file %s: %v", path, err)
		}
		defer file.Close()
		reader = file
		info, err := file.Stat()
		if err != nil {
			return fmt.Errorf("stating file %s: %v", path, err)
		}
		size = info.Size()
	}

	// Create WAV reader
	wavReader, err := wav.NewReader(reader, size)
	if err != nil {
		return fmt.Errorf("WAV reader: %v", err)
	}
	if wavReader == nil {
		return fmt.Errorf("nil WAV reader")
	}

	// Get sample rate
	fileinfo := wavReader.GetFile()
	samplerate := int(fileinfo.SampleRate)
	if samplerate == 0 || samplerate > 100000 {
		samplerate = 44100
	}

	// Less verbose for service
	// fmt.Printf("Playing %s (Sample rate: %d)\n", path, samplerate)

	// Open ALSA playback device
	out, err := alsa.NewPlaybackDevice(
		"default",
		1,
		alsa.FormatS16LE,
		samplerate,
		alsa.BufferParams{},
	)
	if err != nil {
		return fmt.Errorf("alsa: %v", err)
	}
	if out == nil {
		return fmt.Errorf("nil ALSA device")
	}
	defer out.Close()

	// Play samples
	for {
		s, err := wavReader.ReadSampleEvery(2, 0)
		var cvert []int16
		for _, b := range s {
			cvert = append(cvert, int16(b))
		}
		if cvert != nil {
			out.Write(cvert)
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("WAV decode: %v", err)
		}
	}

	return nil
}

func getSoundPath(
	localFile,
	userPath,
	embeddedPath string,
) (path string, isEmbedded bool) {
	// 1. Check for local file
	if _, err := os.Stat(localFile); err == nil {
		return localFile, false
	}
	// 2. Check for user-defined path
	if userPath != "" {
		if _, err := os.Stat(userPath); err == nil {
			return userPath, false
		}
		fmt.Printf(
			"Custom sound file %s not found, falling back to embedded\n",
			userPath,
		)
	}
	// 3. Fallback to embedded file
	return embeddedPath, true
}
