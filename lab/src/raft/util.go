package raft

import "log"

// Debugging
const Debug = 0

func DPrintf(format string, a ...interface{}) (n int, err error) {
	if Debug > 0 {
		log.Printf(format, a...)
	}
	return
}

// service for last func
func min(a int, b int) int {
	if a<b {
		return a
	}
	return b
}

func max_(a int, b int) int {
	if a>b {
		return a
	}else {
		return b
	}
}