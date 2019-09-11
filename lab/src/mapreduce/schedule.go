package mapreduce

import (
	"fmt"
)

//
// schedule() starts and waits for all tasks in the given phase (mapPhase
// or reducePhase). the mapFiles argument holds the names of the files that
// are the inputs to the map phase, one per map task. nReduce is the
// number of reduce tasks. the registerChan argument yields a stream
// of registered workers; each item is the worker's RPC address,
// suitable for passing to call(). registerChan will yield all
// existing registered workers (if any) and new ones as they register.
//
func schedule(jobName string, mapFiles []string, nReduce int, phase jobPhase, registerChan chan string) {
	var ntasks int
	var n_other int // number of inputs (for reduce) or outputs (for map)
	switch phase {
	case mapPhase:
		ntasks = len(mapFiles)
		n_other = nReduce
	case reducePhase:
		ntasks = nReduce
		n_other = len(mapFiles)
	}

	fmt.Printf("Schedule: %v %v tasks (%d I/Os)\n", ntasks, phase, n_other)

	// All ntasks tasks have to be scheduled on workers. Once all tasks
	// have completed successfully, schedule() should return.
	//
	// Your code here (Part III, Part IV).
	//

	//get Worker.DoTask RPC
	makeDoTaskArgs := func(task int, phase jobPhase) DoTaskArgs {

		var doTaskArgs DoTaskArgs
		if phase == mapPhase {
			doTaskArgs.File = mapFiles[task]
		}
		doTaskArgs.JobName = jobName

		// NumOtherPhase is the total number of tasks in other phase; mappers
		// need this to compute the number of output bins, and reducers needs
		// this to know how many input files to collect.
		doTaskArgs.NumOtherPhase = n_other
		doTaskArgs.Phase = phase
		doTaskArgs.TaskNumber = task
		return doTaskArgs
	}

	var taskChan = make(chan int)
	go func() {
		for i:=0; i<ntasks; i++ {
			taskChan <- i
		}
	}()

	finish := make(chan int)
	finishTime := 0

loop:
	for {
		select {
		case task := <- taskChan:
			go func() {
				worker := <- registerChan
				//send an PRC to a worker
				result := call(worker,"Worker.DoTask", makeDoTaskArgs(task, phase), nil)
				if result {
					//schedule successfully, put worker to registerChan for other task
					finish <- 1
					go func() { registerChan <- worker }()
				}else {
					//failure, put task to taskChan, retry
					taskChan <- task
				}
			}()
		case <- finish:
			finishTime += 1
		default:
			if finishTime == ntasks {
				break loop
			}
		}
	}


/*
  //can't solve the situation of worker fault
	var wg sync.WaitGroup
	wg.Add(ntasks)
	for i := 0; i<ntasks; i++ {
		go func() {
			task := <- taskChan
			worker := <- registerChan
			//send an PRC to a worker
			result := call(worker,"Worker.DoTask", makeDoTaskArgs(task, phase), nil)
			if result {
				//schedule successfully, put worker to registerChan for other task
				wg.Done()
				go func() { registerChan <- worker }()
			}else {
				//failure, put task to taskChan, retry
				taskChan <- task
				go func() { ntasks++ }()
			}
		}()
	}
	wg.Wait()

 */

	fmt.Printf("Schedule: %v done\n", phase)
}
