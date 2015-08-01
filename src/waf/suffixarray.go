package waf

import (
	"bytes"
)


type SuffixIndex struct {
	data	[]byte
	index	[][]int
}


func Index(data []byte) SuffixIndex {
	return SuffixIndex { data, index(data) }
}


func index(data[]byte) [][]int{
	Q	:= make([][]int,256)

	for i,b := range data {
		Q[b] = append(Q[b], i )
	}

	return Q
}


func (SA SuffixIndex)Match(token []byte) bool {
	part := SA.index[token[0]]
	if len(part) == 0 {
		return false
	}

	for _,start := range part {
		if bytes.HasPrefix(SA.data[start:], token) {
			return true
		}
	}
	return false
}
