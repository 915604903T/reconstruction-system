package handlers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"fmt"
	"os/exec"
	"syscall"
	"os"
	"io"
	"sync"
	"bufio"

//	"github.com/Fvoiretryzig/faas-wasm/controller"
	"210.28.132.171/ShimeiT/faas-wasm/controller"

	log "github.com/sirupsen/logrus"

	typesv1 "github.com/openfaas/faas-provider/types"

	"github.com/containerd/cgroups"
	specs "github.com/opencontainers/runtime-spec/specs-go"
)

func newController() (*exec.Cmd, io.WriteCloser, io.ReadCloser) {
	cmd := exec.Command("./mycontroller")
	stdin, err := cmd.StdinPipe()
	if err!= nil {
		panic(err)
	}
	stdout, err := cmd.StdoutPipe()
	if err!=nil {
		panic(err)
	}
	err = cmd.Start()
	if err!= nil {
		panic(err)
	}
//	log.Infof("controller success Start")
	return cmd, stdin, stdout
}
// MakeDeployHandler creates a handler to create new functions in the cluster
func MakeDeployHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		log.Info("deployment request")
		defer r.Body.Close()

		body, _ := ioutil.ReadAll(r.Body)
		request := typesv1.FunctionDeployment{}
		if err := json.Unmarshal(body, &request); err != nil {
			log.Errorln("error during unmarshal of create function request. ", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		log.Infof("image: %s\n", request.Image)

		function_name := request.Service
		function_image := request.Image
	/*===========Start controller program============*/
		controller_proc, c_stdin, c_stdout:= newController()
		defer c_stdin.Close()
		stdouts[function_name] = &c_stdout
		pid := controller_proc.Process.Pid
	/*============Add controller to the first core===========*/
		shares := uint64(1024)
		control, err := cgroups.New(cgroups.V1, cgroups.StaticPath("/"+function_name), &specs.LinuxResources{
            CPU: &specs.LinuxCPU{
                Shares: &shares,
                Cpus: "0",  //initially choose the first core
            },
        })
		if err!=nil {
			panic(err)
		}
//		log.Infof("create cgroups successfully")
		control.Add(cgroups.Process{Pid: pid})
//		log.Infof("add %d to this cgroups", pid)
	/*Create fifo file and Use stdin Pipe to let controller get fifo name*/
		pipeFile := "./pipe/" + function_name + ".pipe"
		os.Remove(pipeFile)
		err = syscall.Mkfifo(pipeFile, 0666)
		if err != nil {
			log.Fatal("Make named pipe file error: ", err)
		}
//		log.Infof("after create pipefile before write to stdin")
		c_stdin.Write([]byte(pipeFile + "\n"))
//		log.Infof("after write to controller pipefile name")

	/*============Write to fifo file============*/
		var pipe_lock sync.Mutex
		locks[function_name] = pipe_lock
		pipe_lock.Lock()
		file, err := os.OpenFile(pipeFile, os.O_WRONLY, 0777)
		if err != nil {
			log.Fatalf("deploy opening file: %v", err)
		}
		file.WriteString(fmt.Sprintf("deploy %s %s\n", function_name, function_image))
		pipe_lock.Unlock()
	/*==========Read from controller stdout=========*/
		//go func() {
		buff := bufio.NewScanner(c_stdout)
		for buff.Scan(){
			tmp := buff.Text()
			//fmt.Println(tmp)
			if tmp=="end"{
				break
			}else {
				fmt.Println(tmp)
			}
		}
		//}()
	/*==========Fill infomation========*/
		pchan[function_name] = make(chan bool, 1)
		proc_cgroups[function_name] = &control
		func_controller := &controller.Controller{
							Process: controller_proc,
							Invoke_cnt: 0,
							Name: function_name,
							State: controller.IDLE,
		}
		process[function_name] = func_controller
		functions[request.Service] = requestToStatus(request)
		log.Infof("deployment request for function %s", request.Service)
		w.WriteHeader(http.StatusOK)
	}
}
