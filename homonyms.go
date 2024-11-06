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
var candidates2 []string
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
var top10Longest = map[int]string{}
var longest = 0

var checkAsyncLock = sync.WaitGroup{}

type tempHomsContainer struct {
	mu       sync.Mutex
	counters []string
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
		homonymsArray[0] += tempHoms[i] + " "
	}

	homonymsArray[0] = strings.TrimSuffix(homonymsArray[0], " ")

	if err != nil {
		log.Printf("Error in homonyms stage 1: %s", err)
		return err
	}

	return nil
}

// Helper to turn a string into a list of known words

// Check for ones that are the exact same, no affixes needed
func StageTwo() error {
	tempHoms := []string{}

	err := runOnFile(func(word Word) error {
		standardizedWord := word.Navi

		homList := []string{}
		// If the word can conjugate into something else, record it
		results, err := TranslateFromNaviHash(standardizedWord, true)
		if err == nil && len(results[0]) > 2 {
			allNaviWords := ""
			for i, a := range results[0] {
				if i != 0 { //&& i < 3 {
					tempHoms = append(tempHoms, a.Navi)
					homList = AppendStringAlphabetically(homList, a.Navi)
				}
			}

			for _, a := range homList {
				allNaviWords += a + " "
			}

			homoMap[allNaviWords] = 1
			fmt.Println(strconv.Itoa(len(results[0])) + " " + allNaviWords + " " + standardizedWord)
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
						tempHoms = append(tempHoms, a.Navi)
						allNaviWords += a.Navi + " "
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
func reconjugateNouns(file *os.File, input Word, inputNavi string, prefixCheck int, suffixCheck int, unlenite int, affixCountdown int8) error {
	// End state: Limit to 2 affixes per noun
	if affixCountdown == 0 {
		return nil
	}
	if _, ok := candidates2Map[inputNavi]; !ok {
		fileAppend(file, inputNavi, 100-((3-int(affixCountdown))*20))
		candidates2 = append(candidates2, inputNavi)
	}
	switch prefixCheck {
	case 0:
		for _, element := range stemPrefixes {
			// If it has a lenition-causing prefix
			newWord := element + inputNavi
			reconjugateNouns(file, input, newWord, 1, suffixCheck, 0, affixCountdown-1)
		}
		fallthrough
	case 1:
		fallthrough
	case 2:
		// Non-lenition prefixes for nouns only
		for _, element := range prefixes1Nouns {
			newWord := element + inputNavi
			reconjugateNouns(file, input, newWord, 4, suffixCheck, 0, affixCountdown-1)
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
			reconjugateNouns(file, input, lenited2, 4, suffixCheck, -1, affixCountdown-1)
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
			reconjugateNouns(file, input, newWord, prefixCheck, 2, unlenite, affixCountdown-1)
		}
		fallthrough
	case 2:
		newWord := inputNavi + "o"
		reconjugateNouns(file, input, newWord, prefixCheck, 3, unlenite, affixCountdown-1)
		fallthrough
	case 3:
		newWord := inputNavi + "pe"
		reconjugateNouns(file, input, newWord, prefixCheck, 4, unlenite, affixCountdown-1)
		fallthrough
	case 4:
		vowel := false
		diphthong := false
		consonant := true
		naviRunes := []rune(inputNavi)
		lastRune := naviRunes[len(naviRunes)-1]
		if is_vowel(string(lastRune)) {
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
			reconjugateNouns(file, input, newWord, prefixCheck, 5, unlenite, affixCountdown-1)
		}
		fallthrough
	case 5:
		newWord := inputNavi + "sì"
		reconjugateNouns(file, input, newWord, prefixCheck, 6, unlenite, affixCountdown-1)
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
func reconjugateVerbs(file *os.File, inputNavi string, prefirstUsed bool, firstUsed bool, secondUsed bool, affixLimit int8) error {
	if affixLimit == 0 {
		return nil
	}
	if _, ok := candidates2Map[inputNavi]; !ok {
		fileAppend(file, removeBrackets(inputNavi), 100-((3-int(affixLimit))*20))
		candidates2 = append(candidates2, removeBrackets(inputNavi))
	}

	if !prefirstUsed {
		reconjugateVerbs(file, strings.ReplaceAll(inputNavi, "<0>", ""), true, firstUsed, secondUsed, affixLimit-1)
		for _, a := range prefirst {
			reconjugateVerbs(file, strings.ReplaceAll(inputNavi, "<0>", a), true, firstUsed, secondUsed, affixLimit-1)
		}
		reconjugateVerbs(file, strings.ReplaceAll(inputNavi, "<0>", "äpeyk"), true, firstUsed, secondUsed, affixLimit-1)
	} else if !firstUsed {
		reconjugateVerbs(file, strings.ReplaceAll(inputNavi, "<0>", ""), prefirstUsed, true, secondUsed, affixLimit-1)
		for _, a := range first {
			reconjugateVerbs(file, strings.ReplaceAll(inputNavi, "<1>", a), prefirstUsed, true, secondUsed, affixLimit-1)
		}
	} else if !secondUsed {
		reconjugateVerbs(file, strings.ReplaceAll(inputNavi, "<2>", ""), prefirstUsed, firstUsed, true, affixLimit-1)
		for _, a := range second {
			reconjugateVerbs(file, strings.ReplaceAll(inputNavi, "<2>", a), prefirstUsed, firstUsed, true, affixLimit-1)
		}
	}

	return nil
}

func fileAppend(file *os.File, word string, weight int) {
	// No duplicates
	if _, ok := candidates2Map[strings.ToLower(word)]; ok {
		return
	}
	// Write a string to the file
	_, err := file.WriteString(word + "\t" + strconv.Itoa(weight) + "\n")
	if err != nil {
		fmt.Println("Error writing to file (0 affixes):", err)
		return
	}
	candidates2Map[strings.ToLower(word)] = 1
}

func reconjugate(file *os.File, word Word, allowPrefixes bool, affixLimit int8) {
	// remove "+" and "--", we want to be able to search with and without those!
	word.Navi = strings.ReplaceAll(word.Navi, "+", "")
	word.Navi = strings.ReplaceAll(word.Navi, "--", "")
	word.Navi = strings.ToLower(word.Navi)

	if _, ok := candidates2Map[word.Navi]; !ok {
		fileAppend(file, removeBrackets(word.Navi), 100-((3-int(affixLimit))*20))
		candidates2 = append(candidates2, removeBrackets(word.Navi))
	}

	if word.PartOfSpeech == "pn." {
		if _, ok := candidates2Map["nì"+word.Navi]; !ok {
			candidates2 = append(candidates2, "nì"+word.Navi)
			fileAppend(file, "nì"+word.Navi, 100-((3-int(affixLimit))*20))
		}
	}

	if word.PartOfSpeech == "n." || word.PartOfSpeech == "pn." || word.PartOfSpeech == "Prop.n." || word.PartOfSpeech == "inter." {
		reconjugateNouns(file, word, word.Navi, 0, 0, 0, affixLimit)
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
			candidates2 = append(candidates2, word.Navi)
			reconjugateNouns(file, word, word.Navi, 0, 0, -1, affixLimit-1)
		}
	} else if word.PartOfSpeech[0] == 'v' {
		reconjugateVerbs(file, word.InfixLocations, false, false, false, affixLimit)
		//None of these can productively combine with infixes
		if allowPrefixes {
			// Gerunds
			gerund := removeBrackets("tì" + strings.ReplaceAll(word.InfixLocations, "<1>", "us"))
			candidates2 = append(candidates2, gerund)
			reconjugateNouns(file, word, gerund, 0, 0, 0, affixLimit-1)
			//candidates2 = append(candidates2, removeBrackets("nì"+strings.ReplaceAll(word.InfixLocations, "<1>", "awn")))
			// [verb]-abl
			candidates2 = append(candidates2, "tsuk"+word.Navi)
			fileAppend(file, "tsuk"+word.Navi, 80)
			candidates2 = append(candidates2, "suk"+word.Navi)
			fileAppend(file, "suk"+word.Navi, 60)
			candidates2 = append(candidates2, "atsuk"+word.Navi)
			fileAppend(file, "atsuk"+word.Navi, 60)
			candidates2 = append(candidates2, "tsuk"+word.Navi+"a")
			fileAppend(file, "tsuk"+word.Navi+"a", 60)
			candidates2 = append(candidates2, "ketsuk"+word.Navi)
			fileAppend(file, "ketsuk"+word.Navi, 80)
			candidates2 = append(candidates2, "hetsuk"+word.Navi)
			fileAppend(file, "hetsuk"+word.Navi, 60)
			candidates2 = append(candidates2, "aketsuk"+word.Navi)
			fileAppend(file, "aketsuk"+word.Navi, 60)
			candidates2 = append(candidates2, "ketsuk"+word.Navi+"a")
			fileAppend(file, "ketsuk"+word.Navi+"a", 60)
			candidates2 = append(candidates2, "hetsuk"+word.Navi+"a")
			fileAppend(file, "hetsuk"+word.Navi+"a", 60)

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
				candidates2 = append(candidates2, gerund)
				reconjugateNouns(file, word, gerund, 0, 0, -1, affixLimit-1)
			}
		}
		// Ability to [verb]
		candidates2 = append(candidates2, word.Navi+"tswo")
		reconjugateNouns(file, word, word.Navi+"tswo", 0, 0, 0, affixLimit-1)
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
			candidates2 = append(candidates2, word.Navi+"tswo")
			reconjugateNouns(file, word, word.Navi+"tswo", 0, 0, -1, affixLimit-2)
		}

	} else if word.PartOfSpeech == "adj." {
		if !strings.HasSuffix(word.Navi, "a") {
			candidates2 = append(candidates2, word.Navi+"a")
			// Write a string to the file
			fileAppend(file, word.Navi+"a", 80)
		}

		if allowPrefixes {
			if !strings.HasPrefix(word.Navi, "a") {
				candidates2 = append(candidates2, "a"+word.Navi)
				fileAppend(file, "a"+word.Navi, 80)
			}
			candidates2 = append(candidates2, "nì"+word.Navi)
			fileAppend(file, "nì"+word.Navi, 80)
		}

		//Lenited forms, too
		for _, a := range lenitors {
			if strings.HasPrefix(word.Navi, a) {
				word.Navi = strings.TrimPrefix(word.Navi, a)
				word.Navi = lenitionMap[a] + word.Navi
				fileAppend(file, word.Navi, 80)
				candidates2 = append(candidates2, word.Navi)
				if !strings.HasSuffix(word.Navi, "a") {
					fileAppend(file, word.Navi+"a", 60)
					candidates2 = append(candidates2, word.Navi+"a")
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

func CheckHomsAsync(candidates []string, tempHoms *[]string, word Word, minAffix int, wg *sync.WaitGroup) {
	defer wg.Done()

	sort.Slice(candidates, func(i, j int) bool {
		return len([]rune(candidates[i])) < len([]rune(candidates[j]))
	})

	for _, a := range candidates {
		results, err := TranslateFromNaviHash(a, true)
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
					if c == b.Navi {
						dupe = true
						break
					}
				}
				if !dupe { //&& i < 3 {
					noDupes = AppendStringAlphabetically(noDupes, b.Navi)
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
						fmt.Println(word.PartOfSpeech + ": -" + a + " " + word.Navi + "- -" + allNaviWords + " " + allLengthsString)
					}
					*tempHoms = append(*tempHoms, a)
				}
			}
		}
		/*if len(strings.Split(a, " ")) > 1 {
			fmt.Println("oops " + a)
			continue
		}*/
		runes := []rune(a)
		runeCount := len(runes)
		if runeCount < 40 {
			continue
		}
		if runeCount > longest {
			longest = runeCount
			//fmt.Println(a)
		}
		if _, ok := top10Longest[runeCount]; ok {
			top10Longest[runeCount] = top10Longest[runeCount] + " " + a
		} else {
			top10Longest[runeCount] = a
		}
	}
}

func StageThree(minAffix int, affixLimit int8, startNumber int) (err error) {
	start := time.Now()

	//tempHoms := []string{}

	wordCount := 0

	file, err := os.Create("conjugations.txt")
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer file.Close()

	err = RunOnDict(func(word Word) error {
		wordCount += 1
		//checkAsyncLock.Wait()

		word.Navi = ReefMe(word.IPA, false)[0]
		word.Navi = strings.ReplaceAll(word.Navi, "_", "")
		word.Navi = strings.ReplaceAll(word.Navi, "-", "")

		if wordCount >= startNumber {
			// Progress counter
			if wordCount%100 == 0 {
				total_seconds := time.Since(start)

				log.Printf("On word " + strconv.Itoa(wordCount) + ".  Time elapsed is " +
					strconv.Itoa(int(math.Floor(total_seconds.Hours()))) + " hours, " +
					strconv.Itoa(int(math.Floor(total_seconds.Minutes()))%60) + " minutes and " +
					strconv.Itoa(int(total_seconds.Seconds())%60) + " seconds")
			}
			// save original Navi word, we want to add "+" or "--" later again
			//naviWord := word.Navi
			word.Navi = strings.Trim(word.Navi, " ")

			fileAppend(file, word.Navi, 100)

			// No multiword words
			if !strings.Contains(word.Navi, " ") {
				candidates2 = []string{word.Navi} //empty array of strings

				// Get conjugations
				reconjugate(file, word, true, affixLimit)

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
					candidates2 = append(candidates2, word.Navi)
					reconjugate(file, word, false, affixLimit-1)
				}
				//checkAsyncLock.Wait()
				//checkAsyncLock.Add(1)
				//go CheckHomsAsync(candidates2, &tempHoms, word, minAffix, &checkAsyncLock)
			} else if strings.HasSuffix(word.Navi, " si") {
				// "[word] si" can take the form "[word]tswo"
				siTswo := strings.TrimSuffix(word.Navi, " si")
				siTswo = siTswo + "tswo"
				reconjugateNouns(file, word, siTswo, 0, 0, 0, affixLimit-1)
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
					candidates2 = append(candidates2, siTswo)
					reconjugateNouns(file, word, siTswo, 0, 0, -1, affixLimit-2)
				}
				//checkAsyncLock.Wait()
				//checkAsyncLock.Add(1)
				//go CheckHomsAsync(candidates2, &tempHoms, word, minAffix, &checkAsyncLock)
			}
		}

		return nil
	})

	//fmt.Println(homoMap)
	//fmt.Println(tempHoms)

	total_seconds := time.Since(start)

	log.Printf("Stage three took " + strconv.Itoa(int(math.Floor(total_seconds.Hours()))) + " hours, " +
		strconv.Itoa(int(math.Floor(total_seconds.Minutes()))%60) + " minutes and " +
		strconv.Itoa(int(total_seconds.Seconds())%60) + " seconds")

	return
}

// Do everything
func homonymSearch() {
	fmt.Println("Stage 1:")
	//StageOne()
	fmt.Println("Stage 2:")
	//StageTwo()
	fmt.Println("Stage 3:")
	// minimum affixes, maximum affixes, start at word number N
	StageThree(0, 3, 0)
	fmt.Println("Checked " + strconv.Itoa(len(candidates2Map)) + " total conjugations")
	fmt.Println(longest)
	fmt.Println(top10Longest[longest])
}
