package main

import (
	"fmt"
	"os"
	"os/exec"
	"time"
	"math/rand"
)

type event int
const (
	arrowUp event = iota
	arrowDown
	arrowLeft
	arrowRigth
	fireEvent
	quitEvent
)

var (
	arrowMapping = map[byte]event{65: arrowUp, 66: arrowDown, 67: arrowRigth, 68: arrowLeft}
)

type projectile struct {
	x, y int
}

type bonus struct {
	x, y int
}

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
				if arrowState == 0 {
					switch b[0] {
						case 27: arrowState = 1
						case ' ': inputEvents <- fireEvent
						case 4, 'q': inputEvents <- quitEvent
					}
				} else if arrowState == 1 {
					switch b[0] {
						case 91: arrowState = 2
						case 27: inputEvents <- quitEvent  // second escape in a row
						default: arrowState = 0
					}
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

	displayWidth, displayHeight := 60, 40
	moveDirection := 1
	xpos, ypos := 10, displayHeight-1
	minx, maxx := 3, displayWidth - 4
	hasDoubleGun := false
	
	projectiles := make([]projectile, 0)
	bonuses := make([]bonus, 0)
	
	ticker := time.Tick(20 * time.Millisecond)
	
	// create display buffer
	displayContent := make([][]string, displayHeight)
	for iy := range displayContent {
		displayContent[iy] = make([]string, displayWidth)
	}

	fmt.Println()
	fmt.Println("\x1b[45m                                                                    \x1b[0m")
	fmt.Println("\x1b[45m  \x1b[0;1m                    Welcome here, brave pilot!                  \x1b[45m  \x1b[0m")
	fmt.Println("\x1b[45m                                                                    \x1b[0m")
	
	for {
		eventReading: for {
			select {
			case event := <-inputEvents:
				switch event {
				case arrowLeft:
					moveDirection = -1
				case arrowRigth:
					moveDirection = 1
				case fireEvent:
					if hasDoubleGun {
						projectiles = append(projectiles, projectile{xpos-1, ypos})
						projectiles = append(projectiles, projectile{xpos+1, ypos})
					} else {
						projectiles = append(projectiles, projectile{xpos, ypos-1})
					}
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
		
		leftProjectiles := make([]projectile, 0, len(projectiles)+5)
		for _, projectile := range projectiles {
			projectile.y -= 1
			if projectile.y >= 0 {
				leftProjectiles = append(leftProjectiles, projectile)
			}
		}
		projectiles = leftProjectiles
		
		if rand.Intn(100) == 0 {
			bonuses = append(bonuses, bonus{rand.Intn(displayWidth), -1})
		}
		leftBonuses := make([]bonus, 0, len(bonuses)+5)
		for _, bonus := range bonuses {
			bonus.y += 1
			if bonus.y < displayHeight {
				if bonus.y == ypos && bonus.x - xpos <= 1 && xpos - bonus.x <= 1 {
					hasDoubleGun = true
				} else {
					leftBonuses = append(leftBonuses, bonus)
				}
			}
		}
		bonuses = leftBonuses
		
		// clear display buffer (fill with spaces)
		for iy := 0; iy < displayHeight; iy++ {
			for ix := 0; ix < displayWidth; ix++ {
				displayContent[iy][ix] = " "
			}
		}
		
		
		// draw on the buffer:
		// draw bonuses
		for _, bonus := range bonuses {
			displayContent[bonus.y][bonus.x] = "\x1b[1;32m$\x1b[0m"
		}
		
		// draw projectiles
		for _, projectile := range projectiles {
			displayContent[projectile.y][projectile.x] = "\x1b[1;35m|\x1b[0m"
		}
		
		// draw ship
		displayContent[ypos-1][xpos] = "\x1b[1;44m^\x1b[0m"
		displayContent[ypos][xpos-1] = "\x1b[1;44m<"
		displayContent[ypos][xpos] = "o"
		displayContent[ypos][xpos+1] = ">\x1b[0m"
		
		
		// print buffer to terminal
		for iy := 0; iy < displayHeight; iy++ {
			for ix := 0; ix < displayWidth; ix++ {
				fmt.Print(displayContent[iy][ix])
			}
			fmt.Println()
		}
		fmt.Println()
		
		<-ticker  // zaczekaj na kolejny tick
		
		// clear terminal before next frame
		fmt.Printf("\x1b[%dA\x1b[G", displayHeight+1)
	}
}
