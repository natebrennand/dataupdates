package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/coopernurse/gorp"
	_ "github.com/lib/pq"
	"io"
	"log"
	"os"
	"sync"
)

const (
	jsonFile        = "./doc.json"
	courses_table   = "courses_t"
	courses_table_2 = "courses_v2_t"
)

var done = make(chan bool)

func getEnvVar(name string) string {
	val := os.Getenv(name)
	if val == "" {
		log.Fatal(fmt.Sprintf("%s must be set"), name)
	}
	return val
}

func connectPG() *gorp.DbMap {
	db, err := sql.Open("postgres", fmt.Sprintf("user=%s dbname=%s password=%s host=%s port=%s sslmode=disable",
		getEnvVar("PG_USER"),
		getEnvVar("PG_DBNAME"),
		getEnvVar("PG_PASS"),
		getEnvVar("PG_HOST"),
		getEnvVar("PG_PORT"),
	))
	if err != nil {
		log.Fatalf("failed to connect to Postgres => %s", err.Error())
	}

	gorpdb := &gorp.DbMap{Db: db, Dialect: gorp.PostgresDialect{}}
	gorpdb.AddTableWithName(Course{}, "courses_t").SetKeys(false, "Course", "Term")

	if gorpdb.CreateTablesIfNotExists() != nil {
		log.Fatalf("Error creating databases => %s", err.Error())
	}
	return gorpdb
}

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

func parseCourses(cChan chan Course, wg *sync.WaitGroup) {
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
			wg.Done()
			close(cChan)
			return
		} else if err != nil {
			panic(err)
		}
		log.Print("sending down chan")
		cChan <- c
		log.Print("sent down chan")

		r = io.MultiReader(dec.Buffered(), r)
		if b, err := readByteSkippingSpace(r); err != nil {
			log.Printf("broken, hit %s, err => %s", b, err.Error())
			panic(err)
		} else {
			switch b {
			case ',':
				log.Printf("after courses, %s, hit comma", c.Course)
				continue
			case ']': // end
				log.Print("done reading")
				return
			default:
				panic("Invalid character in JSON data: " + string([]byte{b}))
			}
		}
	}
}

func dbWorker(db *gorp.DbMap, readyCourse chan Course, wg *sync.WaitGroup) {
	var c Course
	var more bool
	for {
		c, more = <-readyCourse
		if err := db.Insert(&c); err != nil {
			log.Printf("While inserting course => %#v\n, database error => %s", c, err.Error())
		}
		log.Printf("Saving %s to the db\n", c.Course)
		if !more {
			wg.Done()
		}
	}
}

func main() {
	var wg sync.WaitGroup
	db := connectPG()
	defer db.Db.Close()

	courseChan := make(chan Course)
	go parseCourses(courseChan, &wg)

	var c Course
	var more bool

	newCourse := make(chan Course, 100)
	readyCourse := make(chan Course)

	// workers downloads descriptions from the bulletin
	for i := 0; i < 10; i++ {
		go bulletinWorker(newCourse, readyCourse)
	}

	wg.Add(1)
	// db worker reads from readyCourse and inserts to the database
	go dbWorker(db, readyCourse, &wg)

	for {
		c, more = <-courseChan
		// uncomment to get description and insert to DB
		// newCourse <- c
		if !more {
			break
		}
		log.Printf("prepped, %s", c.Course)
	}
	wg.Wait()
}
