// Copyright (c) OpenFaaS Author(s) 2019. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package handlers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"fmt"
	"bufio"

	log "github.com/sirupsen/logrus"
)

type deleteFunctionRequest struct {
	FunctionName string `json:"functionName"`
}

// MakeDeleteHandler delete a function
func MakeDeleteHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Info("delete request")
		defer r.Body.Close()

		body, _ := ioutil.ReadAll(r.Body)
		request := deleteFunctionRequest{}
		if err := json.Unmarshal(body, &request); err != nil {
			log.Errorf("error de-serializing request body:%s", body)
			log.Error(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if len(request.FunctionName) == 0 {
			log.Errorln("can not delete a function, request function name is empty")
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		name := request.FunctionName
		if _, ok := functions[name]; !ok {
			w.WriteHeader(http.StatusNotFound)
            w.Write([]byte("{ \"status\" : \"Not found\"}"))
            log.Errorf("%s not found", name)
            return
		}
//		fmt.Println("delete before write to pipe", name)
	/*==========Write to fifo file=========*/
        pipeFile := "./pipe/" + name + ".pipe"
        pipe_lock := locks[name]
        pipe_lock.Lock()
        file, err := os.OpenFile(pipeFile, os.O_WRONLY, 0777)
        if err!=nil {
            log.Fatalf("delete opening file: %v", err)
        }
		file.WriteString(fmt.Sprintf("delete %s\n", name))
        pipe_lock.Unlock()
	/*==========Read from controller stdout==========*/
		c_stdout := *stdouts[name]
		buff := bufio.NewScanner(c_stdout)
		for buff.Scan(){
			tmp := buff.Text()
			if tmp=="end"{
				break
			}else {
				fmt.Println(tmp)
			}
		}
//		log.Infof("delete after read stdout")
	/*==============delete map==============*/
		c_stdout.Close()		//is this must????TODO()
		stdouts[name] = nil
		delete(stdouts, name)
		functions[name] = nil
		delete(functions, name)
		p := process[name].Process.Process
		state, _ := p.Wait()
		log.Infof("this is exit code: %d", state.ExitCode())
		process[name] = nil
		delete(process, name)

		control := *proc_cgroups[name]
		control.Delete()
		delete(proc_cgroups, name)
		log.Infof("delete request %s successful", request.FunctionName)
	}
}
