package controller

import (
	"os/exec"
)
const IDLE int = 0
const READY int = 1
const RUNNING int = 2
const WAITING int = 3

type Controller struct {
	Process *exec.Cmd
	Invoke_cnt int
	Name string
	State int
}

