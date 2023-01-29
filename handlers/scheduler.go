package handlers

import(
//	"fmt"
	"strconv"

//	"os/exec"

	"210.28.132.171/ShimeiT/faas-wasm/controller"
	log "github.com/sirupsen/logrus"
	//"github.com/containerd/cgroups"
	specs "github.com/opencontainers/runtime-spec/specs-go"
)
var ready []*controller.Controller
func rescheduleCPU(){
	cnt := 0
	for num, func_controller := range ready {
		name := func_controller.Name
		running[name] = func_controller
//		fmt.Printf("move %s from ready to running\n", name)
		ready[num] = nil
	}
	ready = ready[0:0]
	average := (CPU_NUM-1)/len(running)
//	fmt.Println("average:", average, "\n")
	var cpus string
	for _, func_controller := range running {
//		fmt.Printf("in reschedule cpu process: %s\n", func_controller.Name)
		name := func_controller.Name
		this_cgroups := *proc_cgroups[name]
		if average==1 {
			cpus = strconv.Itoa(cnt+1)
		}else {
			cpus = strconv.Itoa(average*cnt + 1) + "-" + strconv.Itoa(average*(cnt+1))
		}
		err := this_cgroups.Update(&specs.LinuxResources{
            CPU: &specs.LinuxCPU{
                Cpus: cpus,
            },
        })
		if err != nil {
			log.Errorln("rescheduleCpu error")
			panic(err)
		}
//		fmt.Println("after allocate cpu")
		if func_controller.State == controller.WAITING {
//			pid := func_controller.Process.Process.Pid
//			cmd := exec.Command("kill", "-CONT", string(pid))
//			cmd.Run()
			this_cgroups.Thaw()
			overTimer[name].Reset(overtime)
			func_controller.State = controller.RUNNING
//			fmt.Printf("%s this cgroup is Thaw()\n", name)
		}else if func_controller.State == controller.READY{
//			fmt.Println(name,"before pchan in schedule")
			pchan[name]<-true
			func_controller.State = controller.RUNNING
//			fmt.Println("after pchan in schedule")
		}
//		func_controller.State = controller.RUNNING
		cnt += 1
//		fmt.Printf("%s cpus: %s\n", name, cpus)
	}
//	fmt.Println("exit reschedule cpu\n")
}
/*func PrintQueue() {
	for k, func_controller := range wait_queue {
		fmt.Println("now in wait queue", k, ":", func_controller.Name)
	}
}*/
//scheduler main process
func Schedule() {
	for {
		//fmt.Println("this is schedule")
		select {
		case <-Stimer.C:
		//	fmt.Println("Stimerrrr")
			if len(wait_queue) == 0 {
				Stimer.Reset(time_slice)
                break
            }
//			fmt.Println("wait_queue length is not zero")
            if len(running)<MAX_CON_PROC {
//				fmt.Println("begin schedule")
                remain := MAX_CON_PROC - len(running)
//				fmt.Println("Timer remain process:", remain)

				wait_queue_lock.Lock()
//				fmt.Println("wait_queue length:", len(wait_queue))
//				PrintQueue()
                if len(wait_queue) > remain {
//					fmt.Println("len wait queue > remian")
                    for k, func_controller := range wait_queue[:remain] {
//						fmt.Println("this is append name:", func_controller.Name)
						ready = append(ready, func_controller)
                        wait_queue[k] = nil
                    }
                    wait_queue = wait_queue[remain:]
                }else {
//					fmt.Println("len wait queue =< remain")
                    for k,func_controller := range wait_queue {
						ready = append(ready, func_controller)
                        wait_queue[k] = nil
                    }
					wait_queue = wait_queue[0:0]
//					fmt.Println("ready len:", len(ready))
//					fmt.Println("wait len:", len(wait_queue))
                }
				wait_queue_lock.Unlock()
                rescheduleCPU()
				Stimer.Reset(time_slice)
            }else {
				Stimer.Reset(time_slice)
			}
		case <- overChan:
		case <-proc_finish:
//			fmt.Println("this is proc_finishsss")
			if len(wait_queue) == 0 {
//				fmt.Println("this is proc_finish!!!!wait_queue")
                break
            }
            if len(running)<MAX_CON_PROC {
//				fmt.Println("begin schedule in proc_finish/overChan")
                remain := MAX_CON_PROC - len(running)
//				fmt.Println("Timer remain process:", remain)

				wait_queue_lock.Lock()
//				fmt.Println("wait_queue length:", len(wait_queue))
//				PrintQueue()
                if len(wait_queue) > remain {
//					fmt.Println("len wait queue > remian")
                    for k, func_controller := range wait_queue[:remain] {
//						fmt.Println(func_controller.Name, "append ready queue")
                        ready = append(ready, func_controller)
                        wait_queue[k] = nil
                    }
                    wait_queue = wait_queue[remain:]
                }else {
//					fmt.Println("len wait queue =< remain")
                    for k,func_controller := range wait_queue {
//						fmt.Println(func_controller.Name, "append ready queue")
                        ready = append(ready, func_controller)
                        wait_queue[k] = nil
                    }
                    wait_queue = wait_queue[0:0]
//					fmt.Println("ready len:", len(ready))
//					fmt.Println("wait len:", len(wait_queue))
                }
				wait_queue_lock.Unlock()
                rescheduleCPU()
            }
		}
	}
}
