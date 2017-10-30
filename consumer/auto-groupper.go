package consumer

import (
	"elastictrail/common"
	"sort"
)

type AutoGroupper struct {
	groups []*LineGroup
	//numericCharAcceptRatio float32 // the maximum ratio of numbers/other chars in a term (more will delete the term)
	lineCount int
}

type ByLineCount []*LineGroup

func (a ByLineCount) Len() int           { return len(a) }
func (a ByLineCount) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByLineCount) Less(i, j int) bool { return a[i].LineCount < a[j].LineCount }

func NewAutoGroupper() *AutoGroupper {

	auto := AutoGroupper{groups: []*LineGroup{}}

	//auto.numericCharAcceptRatio = 0.2
	return &auto
}

// Consolidate reduces amount of groups by trying to get the larger groups to consume smaller ones
//this is possible since bigger groups were pruned, and contain less terms so they require less in order for a line to join
func (ag *AutoGroupper) Consolidate() {
	//sort the group list
	sort.Sort(ByLineCount(ag.groups))

	//find the potential eaters and eaten
	sharks := map[*LineGroup]bool{}
	smallFish := map[*LineGroup]bool{}
	for _, gr := range ag.groups {
		gPrecent := float32(gr.LineCount) / float32(ag.lineCount)
		if gPrecent < 0.01 || gr.LineCount <= 10 {
			smallFish[gr] = true
		}
		if (gPrecent > 0.01 || gr.LineCount > 100) && !smallFish[gr] {
			sharks[gr] = true
		}
	}

	// try to assimilate small groups into big ones
	eaten := map[*LineGroup]bool{}
	for shark := range sharks {
		for fish := range smallFish {
			if shark.TryConsumeGroup(fish) {
				eaten[fish] = true
			}
		}
	}

	//remove all eaten groups from real list (by pointer comparison)
	for i := 0; i < len(ag.groups); i++ {
		gr := ag.groups[i]
		if eaten[gr] {
			ag.groups = append(ag.groups[:i], ag.groups[i+1:]...)
			i-- // -1 as the slice just got shorter
		}
	}

	//sort the group list after consolidation
	sort.Sort(ByLineCount(ag.groups))
}

// FindGroup tries to match the given line to existing groups, if a match is not found, the line creates a new group
func (ag *AutoGroupper) FindGroup(line common.LogLine) {
	message := line.Message()
	for _, lg := range ag.groups {
		if lg.TryAddLine(line) {
			lg.lines[message] = true
			ag.lineCount++

			if ag.lineCount%99 == 0 {
				ag.Consolidate()
			}
			return
		}
	}

	//no group found yet - create a new one
	if len(line.GetTerms()) > 0 {
		lg := NewLineGroup(line)
		lg.lines[message] = true
		ag.lineCount++
		lg.generateTemplate(false)
		ag.groups = append(ag.groups, lg)
	}
}
