package fwew_lib

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"math"
	"os"
	"slices"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

var homonymsArray = []string{"", "", ""}
var candidates2 Queue = *CreateQueue(12000)
var first2StageMap = HomoMapStruct{}
var stage3Map = HomoMapStruct{}
var homoMap = HomoMapStruct{}
var resultCount = 0
var lenitors = []string{"px", "p", "ts", "tx", "t", "kx", "k", "'"}
var lenitionMap = map[string]string{
	"ts": "s",
	"t":  "s",
	"tx": "t",
	"p":  "f",
	"px": "p",
	"k":  "h",
	"kx": "k",
	"'":  "",
}

var inefficiencyWarning = false
var nasalAssimilationOnly = false

/*
var top10Longest = map[uint8]string{}
var longest uint8 = 0
*/
var totalCandidates int = 0
var charLimit int = 14
var progressInterval int = 100
var changePOS = map[string]bool{
	"tswo":   true, // ability to [verb]
	"yu":     true, // [verb]er
	"tsuk":   true, //[verb]able
	"ketsuk": true, //un[verb]able
	"us":     true, //[verb]ing (active participle only)
	"awn":    true, //[verb]ed (passive participle only)
}

var resultsFile *os.File
var previous *os.File

//var dupeLengthsMap = map[int]int{}

var finished = queueFinished{false, sync.Mutex{}}
var finishedSentinelValue = "lu hasey srak?"
var wordCount = 0
var dictArray = []*FwewDict{}

type queueFinished struct {
	finished bool
	mu       sync.Mutex
}

type Queue struct {
	mu       sync.Mutex
	capacity int
	q        []string
}

type HomoMapStruct struct {
	mu      sync.Mutex
	homoMap map[string]bool
}

var writeLock sync.Mutex
var addWaitGroup sync.WaitGroup
var makeWaitGroup sync.WaitGroup
var checkWaitGroup sync.WaitGroup
var start time.Time

// FifoQueue
type FifoQueue interface {
	Insert()
	Remove()
}

func (h *HomoMapStruct) Insert(item string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.homoMap[item] = true
}

func (h *HomoMapStruct) Present(item string) bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	_, ok := h.homoMap[item]
	return ok
}

func (h *HomoMapStruct) Clear() {
	h.mu.Lock()
	defer h.mu.Unlock()
	clear(h.homoMap)
}

// Insert inserts the item into the queue
func (q *Queue) Insert(item string) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	if len(q.q) < int(q.capacity) {
		q.q = append(q.q, item)
		return nil
	}
	return errors.New("Queue is full")
}

// Remove removes the oldest element from the queue
func (q *Queue) Remove() (string, error) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if len(q.q) > 0 {
		item := q.q[0]
		q.q = q.q[1:]
		return item, nil
	}
	return "", errors.New("Queue is empty")
}

func (q *Queue) Length() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.q)
}

// CreateQueue creates an empty queue with desired capacity
func CreateQueue(capacity int) *Queue {
	return &Queue{
		capacity: capacity,
		q:        make([]string, 0, capacity),
	}
}

func DuplicateDetector(query string) bool {
	result := false
	query = " " + query + " "

	for i := 0; i < len(homonymsArray); i++ {
		temp := " " + homonymsArray[i] + " "
		if strings.Contains(temp, query) {
			return true
		}
	}

	return result
}

// Check for ones that are the exact same, no affixes needed
func StageOne() error {
	resultsFile.WriteString("Stage 1:\n")

	err := runOnFile(func(word Word) error {
		standardizedWord := word.Navi
		badChars := `~@#$%^&*()[]{}<>_/.,;:!?|+\"„“”«»`

		// remove all the sketchy chars from arguments
		for _, c := range badChars {
			standardizedWord = strings.ReplaceAll(standardizedWord, string(c), "")
		}

		// normalize tìftang character
		standardizedWord = strings.ReplaceAll(standardizedWord, "’", "'")
		standardizedWord = strings.ReplaceAll(standardizedWord, "‘", "'")

		// find everything lowercase
		standardizedWord = strings.ToLower(standardizedWord)

		if strings.Contains(standardizedWord, "é") {
			standardizedWord = strings.ReplaceAll(standardizedWord, "é", "e")
		}

		if !first2StageMap.Present(standardizedWord) {
			// If the word appears more than once, record it
			if entry, ok := dictHash[standardizedWord]; ok {
				if len(entry) > 1 {
					query, _ := QueryHelper(entry)
					foundResult(standardizedWord, query, true)
				}
			}
		}

		return nil
	})

	if err != nil {
		log.Printf("Error in homonyms stage 1: %s", err)
		return err
	}

	return nil
}

// Helper to detect presences of affixes
func AffixCount(word Word) string {
	var fixes strings.Builder

	if len(word.Affixes.Prefix) > 0 {
		fixes.WriteString("1")
	} else {
		fixes.WriteString("0")
	}
	if len(word.Affixes.Infix) > 0 {
		/*for _, fix := range word.Affixes.Infix {
			fixes.WriteString(fix)
		}*/
		fixes.WriteString("1")
	}
	if len(word.Affixes.Suffix) > 0 {
		fixes.WriteString("1")
	} else {
		fixes.WriteString("0")
	}

	//fmt.Println(prefixCount + infixCount + suffixCount)

	return fixes.String()
}

func QueryHelper(results []Word) (string, bool) {
	slices.SortFunc(results, func(i, j Word) int {
		if i.Navi != j.Navi {
			return strings.Compare(i.Navi, j.Navi)
		}

		if len(i.Affixes.Prefix) != len(j.Affixes.Prefix) {
			return len(i.Affixes.Prefix) - len(j.Affixes.Prefix)
		}

		if len(i.Affixes.Suffix) != len(j.Affixes.Suffix) {
			return len(i.Affixes.Suffix) - len(j.Affixes.Suffix)
		}

		return len(i.Affixes.Infix) - len(j.Affixes.Infix)
	})

	var allNaviWords strings.Builder
	//infixFound := false
	noDupes := []string{}

	allPrefixes := [][]string{}
	allInfixes := [][]string{}
	allSuffixes := [][]string{}

	for _, b := range results {
		noDupes = append(noDupes, b.Navi)

		allPrefixes = append(allPrefixes, b.Affixes.Prefix)
		allInfixes = append(allInfixes, b.Affixes.Infix)
		allSuffixes = append(allSuffixes, b.Affixes.Suffix)

		/*if len(b.Affixes.Infix) > 0 {
			infixFound = true
		}*/
	}

	for i, b := range noDupes {
		allNaviWords.WriteString(b)
		if i+1 < len(noDupes) {
			allNaviWords.WriteString(" ")
		}
	}

	// All the conditions you want limit shown results to
	// Commented code is an example
	allGood := true
	/*allGood := false

	hasPrefix := false

	for _, a := range allPrefixes {
		if len(a) > 0 {
			hasPrefix = true
			break
		}
	}

	if hasPrefix {
		for _, a := range allSuffixes {
			if implContainsAny(a, []string{"o"}) {
				allGood = true
				break
			}
			if allGood {
				break
			}
		}
	}*/

	allNaviWords.WriteString("] ")

	preUnique := findUniques(allPrefixes, false)
	allNaviWords.WriteString(preUnique)

	allNaviWords.WriteString("-")

	inUnique := findUniques(allInfixes, false)
	allNaviWords.WriteString(inUnique)

	allNaviWords.WriteString("-")

	sufUnique := findUniques(allSuffixes, true)
	allNaviWords.WriteString(sufUnique)

	return allNaviWords.String(), allGood
}

// Check for ones that are the exact same, no affixes needed
func StageTwo() error {
	resultsFile.WriteString("Stage 2:\n")

	err := runOnFile(func(word Word) error {
		lower := strings.ToLower(word.Navi)
		if !first2StageMap.Present(lower) {
			standardizedWord := word.Navi

			first2StageMap.Insert(lower)

			if len(strings.Split(word.Navi, " ")) == 1 {
				// If the word can conjugate into something else, record it
				results, err := TranslateFromNaviHash(dictArray[0], standardizedWord, true)
				if err == nil && len(results[0]) > 2 {
					results[0] = results[0][1:]
					allNaviWords, show := QueryHelper(results[0])
					foundResult(standardizedWord, allNaviWords, show)
				}

				// Lenited forms should be taken care of
			}
		}

		return nil
	})

	if err != nil {
		log.Printf("Error in homonyms stage 2: %s", err)
		return err
	}

	//fmt.Println(homonymsArray[1])

	return nil
}

// For StageThree, this adds things to the candidates
func addToCandidates(candidates *[][]string, candidate1 string) bool {
	newLength := len([]rune(candidate1))
	inserted := false
	// Is it longer than the words we want to check?
	if newLength > charLimit {
		return false
	}

	// Particularly for nasal assimilation, we want "-pe" words to go before other affixes
	if strings.HasSuffix(candidate1, "pe") {
		newLength--
	}

	// If it's in the range, is it good?
	if !stage3Map.Present(candidate1) {
		inserted = true
		(*candidates)[newLength] = append((*candidates)[newLength], candidate1)
		//totalCandidates++
		stage3Map.Insert(candidate1)
	}

	//Lenited forms, too
	found := false
	lenited := ""
	for _, a := range lenitors {
		if strings.HasPrefix(candidate1, a) {
			lenited = strings.TrimPrefix(candidate1, a)
			lenited = lenitionMap[a] + lenited
			found = true
			break
		}
	}

	if !inserted && first2StageMap.Present(candidate1) {
		inserted = true
	}

	if !found {
		return inserted
	}

	// If it's in the range, is it good?
	if !stage3Map.Present(lenited) {
		inserted = true
		// lenited ones will be sorted to appear later
		(*candidates)[newLength+1] = append((*candidates)[newLength+1], lenited)
		//totalCandidates++
		stage3Map.Insert(lenited)

	}

	if !inserted && first2StageMap.Present(lenited) {
		inserted = true
	}

	return inserted
}

// Helper for StageThree, based on reconstruct from affixes.go
func reconjugateNouns(candidates *[][]string, input Word, inputNavi string, prefixCheck int, suffixCheck int, unlenite int8, affixCountdown int8) error {
	// End state: Limit to 2 affixes per noun
	if affixCountdown == 0 {
		return nil
	}

	runeLen := len([]rune(inputNavi))

	if runeLen > charLimit {
		return nil
	}

	inserted := true
	inserted = addToCandidates(candidates, inputNavi)

	// Do not reconstruct things based on things we already reconstructed
	if !inserted {
		return nil
	}

	switch prefixCheck {
	case 0:
		for _, element := range stemPrefixes {
			if strings.HasSuffix(element, string(inputNavi[0])) {
				// regardless of whether or not it's found
				newWord := element + strings.TrimPrefix(inputNavi, string(inputNavi[0]))
				reconjugateNouns(candidates, input, newWord, 1, suffixCheck, 0, affixCountdown-1)
			} else {
				// regardless of whether or not it's found
				newWord := element + inputNavi
				reconjugateNouns(candidates, input, newWord, 1, suffixCheck, 0, affixCountdown-1)
			}
		}
		fallthrough
	case 1:
		fallthrough
	case 2:
		// Non-lenition prefixes for nouns only
		for _, element := range prefixes1Nouns {
			if strings.HasSuffix(element, string(inputNavi[0])) {
				// regardless of whether or not it's found
				newWord := element + strings.TrimPrefix(inputNavi, string(inputNavi[0]))
				reconjugateNouns(candidates, input, newWord, 4, suffixCheck, 0, affixCountdown-1)
			} else {
				// regardless of whether or not it's found
				newWord := element + inputNavi
				reconjugateNouns(candidates, input, newWord, 4, suffixCheck, 0, affixCountdown-1)
			}
		}
		fallthrough
	case 3:
		// This one will demand this makes it use lenition
		lenited := inputNavi
		if unlenite == 0 {
			for _, a := range lenitors {
				if strings.HasPrefix(lenited, a) {
					lenited = strings.TrimPrefix(lenited, a)
					lenited = lenitionMap[a] + lenited
					break
				}
			}
		}

		for _, element := range append(prefixes1lenition, "tsay") {
			// If it has a lenition-causing prefix
			if strings.HasSuffix(element, string(lenited[0])) {
				// regardless of whether or not it's found
				lenited2 := element + strings.TrimPrefix(lenited, string(lenited[0]))
				reconjugateNouns(candidates, input, lenited2, 4, suffixCheck, -1, affixCountdown-1)
			} else {
				// regardless of whether or not it's found
				lenited2 := element + lenited
				reconjugateNouns(candidates, input, lenited2, 4, suffixCheck, -1, affixCountdown-1)
			}
		}
		fallthrough
	case 4:
		//fallthrough
	}

	switch suffixCheck {
	case 0: // -o "some"

		fallthrough
	case 1:
		for _, element := range stemSuffixes {
			if strings.HasSuffix(inputNavi, string(element[0])) {
				// regardless of whether or not it's found
				newWord := strings.TrimSuffix(inputNavi, string(element[0])) + element
				reconjugateNouns(candidates, input, newWord, prefixCheck, 2, unlenite, affixCountdown-1)
			} else {
				// regardless of whether or not it's found
				newWord := inputNavi + element
				reconjugateNouns(candidates, input, newWord, prefixCheck, 2, unlenite, affixCountdown-1)
			}
		}
		fallthrough
	case 2:
		newWord := inputNavi + "o"
		reconjugateNouns(candidates, input, newWord, prefixCheck, 3, unlenite, affixCountdown-1)
		fallthrough
	case 3:
		if strings.HasSuffix(inputNavi, "p") {
			// regardless of whether or not it's found
			newWord := inputNavi + "e"
			reconjugateNouns(candidates, input, newWord, prefixCheck, 4, unlenite, affixCountdown-1)
		} else {
			// regardless of whether or not it's found
			newWord := inputNavi + "pe"
			reconjugateNouns(candidates, input, newWord, prefixCheck, 4, unlenite, affixCountdown-1)
		}
		fallthrough
	case 4:
		vowel := false
		diphthong := false
		consonant := true
		naviRunes := []rune(inputNavi)
		lastRune := naviRunes[len(naviRunes)-1]
		if is_vowel(lastRune) {
			vowel = true
		} else if lastRune == 'y' || lastRune == 'w' {
			diphthong = true
		} else {
			consonant = true
		}

		// This significantly reduces the amount of conjugations needed to check, about 20% of how many it would check otherwise
		for _, element := range adposuffixes {
			if vowel {
				if implContainsAny([]string{element}, []string{"ìl", "it", "ur", "ìri"}) {
					continue
				} else if element == "ä" {
					if !implContainsAny([]string{string(lastRune)}, []string{"u", "o"}) {
						continue
					}
				} else if element == "yä" {
					if implContainsAny([]string{string(lastRune)}, []string{"u", "o"}) {
						continue
					}
				}
			} else if diphthong {
				if implContainsAny([]string{element}, []string{"l", "ìri", "yä"}) {
					continue
				} else if element == "it" {
					if strings.HasSuffix(inputNavi, "ey") {
						continue
					}
				} else if element == "ur" {
					if strings.HasSuffix(inputNavi, "ew") {
						continue
					}
				}
			} else if consonant {
				if implContainsAny([]string{element}, []string{"l", "t", "r", "ri", "yä"}) {
					continue
				} else if element == "ru" && lastRune != '\'' {
					continue
				}
			}
			newWord := inputNavi + element
			reconjugateNouns(candidates, input, newWord, prefixCheck, 5, unlenite, affixCountdown-1)
		}
		fallthrough
	case 5:
		if strings.HasSuffix(inputNavi, "s") {
			// regardless of whether or not it's found
			newWord := inputNavi + "ì"
			reconjugateNouns(candidates, input, newWord, prefixCheck, 6, unlenite, affixCountdown-1)
		} else {
			// regardless of whether or not it's found
			newWord := inputNavi + "sì"
			reconjugateNouns(candidates, input, newWord, prefixCheck, 6, unlenite, affixCountdown-1)
		}
	}

	return nil
}

// Helper for ReconjugateVerbs
func removeBrackets(input string) string {
	input = strings.ReplaceAll(input, "<0>", "")
	input = strings.ReplaceAll(input, "<1>", "")
	input = strings.ReplaceAll(input, "<2>", "")
	return input
}

// Helper for StageThree, based on reconstruct from affixes.go
func reconjugateVerbs(candidates *[][]string, inputNavi string, prefirstUsed bool, firstUsed bool, secondUsed bool, affixLimit int8, add bool) error {
	if affixLimit == 0 {
		return nil
	}

	if add {
		inserted := true
		noBracket := removeBrackets(inputNavi)
		inserted = addToCandidates(candidates, noBracket)

		if !inserted {
			return nil
		}
	}

	if !prefirstUsed {
		reconjugateVerbs(candidates, strings.ReplaceAll(inputNavi, "<0>", ""), true, firstUsed, secondUsed, affixLimit-1, false)
		for _, a := range prefirst {
			reconjugateVerbs(candidates, strings.ReplaceAll(inputNavi, "<0>", a), true, firstUsed, secondUsed, affixLimit-1, true)
		}
		reconjugateVerbs(candidates, strings.ReplaceAll(inputNavi, "<0>", "äpeyk"), true, firstUsed, secondUsed, affixLimit-1, true)
	} else if !firstUsed {
		reconjugateVerbs(candidates, strings.ReplaceAll(inputNavi, "<1>", ""), prefirstUsed, true, secondUsed, affixLimit-1, false)
		for _, a := range first {
			reconjugateVerbs(candidates, strings.ReplaceAll(inputNavi, "<1>", a), prefirstUsed, true, secondUsed, affixLimit-1, true)
		}
	} else if !secondUsed {
		reconjugateVerbs(candidates, strings.ReplaceAll(inputNavi, "<2>", ""), prefirstUsed, firstUsed, true, affixLimit-1, false)
		for _, a := range second {
			reconjugateVerbs(candidates, strings.ReplaceAll(inputNavi, "<2>", a), prefirstUsed, firstUsed, true, affixLimit-1, true)
		}
	}

	return nil
}

func reconjugate(pigeonhole *[][]string, word Word, allowPrefixes bool, affixLimit int8) {
	// remove "+" and "--", we want to be able to search with and without those!
	word.Navi = strings.ReplaceAll(word.Navi, "+", "")
	word.Navi = strings.ReplaceAll(word.Navi, "--", "")
	word.Navi = strings.ToLower(word.Navi)

	if word.PartOfSpeech == "pn." {
		addToCandidates(pigeonhole, "nì"+word.Navi)
	}

	if word.PartOfSpeech == "n." || word.PartOfSpeech == "pn." || word.PartOfSpeech == "Prop.n." || word.PartOfSpeech == "inter." {
		reconjugateNouns(pigeonhole, word, word.Navi, 0, 0, 0, affixLimit)
	} else if word.PartOfSpeech[0] == 'v' {
		reconjugateVerbs(pigeonhole, word.InfixLocations, false, false, false, affixLimit, false)

		// v<us>erb and v<awn>erb (active and passive participles) with attributive markers
		for _, a := range []string{"us", "awn"} {
			participle := removeBrackets(strings.ReplaceAll(word.InfixLocations, "<1>", a))

			addToCandidates(pigeonhole, participle+"a")
		}

		//None of these can productively combine with infixes
		if allowPrefixes {
			// Gerunds
			gerund := removeBrackets("tì" + strings.ReplaceAll(word.InfixLocations, "<1>", "us"))

			reconjugateNouns(pigeonhole, word, gerund, 0, 0, 0, affixLimit-1)
			//candidates2 = append(candidates2, removeBrackets("nì"+strings.ReplaceAll(word.InfixLocations, "<1>", "awn")))
			// [verb]-able
			abilityVerbs := []string{"tsuk" + word.Navi, "suk" + word.Navi, "atsuk" + word.Navi,
				"tsuk" + word.Navi + "a", "ketsuk" + word.Navi, "hetsuk" + word.Navi, "aketsuk" + word.Navi,
				"ketsuk" + word.Navi + "a", "hetsuk" + word.Navi + "a"}
			for _, a := range abilityVerbs {
				addToCandidates(pigeonhole, a)
			}

			// v<us>erb and v<awn>erb (active and passive participles) with attributive markers
			for _, a := range []string{"us", "awn"} {
				participle := removeBrackets(strings.ReplaceAll(word.InfixLocations, "<1>", a))
				addToCandidates(pigeonhole, "a"+participle)
			}
		}
		// Ability to [verb]
		reconjugateNouns(pigeonhole, word, word.Navi+"tswo", 0, 0, 0, affixLimit-1)
		reconjugateNouns(pigeonhole, word, word.Navi+"yu", 0, 0, 0, affixLimit-1)

	} else if word.PartOfSpeech == "adj." {
		addToCandidates(pigeonhole, word.Navi+"a")

		if allowPrefixes {
			addToCandidates(pigeonhole, "a"+word.Navi)
			addToCandidates(pigeonhole, "nì"+word.Navi)
		}
	}
}

// Make alphabetized lists of strings
// For this specific use case, consistency is more important than accuracy
func AppendStringAlphabetically(array []string, addition string) []string {
	newArray := []string{}
	appended := false
	for _, a := range array {
		if !appended && a > addition {
			newArray = append(newArray, addition)
			appended = true
		}
		newArray = append(newArray, a)
	}
	if !appended {
		newArray = append(newArray, addition)
	}
	return newArray
}

// modified from https://www.slingacademy.com/article/how-to-find-common-elements-of-2-slices-in-go/
func findUniques(affixes [][]string, reverse bool) string {
	var uniques strings.Builder

	if reverse {
		for i := range affixes {
			slices.Reverse(affixes[i])
		}
	}

	all := map[string]bool{}
	checked := map[string]bool{}

	uniqueSlice := []string{}

	if len(affixes) > 1 {
		for _, a := range affixes {
			for _, aPrime := range a {
				all[aPrime] = true
			}
		}
		// compare all of one array
		for i, a := range affixes {
			// to the arrays after
			for _, b := range affixes[i+1:] {
				for _, aPrime := range a {
					if _, ok := checked[aPrime]; ok {
						continue
					}

					checked[aPrime] = true

					// Find kusara, kawnara, tsukkara and kestukkara
					// also kìm, kìmyu and kìmtswo
					if _, ok := changePOS[aPrime]; ok {
						uniqueSlice = append(uniqueSlice, aPrime)
						continue
					}

					for _, bPrime := range b {
						if aPrime == bPrime {
							all[aPrime] = false
							break
						}
					}
				}

				for _, bPrime := range b {
					if _, ok := checked[bPrime]; ok {
						continue
					}

					checked[bPrime] = true
					for _, aPrime := range a {
						if aPrime == bPrime {
							all[aPrime] = false
							break
						}
					}
				}
			}
		}

		for key, val := range all {
			if val {
				uniqueSlice = append(uniqueSlice, key)
			}
		}

		slices.SortFunc(uniqueSlice, func(i, j string) int { return strings.Compare(i, j) })

		for _, a := range uniqueSlice {
			uniques.WriteString(a)
		}
	}

	return uniques.String()
}

func CheckHomsAsync(dict *FwewDict, minAffix int) {
	defer checkWaitGroup.Done()
	wait := false
	firstWait := true
	start2 := time.Now()
	makingFinished := false
	for !makingFinished {

		// Don't pull from empty
		for candidates2.Length() == 0 {
			time.Sleep(time.Millisecond * 5)

			// Make sure it's not finished first
			finished.mu.Lock()
			makingFinished = finished.finished
			finished.mu.Unlock()

			if makingFinished {
				break
			}
		}

		if makingFinished {
			break
		}

		a, err0 := candidates2.Remove()

		if a == finishedSentinelValue {
			finished.mu.Lock()
			finished.finished = true
			finished.mu.Unlock()
			break
		}

		wordNumber, err1 := strconv.Atoi(a)

		if err1 == nil {
			if wordNumber%progressInterval == 0 {
				total_seconds := time.Since(start)

				printMessage := "Word " + strconv.Itoa(wordNumber) + " is in dict " +
					strconv.Itoa(int(dict.dictNum)) + ".  Time elapsed is " +
					strconv.Itoa(int(math.Floor(total_seconds.Hours()))) + " hours, " +
					strconv.Itoa(int(math.Floor(total_seconds.Minutes()))%60) + " minutes and " +
					strconv.Itoa(int(total_seconds.Seconds())%60) + " seconds.  " + strconv.Itoa(totalCandidates) + " conjugations checked"

				log.Printf(printMessage)
				resultsFile.WriteString(printMessage + "\n")
			}
			continue
		}

		if err0 != nil {
			if !wait {
				start2 = time.Now()
				wait = true
				continue
			}
			continue
		}

		if inefficiencyWarning && wait {
			wait = false
			waitedString := "Dictionary " + strconv.Itoa(int(dict.dictNum)) + " waited " + strconv.FormatInt(time.Since(start2).Milliseconds(), 10) + "ms"
			if !firstWait {
				waitedString += "\nThis should only have happened at the beginning"
			} else {
				firstWait = false
			}
			fmt.Println(waitedString)
			resultsFile.WriteString(waitedString + "\n")
		}

		totalCandidates++

		//Nasal assimilation stuff
		if nasalAssimilationOnly {
			invalidSuffix := false
			suffixesThings := []string{"tsyìpna", "tsyìpne", "fkeykna", "fkeykne", "tsyìpnuä", "fkeyknuä", "tsyìpnue", "fkeyknue"}
			for _, suffix := range suffixesThings {
				if strings.HasSuffix(a, suffix) || strings.HasSuffix(a, suffix+"sì") {
					invalidSuffix = true
					break
				}
			}

			if invalidSuffix {
				continue
			}

			tempA := strings.ReplaceAll(a, "nts", "")
			tempA = strings.ReplaceAll(tempA, "mts", "")
			tempA = strings.ReplaceAll(tempA, "ngts", "")

			containsNasal := false

			for _, t := range []string{"t", "k", "p", "tx", "kx", "px"} {
				for _, n := range []string{"n", "ng", "m"} {
					if strings.Contains(tempA, n+t) || strings.Contains(tempA, t+n) {
						containsNasal = true
						break
					}
				}
				if containsNasal {
					break
				}
			}

			if !containsNasal {
				continue
			}
		}

		// These can clog up the search results
		if strings.HasSuffix(a, "rofa") || strings.HasSuffix(a, "rofasì") {
			continue
		}

		results, err := TranslateFromNaviHash(dict, a, true)

		if err == nil && len(results) > 0 && len(results[0]) > 2 {

			results[0] = results[0][1:]

			homoMapQuery, show := QueryHelper(results[0])

			// No duplicates

			if !homoMap.Present(homoMapQuery) {
				homoMap.Insert(homoMapQuery)

				// No duplicates from previous
				if first2StageMap.Present(strings.ToLower(a)) {
					continue
				}

				stringy := "dict " + strconv.Itoa(int(dict.dictNum)) + ": [" + a + " " + results[0][0].Navi + "] [" + homoMapQuery

				err := foundResult(a, stringy, show)
				if err != nil {
					fmt.Println("Error writing to file:", err)
					return
				}
			}
		}
		/*if len(strings.Split(a, " ")) > 1 {
			fmt.Println("oops " + a)
			continue
		}
		runeCount := uint8(len(a.navi))
		if runeCount < 40 {
			continue
		}
		if runeCount > longest {
			longest = runeCount
			//fmt.Println(a)
		}
		if _, ok := top10Longest[runeCount]; ok {
			top10Longest[runeCount] = top10Longest[runeCount] + " " + a.navi
		} else {
			top10Longest[runeCount] = a.navi
		}*/
	}

	printMessage := "Dictionary " + strconv.Itoa(int(dict.dictNum)) + " finished"

	log.Printf(printMessage)
	resultsFile.WriteString(printMessage + "\n")
}

func foundResult(conjugation string, homonymfo string, show bool) error {
	writeLock.Lock()
	defer writeLock.Unlock()
	resultCount++
	if show {
		fmt.Println(homonymfo)
		_, err := resultsFile.WriteString(homonymfo + "\n")
		if err != nil {
			return err
		}
	}
	lowercase := strings.ToLower(conjugation)
	_, err2 := previous.WriteString(lowercase + "\n")

	first2StageMap.Insert(lowercase)

	return err2
}

func addHomsAsync(pigeonhole *[][]string) {
	defer addWaitGroup.Done()
	low := !inefficiencyWarning
	lengthy := candidates2.Length()

	for _, alpha := range *pigeonhole {
		for _, a := range alpha {
			if !low && inefficiencyWarning && lengthy == 0 {
				waitedString := "Queue reached 0.  This should only happen at the beginning"
				fmt.Println(waitedString)
				resultsFile.WriteString(waitedString + "\n")
				low = true
			}
			//start2 := time.Now()
			err3 := candidates2.Insert(a)

			if err3 != nil {
				for candidates2.Length() > 8000 {
					time.Sleep(time.Millisecond * 5)
				}
				candidates2.Insert(a)
			}

			//fmt.Println("waited " + strconv.FormatInt(time.Since(start2).Milliseconds(), 10) + "ms"
		}
	}
}

func makeHomsAsync(affixLimit int8, startNumber int) error {
	defer makeWaitGroup.Done()
	wordCount = 0

	err := RunOnDict(func(word Word) error {
		wordCount += 1
		//checkAsyncLock.Wait()

		if wordCount >= startNumber {
			// Reset dupe detector so it's not taking up all the RAM
			stage3Map.Clear()

			pigeonhole := make([][]string, charLimit+2)

			pigeonhole[1] = append(pigeonhole[1], word.Navi)

			//candidates2slice := []candidate{{navi: word.Navi, length: len([]rune(word.Navi))}} //empty array of strings

			// Let the dictionary threads know that we are on number worcCount
			if wordCount%progressInterval == 0 {
				pigeonhole[0] = append(pigeonhole[0], strconv.Itoa(wordCount))
			}

			// No multiword words
			if !strings.Contains(word.Navi, " ") {

				// Get conjugations
				reconjugate(&pigeonhole, word, true, affixLimit)

				/*slices.SortStableFunc(candidates2slice, func(i, j candidate) int {
					return i.length - j.length
				})*/
			} else if strings.HasSuffix(word.Navi, " si") {
				// "[word] si" can take the form "[word]tswo"
				siless := strings.TrimSuffix(word.Navi, " si")

				reconjugateNouns(&pigeonhole, word, siless+"tswo", 0, 0, 0, affixLimit)
				reconjugateNouns(&pigeonhole, word, siless+"siyu", 0, 0, 0, affixLimit)
			}

			addWaitGroup.Wait()
			addWaitGroup.Add(1)
			go addHomsAsync(&pigeonhole)
		}

		return nil
	})

	candidates2.Insert(finishedSentinelValue)
	fmt.Println("Finished making word candidates")
	resultsFile.WriteString("Finished making word candidates\n")

	return err
}

func StageThree(dictCount uint8, minAffix int, affixLimit int8, charLimitSet int, startNumber int,
	inefficiencyWarningSet bool, progressIntervalSet int) (err error) {

	inefficiencyWarning = inefficiencyWarningSet
	charLimit = charLimitSet
	progressInterval = progressIntervalSet

	start = time.Now()

	resultsFile.WriteString("Stage 3\n")

	if startNumber > len(dictHash) {
		return errors.New("startNumber is longer than the provided dictionary")
	}

	if progressIntervalSet <= 0 {
		return errors.New("progress interval must be 1 or greater")
	}

	resultsFile.WriteString(strconv.Itoa(int(affixLimit)) + " affix and " + strconv.Itoa(int(charLimit)) + " character limits\n")
	fmt.Println(strconv.Itoa(int(affixLimit)) + " affix and " + strconv.Itoa(int(charLimit)) + " character limits")

	makeWaitGroup.Add(1)
	go makeHomsAsync(affixLimit, startNumber)
	for _, dict := range dictArray {
		checkWaitGroup.Add(1)
		go CheckHomsAsync(dict, minAffix)
	}

	makeWaitGroup.Wait()

	checkWaitGroup.Wait()

	fmt.Println("All dictionaries finished")
	resultsFile.WriteString("All dictionaries finished\n")

	//fmt.Println(homoMap)
	//fmt.Println(tempHoms)

	total_seconds := time.Since(start)

	finalString := "Stage three took " + strconv.Itoa(int(math.Floor(total_seconds.Hours()))) + " hours, " +
		strconv.Itoa(int(math.Floor(total_seconds.Minutes()))%60) + " minutes and " +
		strconv.Itoa(int(total_seconds.Seconds())%60) + " seconds"
	log.Printf(finalString)
	resultsFile.WriteString(finalString + "\n")

	checkedString := "Narrowed from " + strconv.Itoa(totalCandidates) + " conjugations to " + strconv.Itoa(resultCount)
	fmt.Println(checkedString)
	resultsFile.WriteString(checkedString + "\n")

	/*fmt.Println(longest)
	resultsFile.WriteString(strconv.Itoa(int(longest)) + "\n")
	fmt.Println(top10Longest[longest])
	resultsFile.WriteString(top10Longest[longest] + "\n")*/

	//fmt.Println(dupeLengthsMap)

	return
}

// Do everything
func homonymSearch() error {
	if _, err := os.Stat("results.txt"); err == nil {
		// path/to/whatever exists
		fmt.Println("results.txt exists.  Please rename or delete it so it's not overwritten")
		return err
	} else if errors.Is(err, os.ErrNotExist) {
		// path/to/whatever does *not* exist
		a, err2 := os.Create("results.txt")
		if err2 != nil {
			fmt.Println("error opening file:", err2)
			return err2
		}
		resultsFile = a
	} else {
		// Schrodinger: file may or may not exist. See err for details.

		// Therefore, do *NOT* use !os.IsNotExist(err) to test for file existence

		fmt.Println("An error occured determining whether or not results.txt exists")
		return err
	}

	defer resultsFile.Close()

	// We'll need this for the previous file
	homoMap.homoMap = map[string]bool{}
	first2StageMap.homoMap = map[string]bool{}
	stage3Map.homoMap = map[string]bool{}

	if _, err := os.Stat("previous.txt"); err == nil {
		// path/to/whatever exists
		b, err2 := os.Open("previous.txt")
		if err2 != nil {
			fmt.Println("error opening file:", err2)
			return err2
		}

		allWords := []string{}

		scanner := bufio.NewScanner(b)
		// This will not read lines over 64k long, but works for Na'vi words just fine
		for scanner.Scan() {
			first2StageMap.Insert(scanner.Text())
			allWords = append(allWords, scanner.Text())
		}

		if err := scanner.Err(); err != nil {
			log.Fatal(err)
		}

		sort.Slice(allWords, func(i, j int) bool {
			return AlphabetizeHelper(allWords[i], allWords[j])
		})

		a, err := os.Create("previous.txt")
		if err != nil {
			fmt.Println("error opening file:", err)
			return err
		}
		ourDict := FwewDictInit(uint8(0))
		for _, word := range allWords {
			resultCount++
			// Make sure it knows the signatures of the older words so it doesn't duplicate them
			results, err := TranslateFromNaviHash(ourDict, word, true)

			if err == nil && len(results) > 0 && len(results[0]) > 2 {

				results[0] = results[0][1:]

				homoMapQuery, _ := QueryHelper(results[0])

				// No duplicates
				if !homoMap.Present(homoMapQuery) {
					homoMap.Insert(homoMapQuery)
				}
			}
			a.WriteString(word + "\n")
		}

		previous = a
	} else if errors.Is(err, os.ErrNotExist) {
		// path/to/whatever does *not* exist
		a, err := os.Create("previous.txt")
		if err != nil {
			fmt.Println("error opening file:", err)
			return err
		}
		previous = a
	} else {
		// Schrodinger: file may or may not exist. See err for details.

		// Therefore, do *NOT* use !os.IsNotExist(err) to test for file existence

		fmt.Println("An error occured determining whether or not previous.txt exists")
		return err
	}

	defer previous.Close()

	dictCount := uint8(4)
	for i := uint8(0); i < dictCount; i++ {
		dictArray = append(dictArray, FwewDictInit(i+1))
	}

	fmt.Println("Stage 1:")
	StageOne()
	fmt.Println("Stage 2:")
	StageTwo()
	fmt.Println("Stage 3:")
	// number of dictionaries, minimum affixes, maximum affixes, maximum word length, start at word number N
	// warn about inefficiencies, Progress updates after checking every N number of words
	StageThree(dictCount, 0, 4, 14, 0, true, 100)
	// For nasal assimilation mode, change nasalAssimilationOnly variable at the top of this file.

	return nil
}
