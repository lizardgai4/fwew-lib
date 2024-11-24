package fwew_lib

import (
	"fmt"
	"log"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

var homonymsArray = []string{"", "", ""}
var candidates2 []candidate
var candidates2Map = map[string]int{}
var homoMap = map[string]int{}
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
var top10Longest = map[uint8]string{}
var longest uint8 = 0

var checkAsyncLock = sync.WaitGroup{}

type tempHomsContainer struct {
	mu       sync.Mutex
	counters []string
}

type candidate struct {
	navi   string
	length uint8
}

func (c *tempHomsContainer) addTempHom(name string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.counters = append(c.counters, name)
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
	tempHoms := []string{}

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

		// If the word appears more than once, record it
		if _, ok := dictHash[standardizedWord]; ok {
			found := false
			for _, a := range tempHoms {
				if a == standardizedWord {
					found = true
					break
				}
			}
			if !found {
				tempHoms = append(tempHoms, standardizedWord)
			}
		}
		if strings.Contains(standardizedWord, "é") {
			noAcute := strings.ReplaceAll(standardizedWord, "é", "e")
			found := false
			for _, a := range tempHoms {
				if a == noAcute {
					found = true
					break
				}
			}
			if !found {
				tempHoms = append(tempHoms, noAcute)
				tempHoms = append(tempHoms, standardizedWord)
			}
		}

		return nil
	})

	// Reverse the order to make accidental and new homonyms easier to see
	// Also make it a string for easier searching
	i := len(tempHoms)
	for i > 0 {
		i--
		homonymsArray[0] += tempHoms[i] + "000 "
	}

	homonymsArray[0] = strings.TrimSuffix(homonymsArray[0], " ")

	if err != nil {
		log.Printf("Error in homonyms stage 1: %s", err)
		return err
	}

	return nil
}

// Helper to detect presences of affixes
func AffixCount(word Word) string {
	prefixCount := "0"
	infixCount := "0"
	suffixCount := "0"

	if len(word.Affixes.Prefix) > 0 {
		prefixCount = "1"
	}
	if len(word.Affixes.Infix) > 0 {
		infixCount = "1"
	}
	if len(word.Affixes.Suffix) > 0 {
		suffixCount = "1"
	}

	//fmt.Println(prefixCount + infixCount + suffixCount)

	return prefixCount + infixCount + suffixCount
}

// Helper to turn a string into a list of known words

// Check for ones that are the exact same, no affixes needed
func StageTwo() error {
	tempHoms := []string{}

	err := runOnFile(func(word Word) error {
		standardizedWord := word.Navi

		candidates2Map[word.Navi] = 1

		homList := []string{}
		// If the word can conjugate into something else, record it
		results, err := TranslateFromNaviHash(standardizedWord, true)
		if err == nil && len(results[0]) > 2 {
			dupes := []string{}

			allNaviWords := ""
			for i, a := range results[0] {
				if i != 0 { //&& i < 3 {
					dupe := false
					dupeToFind := a.Navi + AffixCount(a)
					for _, b := range dupes {
						if b == dupeToFind {
							dupe = true
							break
						}
					}
					if dupe {
						continue
					}

					dupes = append(dupes, dupeToFind)
					tempHoms = append(tempHoms, a.Navi+AffixCount(a))
					homList = AppendStringAlphabetically(homList, a.Navi+AffixCount(a))
				}
			}

			if len(homList) >= 2 {
				for _, a := range homList {
					allNaviWords += a + " "
				}

				homoMap[allNaviWords] = 1
				fmt.Println(strconv.Itoa(len(results[0])) + " " + allNaviWords + " " + standardizedWord)
			}
		}

		//Lenited forms, too
		found := false
		for _, a := range lenitors {
			if strings.HasPrefix(word.Navi, a) {
				word.Navi = strings.TrimPrefix(word.Navi, a)
				word.Navi = lenitionMap[a] + word.Navi
				found = true
				break
			}
		}
		if found {
			// If the word can conjugate into something else, record it
			results, err := TranslateFromNaviHash(word.Navi, true)
			if err == nil && len(results[0]) > 2 {
				allNaviWords := ""
				for i, a := range results[0] {
					if i != 0 { //&& i < 3 {
						tempHoms = append(tempHoms, a.Navi+AffixCount(a))
						allNaviWords += a.Navi + AffixCount(a) + " "
					}
				}

				if _, ok := homoMap[allNaviWords]; !ok {
					homoMap[allNaviWords] = 1
					fmt.Println(strconv.Itoa(len(results[0])) + " " + allNaviWords + " " + word.Navi)
				}
			}
		}

		return nil
	})

	// Reverse the order to make accidental and new homonyms easier to see
	// Also make it a string for easier searching
	i := len(tempHoms)
	for i > 0 {
		i--
		homonymsArray[1] += tempHoms[i] + " "
	}

	homonymsArray[1] = strings.TrimSuffix(homonymsArray[1], " ")

	if err != nil {
		log.Printf("Error in homonyms stage 2: %s", err)
		return err
	}

	//fmt.Println(homonymsArray[1])

	return nil
}

// Helper for StageThree, based on reconstruct from affixes.go
func reconjugateNouns(input Word, inputNavi string, prefixCheck int, suffixCheck int, unlenite int8, affixCountdown int8) error {
	// End state: Limit to 2 affixes per noun
	if affixCountdown == 0 {
		return nil
	}
	if _, ok := candidates2Map[inputNavi]; !ok {
		candidates2 = append(candidates2, candidate{navi: inputNavi, length: uint8(len([]rune(inputNavi)))})
		candidates2Map[inputNavi] = 1
	}
	switch prefixCheck {
	case 0:
		for _, element := range stemPrefixes {
			// If it has a lenition-causing prefix
			newWord := element + inputNavi
			reconjugateNouns(input, newWord, 1, suffixCheck, 0, affixCountdown-1)
		}
		fallthrough
	case 1:
		fallthrough
	case 2:
		// Non-lenition prefixes for nouns only
		for _, element := range prefixes1Nouns {
			newWord := element + inputNavi
			reconjugateNouns(input, newWord, 4, suffixCheck, 0, affixCountdown-1)
		}
		fallthrough
	case 3:
		// This one will demand this makes it use lenition
		lenited := inputNavi
		if unlenite != 0 {
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
			reconjugateNouns(input, lenited2, 4, suffixCheck, -1, affixCountdown-1)
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
			reconjugateNouns(input, newWord, prefixCheck, 2, unlenite, affixCountdown-1)
		}
		fallthrough
	case 2:
		newWord := inputNavi + "o"
		reconjugateNouns(input, newWord, prefixCheck, 3, unlenite, affixCountdown-1)
		fallthrough
	case 3:
		newWord := inputNavi + "pe"
		reconjugateNouns(input, newWord, prefixCheck, 4, unlenite, affixCountdown-1)
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
			reconjugateNouns(input, newWord, prefixCheck, 5, unlenite, affixCountdown-1)
		}
		fallthrough
	case 5:
		newWord := inputNavi + "sì"
		reconjugateNouns(input, newWord, prefixCheck, 6, unlenite, affixCountdown-1)
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
func reconjugateVerbs(inputNavi string, prefirstUsed bool, firstUsed bool, secondUsed bool, affixLimit int8) error {
	if affixLimit == 0 {
		return nil
	}
	if _, ok := candidates2Map[inputNavi]; !ok {
		noBracket := removeBrackets(inputNavi)
		candidates2 = append(candidates2, candidate{navi: noBracket, length: uint8(len([]rune(noBracket)))})
		candidates2Map[inputNavi] = 1
	}

	if !prefirstUsed {
		reconjugateVerbs(strings.ReplaceAll(inputNavi, "<0>", ""), true, firstUsed, secondUsed, affixLimit-1)
		for _, a := range prefirst {
			reconjugateVerbs(strings.ReplaceAll(inputNavi, "<0>", a), true, firstUsed, secondUsed, affixLimit-1)
		}
		reconjugateVerbs(strings.ReplaceAll(inputNavi, "<0>", "äpeyk"), true, firstUsed, secondUsed, affixLimit-1)
	} else if !firstUsed {
		reconjugateVerbs(strings.ReplaceAll(inputNavi, "<1>", ""), prefirstUsed, true, secondUsed, affixLimit-1)
		for _, a := range first {
			reconjugateVerbs(strings.ReplaceAll(inputNavi, "<1>", a), prefirstUsed, true, secondUsed, affixLimit-1)
		}
	} else if !secondUsed {
		reconjugateVerbs(strings.ReplaceAll(inputNavi, "<2>", ""), prefirstUsed, firstUsed, true, affixLimit-1)
		for _, a := range second {
			reconjugateVerbs(strings.ReplaceAll(inputNavi, "<2>", a), prefirstUsed, firstUsed, true, affixLimit-1)
		}
	}

	return nil
}

func reconjugate(word Word, allowPrefixes bool, affixLimit int8) {
	// remove "+" and "--", we want to be able to search with and without those!
	word.Navi = strings.ReplaceAll(word.Navi, "+", "")
	word.Navi = strings.ReplaceAll(word.Navi, "--", "")
	word.Navi = strings.ToLower(word.Navi)

	if _, ok := candidates2Map[word.Navi]; !ok {
		candidates2 = append(candidates2, candidate{navi: word.Navi, length: uint8(len([]rune(word.Navi)))})
		candidates2Map[word.Navi] = 1
	}

	if word.PartOfSpeech == "pn." {
		if _, ok := candidates2Map["nì"+word.Navi]; !ok {
			new := "nì" + word.Navi
			candidates2 = append(candidates2, candidate{navi: new, length: uint8(len([]rune(new)))})
			candidates2Map[new] = 1
		}
	}

	if word.PartOfSpeech == "n." || word.PartOfSpeech == "pn." || word.PartOfSpeech == "Prop.n." || word.PartOfSpeech == "inter." {
		reconjugateNouns(word, word.Navi, 0, 0, 0, affixLimit)
		//Lenited forms, too
		found := false

		for _, a := range lenitors {
			if strings.HasPrefix(word.Navi, a) {
				word.Navi = strings.TrimPrefix(word.Navi, a)
				word.Navi = lenitionMap[a] + word.Navi
				found = true
				break
			}
		}
		if found {
			if _, ok := candidates2Map[word.Navi]; !ok {
				candidates2 = append(candidates2, candidate{navi: word.Navi, length: uint8(len([]rune(word.Navi)))})
				candidates2Map[word.Navi] = 1
			}
			reconjugateNouns(word, word.Navi, 0, 0, -1, affixLimit-1)
		}
	} else if word.PartOfSpeech[0] == 'v' {
		reconjugateVerbs(word.InfixLocations, false, false, false, affixLimit)
		//None of these can productively combine with infixes
		if allowPrefixes {
			// Gerunds
			gerund := removeBrackets("tì" + strings.ReplaceAll(word.InfixLocations, "<1>", "us"))

			if _, ok := candidates2Map[gerund]; !ok {
				candidates2 = append(candidates2, candidate{navi: gerund, length: uint8(len([]rune(gerund)))})
				candidates2Map[gerund] = 1
			}

			reconjugateNouns(word, gerund, 0, 0, 0, affixLimit-1)
			//candidates2 = append(candidates2, removeBrackets("nì"+strings.ReplaceAll(word.InfixLocations, "<1>", "awn")))
			// [verb]-able
			abilityVerbs := []string{"tsuk" + word.Navi, "suk" + word.Navi, "atsuk" + word.Navi,
				"tsuk" + word.Navi + "a", "ketsuk" + word.Navi, "hetsuk" + word.Navi, "aketsuk" + word.Navi,
				"ketsuk" + word.Navi + "a", "hetsuk" + word.Navi + "a"}
			for _, a := range abilityVerbs {
				if _, ok := candidates2Map[a]; !ok {
					candidates2 = append(candidates2, candidate{navi: a, length: uint8(len([]rune(a)))})
					candidates2Map[a] = 1
				}
			}

			//Lenited forms, too
			found := false

			for _, a := range lenitors {
				if strings.HasPrefix(gerund, a) {
					gerund = strings.TrimPrefix(gerund, a)
					gerund = lenitionMap[a] + gerund
					found = true
					break
				}
			}
			if found {
				if _, ok := candidates2Map[gerund]; !ok {
					candidates2 = append(candidates2, candidate{navi: gerund, length: uint8(len([]rune(gerund)))})
					candidates2Map[gerund] = 1
				}
				reconjugateNouns(word, gerund, 0, 0, -1, affixLimit-1)
			}
		}
		// Ability to [verb]
		tswo := word.Navi + "tswo"
		if _, ok := candidates2Map[tswo]; !ok {
			candidates2 = append(candidates2, candidate{navi: tswo, length: uint8(len([]rune(tswo)))})
			candidates2Map[tswo] = 1
		}
		reconjugateNouns(word, tswo, 0, 0, 0, affixLimit-1)

		//Lenited forms, too
		found := false

		for _, a := range lenitors {
			if strings.HasPrefix(word.Navi, a) {
				word.Navi = strings.TrimPrefix(word.Navi, a)
				word.Navi = lenitionMap[a] + word.Navi
				found = true
				break
			}
		}
		if found {
			tswo = word.Navi + "tswo"
			if _, ok := candidates2Map[tswo]; !ok {
				candidates2 = append(candidates2, candidate{navi: tswo, length: uint8(len([]rune(tswo)))})
				candidates2Map[tswo] = 1
			}
			reconjugateNouns(word, tswo, 0, 0, -1, affixLimit-2)
		}

	} else if word.PartOfSpeech == "adj." {
		adjA := word.Navi + "a"
		if _, ok := candidates2Map[adjA]; !strings.HasSuffix(word.Navi, "a") && !ok {
			candidates2 = append(candidates2, candidate{navi: adjA, length: uint8(len([]rune(adjA)))})
			candidates2Map[adjA] = 1
		}

		if allowPrefixes {
			aAdj := "a" + word.Navi
			if _, ok := candidates2Map[aAdj]; !strings.HasPrefix(word.Navi, "a") && !ok {
				candidates2 = append(candidates2, candidate{navi: aAdj, length: uint8(len([]rune(aAdj)))})
				candidates2Map[aAdj] = 1
			}
			nìAdj := "nì" + word.Navi
			if _, ok := candidates2Map[nìAdj]; !ok {
				candidates2 = append(candidates2, candidate{navi: nìAdj, length: uint8(len([]rune(nìAdj)))})
				candidates2Map[nìAdj] = 1
			}
		}

		//Lenited forms, too
		for _, a := range lenitors {
			if strings.HasPrefix(word.Navi, a) {
				word.Navi = strings.TrimPrefix(word.Navi, a)
				word.Navi = lenitionMap[a] + word.Navi
				candidates2 = append(candidates2, candidate{navi: word.Navi, length: uint8(len([]rune(word.Navi)))})
				candidates2Map[word.Navi] = 1
				if !strings.HasSuffix(word.Navi, "a") {
					lenitA := word.Navi + "a"
					candidates2 = append(candidates2, candidate{navi: lenitA, length: uint8(len([]rune(lenitA)))})
					candidates2Map[lenitA] = 1
				}
				break
			}
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

func CheckHomsAsync(file *os.File, candidates []candidate, tempHoms *[]string, word Word, minAffix int, wg *sync.WaitGroup) {
	defer wg.Done()

	for _, a := range candidates {

		//Nasal assimilation stuff
		/*containsNasal := false

		for _, t := range []string{"t", "k", "p", "tx", "kx", "px"} {
			for _, n := range []string{"n", "ng", "m"} {
				if strings.Contains(a.navi, n+t) || strings.Contains(a.navi, t+n) {
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
		results, err := TranslateFromNaviHash(a.navi, true)

		if err == nil && len(results) > 0 && len(results[0]) > 2 {
			allNaviWords := ""
			allLengths := []int{}
			noDupes := []string{}
			atLeast3 := false
			for i, b := range results[0] {
				if i == 0 {
					continue
				}
				dupe := false
				for _, c := range noDupes {
					if c == b.Navi+AffixCount(b) {
						dupe = true
						break
					}
				}
				if !dupe { //&& i < 3 {
					noDupes = AppendStringAlphabetically(noDupes, b.Navi+AffixCount(b))
					lengths := len(b.Affixes.Prefix) + len(b.Affixes.Suffix) + len(b.Affixes.Infix)
					allLengths = append(allLengths, lengths)
					if lengths >= minAffix {
						atLeast3 = true
					}
				}
			}

			for _, b := range noDupes {
				allNaviWords += b + " "
			}

			// No duplicates
			if len(noDupes) > 1 {
				allLengthsString := ""

				if _, ok := homoMap[allNaviWords]; !ok {
					for _, a := range allLengths {
						allLengthsString += strconv.Itoa(a) + " "
					}
					allLengthsString = strings.TrimSuffix(allLengthsString, " ")
					homoMap[allNaviWords] = 1
					if atLeast3 {
						stringy := word.PartOfSpeech + ": -" + a.navi + " " + word.Navi + "- -" + allNaviWords + " " + allLengthsString
						fmt.Println(stringy)
						_, err := file.WriteString(stringy + "\n")
						if err != nil {
							fmt.Println("Error writing to file:", err)
							return
						}
					}
					*tempHoms = append(*tempHoms, a.navi)
				}
			}
		}
		/*if len(strings.Split(a, " ")) > 1 {
			fmt.Println("oops " + a)
			continue
		}*/
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
		}
	}
}

func StageThree(minAffix int, affixLimit int8, startNumber int) (err error) {
	start := time.Now()

	file, err := os.Create("results.txt")
	if err != nil {
		fmt.Println("error opening file:", err)
		return
	}

	defer file.Close()

	tempHoms := []string{}

	wordCount := 0

	file.WriteString("Stage 3\n")

	err = RunOnDict(func(word Word) error {
		wordCount += 1
		//checkAsyncLock.Wait()

		if wordCount >= startNumber {
			// Progress counter
			if wordCount%100 == 0 {
				total_seconds := time.Since(start)

				printMessage := "On word " + strconv.Itoa(wordCount) + ".  Time elapsed is " +
					strconv.Itoa(int(math.Floor(total_seconds.Hours()))) + " hours, " +
					strconv.Itoa(int(math.Floor(total_seconds.Minutes()))%60) + " minutes and " +
					strconv.Itoa(int(total_seconds.Seconds())%60) + " seconds.  " + strconv.Itoa(len(candidates2Map)) + " conjugations checked"

				log.Printf(printMessage)
				file.WriteString(printMessage + "\n")
			}
			// save original Navi word, we want to add "+" or "--" later again
			//naviWord := word.Navi

			// No multiword words
			if !strings.Contains(word.Navi, " ") {
				candidates2 = []candidate{{navi: word.Navi, length: uint8(len([]rune(word.Navi)))}} //empty array of strings

				// Get conjugations
				reconjugate(word, true, affixLimit)

				//Lenited forms, too
				found := false
				for _, a := range lenitors {
					if strings.HasPrefix(word.Navi, a) {
						word.Navi = strings.TrimPrefix(word.Navi, a)
						word.Navi = lenitionMap[a] + word.Navi
						found = true
						break
					}
				}
				if found {
					reconjugate(word, false, affixLimit)
				}

				sort.SliceStable(candidates2, func(i, j int) bool {
					return candidates2[i].length < candidates2[j].length
				})
				checkAsyncLock.Wait()
				checkAsyncLock.Add(1)
				go CheckHomsAsync(file, candidates2, &tempHoms, word, minAffix, &checkAsyncLock)
			} else if strings.HasSuffix(word.Navi, " si") {
				// "[word] si" can take the form "[word]tswo"
				siTswo := strings.TrimSuffix(word.Navi, " si")
				siTswo = siTswo + "tswo"
				reconjugateNouns(word, siTswo, 0, 0, 0, affixLimit)
				//Lenited forms, too
				found := false
				for _, a := range lenitors {
					if strings.HasPrefix(siTswo, a) {
						siTswo = strings.TrimPrefix(siTswo, a)
						siTswo = lenitionMap[a] + siTswo
						found = true
						break
					}
				}
				if found {
					if _, ok := candidates2Map[siTswo]; !ok {
						candidates2 = append(candidates2, candidate{navi: siTswo, length: uint8(len([]rune(siTswo)))})
						candidates2Map[siTswo] = 1
					}
					reconjugateNouns(word, siTswo, 0, 0, -1, affixLimit)
				}

				sort.SliceStable(candidates2, func(i, j int) bool {
					return candidates2[i].length < candidates2[j].length
				})
				checkAsyncLock.Wait()
				checkAsyncLock.Add(1)
				go CheckHomsAsync(file, candidates2, &tempHoms, word, minAffix, &checkAsyncLock)
			}
		}

		return nil
	})

	//fmt.Println(homoMap)
	//fmt.Println(tempHoms)

	total_seconds := time.Since(start)

	finalString := "Stage three took " + strconv.Itoa(int(math.Floor(total_seconds.Hours()))) + " hours, " +
		strconv.Itoa(int(math.Floor(total_seconds.Minutes()))%60) + " minutes and " +
		strconv.Itoa(int(total_seconds.Seconds())%60) + " seconds"
	log.Printf(finalString)
	file.WriteString(finalString + "\n")

	checkedString := "Checked " + strconv.Itoa(len(candidates2Map)) + " total conjugations"
	fmt.Println(checkedString)
	file.WriteString(checkedString + "\n")

	return
}

// Do everything
func homonymSearch() {
	fmt.Println("Stage 1:")
	StageOne()
	fmt.Println("Stage 2:")
	StageTwo()
	fmt.Println("Stage 3:")
	// minimum affixes, maximum affixes, start at word number N
	StageThree(0, 5, 0)

	fmt.Println(longest)
	file.WriteString(strconv.Itoa(longest) + "\n")
	fmt.Println(top10Longest[longest])
	file.WriteString(top10Longest[longest] + "\n")
}
