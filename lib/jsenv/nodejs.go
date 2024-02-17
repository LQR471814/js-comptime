package jsenv

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

//go:embed nodejs-exporter.js
var exporter string

type Nodejs struct {
	Command string
}

func (env Nodejs) Eval(ctx context.Context, code string, results []EvalResult) error {
	conn, err := net.ListenPacket("udp", ":0")
	if err != nil {
		return err
	}
	defer conn.Close()
	_, port, err := net.SplitHostPort(conn.LocalAddr().String())
	if err != nil {
		return err
	}
	executedCode := exporter + code

	err = os.Mkdir(".jscomptime", 0777)
	if err != nil && !os.IsExist(err) {
		return err
	}
	err = os.WriteFile(".jscomptime/code.js", []byte(executedCode), 0777)
	if err != nil {
		return err
	}

	errorc := make(chan error)
	outputc := make(chan eval)

	listenCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	go listenEval(listenCtx, conn, outputc, errorc)

	go func() {
		cmd := exec.Command(env.Command, ".jscomptime/code.js")
		cmd.Env = append(cmd.Env, "JSCOMPTIME_PORT="+port)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		fmt.Println("-------- nodejs comptime code output --------")
		err := cmd.Run()
		if err != nil {
			errorc <- err
		}
	}()

	evaluated := 0
	for evaluated < len(results) {
		select {
		case err := <-errorc:
			if err != nil {
				return err
			}
		case e := <-outputc:
			results[e.id].Result = e.content
			evaluated++
		}
	}

	return nil
}

type eval struct {
	id      int
	content string
}

func listenEval(
	ctx context.Context,
	conn net.PacketConn,
	output chan eval,
	error chan error,
) {
	current := bytes.NewBuffer(nil)
	chunk := make([]byte, 512, 512)
	for {
		select {
		case <-ctx.Done():
			return
		default:
			break
		}

		err := conn.SetReadDeadline(time.Now().Add(time.Second))
		if err != nil {
			error <- err
			return
		}
		n, _, err := conn.ReadFrom(chunk)
		if err != nil {
			error <- err
			return
		}
		_, err = current.Write(chunk[:n])
		if err != nil {
			error <- err
			return
		}

		// is null-terminated
		if n > 0 && string(chunk)[n-1] == 0 {
			text := current.String()
			split := strings.SplitN(text[:len(text)-1], "|", 2)
			id, err := strconv.Atoi(split[0])
			if err != nil {
				error <- err
				return
			}

			output <- eval{
				id:      id,
				content: split[1],
			}
			current.Reset()
		}
	}
}
