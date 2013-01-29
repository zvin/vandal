package main

import (
	"sort"
)

type Website struct {
	Url       string
	UserCount int
}

type ranking []Website

func (r ranking) Len() int {
	return len(r)
}

func (r ranking) Less(i, j int) bool {
	return r[i].UserCount > r[j].UserCount // we want it in the reverse order
}

func (r ranking) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}

func SortWebsites(sites []Website) {
	sort.Sort(ranking(sites))
}
