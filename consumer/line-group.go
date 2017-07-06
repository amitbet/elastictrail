package consumer

import (
	"fmt"
	"math"
	"strconv"
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
	lines                 []string
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

// NewLineGroup creates a new line group
func NewLineGroup(terms []string) *LineGroup {
	prev := ""
	lg := LineGroup{TermSet: map[string]*LineTerm{}}
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
	return &lg
}

func (lg *LineGroup) TryConsumeGroup(other *LineGroup) bool {
	if lg.CanAddByTerms(other.TermSet) {
		lg.LineCount += other.LineCount

		return true
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
	for i, l := range lg.lines {
		linestr += l + "\n"
		if i > 3 {
			break
		}
	}

	return fmt.Sprintf("lines:%d, pruneCount: %d, terms(%d): %s lines:\n%s", lg.LineCount, lg.pruneCounter, len(lg.TermSet), terms, linestr)
}

// pruneTerms removes terms with low grades from the group and may raise the threshold for acceptance
func (lg *LineGroup) pruneTerms() {
	// calculating term grade out of all terms consumed gives more importance to the terms that are always presnet and filters out passing visitors
	for _, term := range lg.TermSet {
		termSeenGrade := float32(term.Counter) / float32(lg.LineCount)

		if termSeenGrade < lg.termDeletionThreshold {
			//remove the term instances from the global counter
			lg.totalTermsConsumed -= term.Counter
			//remove the term from set
			delete(lg.TermSet, term.Term)
			//decrease term count
			lg.TermCount--
		}
	}
	lg.pruneCounter++
	lg.pruneCycle += 3
}

func (lg *LineGroup) CanAddByTerms(terms map[string]*LineTerm) bool {
	var termMatchCounter float32
	var termCountAcc int
	for term := range terms {
		if groupTerm := lg.TermSet[term]; groupTerm != nil {
			termMatchCounter++
			termCountAcc += groupTerm.Counter
		}
	}

	//termMatchCountGrade := termMatchCounter / float32(len(lg.TermSet))
	termWeightedGrade := float32(termCountAcc) / float32(lg.totalTermsConsumed)
	//fmt.Printf("termMatchCountGrade: %f weighted: %f\n", termMatchCountGrade, termWeightedGrade)

	if termWeightedGrade > lg.acceptanceThreshold {
		return true
	}
	return false
}

// TryAddLine adds a line to the group if it fits the group term profile
func (lg *LineGroup) TryAddLine(terms []string) bool {
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
			lg.pruneTerms()
		}
		return true
	}
	return false

}
