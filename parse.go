package main

import (
	"encoding/json"
	"io"
	"log"
	"os"
	"sync"
)

func readByteSkippingSpace(r io.Reader) (b byte, err error) {
	buf := make([]byte, 1)
	for {
		_, err := r.Read(buf)
		if err != nil {
			return 0, err
		}
		b := buf[0]
		switch b {
		// Only handling ASCII white space for now
		case ' ', '\t', '\n', '\v', '\f', '\r':
			continue
		default:
			return b, nil
		}
	}
}

func parseCourses(jsonFile string, cChan chan Course, wg *sync.WaitGroup) {
	// open file for parsing
	file, err := os.OpenFile(jsonFile, os.O_RDONLY, 0644)
	if err != nil {
		log.Fatalf("Failed to open file, %s, with error: %s", jsonFile, err.Error())
	}
	//defer file.Close()
	r := io.Reader(file)

	// Skip whitespace & '['
	if b, err := readByteSkippingSpace(r); err != nil {
		panic(err)
	} else if b != '[' {
		panic("Input is not a JSON array")
	}

	var c Course
	for {
		dec := json.NewDecoder(r)
		if err := dec.Decode(&c); err == io.EOF {
			log.Print("finished parsing json file")
			return
		} else if err != nil {
			panic(err)
		}
		c.fill()
		cChan <- c

		r = io.MultiReader(dec.Buffered(), r)
		if b, err := readByteSkippingSpace(r); err != nil {
			log.Printf("broken, hit %s, err => %s", b, err.Error())
			panic(err)
		} else {
			switch b {
			case ',':
				continue
			case ']':
				log.Print("done reading")
				close(cChan)
				wg.Done()
				return
			default:
				panic("Invalid character in JSON data: " + string([]byte{b}))
			}
		}
	}
}
