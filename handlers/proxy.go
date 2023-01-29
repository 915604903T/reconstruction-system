package handlers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"fmt"
	"bufio"
	"time"
	"strconv"

//	os/exec"

//	"github.com/Fvoiretryzig/faas-wasm/controller"
	"210.28.132.171/ShimeiT/faas-wasm/controller"
	"github.com/containerd/cgroups"
	specs "github.com/opencontainers/runtime-spec/specs-go"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

type response struct {
	Function     string
	ResponseBody string
	HostName     string
}
func Overtime(name string, i float64) {
	flag := false
	for {
//		log.Errorln(name, "this is for overtime")
		select{
		case <-overTimer[name].C:
//			fmt.Println("\nOOOOOOOOOOOOOOO", name, "this is overTimer")
			_, ok := running[name]
			if !ok {
//				fmt.Printf("OOOOOOOOOOOOOOOOOO %s not in running\n", name)
				flag = true
				break
			}
			if len(wait_queue)==0 {
//				fmt.Println("OOOOOOOOOOOOOO", name, "when overtime no process wait continue")
				overTimer[name].Reset(overtime)
				break
			}
			control := *proc_cgroups[name]
			func_controller := process[name]
//			pid := func_controller.Process.Process.Pid
//			cmd := exec.Command("kill", "-STOP", string(pid))
//			cmd.Run()
			control.Freeze()
			func_controller.State = controller.WAITING

			wait_queue_lock.Lock()
			running[name] = nil
			delete(running, name)
			wait_queue = append(wait_queue, func_controller)
			wait_queue_lock.Unlock()

//			fmt.Printf("OOOOOOOOOOOOOOOOO %s this cgroup is Freeze()\n", name)
			overChan <- true
//			fmt.Printf("OOOOOOOOOOOOOOOOOO send chan to %s\n", name)
		}
		if flag == true {
//			fmt.Println("OOOOOOOOOOOOOOOOOOOO exit overtime", i)
			break
		}
	}
}
// MakeProxy creates a proxy for HTTP web requests which can be routed to a function.
func MakeProxy() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		name := vars["name"]
		log.Info("proxy request: " + name)
		v, okay := functions[name]
		if !okay {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("{ \"status\" : \"Not found\"}"))
			log.Errorf("%s not found", name)
			return
		}

		v.InvocationCount = v.InvocationCount + 1
		invokeTime := strconv.Itoa(int(v.InvocationCount))
		defer r.Body.Close()
		body, _ := ioutil.ReadAll(r.Body)
//		log.Infof("body: %s", string(body))
		hostName, _ := os.Hostname()

		func_controller := process[name]

		func_controller.Invoke_cnt++
		c_stdout := *stdouts[name]

		if func_controller.State == controller.IDLE {
			func_controller.State = controller.READY

			wait_queue_lock.Lock()
			wait_queue = append(wait_queue, func_controller)
			wait_queue_lock.Unlock()

//			log.Infof("[%s] before pchan in proxy", name)
			<-pchan[name]	//wait for scheduler(in shceduler rescheduleCPU send this msg)
//			log.Infof("[%s]after pchan in proxy", name)
			overTimer[name] = time.NewTimer(overtime)
			go Overtime(name, v.InvocationCount)
		}
	/*========Write to fifo file========*/
		pipeFile := "./pipe/" + name + ".pipe"
		pipe_lock := locks[name]
		pipe_lock.Lock()
		file, err := os.OpenFile(pipeFile, os.O_WRONLY, 0777)
		if err!=nil {
			log.Fatalf("invoke error opening file: %v", err)
		}
		file.WriteString(fmt.Sprintf("invoke %s %s\n", name, string(body)))
//		log.Infof("[%s] after write to pipe", name)
		pipe_lock.Unlock()
	/*==========read controller stdout==========*/
		buff := bufio.NewScanner(c_stdout)
		var result string
		fmt.Println("invokeTime: ", "end"+invokeTime)
		for buff.Scan() {
//			fmt.Println("PPPPPPPPPPPPPPPPPPP this is", name)
			tmp := buff.Text()
			if tmp==("end"+invokeTime) {
				break
			}else{
				fmt.Println(tmp)
				result = tmp
//				fmt.Println("PPPPPPPPPPPPPPPPPPPP result: ", name, result)
			}
		}
		log.Infof("[%s] this is result:%s",name, result)

		func_controller.Invoke_cnt--
		if func_controller.Invoke_cnt==0 {
//			log.Infof("[%s] every invoke finished!delete from running map", name)
			control, err := cgroups.Load(cgroups.V1, cgroups.StaticPath("/"+name))
            if err!=nil {
				fmt.Println("load cgroups to reset error")
                panic(err)
            }
            err = control.Update(&specs.LinuxResources{
                CPU: &specs.LinuxCPU{
                    Cpus: "0",
                },
            })
//			fmt.Println("this is cgroups err:", err)
//			log.Infof("[%s] reset occupied cpu to origin", name)
			func_controller.State = controller.IDLE
			running[name] = nil
			delete(running, name)
			overTimer[name].Stop()
			overTimer[name].Reset(0)
//			log.Infof("[%s] before send to proc_finish", name)
			proc_finish <- true
//			log.Infof("[%s] after send to proc_finish", name)
//			log.Errorln("[",name, "]", "invoke cnt 0 end", v.InvocationCount)
		}
		d := &response{
            Function:     name,
            //ResponseBody: string(body),
			ResponseBody: result,
            HostName:     hostName,
        }
		responseBody, err := json.Marshal(d)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			log.Errorf("error invoking %s. %v", name, err)
			return
		}
		log.Infof("[%s] responseBody: %s\n", name, responseBody)
		w.Write(responseBody)

		log.Infof("proxy request: %s completed.", name)
	}
}
