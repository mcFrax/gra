package main

import (
	"fmt"
	"os"
	"os/exec"
	"time"
)

type event int
const (
	arrowUp event = iota
	arrowDown
	arrowLeft
	arrowRigth
	quitEvent
)

var (
	arrowMapping = map[byte]event{65: arrowUp, 66: arrowDown, 67: arrowRigth, 68: arrowLeft}
)

func main() {
	exec.Command("stty", "-F", "/dev/tty", "cbreak", "min", "1").Run()
	exec.Command("stty", "-F", "/dev/tty", "-echo").Run()
	defer exec.Command("stty", "-F", "/dev/tty", "-echo").Run()

	inputEvents := make(chan event, 100)
	go func(){
		b := make([]byte, 1)
		arrowState := 0
		for {
			_, err := os.Stdin.Read(b)
			if err == nil {
				if arrowState == 0 && b[0] == 27 {
					arrowState = 1
				} else if arrowState == 1 && b[0] == 91 {
					arrowState = 2
				} else if arrowState == 2 {
					event, matched := arrowMapping[b[0]]
					if matched {
						inputEvents <- event
					}
					arrowState = 0
				} else {
					arrowState = 0
				}
			} else {
				inputEvents <- quitEvent
				break
			}
		}
	}()

	displayWidth, displayHeight := 60, 20
	moveDirection := 1
	xpos := 10
	minx, maxx := 3, displayWidth - 4
	ticker := time.Tick(20 * time.Millisecond)
	
	// create display buffer
	displayContent := make([][]string, displayHeight)
	for iy := range displayContent {
		displayContent[iy] = make([]string, displayWidth)
	}
	
	for {
		eventReading: for {
			select {
			case event := <-inputEvents:
				switch event {
				case arrowLeft:
					moveDirection = -1
				case arrowRigth:
					moveDirection = 1
				case quitEvent:
					fmt.Println("\x1b[1;45m                                                                    \x1b[0m")
					fmt.Println("\x1b[1;43m                                                                    \x1b[0m")
					fmt.Println("\x1b[32m ----------------------------------------------------------------  \x1b[0m")
					return
				}
			default:
				break eventReading  // nie ma więcej zdarzeń
			}
		}
		
		xpos += moveDirection
		switch {
			case xpos < minx: xpos = minx
			case xpos > maxx: xpos = maxx
		}
		
		// clear display buffer (fill with spaces)
		for iy := 0; iy < displayHeight; iy++ {
			for ix := 0; ix < displayWidth; ix++ {
				displayContent[iy][ix] = " "
			}
		}
		
		
		// draw on the buffer
		displayContent[displayHeight-2][xpos] = "\x1b[1;44m^\x1b[0m"
		displayContent[displayHeight-1][xpos-1] = "\x1b[1;44m<"
		displayContent[displayHeight-1][xpos] = "o"
		displayContent[displayHeight-1][xpos+1] = ">\x1b[0m"
		
		
		// print buffer to terminal
		fmt.Print("\x1b[H\x1b[J")
		for iy := 0; iy < displayHeight; iy++ {
			for ix := 0; ix < displayWidth; ix++ {
				fmt.Print(displayContent[iy][ix])
			}
			fmt.Println()
		}
		fmt.Println()
		
		<-ticker  // zaczekaj na kolejny tick
	}
}
