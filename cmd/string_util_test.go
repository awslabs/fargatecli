package cmd

import "fmt"

func ExampleHumanize() {
	fmt.Println(Humanize("HELLO_COMPUTER"))
	// Output: Hello Computer
}

func ExampleMap() {
	reverse := func(s string) string {
		r := []rune(s)
		for i, j := 0, len(r)-1; i < len(r)/2; i, j = i+1, j-1 {
			r[i], r[j] = r[j], r[i]
		}
		return string(r)
	}

	fmt.Printf("%v", Map([]string{"Pippin", "Merry"}, reverse))
	// Output: [nippiP yrreM]
}
