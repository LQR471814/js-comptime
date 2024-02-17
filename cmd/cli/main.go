package main

import (
	"context"
	"fmt"
	"io"
	"jscomptime/lib/comptime"
	"jscomptime/lib/jsenv"
	"log"
	"os"
)

/*
a comptime value can only be brought into runtime code in only 2 ways:
- a comptime variable is referenced directly (by identifier)
- a comptime function is called
*/

func main() {
	log.SetFlags(log.Ltime | log.Lshortfile)

	buff, err := io.ReadAll(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}

	env := jsenv.Nodejs{
		Command: "node",
	}

	compiled, err := comptime.Compile(context.Background(), buff, env)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(compiled)
}
