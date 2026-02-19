package main

import "os/exec"

func openURL(url string) error {
	return exec.Command("open", url).Run()
}
