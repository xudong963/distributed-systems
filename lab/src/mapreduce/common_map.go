package mapreduce

import (
	"encoding/json"
	"hash/fnv"
	"io/ioutil"
	"os"
)

func doMap(
	jobName string, // the name of the MapReduce job
	mapTask int,    // which map task this is
	inFile string,
	nReduce int, // the number of reduce task that will be run ("R" in the paper)
	mapF func(filename string, contents string) []KeyValue,
) {
	// Your code here (Part I).
	//
	infile, err := ioutil.ReadFile(inFile)
	if err != nil {
		panic(err)
	}
	keyValues := mapF(inFile, string(infile))
	outputFiles := make([]*os.File, nReduce)

	for i:=0; i<nReduce; i++ {
		// Each file name contains a prefix,
		// the map task number, and the reduce task number
		filename := reduceName(jobName, mapTask, i)
		outputFiles[i], err = os.Create(filename)
		if err != nil{
			panic(err)
		}
	}

	// hash each key to pick the intermediate
	// file and thus the reduce task that will process the key
	for _, kv := range keyValues{
		index := ihash(kv.Key) % nReduce
		err1 := json.NewEncoder(outputFiles[index]).Encode(&kv)
		if err1 != nil{
			panic(err1)
		}
	}

	for _, file := range outputFiles{
		file.Close()
	}
}

func ihash(s string) int {
	h := fnv.New32a()
	h.Write([]byte(s))
	return int(h.Sum32() & 0x7fffffff)
}
