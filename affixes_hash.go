package main

import (
	"slices"
	"strings"
)

type PreferenceOfSpeech uint8

const (
	ANY PreferenceOfSpeech = iota
	NOUN
	VERB
	ADJ
	NÌ
	PN
)

type ConjugationCandidate struct {
	word      string
	lenition  []string
	prefixes  []string
	suffixes  []string
	infixes   []string
	insistPOS PreferenceOfSpeech
}

func candidateDupe(candidate ConjugationCandidate) (c ConjugationCandidate) {
	a := ConjugationCandidate{}
	a.word = candidate.word
	a.lenition = candidate.lenition
	a.prefixes = candidate.prefixes
	a.infixes = candidate.infixes
	a.suffixes = candidate.suffixes
	a.insistPOS = candidate.insistPOS
	return a
}

var unlenitionLetters = []string{
	"ts", "kx", "tx", "px", // traps digraphs because they cannot unlenite
	"f", "p", "h", "k", "s",
	"t", "a", "ä", "e", "i",
	"ì", "o", "u", "ù",
}

// "ts" is there to prevent "ts" from becoming "txs"
var unlenition = map[string][]string{
	// digraphs cannot unlenite
	"ts": {}, // here to trap the "ts" ahead of the "t"
	"px": {}, // here to trap the "px" ahead of the "p"
	"kx": {}, // here to trap the "kx" ahead of the "k"
	"tx": {}, // here to trap the "tx" ahead of the "t"
	"f":  {"f", "p"},
	"p":  {"px"},
	"h":  {"h", "k"},
	"k":  {"kx"},
	"s":  {"s", "t", "ts"},
	"t":  {"tx"},
	"a":  {"a", "'a"},
	"ä":  {"ä", "'ä"},
	"e":  {"e", "'e"},
	"i":  {"i", "'i"},
	"ì":  {"ì", "'ì"},
	"o":  {"o", "'o"},
	"u":  {"u", "'u"},
	"ù":  {"ù", "'ù"},
}

var prefixes1Nouns = []string{"fì", "tsa"}
var prefixes1NounsLenition = []string{"pay", "fay"}
var prefixes1lenition = []string{"ay", "me", "pxe"}
var stemPrefixes = []string{"fne", "sna", "munsna"}
var verbPrefixes = []string{"tsuk", "ketsuk"}
var caseEndings = map[string]bool{
	"ìl":  true,
	"l":   true,
	"it":  true,
	"ti":  true,
	"t":   true,
	"ur":  true,
	"ru":  true,
	"r":   true,
	"yä":  true,
	"ä":   true,
	"ìri": true,
	"ri":  true,
	"ye":  true,
	"e":   true,
	"il":  true,
	"iri": true,
}

var adposuffixes = []string{
	// adpositions that can be mistaken for case endings
	"pxel",                                                     //"agentive"
	"mungwrr",                                                  //"dative"
	"kxamlä", "ìlä", "wä", "nuä", "kxamle", "ìle", "we", "nue", //"genitive"
	"teri", //"topical"
	// Case endings
	"ìl", "l", "it", "ti", "t", "ur", "ru", "r", "yä", "ä", "e", "ye", "ìri", "ri",
	// Sorted alphabetically by their reverse forms
	"ftumfa", "nemfa", "rofa", "ka", "fa", "na", "ta", "ya", "yoa", "krrka", "ftuopa", //-a
	"lisre", "pxisre", "sre", "luke", "ne", //-e
	"fpi",          //-i
	"mì",           //-ì
	"lok",          //-k
	"mìkam", "kam", //-m
	"ken", "sìn", "talun", //-n
	"äo", "eo", "io", "uo", "ro", "to", "sko", //-o
	"tafkip", "takip", "fkip", "kip", //-p
	"ftu", "hu", //-u
	"pximaw", "maw", "pxaw", "few", "raw", //-w
	"vay", "kay", //-y
}

var lenitionAdposition = map[string]string{
	"pel":   "pxel",
	"kamlä": "kxamlä",
	"kamle": "kxamle",
	"pisre": "pxisre",
	"pimaw": "pximaw",
}

var vowelSuffixes = map[string][]string{
	"äo":  {"ä", "e"},
	"eo":  {"e"},
	"io":  {"i"},
	"uo":  {"u"},
	"ìlä": {"ì"},
	"o":   {"o"},
}
var stemSuffixes = []string{"tsyìp", "fkeyk"}
var verbSuffixes = []string{"tswo", "yu", "tseng"}

var infixes = map[rune][]string{
	rune('a'): {"ay", "asy", "aly", "ary", "am", "alm", "arm", "ats", "awn"},
	rune('ä'): {"äng", "äpeyk", "äp"},
	rune('e'): {"epeyk", "ep", "er", "ei", "eiy", "eng", "eyk"},
	rune('i'): {"iv", "ilv", "irv", "imv", "iyev"},
	rune('ì'): {"ìy", "ìsy", "ìly", "ìry", "ìm", "ìlm", "ìrm", "ìyev"},
	rune('o'): {"ol"},
	rune('u'): {"us", "uy"},
}

var prefirst = []string{"äp", "äpeyk", "ep", "epeyk", "eyk"}
var first = []string{"ay", "asy", "aly", "ary", "ìy", "ìsy", "ìly", "ìry", "ol", "er", "ìm",
	"ìlm", "ìrm", "am", "alm", "arm", "ìyev", "iyev", "iv", "ilv", "irv", "imv", "us", "awn"}
var second = []string{"ei", "eiy", "äng", "eng", "uy", "ats"}

var prefirstMap = map[string]bool{"äp": true, "äpeyk": true, "ep": true, "epeyk": true, "eyk": true}
var firstMap = map[string]bool{"ay": true, "asy": true, "aly": true, "ary": true, "ìy": true, "ìsy": true,
	"ìly": true, "ìry": true, "ol": true, "er": true, "ìm": true, "ìlm": true,
	"ìrm": true, "am": true, "alm": true, "arm": true, "ìyev": true, "iyev": true,
	"iv": true, "ilv": true, "irv": true, "imv": true, "us": true, "awn": true}
var secondMap = map[string]bool{"ei": true, "eiy": true, "äng": true, "eng": true, "uy": true, "ats": true}

var unreefFixes = map[string]string{
	"eng":    "äng",
	"ep":     "äp",
	"ye":     "yä",
	"e":      "ä",
	"we":     "wä",
	"ìle":    "ìlä",
	"nue":    "nuä",
	"kxamle": "kxamlä",
}

var weirdNounSuffixes = map[string]string{
	// For "tsa" with case endings
	// Canonized in:
	// https://naviteri.org/2011/08/new-vocabulary-clothing/comment-page-1/#comment-912
	"tsa":   "tsaw",
	"teyng": "tì'eyng",
	// The a re-appears when case endings are added (it uses a instead of ì)
	"oenga": "oeng",
	// Foreign nouns
	"'ìnglìs":      "'ìnglìsì",
	"keln":         "kelnì",
	"kerìsmìs":     "kerìsmìsì",
	"kìreys":       "kìreysì", // https://naviteri.org/2011/09/miscellaneous-vocabulary/
	"tsìräf":       "tsìräfì",
	"nìyu york":    "nìyu yorkì",
	"nu york":      "nu yorkì", // https://naviteri.org/2013/01/awvea-posti-zisita-amip-first-post-of-the-new-year/
	"päts":         "pätsì",
	"post":         "postì",
	"losäntsyeles": "losäntsyelesì",
	"york":         "yorkì", // For a program called Litxap
}

func isDuplicate(candidateMap *map[string]ConjugationCandidate, input ConjugationCandidate) bool {
	map2 := (*candidateMap)
	if a, ok := map2[input.word]; ok {
		if input.insistPOS == a.insistPOS {
			if len(input.prefixes) == len(a.prefixes) && len(input.suffixes) == len(a.suffixes) {
				if len(input.infixes) == len(a.infixes) {
					return true
				}
			}
		}
	}
	return false
}

func isDuplicateFix(fixes []string, fix string) (newFixes []string) {
	if newfix, ok := unreefFixes[fix]; ok {
		fix = newfix
	}
	if slices.Contains(fixes, fix) {
		return fixes
	}
	fixes = append(fixes, fix)
	return fixes
}

// fuction to check given string is in array or not
// modified from https://www.golinuxcloud.com/golang-array-contains/
func implContainsAny(sl []string, names []string) bool {
	// iterate over the array and compare given string to each element
	for _, value := range sl {
		if slices.Contains(names, value) {
			return true
		}
	}
	return false
}

// Helper for infix detection
func verifyInfix(existing []string, new string) (bool, []string) {
	if _, ok := prefirstMap[new]; ok {
		if existing[0] == "" {
			return true, []string{new, existing[1], existing[2]}
		}
	} else if _, ok := firstMap[new]; ok {
		if existing[1] == "" {
			return true, []string{existing[0], new, existing[2]}
		}
	} else if _, ok := secondMap[new]; ok {
		if existing[2] == "" {
			return true, []string{existing[0], existing[1], new}
		}
	}

	return false, existing
}

func verifyCaseEnding(noun string, ending string) bool {
	// error prevention
	if len(noun) == 0 {
		return false
	}

	if get_last_rune(noun, 1) == 'i' && (ending == "ä" || ending == "e") {
		//soaiä, tìftiä
		return true
	}
	// Don't check adpositions
	if _, ok := caseEndings[ending]; !ok {
		return true
	}
	// Non-standard conjugations
	if noun == "omatikaya" && ending == "ä" {
		return true
	}
	diphthongs := map[string]bool{
		"ay": true,
		"aw": true,
		"ey": true,
		"ew": true,
	}
	vowels := map[string]bool{
		"a": true,
		"ä": true,
		"e": true,
		"i": true,
		"ì": true,
		"o": true,
		"u": true,
		"ù": true,
	}
	nounEnding := ""
	if len(noun) >= 2 {
		nounEnding = noun[len(noun)-2:]
	}
	if _, ok := diphthongs[nounEnding]; ok {
		nounEnding := noun[len(noun)-2:]
		//ewur isn't valid
		if nounEnding == "ew" && ending == "ur" {
			return false
		}
		// Diphthong
		diphthongEndings := map[string]bool{
			"ìl": true,
			"ti": true,
			"it": true,
			"ru": true,
			"ur": true,
			"ä":  true,
			"e":  true,
			"ri": true,
		}
		if _, ok := diphthongEndings[ending]; ok {
			return true
		} else {
			lastRune := get_last_rune(noun, 1)
			switch lastRune {
			case 'y':
				// ayt, eyt
				if ending == "t" {
					return true
				}
			case 'w':
				// ewr, awr
				if ending == "r" {
					return true
				}
			}
		}
	} else if _, ok := vowels[string(get_last_rune(noun, 1))]; ok {
		lastVowel := get_last_rune(noun, 1)
		if lastVowel == 'u' || lastVowel == 'o' {
			// No oyä or ayä
			switch ending {
			case "yä", "ye":
				return false
			case "ä", "e":
				return true
			}
		}
		vowelEndings := map[string]bool{
			"l":  true,
			"t":  true,
			"ti": true,
			"ru": true,
			"r":  true,
			"yä": true,
			"ye": true,
			"ri": true,
		}
		if _, ok := vowelEndings[ending]; ok {
			return true
		}
	} else {
		// Consonant or psuedovowel
		otherEndings := map[string]bool{
			"ìl":  true,
			"ti":  true,
			"it":  true,
			"ur":  true,
			"ä":   true,
			"e":   true,
			"ìri": true,
		}
		if _, ok := otherEndings[ending]; ok {
			return true
		}
		//'ri
		if get_last_rune(noun, 1) == '\'' && (ending == "ru" || ending == "ri") {
			return true
		}
	}
	return false
}

func deconjugateHelper(input ConjugationCandidate, dupes *map[string]ConjugationCandidate, candidates *[]ConjugationCandidate, prefixCheck int, suffixCheck int, unlenite int8, checkInfixes []string, lastPrefix string, lastSuffix string) {
	candidateMap := *dupes

	if isDuplicate(dupes, input) {
		return
	}

	vowels := "aäeiìouù"

	// For double letter homonyms like ayyoka or ayokka
	suffixRunes := []rune(lastSuffix)
	if len(suffixRunes) > 1 && !is_vowel(suffixRunes[0]) {
		newCandidate := candidateDupe(input)
		newCandidate.word = newCandidate.word + string(suffixRunes[0])
		deconjugateHelper(newCandidate, dupes, candidates, prefixCheck, suffixCheck, unlenite, checkInfixes, "", "")

		// txeppxel is pronounced txepel, so account for that
		if slices.Contains([]rune{'p', 't', 'k'}, suffixRunes[0]) {
			shortUnlenition := map[rune]string{
				'p': "px",
				't': "tx",
				'k': "kx",
			}
			newCandidate := candidateDupe(input)
			newCandidate.word = newCandidate.word + string(shortUnlenition[suffixRunes[0]])
			deconjugateHelper(newCandidate, dupes, candidates, prefixCheck, suffixCheck, unlenite, checkInfixes, "", "")
		}

		// Ejectives before nasals are softened, so tokxmì is pronounced tokmì
		if !is_vowel(suffixRunes[0]) {
			unvoicedPlosives := []string{"p", "t", "k"}
			for _, plosive := range unvoicedPlosives {
				if strings.HasSuffix(input.word, plosive) {
					newCandidate := candidateDupe(input)
					newCandidate.word = newCandidate.word + "x"
					deconjugateHelper(newCandidate, dupes, candidates, prefixCheck, suffixCheck, unlenite, checkInfixes, "", "")
				}
			}
		}
	}
	prefixRunes := []rune(lastPrefix)
	if len(prefixRunes) > 1 && !is_vowel(prefixRunes[len(prefixRunes)-1]) {
		newCandidate := candidateDupe(input)
		newCandidate.word = string(prefixRunes[len(prefixRunes)-1]) + newCandidate.word
		deconjugateHelper(newCandidate, dupes, candidates, prefixCheck, suffixCheck, unlenite, checkInfixes, "", "")

		// tsukkxìm
		if slices.Contains([]rune{'p', 't', 'k'}, prefixRunes[len(prefixRunes)-1]) {
			shortUnlenition := map[rune]string{
				'p': "px",
				't': "tx",
				'k': "kx",
			}
			newCandidate := candidateDupe(input)
			newCandidate.word = string(shortUnlenition[prefixRunes[len(prefixRunes)-1]]) + newCandidate.word
			deconjugateHelper(newCandidate, dupes, candidates, prefixCheck, suffixCheck, unlenite, checkInfixes, "", "")
		}
	}

	// fneu checking for fne-'u
	if len(lastPrefix) > 0 && len(input.word) > 0 && hasAt(vowels, lastPrefix, -1) && hasAt(vowels, input.word, 0) {
		if !implContainsAny(prefixes1lenition, []string{lastPrefix}) { // do not do this for leniting prefixes
			newCandidate := candidateDupe(input)
			newCandidate.word = "'" + newCandidate.word
			deconjugateHelper(newCandidate, dupes, candidates, prefixCheck, suffixCheck, unlenite, checkInfixes, "", "")
		}
	}

	// fea checkeing for fe'a
	if len(lastSuffix) > 0 && len(input.word) > 0 {
		if hasAt(vowels, lastSuffix, 0) && (suffixRunes[0] == 'r' || hasAt(vowels, input.word, -1)) {
			//reef dialect has olori and oloru
			//source: https://naviteri.org/2026/04/hiia-tisung-postiya-aham-follow-up-to-the-previous-post/#comment-68711
			newCandidate := candidateDupe(input)
			newCandidate.word += "'"
			deconjugateHelper(newCandidate, dupes, candidates, prefixCheck, suffixCheck, unlenite, checkInfixes, "", "")
		}
	}

	// Exceptions for how words conjugate
	if len(input.suffixes) == 1 {
		if validWord, ok := weirdNounSuffixes[input.word]; ok {
			input.word = validWord
			if !isDuplicate(dupes, input) {
				*candidates = append(*candidates, input)
				candidateMap[input.word] = input
			}
			return
		}
	}

	if len(input.infixes) > 0 && implContainsAny(input.infixes, []string{"ats", "uy"}) {
		// for the cases of zen<ats>eke and zen<uy>eke
		// confirmed in here: https://forum.learnnavi.org/index.php?msg=493217
		if input.word == "zeneke" {
			input.word = "zenke"
			if !isDuplicate(dupes, input) {
				*candidates = append(*candidates, input)
				candidateMap[input.word] = input
			}
			return
		}
	}

	*candidates = append(*candidates, input)
	candidateMap[input.word] = input

	// Add a way for e to become ä again if we're down to 1 syllable
	if len([]rune(input.word)) < 8 && (len(input.prefixes) > 0 || len(input.infixes) > 0 || len(input.suffixes) > 0) && strings.Contains(input.word, "e") {
		// could be tskxäpx (7 letters 1 syllable)
		newCandidate := candidateDupe(input)
		newCandidate.word = strings.ReplaceAll(newCandidate.word, "e", "ä")
		deconjugateHelper(newCandidate, dupes, candidates, prefixCheck, suffixCheck, unlenite, checkInfixes, "", "")
	}

	newString := ""

	if input.insistPOS == NOUN || input.insistPOS == ANY {
		// For [word] si becoming [word]tswo
		if strings.HasSuffix(input.word, "tswo") {
			newCandidate := candidateDupe(input)
			newCandidate.word = strings.TrimSuffix(input.word, "tswo") + " si"
			newCandidate.insistPOS = VERB
			newCandidate.suffixes = isDuplicateFix(newCandidate.suffixes, "tswo")
			if !isDuplicate(dupes, newCandidate) {
				*candidates = append(*candidates, newCandidate)
				candidateMap[input.word] = input
			}
		}
	}

	if input.insistPOS == ADJ || input.insistPOS == ANY {
		// For lrrtok-susi and others
		if strings.HasSuffix(input.word, "-susi") || strings.HasSuffix(input.word, "-susia") {
			found := false
			trimmedWord := strings.TrimSuffix(input.word, "-susi")
			aPosition := 0
			if before, ok := strings.CutSuffix(input.word, "-susia"); ok {
				trimmedWord = before
				aPosition = 1
			}

			for _, pairWordSet := range multiword_words[trimmedWord] {
				if slices.Contains(pairWordSet, "si") {
					found = true
				}
				if found {
					break
				}
			}

			if !found && aPosition == 0 && strings.HasPrefix(trimmedWord, "a") {
				noA := strings.TrimPrefix(trimmedWord, "a")
				for _, pairWordSet := range multiword_words[noA] {
					if slices.Contains(pairWordSet, "si") {
						found = true
					}
					if found {
						aPosition = -1
						break
					}
				}
			}

			if !isDuplicate(dupes, input) {
				*candidates = append(*candidates, input)
				candidateMap[input.word] = input
			} // to bump the real candidate into recognition

			if found {
				newCandidate := candidateDupe(input)
				newCandidate.word = trimmedWord + " si"
				if aPosition == -1 {
					newCandidate.word = strings.TrimPrefix(trimmedWord, "a") + " si"
					newCandidate.prefixes = append(newCandidate.prefixes, "a")
				}
				newCandidate.infixes = []string{"us"}
				newCandidate.insistPOS = VERB
				if aPosition == 1 {
					newCandidate.suffixes = append(newCandidate.suffixes, "a")
				}
				if !isDuplicate(dupes, newCandidate) {
					*candidates = append(*candidates, newCandidate)
					candidateMap[input.word] = input
				}
			}
			return
		}
	}

	// Make sure that the first set of prefices (a, nì, ke) aren't combined with suffixes
	newPrefixCheck := max(prefixCheck, 1)

	// For making sure only the top one can check suffixes like this
	newSuffixCheck := max(suffixCheck, 2)

	switch prefixCheck {
	case 0:
		if strings.HasPrefix(input.word, "a") && input.insistPOS != NOUN && input.insistPOS != ADJ {
			// No nouns, adpositions or adverbs
			newCandidate := candidateDupe(input)
			newCandidate.word = input.word[1:]
			newCandidate.prefixes = isDuplicateFix(newCandidate.prefixes, "a")
			newCandidate.insistPOS = ADJ
			deconjugateHelper(newCandidate, dupes, candidates, 1, newSuffixCheck, -1, []string{}, "a", "")
			newCandidate.insistPOS = VERB
			deconjugateHelper(newCandidate, dupes, candidates, 1, newSuffixCheck, -1, []string{"", "", ""}, "a", "")
		} else if strings.HasPrefix(input.word, "nì") {
			newCandidate := candidateDupe(input)
			newCandidate.word = strings.TrimPrefix(input.word, "nì")
			newCandidate.prefixes = isDuplicateFix(newCandidate.prefixes, "nì")
			newCandidate.insistPOS = NÌ
			// No other affixes allowed
			deconjugateHelper(newCandidate, dupes, candidates, 10, 10, -1, []string{}, "nì", "") // No other fixes
		}
		fallthrough
	case 1:
		if input.insistPOS == ANY || input.insistPOS == ADJ {
			for _, element := range verbPrefixes {
				// If it has a prefix
				if strings.HasPrefix(input.word, element) {
					// remove it
					newCandidate := candidateDupe(input)
					newCandidate.word = strings.TrimPrefix(input.word, element)
					newCandidate.insistPOS = VERB
					newCandidate.prefixes = isDuplicateFix(newCandidate.prefixes, element)
					deconjugateHelper(newCandidate, dupes, candidates, 10, 10, -1, []string{}, element, "")

					// check "tsatan", "tan" and "atan"
					newCandidate.word = string(get_last_rune(element, 1)) + newCandidate.word
					deconjugateHelper(newCandidate, dupes, candidates, 10, 10, -1, []string{}, element, "")
				}
			}
		}
		fallthrough
	case 2:
		// Non-lenition prefixes for nouns only
		if input.insistPOS == ANY || input.insistPOS == NOUN {
			for _, element := range prefixes1Nouns {
				// If it has a prefix
				if newString, ok := strings.CutPrefix(input.word, element); ok {
					// remove it
					newCandidate := candidateDupe(input)
					newCandidate.word = newString
					newCandidate.insistPOS = NOUN
					newCandidate.prefixes = isDuplicateFix(newCandidate.prefixes, element)
					deconjugateHelper(newCandidate, dupes, candidates, 3, newSuffixCheck, -1, []string{}, element, "")

					// check "tsatan", "tan" and "atan"
					newCandidate.word = string(get_last_rune(element, 1)) + newString
					deconjugateHelper(newCandidate, dupes, candidates, 3, newSuffixCheck, -1, []string{}, element, "")
				}
			}

			// This one will demand this makes it use lenition
			for _, element := range prefixes1NounsLenition {
				// If it has a lenition-causing prefix
				if strings.HasPrefix(input.word, element) {
					lenited := false
					newString = strings.TrimPrefix(input.word, element)

					newCandidate := candidateDupe(input)
					newCandidate.word = newString
					newCandidate.prefixes = isDuplicateFix(newCandidate.prefixes, element)
					newCandidate.insistPOS = NOUN

					// Could it be pekoyu (pe + 'ekoyu, not pe + kxoyu)
					if hasAt(vowels, element, -1) {
						// check "pxeyktan", "yktan" and "eyktan"
						newCandidate.word = string(get_last_rune(element, 1)) + newString
						deconjugateHelper(newCandidate, dupes, candidates, 5, newSuffixCheck, -1, []string{}, element, "")

						// check "pxeylan", "ylan" and "'eylan"
						newCandidate.word = "'" + newCandidate.word
						deconjugateHelper(newCandidate, dupes, candidates, 5, newSuffixCheck, -1, []string{}, element, "")
					}

					// find out the possible unlenited forms
					for _, oldPrefix := range unlenitionLetters {
						// If it has a letter that could have changed for lenition,
						if strings.HasPrefix(newString, oldPrefix) {
							// put all possibilities in the candidates
							lenited = true

							for _, newPrefix := range unlenition[oldPrefix] {
								newCandidate.word = newPrefix + strings.TrimPrefix(newString, oldPrefix)
								if oldPrefix != newPrefix {
									newCandidate.lenition = []string{newPrefix + "→" + oldPrefix}
								}
								deconjugateHelper(newCandidate, dupes, candidates, 5, newSuffixCheck, -1, []string{}, oldPrefix, "")
							}
							break // We don't want the "ts" to become "txs"
						}
					}
					if !lenited {
						newCandidate.word = newString
						deconjugateHelper(newCandidate, dupes, candidates, 5, newSuffixCheck, -1, []string{}, element, "")
					}
				}
			}

			// If it has a lenition-causing prefix
			if strings.HasPrefix(input.word, "pe") {
				lenited := false
				newString = strings.TrimPrefix(input.word, "pe")

				newCandidate := candidateDupe(input)
				newCandidate.word = newString
				newCandidate.prefixes = isDuplicateFix(newCandidate.prefixes, "pe")
				newCandidate.insistPOS = NOUN

				// Could it be pekoyu (pe + 'ekoyu, not pe + kxoyu)
				if hasAt(vowels, "pe", -1) {
					// check "pxeyktan", "yktan" and "eyktan"
					newCandidate.word = string(get_last_rune("pe", 1)) + newString
					deconjugateHelper(newCandidate, dupes, candidates, 3, newSuffixCheck, -1, []string{}, "pe", "")

					// check "pxeylan", "ylan" and "'eylan"
					newCandidate.word = "'" + newCandidate.word
					deconjugateHelper(newCandidate, dupes, candidates, 3, newSuffixCheck, -1, []string{}, "pe", "")
				}

				// find out the possible unlenited forms
				for _, oldPrefix := range unlenitionLetters {
					// If it has a letter that could have changed for lenition,
					if strings.HasPrefix(newString, oldPrefix) {
						// put all possibilities in the candidates
						lenited = true

						for _, newPrefix := range unlenition[oldPrefix] {
							newCandidate.word = newPrefix + strings.TrimPrefix(newString, oldPrefix)
							if oldPrefix != newPrefix {
								newCandidate.lenition = []string{newPrefix + "→" + oldPrefix}
							}
							deconjugateHelper(newCandidate, dupes, candidates, 3, newSuffixCheck, -1, []string{}, oldPrefix, "")
						}
						break // We don't want the "ts" to become "txs"
					}
				}
				if !lenited {
					newCandidate.word = newString
					deconjugateHelper(newCandidate, dupes, candidates, 3, newSuffixCheck, -1, []string{}, "pe", "")
				}
			}
		}
		fallthrough
	case 3:
		if input.insistPOS == ANY || input.insistPOS == NOUN {
			// If it has a prefix
			if after, ok := strings.CutPrefix(input.word, "fra"); ok {
				// remove it
				newString = after

				newCandidate := candidateDupe(input)
				newCandidate.word = newString
				newCandidate.insistPOS = NOUN
				newCandidate.prefixes = isDuplicateFix(newCandidate.prefixes, "fra")

				deconjugateHelper(newCandidate, dupes, candidates, 4, newSuffixCheck, -1, []string{}, "fra", "")

				// check "tsatan", "tan" and "atan"
				newCandidate.word = "a" + newString
				deconjugateHelper(newCandidate, dupes, candidates, 4, newSuffixCheck, -1, []string{}, "fra", "")
			}
		}
		fallthrough
	case 4:
		if input.insistPOS == ANY || input.insistPOS == NOUN {
			// This one will demand this makes it use lenition
			for _, element := range prefixes1lenition {
				// If it has a lenition-causing prefix
				if strings.HasPrefix(input.word, element) {
					lenited := false
					newString = strings.TrimPrefix(input.word, element)

					newCandidate := candidateDupe(input)
					newCandidate.word = newString
					newCandidate.prefixes = isDuplicateFix(newCandidate.prefixes, element)
					newCandidate.insistPOS = NOUN

					// Could it be pekoyu (pe + 'ekoyu, not pe + kxoyu)
					// check "pxeyktan", "yktan" and "eyktan"
					newCandidate.word = string(get_last_rune(element, 1)) + newString
					deconjugateHelper(newCandidate, dupes, candidates, 5, newSuffixCheck, -1, []string{}, element, "")

					// check "pxeylan", "ylan" and "'eylan"
					newCandidate.word = "'" + newCandidate.word
					deconjugateHelper(newCandidate, dupes, candidates, 5, newSuffixCheck, -1, []string{}, element, "")

					// find out the possible unlenited forms
					for _, oldPrefix := range unlenitionLetters {
						// If it has a letter that could have changed for lenition,
						if strings.HasPrefix(newString, oldPrefix) {
							// put all possibilities in the candidates
							lenited = true

							for _, newPrefix := range unlenition[oldPrefix] {
								newCandidate.word = newPrefix + strings.TrimPrefix(newString, oldPrefix)
								if oldPrefix != newPrefix {
									newCandidate.lenition = []string{newPrefix + "→" + oldPrefix}
								}
								deconjugateHelper(newCandidate, dupes, candidates, 5, newSuffixCheck, -1, []string{}, oldPrefix, "")
							}
							break // We don't want the "ts" to become "txs"
						}
					}
					if !lenited {
						newCandidate.word = newString
						deconjugateHelper(newCandidate, dupes, candidates, 5, newSuffixCheck, -1, []string{}, element, "")
					}
				}
			}
		}
		fallthrough
	case 5:
		if input.insistPOS == ANY || input.insistPOS == NOUN {
			for _, element := range stemPrefixes {
				// If it has a prefix
				if strings.HasPrefix(input.word, element) {
					// remove it
					newCandidate := candidateDupe(input)
					newCandidate.word = strings.TrimPrefix(input.word, element)
					newCandidate.insistPOS = NOUN
					newCandidate.prefixes = isDuplicateFix(newCandidate.prefixes, element)
					deconjugateHelper(newCandidate, dupes, candidates, 6, newSuffixCheck, -1, []string{}, element, "")

					// check "tsatan", "tan" and "atan"
					newCandidate.word = string(get_last_rune(element, 1)) + newCandidate.word
					deconjugateHelper(newCandidate, dupes, candidates, 6, newSuffixCheck, -1, []string{}, element, "")
				}
			}
		}
		fallthrough
	case 6:
		if newString, ok := strings.CutPrefix(input.word, "tì"); ok {
			if input.insistPOS == ANY || input.insistPOS == NOUN {
				// remove it
				newCandidate := candidateDupe(input)
				newCandidate.word = newString
				newCandidate.insistPOS = VERB
				newCandidate.prefixes = isDuplicateFix(newCandidate.prefixes, "tì")
				deconjugateHelper(newCandidate, dupes, candidates, 10, 10, -1, []string{"", "", ""}, "tì", "") // No other prefixes allowed

				newCandidate.word = "ì" + newCandidate.word
				deconjugateHelper(newCandidate, dupes, candidates, 10, 10, -1, []string{"", "", ""}, "tì", "") // Or any additional suffixes
			}
		}
	}

	switch suffixCheck {
	case 0:
		// Made sì its own suffix and no suffixes can come after it
		if len(input.suffixes) == 0 && strings.HasSuffix(input.word, "sì") {
			newCandidate := candidateDupe(input)
			newCandidate.word = strings.TrimSuffix(newCandidate.word, "sì")
			newCandidate.suffixes = append(newCandidate.suffixes, "sì")
			deconjugateHelper(newCandidate, dupes, candidates, newPrefixCheck, 1, unlenite, checkInfixes, "", "sì")
		}
		// special case: short genitives of pronouns like "oey" and "ngey"
		if input.insistPOS == ANY || input.insistPOS == NOUN {
			if strings.HasSuffix(input.word, "y") {
				// oey to oe
				newCandidate := candidateDupe(input)
				newCandidate.word = strings.TrimSuffix(input.word, "y")
				newCandidate.insistPOS = PN
				newCandidate.suffixes = isDuplicateFix(newCandidate.suffixes, "y")
				deconjugateHelper(newCandidate, dupes, candidates, newPrefixCheck, 10, unlenite, []string{}, "", "y")

				// ngey to nga
				if before, ok := strings.CutSuffix(newCandidate.word, "e"); ok {
					newCandidate.word = before + "a"
					newCandidate.insistPOS = PN
					deconjugateHelper(newCandidate, dupes, candidates, newPrefixCheck, 10, unlenite, []string{}, "", "y")
				}
			}
		}
		fallthrough
	case 1:
		if input.insistPOS == ANY || input.insistPOS == NOUN {
			for oldSuffix, actual := range lenitionAdposition {
				if newString, ok := strings.CutSuffix(input.word, oldSuffix); ok {
					newCandidate := candidateDupe(input)
					newCandidate.word = newString + actual[:2]
					newCandidate.insistPOS = NOUN
					newCandidate.suffixes = isDuplicateFix(newCandidate.suffixes, actual)
					// all set to 2 to avoid mengeyä -> mengo -> me + 'eng + o
					deconjugateHelper(newCandidate, dupes, candidates, newPrefixCheck, 2, unlenite, []string{}, "", actual)

					newCandidate.word = newString + actual[:1]
					deconjugateHelper(newCandidate, dupes, candidates, newPrefixCheck, 2, unlenite, []string{}, "", actual)
				}
			}
			for _, oldSuffix := range adposuffixes {
				// If it has one of them,
				if before, ok := strings.CutSuffix(input.word, oldSuffix); ok {
					newString = before

					// Make sure you're using a valid case ending
					if !verifyCaseEnding(newString, oldSuffix) {
						continue
					}

					newCandidate := candidateDupe(input)
					newCandidate.word = newString
					newCandidate.insistPOS = NOUN
					newCandidate.suffixes = isDuplicateFix(newCandidate.suffixes, oldSuffix)
					// all set to 2 to avoid mengeyä -> mengo -> me + 'eng + o
					deconjugateHelper(newCandidate, dupes, candidates, newPrefixCheck, 2, unlenite, []string{}, "", oldSuffix)

					if oldSuffix == "ä" && !strings.HasSuffix(input.word, "yä") && strings.HasSuffix(input.word, "iä") { // Don't make peyä -> yä -> ya (air)
						// soaiä, tìftiä, etx.
						newString += "a"
						newCandidate.word = newString
						deconjugateHelper(newCandidate, dupes, candidates, newPrefixCheck, 2, unlenite, []string{}, "", oldSuffix)
					} else if oldSuffix == "e" && !strings.HasSuffix(input.word, "ye") && strings.HasSuffix(input.word, "ie") {
						// reef of above
						newString += "a"
						newCandidate.word = newString
						deconjugateHelper(newCandidate, dupes, candidates, newPrefixCheck, 2, unlenite, []string{}, "", "ä")
					} else if oldSuffix == "yä" && strings.HasSuffix(newString, "e") {
						// A one-off
						if newString == "tse" {
							newCandidate.word = "tsaw"
							deconjugateHelper(newCandidate, dupes, candidates, newPrefixCheck, 2, unlenite, []string{}, "", oldSuffix)
						}
						// ngeyä -> nga
						newCandidate.word = strings.TrimSuffix(newString, "e") + "a"
						deconjugateHelper(newCandidate, dupes, candidates, newPrefixCheck, 2, unlenite, []string{}, "", oldSuffix)
						// oengeyä
						newCandidate.word = strings.TrimSuffix(newString, "e")
						if newCandidate.word == "oeng" { //no mengeyä -> meng -> me + 'eng
							deconjugateHelper(newCandidate, dupes, candidates, newPrefixCheck, 2, unlenite, []string{}, "", oldSuffix)
						}
						// sneyä -> sno
						newCandidate.word = strings.TrimSuffix(newString, "e") + "o"
						deconjugateHelper(newCandidate, dupes, candidates, newPrefixCheck, 2, unlenite, []string{}, "", oldSuffix)
					} else if oldSuffix == "ye" && strings.HasSuffix(newString, "e") {
						// reef of above
						if newString == "tse" {
							newCandidate.word = "tsaw"
							deconjugateHelper(newCandidate, dupes, candidates, newPrefixCheck, 2, unlenite, []string{}, "", "yä")
						}
						// ngeye -> nga
						newCandidate.word = strings.TrimSuffix(newString, "e") + "a"
						deconjugateHelper(newCandidate, dupes, candidates, newPrefixCheck, 2, unlenite, []string{}, "", "yä")
						// oengeye
						newCandidate.word = strings.TrimSuffix(newString, "e")
						if newCandidate.word == "oeng" { //no mengeyä -> meng -> me + 'eng
							deconjugateHelper(newCandidate, dupes, candidates, newPrefixCheck, 2, unlenite, []string{}, "", "yä")
						}
						// sneye -> sno
						newCandidate.word = strings.TrimSuffix(newString, "e") + "o"
						deconjugateHelper(newCandidate, dupes, candidates, newPrefixCheck, 2, unlenite, []string{}, "", "yä")
					} else if vowels, ok := vowelSuffixes["yä"]; ok {
						for _, vowel := range vowels {
							// Make sure zekwä-äo is recognized
							if strings.HasSuffix(newString, vowel+"-") {
								newString = strings.TrimSuffix(newString, "-")
								newCandidate.word = newString
								deconjugateHelper(newCandidate, dupes, candidates, newPrefixCheck, 2, unlenite, []string{}, "", "yä")
							}
						}
					}
				}
			}
		}
		fallthrough
	case 2:
		if input.insistPOS == ANY || input.insistPOS == NOUN {
			if newString, ok := strings.CutSuffix(input.word, "pe"); ok {
				newCandidate := candidateDupe(input)
				newCandidate.word = newString
				newCandidate.insistPOS = NOUN
				newCandidate.suffixes = isDuplicateFix(newCandidate.suffixes, "pe")
				deconjugateHelper(newCandidate, dupes, candidates, newPrefixCheck, 4, unlenite, []string{}, "", "pe")
			}
		}
		fallthrough
	case 3:
		// If it has one of them,
		if strings.HasSuffix(input.word, "a") && input.insistPOS != NOUN && input.insistPOS != ADJ {
			// No nouns, adpositions or adverbs
			newString = strings.TrimSuffix(input.word, "a")

			newCandidate := candidateDupe(input)
			newCandidate.word = newString
			newCandidate.insistPOS = ADJ
			newCandidate.suffixes = isDuplicateFix(newCandidate.suffixes, "a")
			deconjugateHelper(newCandidate, dupes, candidates, newPrefixCheck, 4, unlenite, []string{"", "", ""}, "", "a")
			newCandidate.insistPOS = VERB
			deconjugateHelper(newCandidate, dupes, candidates, newPrefixCheck, 4, unlenite, []string{"", "", ""}, "", "a")
		}

		fallthrough
	case 4: // -o suffix "some"
		if input.insistPOS == ANY || input.insistPOS == NOUN {
			if newString, ok := strings.CutSuffix(input.word, "o"); ok {
				newCandidate := candidateDupe(input)
				newCandidate.word = newString
				newCandidate.insistPOS = NOUN
				newCandidate.suffixes = isDuplicateFix(newCandidate.suffixes, "o")
				deconjugateHelper(newCandidate, dupes, candidates, newPrefixCheck, 5, unlenite, []string{}, "", "o")

				// Make sure fya'o-o is recognized
				if vowels, ok := vowelSuffixes["o"]; ok {
					for _, vowel := range vowels {
						// Make sure fya'o-o is recognized
						if strings.HasSuffix(newString, vowel+"-") {
							newString = strings.TrimSuffix(newString, "-")
							newCandidate.word = newString
							deconjugateHelper(newCandidate, dupes, candidates, newPrefixCheck, 5, unlenite, []string{}, "", "o")
						}
					}
				}
			}
		}
		fallthrough
	case 5:
		if input.insistPOS == ANY || input.insistPOS == NOUN {
			for _, oldSuffix := range stemSuffixes {
				// If it has one of them,
				if before, ok := strings.CutSuffix(input.word, oldSuffix); ok {
					newString = before

					//candidates = append(candidates, newString)
					newCandidate := candidateDupe(input)
					newCandidate.word = newString
					newCandidate.insistPOS = NOUN
					newCandidate.suffixes = isDuplicateFix(newCandidate.suffixes, oldSuffix)
					deconjugateHelper(newCandidate, dupes, candidates, newPrefixCheck, 6, unlenite, []string{}, "", oldSuffix)
				}
			}
		}
		fallthrough
	case 6:
		// If it has one of them,
		if input.insistPOS == ANY || input.insistPOS == NOUN {
			// verb suffixes change things from verbs to nouns, that's why we check for noun status
			for _, oldSuffix := range verbSuffixes {
				// If it has one of them,
				if newString, ok := strings.CutSuffix(input.word, oldSuffix); ok {
					newCandidate := candidateDupe(input)
					newCandidate.word = newString
					newCandidate.insistPOS = VERB

					newCandidate.suffixes = isDuplicateFix(newCandidate.suffixes, oldSuffix)
					deconjugateHelper(newCandidate, dupes, candidates, 10, 10, unlenite, []string{"", "", ""}, "", oldSuffix) // Don't allow any other prefixes
					// They may turn the insistPOS back into a noun

					if oldSuffix == "yu" && strings.HasSuffix(newString, "si") {
						newCandidate.word = strings.TrimSuffix(newString, "si") + " si"
						deconjugateHelper(newCandidate, dupes, candidates, 10, 10, unlenite, []string{}, "", oldSuffix) // don't allow any other prefixes or suffixes
					}
				}
			}
		}
	}

	// Short lenition check
	if unlenite != -1 {
		for _, oldPrefix := range unlenitionLetters {
			// If it has a letter that could have changed for lenition,
			if after, ok := strings.CutPrefix(input.word, oldPrefix); ok {
				// put all possibilities in the candidates
				for _, newPrefix := range unlenition[oldPrefix] {
					newCandidate := candidateDupe(input)
					newString = newPrefix + after
					newCandidate.word = newString
					if oldPrefix != newPrefix {
						newCandidate.lenition = []string{newPrefix + "→" + oldPrefix}
					}
					deconjugateHelper(newCandidate, dupes, candidates, prefixCheck, suffixCheck, -1, []string{}, "", "")
				}
				break // We don't want the "ts" to become "txs"
			}
		}
	}

	if len(checkInfixes) == 3 && len(input.infixes) < 3 {
		// Maybe someone else came in with stripped infixes
		if len(input.word) > 2 && input.word[len(input.word)-3] != ' ' &&
			strings.HasSuffix(input.word, "si") && !strings.HasSuffix(input.word, "usi") &&
			!strings.HasSuffix(input.word, "atsi") {
			newCandidate := candidateDupe(input)
			newCandidate.word = strings.TrimSuffix(input.word, "si") + " si"
			newCandidate.insistPOS = VERB
			deconjugateHelper(newCandidate, dupes, candidates, newPrefixCheck, suffixCheck, unlenite, []string{"", "", ""}, "", "")
		} else { // If there is a "si", we don't need to check for infixes
			// Check for infixes
			runes := []rune(input.word)
			for i, c := range runes {
				// Infixes can only begin with vowels
				if is_vowel(c) {
					shortString := string(runes[i:])
					for _, infix := range infixes[c] {
						available, newInfixes := verifyInfix(checkInfixes, infix)
						if available && strings.HasPrefix(shortString, infix) {
							newCandidate := candidateDupe(input)
							newCandidate.word = string(runes[:i]) + strings.TrimPrefix(shortString, infix)
							newCandidate.infixes = isDuplicateFix(newCandidate.infixes, infix)
							newCandidate.insistPOS = VERB
							deconjugateHelper(newCandidate, dupes, candidates, newPrefixCheck, suffixCheck, unlenite, newInfixes, "", "")

							switch infix {
							case "ol":
								newCandidate := candidateDupe(input)
								newCandidate.word = string(runes[:i]) + "ll" + strings.TrimPrefix(shortString, infix)
								newCandidate.infixes = isDuplicateFix(newCandidate.infixes, infix)
								newCandidate.insistPOS = VERB
								deconjugateHelper(newCandidate, dupes, candidates, newPrefixCheck, suffixCheck, unlenite, newInfixes, "", "")
							case "er":
								newCandidate := candidateDupe(input)
								newCandidate.word = string(runes[:i]) + "rr" + strings.TrimPrefix(shortString, infix)
								newCandidate.infixes = isDuplicateFix(newCandidate.infixes, infix)
								newCandidate.insistPOS = VERB
								deconjugateHelper(newCandidate, dupes, candidates, newPrefixCheck, suffixCheck, unlenite, newInfixes, "", "")
							}
						}
					}
				}
			}
		}
	}
}

func deconjugate(input string, candidates *[]ConjugationCandidate) {
	newCandidate := ConjugationCandidate{}
	newCandidate.word = input
	newCandidate.insistPOS = ANY
	deconjugateHelper(newCandidate, &map[string]ConjugationCandidate{}, candidates, 0, 0, 0, []string{"", "", ""}, "", "")
}

func TestDeconjugations(searchNaviWord string) (results []Word) {
	conjugations := []ConjugationCandidate{}
	deconjugate(searchNaviWord, &conjugations)
	for i, candidate := range conjugations {
		if i == 0 {
			continue
		}
		a := strings.ReplaceAll(candidate.word, "ù", "u")
		standardizedWordArray := dialectCrunch(strings.Split(a, " "), false)
		a = ""
		for i, b := range standardizedWordArray {
			if i != 0 {
				a += " "
			}
			a += b
		}

		for _, c := range dictHash[a] {
			for pos := range strings.SplitSeq(c.PartOfSpeech, ",") {
				pos = strings.ReplaceAll(pos, " ", "")

				// An inter. can act like a noun or an adjective, so it gets special treatment
				if pos == "inter." && candidate.insistPOS == VERB && len(candidate.infixes) == 0 {
					dupe := false
					for _, b := range results {
						if b.Navi == c.Navi {
							dupe = true
							break
						}
					}
					if !dupe {
						a := c
						a.Affixes.Lenition = candidate.lenition
						a.Affixes.Prefix = candidate.prefixes
						a.Affixes.Infix = candidate.infixes
						a.Affixes.Suffix = candidate.suffixes
						results = AppendAndAlphabetize(results, a)
						continue
					}
				}

				// Find gerunds (tì-v<us>erb, treated like a noun)
				gerund := false
				infixBan := false
				doubleBan := false
				attributed := false
				participle := false

				// Find gerunds (tì-v<us>erb, the act of [verb]ing)
				if len(candidate.infixes) == 1 && candidate.infixes[0] == "us" {
					// Reverse search is more likely to find it immediately
					for _, v := range slices.Backward(candidate.prefixes) {
						if v == "tì" {
							gerund = true
							break
						}
					}
					if !gerund {
						participle = true
					}
				} else if len(candidate.infixes) > 0 {
					// Now reverse search is just gratuitous
					for _, v := range slices.Backward(candidate.infixes) {
						if v == "us" || v == "awn" {
							participle = true
							break
						}
					}
				}

				// If the insistPOS and found word agree they are nouns
				if len(candidate.suffixes) < 3 && len(candidate.suffixes) > 0 && candidate.suffixes[0] == "tswo" {
					if pos[0] == 'v' {
						siVerb := false
						if len(candidate.infixes) == 0 {
							if _, ok := multiword_words[candidate.word]; ok {
								for _, b := range multiword_words[candidate.word] {
									if b[0] == "si" {
										siVerb = true
										a := c
										a.Navi = candidate.word + " si"
										a.Affixes.Lenition = candidate.lenition
										a.Affixes.Prefix = candidate.prefixes
										a.Affixes.Infix = candidate.infixes
										a.Affixes.Suffix = candidate.suffixes
										results = AppendAndAlphabetize(results, a)
										break
									}
								}
							}
							if !siVerb {
								a := c
								a.Navi = candidate.word
								a.Affixes.Lenition = candidate.lenition
								a.Affixes.Prefix = candidate.prefixes
								a.Affixes.Infix = candidate.infixes
								a.Affixes.Suffix = candidate.suffixes
								results = AppendAndAlphabetize(results, a)
							}
						}
					}
				} else if gerund {
					if pos[0] == 'v' {
						// Make sure the <us> is in the correct place
						rebuiltVerb := strings.ReplaceAll(c.InfixLocations, "<0>", "")
						rebuiltVerb = strings.ReplaceAll(rebuiltVerb, "<1>", "us")
						rebuiltVerb = strings.ReplaceAll(rebuiltVerb, "<2>", "")

						// Does the noun actually contain the verb?
						noTìftang := strings.TrimPrefix(rebuiltVerb, "'")
						if strings.Contains(searchNaviWord, noTìftang) || strings.Contains(searchNaviWord, dialectCrunch([]string{rebuiltVerb}, false)[0]) {
							a := c
							a.Affixes.Lenition = candidate.lenition
							a.Affixes.Prefix = candidate.prefixes
							a.Affixes.Infix = candidate.infixes
							a.Affixes.Suffix = candidate.suffixes
							results = AppendAndAlphabetize(results, a)
						} /*else if len(results) == 0 {
							results = AppendAndAlphabetize(results, infixError(searchNaviWord, "tì"+rebuiltVerb, c.IPA))
						}*/
					}
				} else if candidate.insistPOS == NOUN {
					// n., pn., Prop.n. and inter. (but not vin.)
					if len(candidate.infixes) == 0 {
						if (pos[0] != 'v' && strings.HasSuffix(pos, "n.")) || pos == "inter." {
							a := c
							a.Affixes.Lenition = candidate.lenition
							a.Affixes.Prefix = candidate.prefixes
							a.Affixes.Suffix = candidate.suffixes
							results = AppendAndAlphabetize(results, a)
						}
					}
				} else if candidate.insistPOS == ADJ {
					posNoun := pos
					if len(candidate.infixes) == 0 && (posNoun == "adj." || posNoun == "num.") {
						a := c
						a.Affixes.Lenition = candidate.lenition
						a.Affixes.Prefix = candidate.prefixes
						a.Affixes.Suffix = candidate.suffixes
						results = AppendAndAlphabetize(results, a)
					}
				} else if candidate.insistPOS == VERB {
					posNoun := pos
					if strings.HasPrefix(posNoun, "v") {
						// Verbs with -tswo or -yu cannot have infixes
						if len(candidate.suffixes) > 0 {
							for _, v := range slices.Backward(candidate.suffixes) {
								if v == "a" {
									attributed = true
									break
								}
							}
							// Forward search fixs the "a" before "yu" and "tswo"
							for _, v := range slices.Backward(candidate.suffixes) {
								if slices.Contains(verbSuffixes, v) {
									infixBan = true
								}

								if infixBan {
									break
								}
							}
						}

						// Assuming v<äp>erb-yu is productive
						if len(candidate.infixes) == 1 {
							if candidate.infixes[0] == "äp" || candidate.infixes[0] == "eyk" {
								infixBan = false
							}
						}

						looseTì := false
						tsuk := false

						if len(candidate.prefixes) > 0 {
							// Reverse search is more likely to find it immediately
							for _, v := range slices.Backward(candidate.prefixes) {
								if v == "a" {
									attributed = true
								} else if v == "tì" {
									// we found gerunds up top, so this isn't needed
									looseTì = true
									break
								} else {
									for _, j := range verbPrefixes {
										if v == j {
											if infixBan {
												doubleBan = true
												break
											}
											infixBan = true
											tsuk = true
											break
										}
									}
								}

								if infixBan || doubleBan || looseTì {
									break
								}
							}
						}

						// Assuming v<äp>erb-yu is productive
						if len(candidate.infixes) == 1 {
							if candidate.infixes[0] == "äp" || candidate.infixes[0] == "eyk" {
								infixBan = false
							}
						}

						// Don't want a[verb] and [verb]a
						if attributed && (len(candidate.infixes) == 0 || infixBan) && !tsuk {
							continue
						}

						// Take action on tsuk-verb-yus and a-verb-tswos
						if doubleBan || (attributed && !tsuk && infixBan) || looseTì {
							continue
						}

						a := c
						a.Affixes.Lenition = candidate.lenition
						a.Affixes.Prefix = candidate.prefixes
						a.Affixes.Suffix = candidate.suffixes
						a.Affixes.Infix = candidate.infixes

						if infixBan {
							if len(candidate.infixes) > 0 {
								continue // No nonsense here
							} else {
								results = AppendAndAlphabetize(results, a)
							}
						}

						// Make it verify the infixes are in the correct place
						ol := false
						er := false

						// pre-first position infixes
						rebuiltVerb := c.InfixLocations
						if c.InfixLocations == "z<0><1>en<2>ke" && implContainsAny(candidate.infixes, []string{"ats", "uy"}) {
							rebuiltVerb = "z<0><1>en<2>eke"
						}
						firstInfixes := ""

						for _, newInfix := range candidate.infixes {
							if implContainsAny(prefirst, []string{newInfix}) {
								firstInfixes += newInfix
								rebuiltVerb = strings.ReplaceAll(rebuiltVerb, "<0>", firstInfixes)
								if newInfix == "epeyk" || newInfix == "äpeyk" {
									newCandidateInfixes := []string{}
									for _, newInfix2 := range candidate.infixes {
										// äpeyk gets split
										if newInfix2 == "epeyk" || newInfix2 == "äpeyk" {
											newCandidateInfixes = append(newCandidateInfixes, "äp")
											newCandidateInfixes = append(newCandidateInfixes, "eyk")
										} else {
											newCandidateInfixes = append(newCandidateInfixes, newInfix2)
										}
									}
									a.Affixes.Infix = newCandidateInfixes
								}
								break
							}
						}
						rebuiltVerb = strings.ReplaceAll(rebuiltVerb, "<0>", "")

						// first position infixes
						firstInfixes = ""
						for _, newInfix := range candidate.infixes {
							if implContainsAny(first, []string{newInfix}) {
								rebuiltVerb = strings.ReplaceAll(rebuiltVerb, "<1>", newInfix)
								firstInfixes = newInfix
								switch newInfix {
								case "ol":
									ol = true
								case "er":
									er = true
								}
								break
							}
						}
						rebuiltVerb = strings.ReplaceAll(rebuiltVerb, "<1>", "")

						// second position infixes
						for _, newInfix := range candidate.infixes {
							if newInfix == "eng" {
								rebuiltVerb = strings.ReplaceAll(rebuiltVerb, "<2>", "äng")
								break
							} else if implContainsAny(second, []string{newInfix}) {
								rebuiltVerb = strings.ReplaceAll(rebuiltVerb, "<2>", newInfix)
								break
							}
						}
						rebuiltVerb = strings.ReplaceAll(rebuiltVerb, "<2>", "")

						rebuiltVerb = strings.TrimSpace(rebuiltVerb)

						if ol && strings.Contains(rebuiltVerb, "olll") {
							rebuiltVerb = strings.ReplaceAll(rebuiltVerb, "olll", "ol")
						}
						if er && strings.Contains(rebuiltVerb, "errr") {
							rebuiltVerb = strings.ReplaceAll(rebuiltVerb, "errr", "er")
						}

						if slices.Contains(candidate.suffixes, "yu") {
							rebuiltVerb += "yu"
						}

						//rebuiltVerbForest := rebuiltVerb
						rebuiltVerbArray := dialectCrunch(strings.Split(rebuiltVerb, " "), false)
						rebuiltVerb = ""
						for k, x := range rebuiltVerbArray {
							if k != 0 {
								rebuiltVerb += " "
							}
							rebuiltVerb += x
						}

						if len(candidate.infixes) == 0 || identicalRunes(rebuiltVerb, strings.ReplaceAll(searchNaviWord, "-", " ")) {
							results = AppendAndAlphabetize(results, a)
						} else if participle {
							// In case we have a [word]-susi
							rebuiltHyphen := strings.ReplaceAll(searchNaviWord, "-", " ")
							if identicalRunes("a"+rebuiltVerb, rebuiltHyphen) {
								// a-v<us>erb and a-v<awn>erb
								results = AppendAndAlphabetize(results, a)
							} else if identicalRunes(rebuiltVerb+"a", rebuiltHyphen) {
								// v<us>erb-a and v<awn>erb-a
								results = AppendAndAlphabetize(results, a)
							} else if rebuiltVerb[0] == '\'' && identicalRunes("a"+rebuiltVerb[1:], rebuiltHyphen) {
								// a-'<us>em
								results = AppendAndAlphabetize(results, a)
							} else if rebuiltVerb[len(rebuiltVerb)-1] == '\'' && identicalRunes(rebuiltVerb[:len(rebuiltVerb)-1]+"a", rebuiltHyphen) {
								// fp<us>e'a
								results = AppendAndAlphabetize(results, a)
							} /*else if firstInfixes == "us" {
								if len(results) == 0 {
									results = AppendAndAlphabetize(results, infixError(searchNaviWord, rebuiltVerbForest, c.IPA))
								}
							}*/
						} /*else if gerund { // ti is needed to weed out non-productive tì-verbs
							if len(results) == 0 {
								results = AppendAndAlphabetize(results, infixError(searchNaviWord, rebuiltVerbForest, c.IPA))
							}
						} else {
							if len(results) == 0 {
								results = AppendAndAlphabetize(results, infixError(searchNaviWord, rebuiltVerbForest, c.IPA))
							}
						}*/
					}
				} else if candidate.insistPOS == NÌ {
					posNoun := pos
					if len(candidate.infixes) == 0 && (posNoun == "adj." || posNoun == "pn.") {
						a := c
						a.Affixes.Lenition = candidate.lenition
						a.Affixes.Prefix = candidate.prefixes
						a.Affixes.Suffix = candidate.suffixes
						results = AppendAndAlphabetize(results, a)
					}
				} else if len(candidate.infixes) == 0 {
					a := c
					a.Affixes.Lenition = candidate.lenition
					a.Affixes.Prefix = candidate.prefixes
					a.Affixes.Suffix = candidate.suffixes
					results = AppendAndAlphabetize(results, a)
				}
			}
		}
	}
	return
}
