package main

// A buffered channel can be used like a semaphore

/*func main()  {

	var sem = make(chan int, MaxOutstanding)

	func handle(r *Request) {
		sem <- 1    // Wait for active queue to drain.
		process(r)  // May take a long time.
		<-sem       // Done; enable next request to run.
	}

	func Serve(queue chan *Request) {
		for {
			req := <-queue
			go handle(req)  // Don't wait for handle to finish.
		}
	}
}
*/
// ..more example to see doc