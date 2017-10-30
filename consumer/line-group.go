package consumer

import (
	"elastictrail/common"
	"fmt"
	"math"
	"strconv"
	"strings"
)

type LineTerm struct {
	Term                   string
	Counter                int
	prevTerm               string
	prevTermCorrectCounter int
}

// LineGroup represents a group of log lines that have been bunched together by the algorithm
type LineGroup struct {
	Template              string
	lines                 map[string]bool
	TermSet               map[string]*LineTerm //the actual terms that exist in the log line
	TermCount             int                  //if a term exists several times this can be different than len(termSet)
	separators            map[rune]bool
	LineCount             int
	pruneCycle            int     // indicates how often to prune the terms (do a pruning after X number of additional lines accepted to the group)
	acceptanceThreshold   float32 // starts with 0.8 so: 80% of terms should be matched with group for accepting line
	termDeletionThreshold float32 // default 0.7: a term which appears in less than 70% of lines will be removed
	pruneCycleWidening    int     // makes pruning less frequent each time it runs, postpones the next prune run by X more lines added to the group in addition to the exiting pruneCycle value
	pruneCounter          int     //counts the number of prunes ran on the group
	totalTermsConsumed    int
	origTokenCount        int //number of terms upon group creation
}

// // represents a group (cluster) of log lines
// type ESLogGroupLine struct {
// 	Title     string
// 	LineCount int
// }

// Message returns the first message field that exists in the line definition
func (line *LineGroup) Message() string {
	return line.Template
}

func (line *LineGroup) GetField(fieldName string) string {
	return line.Template
}

func (line *LineGroup) Count() int {
	return line.LineCount
}

// NewLineGroup creates a new line group
func NewLineGroup(ln common.LogLine) *LineGroup {
	terms := ln.GetTerms()
	prev := ""
	lg := LineGroup{TermSet: map[string]*LineTerm{}}
	lg.lines = make(map[string]bool, 2)
	for _, term := range terms {
		lg.TermSet[term] = &LineTerm{Term: term, Counter: 1, prevTerm: prev, prevTermCorrectCounter: 1}
		prev = term
	}
	lg.origTokenCount = len(terms)
	lg.totalTermsConsumed += len(lg.TermSet)
	lg.TermCount = len(terms)
	lg.acceptanceThreshold = 0.8
	lg.termDeletionThreshold = 0.55
	lg.pruneCycle = 5
	lg.LineCount = 1
	lg.pruneCycleWidening = 3
	//lg.maxNumberOfDeletedTerms
	lg.generateTemplate(false)
	return &lg
}

func (lg *LineGroup) TryConsumeGroup(other *LineGroup) bool {
	var termMatchCounter float32
	var termCountAcc int
	for term := range other.TermSet {
		if groupTerm := lg.TermSet[term]; groupTerm != nil {
			termMatchCounter++
			termCountAcc += groupTerm.Counter
		}
	}

	//termMatchCountGrade := termMatchCounter / float32(len(lg.TermSet))
	termWeightedGrade := float32(termCountAcc) / float32(lg.totalTermsConsumed)
	//fmt.Printf("termMatchCountGrade: %f weighted: %f\n", termMatchCountGrade, termWeightedGrade)

	// do the actual merging of the smaller group into the large one:
	if termWeightedGrade > lg.acceptanceThreshold {
		return true
		lg.LineCount += other.LineCount

		for _, oterm := range other.TermSet {
			lg.TermSet[oterm.Term].Counter += oterm.Counter
		}
		for line := range other.lines {
			lg.lines[line] = true
		}
	}
	return false
}

func (lg *LineGroup) String() string {
	terms := "("
	for _, t := range lg.TermSet {
		terms += t.Term + ":" + strconv.Itoa(t.Counter) + "),("
	}
	if len(terms) > 3 {
		terms = terms[:len(terms)-2]
	} else {
		terms = ""
	}
	linestr := ""
	i := 0
	for l := range lg.lines {
		linestr += l + "\n"
		i++
		if i > 3 {
			break
		}
	}

	return fmt.Sprintf("template: %s\n lines:%d, pruneCount: %d, terms(%d): %s lines:\n%s", lg.Template, lg.LineCount, lg.pruneCounter, len(lg.TermSet), terms, linestr)
}

// pruneTerms removes terms with low grades from the group and may raise the threshold for acceptance
func (lg *LineGroup) tryPruneTerms() bool {
	// calculating term grade out of all terms consumed gives more importance to the terms that are always presnet and filters out passing visitors
	termsForDeletion := []*LineTerm{}

	for _, term := range lg.TermSet {
		termSeenGrade := float32(term.Counter) / float32(lg.LineCount)

		if termSeenGrade < lg.termDeletionThreshold {
			termsForDeletion = append(termsForDeletion, term)
		}
		//set some conditions so we don't lose the group's identity for a stray line that slips in..
		if lg.TermCount-len(termsForDeletion) <= 0 || len(termsForDeletion) > 3 {
			return false
		}

		for _, term1 := range termsForDeletion {
			//remove the term instances from the global counter
			lg.totalTermsConsumed -= term1.Counter
			//remove the term from set
			delete(lg.TermSet, term1.Term)
			//decrease term count
			lg.TermCount--
			lg.generateTemplate(false)
		}
	}
	lg.pruneCounter++
	lg.pruneCycle += lg.pruneCycleWidening
	return true
}

// TryAddLine adds a line to the group if it fits the group term profile
func (lg *LineGroup) TryAddLine(line common.LogLine) bool {
	terms := line.GetTerms()
	//var termMatchCounter float32
	var termCountAcc int
	seenTerms := map[string]bool{}
	for _, term := range terms {
		seenTerms[term] = true
	}

	// make sure we don't push totally different lines together...
	sizeDiff := math.Abs(float64(lg.origTokenCount - len(terms)))
	if float32(sizeDiff)/float32(lg.origTokenCount) > 0.25 {
		return false
	}

	//check given terms against the terms already in the group, see if it matches
	for term := range seenTerms {
		if groupTerm := lg.TermSet[term]; groupTerm != nil {
			termCountAcc += groupTerm.Counter
			seenTerms[groupTerm.Term] = true
		}
	}
	termWeightedGrade := float32(termCountAcc) / float32(lg.totalTermsConsumed)
	//if termMatchCounter/float32(len(lg.TermSet)) > lg.acceptanceThreshold {

	//if the grade is enough to accept this line, add it
	if termWeightedGrade > lg.acceptanceThreshold {
		lg.LineCount++
		for termStr := range seenTerms {
			gTerm := lg.TermSet[termStr]
			if gTerm != nil {
				gTerm.Counter++
				lg.totalTermsConsumed++
			}
		}
		if lg.LineCount%lg.pruneCycle == 0 {
			return lg.tryPruneTerms()
		}
		return true
	}
	return false

}

func (lg *LineGroup) generateTemplate(preserveStructure bool) {
	if len(lg.lines) == 0 {
		return
	}
	var line = ""
	for ln := range lg.lines {
		line = ln
		break
	}

	//logger.Debug(">>>>>>> line:" + line)
	if !preserveStructure {
		lg.Template = "*"
	}

	tokens := SplitWithMultiDelims(line, "`:\\/;'=-_+~<>[]{}!@#$%^&*().,?\"| \t\n")
	for _, t := range tokens {
		if groupTerm := lg.TermSet[t]; groupTerm != nil {
			//replace all
			if preserveStructure {
				lg.Template = strings.Replace(lg.Template, t, "*", 1)
			} else {
				lg.Template += groupTerm.Term + "*"
			}
		}
	}
	//logger.Debug(">>>>>>> temp:" + lg.Template)
}

func SplitWithMultiDelims(input string, delims string) []string {
	delimiters := make(map[rune]bool)
	for _, ru := range delims {
		delimiters[ru] = true
	}

	SplitFunc := func(r rune) bool {
		return delimiters[r]
	}

	return strings.FieldsFunc(input, SplitFunc)
}
