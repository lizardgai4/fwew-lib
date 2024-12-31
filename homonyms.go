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
var candidates2 Queue = *CreateQueue(30000)
var candidates2Map = map[string]bool{}
var homoMap = HomoMapStruct{}
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

/*
var top10Longest = map[uint8]string{}
var longest uint8 = 0
*/
var totalCandidates int = 0
var charLimit int = 14
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
var previousWords = map[string]bool{}

//var dupeLengthsMap = map[int]int{}

var finished = queueFinished{false, sync.Mutex{}}

var dictArray = []*FwewDict{}

type queueFinished struct {
	finished bool
	mu       sync.Mutex
}

type candidate struct {
	navi   string
	length uint8
}

type Queue struct {
	mu       sync.Mutex
	capacity int
	q        []string
}

type HomoMapStruct struct {
	mu      sync.Mutex
	homoMap map[string]int
}

var writeLock sync.Mutex
var makeWaitGroup sync.WaitGroup
var checkWaitGroup sync.WaitGroup

// FifoQueue
type FifoQueue interface {
	Insert()
	Remove()
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

		if _, ok := previousWords[standardizedWord]; !ok {
			// If the word appears more than once, record it
			if entry, ok := dictHash[standardizedWord]; ok {
				if len(entry) > 1 {
					query := QueryHelper(entry)
					foundResult(standardizedWord, query)
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

func QueryHelper(results []Word) string {
	sort.Slice(results, func(i, j int) bool {
		if results[i].Navi != results[j].Navi {
			return results[i].Navi < results[j].Navi
		}

		if len(results[i].Affixes.Prefix) != len(results[j].Affixes.Prefix) {
			return len(results[i].Affixes.Prefix) < len(results[j].Affixes.Prefix)
		}

		if len(results[i].Affixes.Suffix) != len(results[j].Affixes.Suffix) {
			return len(results[i].Affixes.Suffix) < len(results[j].Affixes.Suffix)
		}

		return len(results[i].Affixes.Infix) < len(results[j].Affixes.Infix)
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

	allNaviWords.WriteString("] ")

	preUnique := findUniques(allPrefixes, false)
	allNaviWords.WriteString(preUnique)

	allNaviWords.WriteString("-")

	inUnique := findUniques(allInfixes, false)
	allNaviWords.WriteString(inUnique)

	allNaviWords.WriteString("-")

	sufUnique := findUniques(allSuffixes, true)
	allNaviWords.WriteString(sufUnique)

	return allNaviWords.String()
}

// Helper to turn a string into a list of known words

// Check for ones that are the exact same, no affixes needed
func StageTwo() error {
	resultsFile.WriteString("Stage 2:\n")

	err := runOnFile(func(word Word) error {
		if _, ok := previousWords[strings.ToLower(word.Navi)]; !ok {
			standardizedWord := word.Navi

			candidates2Map[word.Navi] = true

			if len(strings.Split(word.Navi, " ")) == 1 {
				allNaviWords := ""

				// If the word can conjugate into something else, record it
				results, err := TranslateFromNaviHash(dictArray[0], standardizedWord, true)
				if err == nil && len(results[0]) > 2 {
					results[0] = results[0][1:]
					allNaviWords = QueryHelper(results[0])
					foundResult(standardizedWord, allNaviWords)
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
func addToCandidates(candidates []candidate, candidate1 string) []candidate {
	newLength := len([]rune(candidate1))
	// Is it longer than the words we want to check?
	if newLength > charLimit {
		return candidates
	}

	// If it's in the range, is it good?
	if _, ok := candidates2Map[candidate1]; !ok {
		candidates = append(candidates, candidate{navi: candidate1, length: uint8(newLength)})
		totalCandidates++
		candidates2Map[candidate1] = true
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
	
	if !found {
		return candidates
	}
	
	// If it's in the range, is it good?
	if _, ok := candidates2Map[lenited]; !ok {
		candidates = append(candidates, candidate{navi: lenited], length: uint8(len([]rune(lenited))})
		totalCandidates++
		candidates2Map[candidate1] = true
	}

	return candidates
}

// Helper for StageThree, based on reconstruct from affixes.go
func reconjugateNouns(candidates *[]candidate, input Word, inputNavi string, prefixCheck int, suffixCheck int, unlenite int8, affixCountdown int8) error {
	// End state: Limit to 2 affixes per noun
	if affixCountdown == 0 {
		return nil
	}

	runeLen := len([]rune(inputNavi))

	if runeLen > charLimit {
		return nil
	}

	*candidates = addToCandidates(*candidates, inputNavi)

	switch prefixCheck {
	case 0:
		for _, element := range stemPrefixes {
			// If it has a lenition-causing prefix
			newWord := element + inputNavi
			reconjugateNouns(candidates, input, newWord, 1, suffixCheck, 0, affixCountdown-1)
		}
		fallthrough
	case 1:
		fallthrough
	case 2:
		// Non-lenition prefixes for nouns only
		for _, element := range prefixes1Nouns {
			newWord := element + inputNavi
			reconjugateNouns(candidates, input, newWord, 4, suffixCheck, 0, affixCountdown-1)
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

			// regardless of whether or not it's found
			lenited2 := element + lenited
			reconjugateNouns(candidates, input, lenited2, 4, suffixCheck, -1, affixCountdown-1)
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
			newWord := inputNavi + element
			reconjugateNouns(candidates, input, newWord, prefixCheck, 2, unlenite, affixCountdown-1)
		}
		fallthrough
	case 2:
		newWord := inputNavi + "o"
		reconjugateNouns(candidates, input, newWord, prefixCheck, 3, unlenite, affixCountdown-1)
		fallthrough
	case 3:
		newWord := inputNavi + "pe"
		reconjugateNouns(candidates, input, newWord, prefixCheck, 4, unlenite, affixCountdown-1)
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
		newWord := inputNavi + "sì"
		reconjugateNouns(candidates, input, newWord, prefixCheck, 6, unlenite, affixCountdown-1)
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
func reconjugateVerbs(candidates *[]candidate, inputNavi string, prefirstUsed bool, firstUsed bool, secondUsed bool, affixLimit int8) error {
	if affixLimit == 0 {
		return nil
	}

	noBracket := removeBrackets(inputNavi)
	*candidates = addToCandidates(*candidates, noBracket)

	if !prefirstUsed {
		reconjugateVerbs(candidates, strings.ReplaceAll(inputNavi, "<0>", ""), true, firstUsed, secondUsed, affixLimit-1)
		for _, a := range prefirst {
			reconjugateVerbs(candidates, strings.ReplaceAll(inputNavi, "<0>", a), true, firstUsed, secondUsed, affixLimit-1)
		}
		reconjugateVerbs(candidates, strings.ReplaceAll(inputNavi, "<0>", "äpeyk"), true, firstUsed, secondUsed, affixLimit-1)
	} else if !firstUsed {
		reconjugateVerbs(candidates, strings.ReplaceAll(inputNavi, "<1>", ""), prefirstUsed, true, secondUsed, affixLimit-1)
		for _, a := range first {
			reconjugateVerbs(candidates, strings.ReplaceAll(inputNavi, "<1>", a), prefirstUsed, true, secondUsed, affixLimit-1)
		}
	} else if !secondUsed {
		reconjugateVerbs(candidates, strings.ReplaceAll(inputNavi, "<2>", ""), prefirstUsed, firstUsed, true, affixLimit-1)
		for _, a := range second {
			reconjugateVerbs(candidates, strings.ReplaceAll(inputNavi, "<2>", a), prefirstUsed, firstUsed, true, affixLimit-1)
		}
	}

	return nil
}

func reconjugate(word Word, allowPrefixes bool, affixLimit int8) []candidate {
	// remove "+" and "--", we want to be able to search with and without those!
	word.Navi = strings.ReplaceAll(word.Navi, "+", "")
	word.Navi = strings.ReplaceAll(word.Navi, "--", "")
	word.Navi = strings.ToLower(word.Navi)

	candidatesSlice := []candidate{}

	candidatesSlice = addToCandidates(candidatesSlice, word.Navi)

	if word.PartOfSpeech == "pn." {
		candidatesSlice = addToCandidates(candidatesSlice, "nì"+word.Navi)
	}

	if word.PartOfSpeech == "n." || word.PartOfSpeech == "pn." || word.PartOfSpeech == "Prop.n." || word.PartOfSpeech == "inter." {
		reconjugateNouns(&candidatesSlice, word, word.Navi, 0, 0, 0, affixLimit)
	} else if word.PartOfSpeech[0] == 'v' {
		reconjugateVerbs(&candidatesSlice, word.InfixLocations, false, false, false, affixLimit)

		// v<us>erb and v<awn>erb (active and passive participles) with attributive markers
		for _, a := range []string{"us", "awn"} {
			participle := removeBrackets(strings.ReplaceAll(word.InfixLocations, "<1>", a))

			candidatesSlice = addToCandidates(candidatesSlice, participle+"a")
		}

		//None of these can productively combine with infixes
		if allowPrefixes {
			// Gerunds
			gerund := removeBrackets("tì" + strings.ReplaceAll(word.InfixLocations, "<1>", "us"))
			lenitedGerund := "s" + strings.TrimPrefix(gerund, "t")

			candidatesSlice = addToCandidates(candidatesSlice, gerund)

			reconjugateNouns(&candidatesSlice, word, gerund, 0, 0, 0, affixLimit-1)
			//candidates2 = append(candidates2, removeBrackets("nì"+strings.ReplaceAll(word.InfixLocations, "<1>", "awn")))
			// [verb]-able
			abilityVerbs := []string{"tsuk" + word.Navi, "suk" + word.Navi, "atsuk" + word.Navi,
				"tsuk" + word.Navi + "a", "ketsuk" + word.Navi, "hetsuk" + word.Navi, "aketsuk" + word.Navi,
				"ketsuk" + word.Navi + "a", "hetsuk" + word.Navi + "a"}
			for _, a := range abilityVerbs {
				candidatesSlice = addToCandidates(candidatesSlice, a)
			}

			// v<us>erb and v<awn>erb (active and passive participles) with attributive markers
			for _, a := range []string{"us", "awn"} {
				participle := removeBrackets(strings.ReplaceAll(word.InfixLocations, "<1>", a))
				candidatesSlice = addToCandidates(candidatesSlice, "a"+participle)
			}

			//Lenited forms, too
			candidatesSlice = addToCandidates(candidatesSlice, lenitedGerund)
			reconjugateNouns(&candidatesSlice, word, lenitedGerund, 10, 0, -1, affixLimit-1)
		}
		// Ability to [verb]
		reconjugateNouns(&candidatesSlice, word, word.Navi+"tswo", 0, 0, 0, affixLimit-1)
		reconjugateNouns(&candidatesSlice, word, word.Navi+"yu", 0, 0, 0, affixLimit-1)

	} else if word.PartOfSpeech == "adj." {
		candidatesSlice = addToCandidates(candidatesSlice, word.Navi+"a")

		if allowPrefixes {
			candidatesSlice = addToCandidates(candidatesSlice, "a"+word.Navi)
			candidatesSlice = addToCandidates(candidatesSlice, "nì"+word.Navi)
		}
	}

	return candidatesSlice
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
		for i, _ := range affixes {
			slices.Reverse(affixes[i])
		}
	}

	all := map[string]bool{}
	checked := map[string]bool{}

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
						uniques.WriteString(aPrime)
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
				uniques.WriteString(key)
			}
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
		for len(candidates2.q) == 0 {
			time.Sleep(time.Millisecond * 5)
		}

		a, _ := candidates2.Remove()

		if a == "" {
			if !wait {
				start2 = time.Now()
				wait = true
				continue
			}
			continue
		}

		if wait {
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

		/*if strings.HasSuffix(a.navi, "tsyìpna") {
			continue
		}

		tempA := strings.ReplaceAll(a.navi, "nts", "")
		tempA = strings.ReplaceAll(tempA, "mts", "")
		tempA = strings.ReplaceAll(tempA, "ngts", "")

		//Nasal assimilation stuff
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
		}*/

		// These can clog up the search results
		if strings.HasSuffix(a, "rofa") || strings.HasSuffix(a, "rofasì") {
			continue
		}

		results, err := TranslateFromNaviHash(dict, a, true)

		if err == nil && len(results) > 0 && len(results[0]) > 2 {

			results[0] = results[0][1:]

			homoMapQuery := QueryHelper(results[0])

			// No duplicates

			homoMap.mu.Lock()
			if _, ok := homoMap.homoMap[homoMapQuery]; !ok {
				homoMap.homoMap[homoMapQuery] = 1

				// No duplicates from previous
				if _, ok := previousWords[strings.ToLower(a)]; ok {
					homoMap.mu.Unlock()
					continue
				}

				stringy := "dict " + strconv.Itoa(int(dict.dictNum)) + ": [" + a + " " + results[0][0].Navi + "] [" + homoMapQuery

				err := foundResult(a, stringy)
				if err != nil {
					fmt.Println("Error writing to file:", err)
					homoMap.mu.Unlock()
					return
				}
			}
			homoMap.mu.Unlock()
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

		finished.mu.Lock()
		makingFinished = finished.finished
		finished.mu.Unlock()
	}
}

func foundResult(conjugation string, homonymfo string) error {
	writeLock.Lock()
	defer writeLock.Unlock()
	fmt.Println(homonymfo)
	_, err := resultsFile.WriteString(homonymfo + "\n")
	if err != nil {
		return err
	}
	lowercase := strings.ToLower(conjugation)
	_, err2 := previous.WriteString(lowercase + "\n")
	previousWords[strings.ToLower(lowercase)] = true
	return err2
}

func makeHomsAsync(affixLimit int8, startNumber int, start time.Time) error {
	defer makeWaitGroup.Done()
	wordCount := 0

	err := RunOnDict(func(word Word) error {
		wordCount += 1
		//checkAsyncLock.Wait()

		if wordCount >= startNumber {
			// Reset dupe detector so it's not taking up all the RAM
			clear(candidates2Map)
			candidates2slice := []candidate{{navi: word.Navi, length: uint8(len([]rune(word.Navi)))}} //empty array of strings

			// Progress counter
			if wordCount%100 == 0 {
				total_seconds := time.Since(start)

				printMessage := "On word " + strconv.Itoa(wordCount) + ".  Time elapsed is " +
					strconv.Itoa(int(math.Floor(total_seconds.Hours()))) + " hours, " +
					strconv.Itoa(int(math.Floor(total_seconds.Minutes()))%60) + " minutes and " +
					strconv.Itoa(int(total_seconds.Seconds())%60) + " seconds.  " + strconv.Itoa(totalCandidates) + " conjugations checked"

				log.Printf(printMessage)
				resultsFile.WriteString(printMessage + "\n")
			}

			// No multiword words
			if !strings.Contains(word.Navi, " ") {

				// Get conjugations
				candidates2slice = append(candidates2slice, reconjugate(word, true, affixLimit)...)

				sort.SliceStable(candidates2slice, func(i, j int) bool {
					return candidates2slice[i].length < candidates2slice[j].length
				})
			} else if strings.HasSuffix(word.Navi, " si") {
				// "[word] si" can take the form "[word]tswo"
				siless := strings.TrimSuffix(word.Navi, " si")
				
				reconjugateNouns(&candidates2slice, word, siless+"tswo", 0, 0, 0, affixLimit)
				reconjugateNouns(&candidates2slice, word, siless+"siyu", 0, 0, 0, affixLimit)

				sort.SliceStable(candidates2slice, func(i, j int) bool {
					return candidates2slice[i].length < candidates2slice[j].length
				})
			}

			for _, a := range candidates2slice {
				err3 := candidates2.Insert(a.navi)
				if err3 != nil {
					//start2 := time.Now()
					for len(candidates2.q) > 15000 {
						time.Sleep(time.Millisecond * 5)
					}
					candidates2.Insert(a.navi)
					//fmt.Println("waited " + strconv.FormatInt(time.Since(start2).Milliseconds(), 10) + "ms"
				}
			}
		}

		return nil
	})

	finished.mu.Lock()
	finished.finished = true
	finished.mu.Unlock()

	return err
}

func StageThree(dictCount uint8, minAffix int, affixLimit int8, charLimitSet int, startNumber int) (err error) {
	homoMap.mu.Lock()
	homoMap.homoMap = map[string]int{}
	homoMap.mu.Unlock()

	charLimit = charLimitSet
	start := time.Now()

	resultsFile.WriteString("Stage 3\n")

	if startNumber > len(dictHash) {
		return errors.New("startNumber is longer than the provided dictionary")
	}

	makeWaitGroup.Add(1)
	go makeHomsAsync(affixLimit, startNumber, start)
	for _, dict := range dictArray {
		checkWaitGroup.Add(1)
		go CheckHomsAsync(dict, minAffix)
	}

	makeWaitGroup.Wait()
	checkWaitGroup.Wait()

	finishedBool := false
	for !finishedBool {
		time.Sleep(time.Second)
		finished.mu.Lock()
		finishedBool = finished.finished
		finished.mu.Unlock()
	}

	//fmt.Println(homoMap)
	//fmt.Println(tempHoms)

	total_seconds := time.Since(start)

	finalString := "Stage three took " + strconv.Itoa(int(math.Floor(total_seconds.Hours()))) + " hours, " +
		strconv.Itoa(int(math.Floor(total_seconds.Minutes()))%60) + " minutes and " +
		strconv.Itoa(int(total_seconds.Seconds())%60) + " seconds"
	log.Printf(finalString)
	resultsFile.WriteString(finalString + "\n")

	checkedString := "Checked " + strconv.Itoa(totalCandidates) + " total conjugations"
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
			previousWords[scanner.Text()] = true
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
		for _, word := range allWords {
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

	dictCount := uint8(8)
	for i := uint8(0); i < dictCount; i++ {
		dictArray = append(dictArray, FwewDictInit(i+1))
	}

	fmt.Println("Stage 1:")
	StageOne()
	fmt.Println("Stage 2:")
	StageTwo()
	fmt.Println("Stage 3:")
	// number of dictionaries, minimum affixes, maximum affixes, maximum word length, start at word number N
	StageThree(dictCount, 0, 127, 127, 0)

	return nil
}
