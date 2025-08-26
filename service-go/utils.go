package main

import (
	"strconv"
	"strings"
)

func ParseInts(values string) ([]int, error) {
	var results []int
	for _, v := range strings.Split(values, ",") {
		i, err := strconv.Atoi(v)
		if err != nil {
			return nil, err
		}
		results = append(results, i)
	}
	return results, nil
}