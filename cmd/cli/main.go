package main

import (
	"io"
	"jscomptime/lib/comptime"
	"log"
	"os"
)

/*
a comptime value can only be brought into runtime code in only 2 ways:
- a comptime variable is referenced directly (by identifier)
- a comptime function is called
*/

func main() {
	buff, err := io.ReadAll(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}

	err = comptime.Compile(buff)
	if err != nil {
		log.Fatal(err)
	}
}
