package main

import (
	"strconv"
	"os"
	"path/filepath"
//	"fmt"
	"strings"
	"syscall"
	"time"
	"bytes"
	"log"
	"os/exec"
)

func runCommandWithTimeout(timeout int, command string, args ...string) (stdout, stderr string, isKilled bool, err error) {
	var stdoutBuf, stderrBuf bytes.Buffer
	cmd := exec.Command(command, args...)
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf
	stdin, err := cmd.StdinPipe()
	if err != nil {
		log.Fatal(err)
	}
	cmd.Start()
	_, err = stdin.Write([]byte(`export SHIT=1`))
	if err != nil {
		log.Fatal(err)
	}
	done := make(chan error)
	go func() {
		done <- cmd.Wait()
	}()
	after := time.After(time.Duration(timeout) * time.Millisecond)
	var exitErr error
	select {
		case <- after:
			cmd.Process.Signal(syscall.SIGINT)
			time.Sleep(10*time.Millisecond)
			cmd.Process.Kill()
			isKilled = true
			err = nil
		case exitErr = <- done:
			isKilled = false
			err = exitErr
	}
	stdout = string(bytes.TrimSpace(stdoutBuf.Bytes())) // Remove \n
	stderr = string(bytes.TrimSpace(stderrBuf.Bytes())) // Remove \n
	return
}

func getTaintResult(timeout int, command string, args ...string) (webshellRisk int){
	_, resultErr, isKilled, err := runCommandWithTimeout(timeout, command, args...)
//	log.Print(fmt.Sprintf(`Run: %s %s`, command, strings.Join(args, "/")))
//	log.Print("Is Killed: ", isKilled)
//	log.Print("Res: \n", resultOut)
//	log.Print("Err: \n", resultErr)
	webshellRisk = 0
	if isKilled {
		webshellRisk = -1
	} else {
		if len(resultErr) >= 12 && resultErr[len(resultErr)-12:len(resultErr)-4] == "WEBSHELL" {
			webshellRiskStr := resultErr[len(resultErr)-3:len(resultErr)-1]
			webshellRisk, _ = strconv.Atoi(webshellRiskStr)
		} else {
			webshellRisk = -1
			if (err != nil) {
				log.Print(err)
			}
		}
	}
	return
}

func walk(path string, info os.FileInfo, _ error) error {
	if strings.ToLower(filepath.Ext(path)) != ".php" &&
		strings.ToLower(filepath.Ext(path)) != ".phpt" &&
		strings.ToLower(filepath.Ext(path)) != ".php3" &&
		strings.ToLower(filepath.Ext(path)) != ".php4" &&
		strings.ToLower(filepath.Ext(path)) != ".php5" &&
		strings.ToLower(filepath.Ext(path)) != ".txt" &&
		strings.ToLower(filepath.Ext(path)) != ".bak" {
		return nil
	}
	res := getTaintResult(5000, "php", "-dvld.active=1", "-dvld.execute=1", "-dvld.webshell_test", "-dvld.verbosity=1", "-dvld.noprocess=1", path)
	switch res {
		case -1:
			log.Print("Timeout\t", path)
		case -2:
			log.Print("Error\t", path)	
		default:
			log.Printf("Risk=%d \t%s", res, path)
	}
	return nil
}

func main() {
	filepath.Walk("/Users/cyrus/Dev", walk)
}