package jsenv

import (
	"os"
	"os/exec"
)

type Node struct {
	Command string
}

func (n Node) Eval(code string, results EvalResult) error {
	err := os.Mkdir(".jscomptime", 0777)
	if err != nil {
		return err
	}

	err = os.WriteFile(".jscomptime/code.js", []byte(code), 0777)
	if err != nil {
		return err
	}

    cmd := exec.Command(n.Command, ".jscomptime/code.js")
    err = cmd.Run()
    if err != nil {
        return err
    }

    return nil
}
