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

	/*start := time.Now()

	pos_map := map[string]uint8{}
	if _, err := os.Stat("dictionary-v2.txt"); err == nil {
		// path/to/whatever exists
		b, err2 := os.Open("dictionary-v2.txt")
		if err2 != nil {
			fmt.Println("error opening file:", err2)
			return
		}

		scanner := bufio.NewScanner(b)
		// This will not read lines over 64k long, but works for Na'vi words just fine
		for scanner.Scan() {
			everything_slice := strings.Split(scanner.Text(), "\t")

			if _, ok := pos_map[everything_slice[4]]; !ok {
				pos_map[everything_slice[4]] = uint8(0)
			}
		}

		if err := scanner.Err(); err != nil {
			log.Fatal(err)
		}

		b.Close()
	}

	all_keys := []string{}
	for key := range pos_map {
		all_keys = append(all_keys, key)
	}
	fmt.Println(len(all_keys))
	sort.Strings(all_keys)
	fmt.Println(len(all_keys))

	counter := uint8(0)

	all_maps := []map[string]uint8{}

	if _, err := os.Stat("dictionary-v2.txt"); err == nil {
		// path/to/whatever exists
		b, err2 := os.Open("dictionary-v2.txt")
		if err2 != nil {
			fmt.Println("error opening file:", err2)
			return
		}

		scanner := bufio.NewScanner(b)
		// This will not read lines over 64k long, but works for Na'vi words just fine
		for scanner.Scan() {
			everything_slice := strings.Split(scanner.Text(), "\t")

			if _, ok := pos_map[everything_slice[4]]; ok {
				pos_map[everything_slice[4]] = pos_map[everything_slice[4]] + 1
			}

			counter += 1

			if counter == 100 {
				counter = 0

				all_maps = append(all_maps, pos_map)

				pos_map = map[string]uint8{}

				for _, key := range all_keys {
					pos_map[key] = uint8(0)
				}
			}
		}

		if err := scanner.Err(); err != nil {
			log.Fatal(err)
		}

		// Print the maps as a CSV
		var buffer strings.Builder

		for _, key := range all_keys {
			buffer.WriteString("\"" + key + "\"")
			buffer.WriteString("\t")
		}

		buffer.WriteString("\n")
		for _, my_map := range all_maps {
			var buffer_2 strings.Builder
			for _, key := range all_keys {
				buffer_2.WriteString(key)
				buffer_2.WriteString("\t")
				buffer.WriteString(strconv.Itoa(int(my_map[key])))
				buffer.WriteString("\t")
			}
			buffer.WriteString("\n")
		}

		fmt.Println(buffer.String())

	}

	total := time.Since(start)
	fmt.Println(total.Seconds())*/
}
