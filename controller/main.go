package controller

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"strconv"
	"strings"

	//	"time"

	log "github.com/sirupsen/logrus"
	wasmer "github.com/wasmerio/wasmer-go/wasmer"
)

var engine *wasmer.Engine
var store *wasmer.Store
var module *wasmer.Module
var wasiEnv *wasmer.WasiEnvironment
var importObject *wasmer.ImportObject
var invokeTimes = 0
var active = 0

//var first = true
func handleCmd(s string) (string, string, string) {
	in := strings.Fields(s)
	cnt := len(in)
	if cnt == 3 {
		return in[0], in[1], in[2]
	} else if cnt == 2 {
		return in[0], in[1], ""
	}
	return "", "", ""
}
func deployFunction(name string, image string) {
	wasmBytes, _ := ioutil.ReadFile(image)
	engine := wasmer.NewEngine()
	store = wasmer.NewStore(engine)
	module, _ = wasmer.NewModule(store, wasmBytes)
	/*=======init wasi environment=======*/
	preopenDir := "./file/" + name
	err := os.MkdirAll(preopenDir, 644)
	if err != nil {
		log.Fatal("%s file already existed!!")
	}
	wasiEnv, err = wasmer.NewWasiStateBuilder(name).
		PreopenDirectory("./file/" + name).
		CaptureStdout().
		Finalize()
	if err != nil {
		panic(err)
	}
	importObject, err = wasiEnv.GenerateImportObject(store, module)
	if err != nil {
		panic(err)
	}
	/*======add cuda support=======*/
	/*	cudaEnv := wasmer.NewCudaEnvironment()
		err = cudaEnv.AddImportObject(store, importObject)
		if err != nil {
		    panic(err)
		}
	*/
	//    fmt.Println("deployment finish")
	fmt.Println("end")
}
func invokeFunction(name string) {
	invoke := invokeTimes + 1
	invokeTimes += 1
	active++
	instance, _ := wasmer.NewInstance(module, importObject)
	function, _ := instance.Exports.GetRawFunction("_start")
	result, err := function.Call()
	fmt.Println("result:", result, err)
	fmt.Println(string(wasiEnv.ReadStdout()))
	active--
	if active == 0 {
		runtime.GC()
		fmt.Println("after gc")
	}
	fmt.Println("end" + strconv.Itoa(invoke))
}
func deleteFunction(name string) {
	fmt.Println("in delete function")
	engine = nil
	store = nil
	module = nil
	os.RemoveAll("./file/" + name)
	fmt.Println("delete finish")
}
func main() {
	var pipeFile string
	fmt.Scanf("%s", &pipeFile) //get pipeFile name
	fmt.Println("pipeFile: ", pipeFile)
	file, err := os.OpenFile(pipeFile, os.O_RDONLY, os.ModeNamedPipe)
	if err != nil {
		log.Fatal("open named pipe file error: ", err)
	}
	reader := bufio.NewReader(file)
	flag := false
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			log.Fatal("read named pipe file error: ", err)
		}
		s_line := string(line)
		cmd, name, image_or_arg := handleCmd(s_line)
		switch cmd {
		case "deploy":
			deployFunction(name, image_or_arg)
		case "invoke":
			go invokeFunction(name)
		case "delete":
			deleteFunction(name)
			flag = true
			//				fmt.Println("after delete function")
			break
		}
		if flag == true {
			break
		}
	}
	file.Close()
	fmt.Println("before exit 0")
	fmt.Println("end")
	os.Exit(0)
}
