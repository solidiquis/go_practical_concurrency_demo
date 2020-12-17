/*
Practical Use of Concurrency: For educational purposes.

This program takes three different sized text files - each
with lines of varying lengths - reads their contents into
memory, formats them so that each line of text is no more
than 70 chars in length, then writes them to a new file
in the tmp directory.

The files (located in the assets directory):
- at_the_mountains_of_madness.txt: 249kb
- the_call_of_cthulhu.txt: 70kb
- the_shadow_over_innsmouth.txt: 155kb

The traditional approach:
- Read text file, format data, and write to new file one by one.

Pros to the traditional approach:
- Synchronous programming is generally an easier mental model
  which makes writing the program easier for most.
- Generally safer as you won't run into data race issues, that is
  you'll never have two or more independent processes try to access
  the same piece of data when at least one of the access is a write.

Cons to the traditional approach:
- Because things are being done one-by-one, larger files
  can cause bottlenecks. Imagine if a chef decided to cook your main
  dish first, THEN cooked your appetizer, because for some reason the
  waiter listed the main as the first item in the list; now your appetizer
  is going to awkwardly come out after you main dish.

Concurrency approach:
- Spin up a goroutine for each file to handle the reading and
  formatting concurrently without waiting for the other to finish.
  The goroutines will send their data through a channel to the writer
  in the order which they finish.

Pros to the concurrency approach:
- No bottlenecks as each file is being worked on in individual goroutines,
  which means that, in the context of this program, the smallest file will
  finish first, even if it's the last to be placed in a goroutine. The chef
  basically now has an assistant who works on the appetizer while the chef
  works concurrently on the main (which takes longer), meaning you'll get
  your appetizer first.
- Overall performance.

Cons to the concurrency approach:
- Unfamiliar mental model for programmers, generally, which can make it
  challenging to implement.
- You may unknowingly create data races.
- Goroutine leaks can happen, in which a goroutine that you thought terminated,
  lives for the lifetime of the programm which holds memory hostage; the
  garbage collector also won't clean up goroutines, making this memory
  leak especially dangerous.

The following code demonstrates the concurrency approach:
*/

package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"regexp"
)

func must(e error) {
	if e != nil {
		panic(e)
	}
}

func formatText(filename string, hplCh chan<- []byte) {
	filePath := fmt.Sprintf("assets/%s", filename)
	data, err := ioutil.ReadFile(filePath)
	must(err)

	// Remove blank lines
	re, err := regexp.Compile(`\n\n`)
	must(err)
	data = re.ReplaceAll(data, []byte("\n"))

	// Remove tabs and/or spaces
	re, err = regexp.Compile(`^\s+`)
	must(err)
	data = re.ReplaceAll(data, []byte(""))

	re, err = regexp.Compile(`\n\s+`)
	must(err)
	data = re.ReplaceAll(data, []byte("\n"))

	re, err = regexp.Compile(`\s*\n`)
	must(err)
	data = re.ReplaceAll(data, []byte(" "))

	// Make every line length 70
	intermediate := make([][]byte, 1+len(data)/70)
	i, j := 0, 70
	for k := 0; k < len(intermediate); k++ {
		intermediate[k] = data[i:j]
		i += 70
		j += 70
		if j >= len(data) {
			break
		}
	}
	intermediate[len(intermediate)-1] = data[i:len(data)]

	// Join the byte slices
	result := bytes.Join(intermediate, []byte("\n"))

	// Send the byte slice through channel
	hplCh <- result
}

func writeFormattedText(data []byte) {
	fmt.Printf("Writing file of length: %v\n", len(data))
	file := fmt.Sprintf("tmp/%d.txt", len(data))
	err := ioutil.WriteFile(file, data, 0644)
	must(err)
}

func main() {
	// Names of txt files in assets dir
	eldritchTexts := []string{
		"at_the_mountains_of_madness.txt", // longest
		"the_shadow_over_innsmouth.txt",   // median
		"the_call_of_cthulhu.txt",         // shortest
	}

	// Create buffered channel instance
	hplCh := make(chan []byte, len(eldritchTexts))

	// Format each text file in its own goroutine
	for _, filename := range eldritchTexts {
		go formatText(filename, hplCh)
	}

	// Byte slices written to channel in the order in which
	// they are received from the goroutines.
	for i := 0; i < len(eldritchTexts); i++ {
		writeFormattedText(<-hplCh)
	}
}
