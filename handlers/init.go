package handlers

import (
	"fmt"
	"runtime"
	"os"
	"time"
	"sync"

	"210.28.132.171/ShimeiT/faas-wasm/controller"
	typesv1 "github.com/openfaas/faas-provider/types"
	"io"
	"github.com/containerd/cgroups"
)
var CPU_NUM int = runtime.NumCPU()
var MAX_CON_PROC int = runtime.NumCPU() - 1

var process = map[string]*controller.Controller{}
var functions = map[string]*typesv1.FunctionStatus{}
var proc_cgroups = map[string]*cgroups.Cgroup{}

var Stimer *time.Timer
var time_slice time.Duration
var proc_finish chan bool   //is it necessay?????TODO()
var running = map[string]*controller.Controller{}//controller running function instance
var wait_queue []*controller.Controller
var pchan = map[string]chan bool{}	//schedule->proxy, allow function to run
//var pchan chan bool
//var overChan = map[string]chan bool //proxy->schedule, overtime function pause
var overChan chan bool
var overTimer = map[string]*time.Timer{}
var overtime time.Duration

var stdouts = map[string]*io.ReadCloser{}	//read controller outs
var locks = map[string]sync.Mutex{}	//pipe lock
var wait_queue_lock sync.Mutex
//var remain_cores int
func Init() {
	if CPU_NUM == 1{
		fmt.Println("ERROR: not support single cpu core")
		os.Exit(1)
	}
	fmt.Println("CPU_NUM:",CPU_NUM)
	time_slice = time.Millisecond * 100

	Stimer = time.NewTimer(time_slice)
	overtime = time.Second * 3

	overChan = make(chan bool, 1)
	proc_finish = make(chan bool , 1)
}
