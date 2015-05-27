/*

* update the Readme with info on what I decide to use.
Options include:
	- termbox go - https://github.com/nsf/termbox-go - this could lead to a nice cli interface but may be overkill.
	- just using bufio.ReadLine(),
		- Add color to my output, https://github.com/aybabtme/rgbterm or https://github.com/alecthomas/colour
	- Rejected options
		- Wrapping of the readline library, https://github.com/shavac/readline - not universally compatible.
*/
package cli

import (
	"fmt"
)

func CLIInterface(dbFile string) int {
	fmt.Println("CLI not yet implmented")
	return 0
}
