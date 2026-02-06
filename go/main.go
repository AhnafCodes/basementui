package main

import (
	"basement/basement"
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

func main() {
	info, err := os.Stdin.Stat()

	if len(os.Args) > 1 {
		if os.Args[1] == "-h" || os.Args[1] == "--help" {
			demo()
			return
		}
		input := strings.Join(os.Args[1:], " ")
		fmt.Println(basement.Parse(input))
	} else if err == nil && (info.Mode() & os.ModeCharDevice) == 0 {
		reader := bufio.NewReader(os.Stdin)
		var builder strings.Builder
		for {
			line, err := reader.ReadString('\n')
			builder.WriteString(line)
			if err == io.EOF {
				break
			}
		}
		input := builder.String()
		fmt.Print(basement.Parse(input))
	} else {
		fmt.Fprintln(os.Stderr, "Usage: basement <markdown> or pipe input")
	}
}

func demo() {
	output := basement.Parse(`
# # Bringing MD Like Syntax To Bash Shell
It should be something as *** easy ***
and as ___ natural ___ as writing text.

> > Keep It Simple

Is the idea

  *  * behind
  *  * all this

~~~ striking ~~~ UX also for ` + "`shell`" + ` users.
--- dimmed ---, !!! inverted !!!, or ::: blinking ::: too!
`)
	output += "- - -" + basement.Parse("- - -")
	output += "\n./basement #green(" + basement.Parse("#green(v0.1.2)") + ")"
	fmt.Print(output + "\n\n")
}
