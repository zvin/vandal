package main

import (
	"sort"
)

type Website struct {
	Url       string
	UserCount int
}

type ranking []Website

var CurrentlyUsedSites []Website

func (r ranking) Len() int {
	return len(r)
}

func (r ranking) Less(i, j int) bool {
	return r[i].UserCount > r[j].UserCount // we went it in the reverse order
}

func (r ranking) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}

func UpdateCurrentlyUsedSites() {
	var sites ranking
	Log.Println("UpdateRanking", "want Lock")
	GlobalLock.Lock()
	Log.Println("UpdateRanking", "got Lock")
	for _, location := range Locations {
		sites = append(sites, Website{Url: location.Url, UserCount: len(location.Users)})
	}
	GlobalLock.Unlock()
	Log.Println("UpdateRanking", "released Lock")
	sort.Sort(sites)
	CurrentlyUsedSites = sites[:MinInt(len(sites), 10)]
}
