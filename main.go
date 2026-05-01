package main

func main() {
	StartEverything()
	/*var everything_buffer strings.Builder
	if _, err := os.Stat("sorted-anagrams-forumat-v0.0.2.txt"); err == nil {
		// path/to/whatever exists
		b, err2 := os.Open("sorted-anagrams-forumat-v0.0.2.txt")
		if err2 != nil {
			fmt.Println("error opening file:", err2)
			return
		}

		scanner := bufio.NewScanner(b)
		// This will not read lines over 64k long, but works for Na'vi words just fine
		for scanner.Scan() {
			everything_buffer.WriteString(scanner.Text() + "\n")
		}

		if err := scanner.Err(); err != nil {
			log.Fatal(err)
		}

		everything_slice := strings.Split(everything_buffer.String(), "[/tr]\n[tr]")

		sort.Slice(everything_slice, func(i, j int) bool {
			return AlphabetizeHelper(strings.ReplaceAll(everything_slice[i], "*", ""), strings.ReplaceAll(everything_slice[j], "*", ""))
		})

		a, err := os.Create("sorted-anagrams-forumat-part2.txt")
		if err != nil {
			fmt.Println("error opening file:", err)
			return
		}

		for _, word := range everything_slice {
			a.WriteString("[tr]" + word + "[/tr]")
			fmt.Println(word)
		}
	}*/
}
