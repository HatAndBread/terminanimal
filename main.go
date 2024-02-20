package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"slices"
	"time"
)

type frames [][][]interface{}

var defaultBGColor uint8
var frameTime int
var isInfinite bool

func main() {
	frameTimePtr := flag.Int("f", 100, "Time between frames in milliseconds")
	bgColorPtr := flag.Int("bg", 0, "Background color")
	infinityPtr := flag.Bool("i", false, "Infinite loop")
	flag.Parse()
	frameTime = *frameTimePtr
	defaultBGColor = uint8(*bgColorPtr)
	isInfinite = *infinityPtr

	args := flag.Args()
	if len(args) == 0 {
		fmt.Println("No filename provided.")
		return
	}
	filename := args[0]
	f, err := getFrames(filename)
	if err != nil {
		fmt.Println(err)
		return
	}
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for sig := range c {
			// clean up on ctrl+c
			if sig == os.Interrupt {
				cleanup()
				os.Exit(0)
			}
		}
	}()

	animate(f)
}

func cleanup() {
	fmt.Print("\033[0m\033[H\033[2J\033[?25h")
}

func animate(frames frames) {
	for _, frame := range frames {
		time.Sleep(time.Duration(frameTime) * time.Millisecond)
		//clear screen
		fmt.Print("\033[0m\033[H\033[2J\033[?25l")

	rowloop:
		for rowIndex, row := range frame {
			if rowIndex%2 == 1 {
				continue rowloop
			}
			str := ""
			var nextRow []interface{}
			if rowIndex != len(frame)-1 {
				nextRow = frame[rowIndex+1]
			}

			lengths := []int{len(row), len(nextRow)}
			maxPixelCount := slices.Max(lengths)
			for i := 0; i < maxPixelCount; i++ {
				var topColor, bottomColor uint8
				if maxPixelCount > len(row) {
					topColor = defaultBGColor
				} else {
					number, ok := row[i].(float64)
					if ok && number < 256 && number >= 0 {
						topColor = uint8(number)
					} else {
						topColor = defaultBGColor
					}
				}
				if maxPixelCount > len(nextRow) {
					bottomColor = defaultBGColor
				} else {
					number, ok := nextRow[i].(float64)
					if ok && number < 256 && number >= 0 {
						bottomColor = uint8(number)
					} else {
						bottomColor = defaultBGColor
					}
				}
				str += fmt.Sprintf("\033[0m\033[48;5;%d;38;5;%dm▄", topColor, bottomColor)
			}
			fmt.Println(str)
		}
	}
	if isInfinite {
		animate(frames)
	} else {
		cleanup()
	}
}

func getFrames(filename string) (frames, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	reader := io.Reader(file)
	var f frames
	decoder := json.NewDecoder(reader)
	decoder.DisallowUnknownFields()
	err = decoder.Decode(&f)
	if err != nil {
		return nil, err
	}
	return f, nil
}
