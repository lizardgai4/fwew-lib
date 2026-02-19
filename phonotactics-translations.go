package main

import (
	"strconv"
	"strings"
)

var message_non_navi_letters = map[string]string{
	"en": "**{oldWord}** Has letters not in Na'vi: `{nonNaviLetters}`",       // English
	"de": "**{oldWord}** 🇩🇪 Has letters not in Na'vi: `{nonNaviLetters}`",    // German (Deutsch)
	"es": "**{oldWord}** 🇪🇦 Has letters not in Na'vi: `{nonNaviLetters}`",    // Spanish (Español)
	"et": "**{oldWord}** 🇪🇪 Has letters not in Na'vi: `{nonNaviLetters}`",    // Estonian (Eesti)
	"fr": "**{oldWord}** 🇫🇷 Has letters not in Na'vi: `{nonNaviLetters}`",    // French (Français)
	"hu": "**{oldWord}** 🇭🇺 Has letters not in Na'vi: `{nonNaviLetters}`",    // Hungarian (Magyar)
	"ko": "**{oldWord}**에는 나비어에 존재하지 않는 낱말이 포함되어 있습니다. - `{nonNaviLetters}`", // Korean (한국어)
	"nl": "**{oldWord}** 🇳🇱 Has letters not in Na'vi: `{nonNaviLetters}`",    // Dutch (Nederlands)
	"pl": "**{oldWord}** 🇵🇱 Has letters not in Na'vi: `{nonNaviLetters}`",    // Polish (Polski)
	"pt": "**{oldWord}** 🇵🇹 Has letters not in Na'vi: `{nonNaviLetters}`",    // Portuguese (Português)
	"ru": "**{oldWord}** 🇷🇺 Has letters not in Na'vi: `{nonNaviLetters}`",    // Russian (Русский)
	"sv": "**{oldWord}** 🇸🇪 Has letters not in Na'vi: `{nonNaviLetters}`",    // Swedish (Svenska)
	"tr": "**{oldWord}** 🇹🇷 Has letters not in Na'vi: `{nonNaviLetters}`",    // Turkish (Türkçe)
	"uk": "**{oldWord}** 🇺🇦 Has letters not in Na'vi: `{nonNaviLetters}`",    // Ukrainian (Українська)
}

var message_no_nuclei = map[string]string{
	"en": "**{oldWord}** Error: could not find any syllable nuclei",    // English
	"de": "**{oldWord}** 🇩🇪 Error: could not find any syllable nuclei", // German (Deutsch)
	"es": "**{oldWord}** 🇪🇦 Error: could not find any syllable nuclei", // Spanish (Español)
	"et": "**{oldWord}** 🇪🇪 Error: could not find any syllable nuclei", // Estonian (Eesti)
	"fr": "**{oldWord}** 🇫🇷 Error: could not find any syllable nuclei", // French (Français)
	"hu": "**{oldWord}** 🇭🇺 Error: could not find any syllable nuclei", // Hungarian (Magyar)
	"ko": "**{oldWord}**에서 음절핵(중성)에 해당하는 요소를 찾을 수 없습니다.",               // Korean (한국어)
	"nl": "**{oldWord}** 🇳🇱 Error: could not find any syllable nuclei", // Dutch (Nederlands)
	"pl": "**{oldWord}** 🇵🇱 Error: could not find any syllable nuclei", // Polish (Polski)
	"pt": "**{oldWord}** 🇵🇹 Error: could not find any syllable nuclei", // Portuguese (Português)
	"ru": "**{oldWord}** 🇷🇺 Error: could not find any syllable nuclei", // Russian (Русский)
	"sv": "**{oldWord}** 🇸🇪 Error: could not find any syllable nuclei", // Swedish (Svenska)
	"tr": "**{oldWord}** 🇹🇷 Error: could not find any syllable nuclei", // Turkish (Türkçe)
	"uk": "**{oldWord}** 🇺🇦 Error: could not find any syllable nuclei", // Ukrainian (Українська)
}

var message_invalid_consonants = map[string]string{
	"en": "**{oldWord}** Invalid consonant combination: `{badConsonants}`",    // English
	"de": "**{oldWord}** 🇩🇪 Invalid consonant combination: `{badConsonants}`", // German (Deutsch)
	"es": "**{oldWord}** 🇪🇦 Invalid consonant combination: `{badConsonants}`", // Spanish (Español)
	"et": "**{oldWord}** 🇪🇪 Invalid consonant combination: `{badConsonants}`", // Estonian (Eesti)
	"fr": "**{oldWord}** 🇫🇷 Invalid consonant combination: `{badConsonants}`", // French (Français)
	"hu": "**{oldWord}** 🇭🇺 Invalid consonant combination: `{badConsonants}`", // Hungarian (Magyar)
	"ko": "**{oldWord}**에 유효하지 않은 조합이 발견되었습니다. - `{badConsonants}`",           // Korean (한국어)
	"nl": "**{oldWord}** 🇳🇱 Invalid consonant combination: `{badConsonants}`", // Dutch (Nederlands)
	"pl": "**{oldWord}** 🇵🇱 Invalid consonant combination: `{badConsonants}`", // Polish (Polski)
	"pt": "**{oldWord}** 🇵🇹 Invalid consonant combination: `{badConsonants}`", // Portuguese (Português)
	"ru": "**{oldWord}** 🇷🇺 Invalid consonant combination: `{badConsonants}`", // Russian (Русский)
	"sv": "**{oldWord}** 🇸🇪 Invalid consonant combination: `{badConsonants}`", // Swedish (Svenska)
	"tr": "**{oldWord}** 🇹🇷 Invalid consonant combination: `{badConsonants}`", // Turkish (Türkçe)
	"uk": "**{oldWord}** 🇺🇦 Invalid consonant combination: `{badConsonants}`", // Ukrainian (Українська)
}

var message_needed_vowel = map[string]string{
	"en": "**{oldWord}** Needs a vowel, diphthong or psuedovowel here: `{breakdown}`",               // English
	"de": "**{oldWord}** 🇩🇪 Needs a vowel, diphthong or psuedovowel here: `{breakdown}`",            // German (Deutsch)
	"es": "**{oldWord}** 🇪🇦 Needs a vowel, diphthong or psuedovowel here: `{breakdown}`",            // Spanish (Español)
	"et": "**{oldWord}** 🇪🇪 Needs a vowel, diphthong or psuedovowel here: `{breakdown}`",            // Estonian (Eesti)
	"fr": "**{oldWord}** 🇫🇷 Needs a vowel, diphthong or psuedovowel here: `{breakdown}`",            // French (Français)
	"hu": "**{oldWord}** 🇭🇺 Needs a vowel, diphthong or psuedovowel here: `{breakdown}`",            // Hungarian (Magyar)
	"ko": "**{oldWord}** 에 유효하지 않은 자음 조합이 발견되었습니다. 다음 위치에 모음 또는 준모음(음절자음)을 추가해주세요. - `{breakdown}`", // Korean (한국어)
	"nl": "**{oldWord}** 🇳🇱 Needs a vowel, diphthong or psuedovowel here: `{breakdown}`",            // Dutch (Nederlands)
	"pl": "**{oldWord}** 🇵🇱 Needs a vowel, diphthong or psuedovowel here: `{breakdown}`",            // Polish (Polski)
	"pt": "**{oldWord}** 🇵🇹 Needs a vowel, diphthong or psuedovowel here: `{breakdown}`",            // Portuguese (Português)
	"ru": "**{oldWord}** 🇷🇺 Needs a vowel, diphthong or psuedovowel here: `{breakdown}`",            // Russian (Русский)
	"sv": "**{oldWord}** 🇸🇪 Needs a vowel, diphthong or psuedovowel here: `{breakdown}`",            // Swedish (Svenska)
	"tr": "**{oldWord}** 🇹🇷 Needs a vowel, diphthong or psuedovowel here: `{breakdown}`",            // Turkish (Türkçe)
	"uk": "**{oldWord}** 🇺🇦 Needs a vowel, diphthong or psuedovowel here: `{breakdown}`",            // Ukrainian (Українська)
}

var message_psuedovowels_cant_coda = map[string]string{
	"en": "**{oldWord}** Psuedovowels can't accept codas: `{breakdown}`",                         // English
	"de": "**{oldWord}** 🇩🇪 Psuedovowels can't accept codas: `{breakdown}`",                      // German (Deutsch)
	"es": "**{oldWord}** 🇪🇦 Psuedovowels can't accept codas: `{breakdown}`",                      // Spanish (Español)
	"et": "**{oldWord}** 🇪🇪 Psuedovowels can't accept codas: `{breakdown}`",                      // Estonian (Eesti)
	"fr": "**{oldWord}** 🇫🇷 Psuedovowels can't accept codas: `{breakdown}`",                      // French (Français)
	"hu": "**{oldWord}** 🇭🇺 Psuedovowels can't accept codas: `{breakdown}`",                      // Hungarian (Magyar)
	"ko": "**{oldWord}**에 유효하지 않은 자음 조합이 발견되었습니다. 준모음(음절자음)은 말음(종성)을 가질 수 없습니다. - `{breakdown}`", // Korean (한국어)
	"nl": "**{oldWord}** 🇳🇱 Psuedovowels can't accept codas: `{breakdown}`",                      // Dutch (Nederlands)
	"pl": "**{oldWord}** 🇵🇱 Psuedovowels can't accept codas: `{breakdown}`",                      // Polish (Polski)
	"pt": "**{oldWord}** 🇵🇹 Psuedovowels can't accept codas: `{breakdown}`",                      // Portuguese (Português)
	"ru": "**{oldWord}** 🇷🇺 Psuedovowels can't accept codas: `{breakdown}`",                      // Russian (Русский)
	"sv": "**{oldWord}** 🇸🇪 Psuedovowels can't accept codas: `{breakdown}`",                      // Swedish (Svenska)
	"tr": "**{oldWord}** 🇹🇷 Psuedovowels can't accept codas: `{breakdown}`",                      // Turkish (Türkçe)
	"uk": "**{oldWord}** 🇺🇦 Psuedovowels can't accept codas: `{breakdown}`",                      // Ukrainian (Українська)
}

var message_psuedovowels_must_onset = map[string]string{
	"en": "**{oldWord}** Psuedovowels must have onsets: `{breakdown}`",                           // English
	"de": "**{oldWord}** 🇩🇪 Psuedovowels must have onsets: `{breakdown}`",                        // German (Deutsch)
	"es": "**{oldWord}** 🇪🇦 Psuedovowels must have onsets: `{breakdown}`",                        // Spanish (Español)
	"et": "**{oldWord}** 🇪🇪 Psuedovowels must have onsets: `{breakdown}`",                        // Estonian (Eesti)
	"fr": "**{oldWord}** 🇫🇷 Psuedovowels must have onsets: `{breakdown}`",                        // French (Français)
	"hu": "**{oldWord}** 🇭🇺 Psuedovowels must have onsets: `{breakdown}`",                        // Hungarian (Magyar)
	"ko": "**{oldWord}**에 유효하지 않은 자음 조합이 발견되었습니다. 준모음(음절자음)은 반드시 두음(초성)이 필요합니다. - `{breakdown}`", // Korean (한국어)
	"nl": "**{oldWord}** 🇳🇱 Psuedovowels must have onsets: `{breakdown}`",                        // Dutch (Nederlands)
	"pl": "**{oldWord}** 🇵🇱 Psuedovowels must have onsets: `{breakdown}`",                        // Polish (Polski)
	"pt": "**{oldWord}** 🇵🇹 Psuedovowels must have onsets: `{breakdown}`",                        // Portuguese (Português)
	"ru": "**{oldWord}** 🇷🇺 Psuedovowels must have onsets: `{breakdown}`",                        // Russian (Русский)
	"sv": "**{oldWord}** 🇸🇪 Psuedovowels must have onsets: `{breakdown}`",                        // Swedish (Svenska)
	"tr": "**{oldWord}** 🇹🇷 Psuedovowels must have onsets: `{breakdown}`",                        // Turkish (Türkçe)
	"uk": "**{oldWord}** 🇺🇦 Psuedovowels must have onsets: `{breakdown}`",                        // Ukrainian (Українська)
}

var message_triple_liquid = map[string]string{
	"en": "**{oldWord}** Triple Rs or Ls aren't allowed: `{breakdown}`",    // English
	"de": "**{oldWord}** 🇩🇪 Triple Rs or Ls aren't allowed: `{breakdown}`", // German (Deutsch)
	"es": "**{oldWord}** 🇪🇦 Triple Rs or Ls aren't allowed: `{breakdown}`", // Spanish (Español)
	"et": "**{oldWord}** 🇪🇪 Triple Rs or Ls aren't allowed: `{breakdown}`", // Estonian (Eesti)
	"fr": "**{oldWord}** 🇫🇷 Triple Rs or Ls aren't allowed: `{breakdown}`", // French (Français)
	"hu": "**{oldWord}** 🇭🇺 Triple Rs or Ls aren't allowed: `{breakdown}`", // Hungarian (Magyar)
	"ko": "**{oldWord}** 연속되는 세개의 R 또는 L은 사용 불가능합니다. - `{breakdown}`",      // Korean (한국어)
	"nl": "**{oldWord}** 🇳🇱 Triple Rs or Ls aren't allowed: `{breakdown}`", // Dutch (Nederlands)
	"pl": "**{oldWord}** 🇵🇱 Triple Rs or Ls aren't allowed: `{breakdown}`", // Polish (Polski)
	"pt": "**{oldWord}** 🇵🇹 Triple Rs or Ls aren't allowed: `{breakdown}`", // Portuguese (Português)
	"ru": "**{oldWord}** 🇷🇺 Triple Rs or Ls aren't allowed: `{breakdown}`", // Russian (Русский)
	"sv": "**{oldWord}** 🇸🇪 Triple Rs or Ls aren't allowed: `{breakdown}`", // Swedish (Svenska)
	"tr": "**{oldWord}** 🇹🇷 Triple Rs or Ls aren't allowed: `{breakdown}`", // Turkish (Türkçe)
	"uk": "**{oldWord}** 🇺🇦 Triple Rs or Ls aren't allowed: `{breakdown}`", // Ukrainian (Українська)
}

var message_reef_dialect = map[string]string{
	"en": " (In reef dialect.  Forest dialect {breakdown})", // English
	"de": " (In reef dialect.  Forest dialect {breakdown})", // German (Deutsch)
	"es": " (In reef dialect.  Forest dialect {breakdown})", // Spanish (Español)
	"et": " (In reef dialect.  Forest dialect {breakdown})", // Estonian (Eesti)
	"fr": " (In reef dialect.  Forest dialect {breakdown})", // French (Français)
	"hu": " (In reef dialect.  Forest dialect {breakdown})", // Hungarian (Magyar)
	"ko": " (산호초 방언 한정 - 숲 방언: {breakdown})",                // Korean (한국어)
	"nl": " (In reef dialect.  Forest dialect {breakdown})", // Dutch (Nederlands)
	"pl": " (In reef dialect.  Forest dialect {breakdown})", // Polish (Polski)
	"pt": " (In reef dialect.  Forest dialect {breakdown})", // Portuguese (Português)
	"ru": " (In reef dialect.  Forest dialect {breakdown})", // Russian (Русский)
	"sv": " (In reef dialect.  Forest dialect {breakdown})", // Swedish (Svenska)
	"tr": " (In reef dialect.  Forest dialect {breakdown})", // Turkish (Türkçe)
	"uk": " (In reef dialect.  Forest dialect {breakdown})", // Ukrainian (Українська)
}

var message_warning = map[string]string{
	"en": "**{oldWord}** Warning: `{breakdown}` `{boundary}` works in a productive compound, but not as a name or root word",    // English
	"de": "**{oldWord}** 🇩🇪 Warning: `{breakdown}` `{boundary}` works in a productive compound, but not as a name or root word", // German (Deutsch)
	"es": "**{oldWord}** 🇪🇦 Warning: `{breakdown}` `{boundary}` works in a productive compound, but not as a name or root word", // Spanish (Español)
	"et": "**{oldWord}** 🇪🇪 Warning: `{breakdown}` `{boundary}` works in a productive compound, but not as a name or root word", // Estonian (Eesti)
	"fr": "**{oldWord}** 🇫🇷 Warning: `{breakdown}` `{boundary}` works in a productive compound, but not as a name or root word", // French (Français)
	"hu": "**{oldWord}** 🇭🇺 Warning: `{breakdown}` `{boundary}` works in a productive compound, but not as a name or root word", // Hungarian (Magyar)
	"ko": "**{oldWord}** 🇰🇷 Warning: `{breakdown}` `{boundary}` works in a productive compound, but not as a name or root word",      // Korean (한국어)
	"nl": "**{oldWord}** 🇳🇱 Warning: `{breakdown}` `{boundary}` works in a productive compound, but not as a name or root word", // Dutch (Nederlands)
	"pl": "**{oldWord}** 🇵🇱 Warning: `{breakdown}` `{boundary}` works in a productive compound, but not as a name or root word", // Polish (Polski)
	"pt": "**{oldWord}** 🇵🇹 Warning: `{breakdown}` `{boundary}` works in a productive compound, but not as a name or root word", // Portuguese (Português)
	"ru": "**{oldWord}** 🇷🇺 Warning: `{breakdown}` `{boundary}` works in a productive compound, but not as a name or root word", // Russian (Русский)
	"sv": "**{oldWord}** 🇸🇪 Warning: `{breakdown}` `{boundary}` works in a productive compound, but not as a name or root word", // Swedish (Svenska)
	"tr": "**{oldWord}** 🇹🇷 Warning: `{breakdown}` `{boundary}` works in a productive compound, but not as a name or root word", // Turkish (Türkçe)
	"uk": "**{oldWord}** 🇺🇦 Warning: `{breakdown}` `{boundary}` works in a productive compound, but not as a name or root word", // Ukrainian (Українська)
}

var message_valid = map[string]string{
	"en": "**{oldWord}** Valid: `{breakdown}` with {syllable_count} syllables {syllable_forest}",    // English
	"de": "**{oldWord}** 🇩🇪 Valid: `{breakdown}` with {syllable_count} syllables {syllable_forest}", // German (Deutsch)
	"es": "**{oldWord}** 🇪🇦 Valid: `{breakdown}` with {syllable_count} syllables {syllable_forest}", // Spanish (Español)
	"et": "**{oldWord}** 🇪🇪 Valid: `{breakdown}` with {syllable_count} syllables {syllable_forest}", // Estonian (Eesti)
	"fr": "**{oldWord}** 🇫🇷 Valid: `{breakdown}` with {syllable_count} syllables {syllable_forest}", // French (Français)
	"hu": "**{oldWord}** 🇭🇺 Valid: `{breakdown}` with {syllable_count} syllables {syllable_forest}", // Hungarian (Magyar)
	"ko": "**{oldWord}**는 `{breakdown}`의 {syllable_count}음절로 구성된 유효한 단어입니다. {syllable_forest}",      // Korean (한국어)
	"nl": "**{oldWord}** 🇳🇱 Valid: `{breakdown}` with {syllable_count} syllables {syllable_forest}", // Dutch (Nederlands)
	"pl": "**{oldWord}** 🇵🇱 Valid: `{breakdown}` with {syllable_count} syllables {syllable_forest}", // Polish (Polski)
	"pt": "**{oldWord}** 🇵🇹 Valid: `{breakdown}` with {syllable_count} syllables {syllable_forest}", // Portuguese (Português)
	"ru": "**{oldWord}** 🇷🇺 Valid: `{breakdown}` with {syllable_count} syllables {syllable_forest}", // Russian (Русский)
	"sv": "**{oldWord}** 🇸🇪 Valid: `{breakdown}` with {syllable_count} syllables {syllable_forest}", // Swedish (Svenska)
	"tr": "**{oldWord}** 🇹🇷 Valid: `{breakdown}` with {syllable_count} syllables {syllable_forest}", // Turkish (Türkçe)
	"uk": "**{oldWord}** 🇺🇦 Valid: `{breakdown}` with {syllable_count} syllables {syllable_forest}", // Ukrainian (Українська)
}

func valid_message(syllable_count int, lang string) string {
	if lang == "en" {
		if syllable_count == 1 {
			message := strings.ReplaceAll(message_valid[lang], "syllables", "syllable")
			message = strings.ReplaceAll(message, "{syllable_count}", strconv.Itoa(syllable_count))
			return message
		}
	}
	return strings.ReplaceAll(message_valid[lang], "{syllable_count}", strconv.Itoa(syllable_count))
}

var message_too_big = map[string]string{
	"en": "⛔ (stopped at {count}. 2000 Character limit)",    // English
	"de": "⛔ (stopped at {count}. 2000 Character limit) 🇩🇪", // German (Deutsch)
	"es": "⛔ (stopped at {count}. 2000 Character limit) 🇪🇦", // Spanish (Español)
	"et": "⛔ (stopped at {count}. 2000 Character limit) 🇪🇪", // Estonian (Eesti)
	"fr": "⛔ (stopped at {count}. 2000 Character limit) 🇫🇷", // French (Français)
	"hu": "⛔ (stopped at {count}. 2000 Character limit) 🇭🇺", // Hungarian (Magyar)
	"ko": "⛔ (stopped at {count}. 2000 Character limit) 🇰🇷", // Korean (한국어)
	"nl": "⛔ (stopped at {count}. 2000 Character limit) 🇳🇱", // Dutch (Nederlands)
	"pl": "⛔ (stopped at {count}. 2000 Character limit) 🇵🇱", // Polish (Polski)
	"pt": "⛔ (stopped at {count}. 2000 Character limit) 🇵🇹", // Portuguese (Português)
	"ru": "⛔ (stopped at {count}. 2000 Character limit) 🇷🇺", // Russian (Русский)
	"sv": "⛔ (stopped at {count}. 2000 Character limit) 🇸🇪", // Swedish (Svenska)
	"tr": "⛔ (stopped at {count}. 2000 Character limit) 🇹🇷", // Turkish (Türkçe)
	"uk": "⛔ (stopped at {count}. 2000 Character limit) 🇺🇦", // Ukrainian (Українська)
}
