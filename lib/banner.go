package lib

import (
	"fmt"
	"log"

	"github.com/fatih/color"
)

const Version = "1.0.2"
const Author = "black5ugar"
const GitHub = "https://github.com/black5ugar/HostCollision"

func Banner() {
	banner := fmt.Sprintf(`

 ██      ██   ███████    ████████ ██████████   ██████    ███████   ██       ██       ██  ████████ ██   ███████   ████     ██
░██     ░██  ██░░░░░██  ██░░░░░░ ░░░░░██░░░   ██░░░░██  ██░░░░░██ ░██      ░██      ░██ ██░░░░░░ ░██  ██░░░░░██ ░██░██   ░██
░██     ░██ ██     ░░██░██           ░██     ██    ░░  ██     ░░██░██      ░██      ░██░██       ░██ ██     ░░██░██░░██  ░██
░██████████░██      ░██░█████████    ░██    ░██       ░██      ░██░██      ░██      ░██░█████████░██░██      ░██░██ ░░██ ░██
░██░░░░░░██░██      ░██░░░░░░░░██    ░██    ░██       ░██      ░██░██      ░██      ░██░░░░░░░░██░██░██      ░██░██  ░░██░██
░██     ░██░░██     ██        ░██    ░██    ░░██    ██░░██     ██ ░██      ░██      ░██       ░██░██░░██     ██ ░██   ░░████
░██     ░██ ░░███████   ████████     ░██     ░░██████  ░░███████  ░████████░████████░██ ████████ ░██ ░░███████  ░██    ░░███
░░      ░░   ░░░░░░░   ░░░░░░░░      ░░       ░░░░░░    ░░░░░░░   ░░░░░░░░ ░░░░░░░░ ░░ ░░░░░░░░  ░░   ░░░░░░░   ░░      ░░░ 

Version: %s 
Author: %s 
Github: %s
	`, Version, Author, GitHub)

	color.HiBlue(banner)
	log.Println(banner)
}
