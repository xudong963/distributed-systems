package mapreduce

import (
	"encoding/json"
	"os"
	"sort"
)

func doReduce(
	jobName string, // the name of the whole MapReduce job
	reduceTask int, // which reduce task this is
	outFile string, // write the output here
	nMap int, // the number of map tasks that were run ("M" in the paper)
	/*
	reduceF() is the application's reduce function. You should
    call it once per distinct key, with a slice of all the values
    for that key. reduceF() returns the reduced value for that key.
	*/
	reduceF func(key string, values []string) string,
) {

	// Your code here (Part I).

	var kv KeyValue
	var distMap = map[string][]string{}
	for i:=0; i<nMap; i++ {
		// for reduce task reduceTask  collects
		// the reduceTask'th intermediate file from each map task
		filename := reduceName(jobName, i, reduceTask)
		file, err := os.Open(filename)
		if err != nil {
			panic(err)
		}
		decode := json.NewDecoder(file)
		for {
			err := decode.Decode(&kv)
			if err != nil{
				break
			}
			distMap[kv.Key] = append(distMap[kv.Key], kv.Value)
		}
		err = file.Close()
		if err!=nil{
			panic(err)
		}
	}

	keys := make([]string, 0, len(distMap))
	for key := range distMap {
		keys = append(keys, key)
	}

	// sort the intermediate key/value pairs by key
	// incomprehension
	sort.Strings(keys)

	outfile, err := os.Create(outFile)
	if err!=nil {
		panic(err)
	}
	enc := json.NewEncoder(outfile)
	for _, key := range keys{
		err = enc.Encode(KeyValue{key, reduceF(key, distMap[key])})
		if err != nil{
			panic(err)
		}
	}
	err = outfile.Close()
	if err!=nil{
		panic(err)
	}
}
