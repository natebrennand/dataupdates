package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"sync"

	_ "github.com/lib/pq" // register the postgres driver w/ sql
)

const (
	// MaxHTTPRequests dicatates the number of open HTTP requests to the bulletin
	MaxHTTPRequests = 10
)

func getEnvVar(name string) string {
	val := os.Getenv(name)
	if val == "" {
		log.Fatal(fmt.Sprintf("%s must be set", name))
	}
	return val
}

func connectPG() *sql.DB {
	db, err := sql.Open("postgres", fmt.Sprintf("user=%s dbname=%s password=%s host=%s port=%s sslmode=disable",
		getEnvVar("PG_USER"),
		getEnvVar("PG_DB"),
		getEnvVar("PG_PASSWORD"),
		getEnvVar("PG_HOST"),
		getEnvVar("PG_PORT"),
	))
	if err != nil {
		log.Fatalf("failed to connect to Postgres => %s", err.Error())
	}

	return db
}

func main() {
	if len(os.Args) != 2 {
		panic("must pass path to JSON file as an argument")
	}
	jsonFile := os.Args[1]
	var wg sync.WaitGroup

	// open database connection
	db := connectPG()
	defer db.Close()

	// parse the json file of Courses
	courseChan := make(chan Course)
	go parseCourses(jsonFile, courseChan, &wg)

	// db worker reads from dbQueue and inserts to the database
	wg.Add(1)
	dbQueue := make(chan Course, 50)
	descCache := make(map[string]string) //  CourseFull --> description
	go dbWorker(db, dbQueue, &wg, descCache)

	// process courses as they come from the parser
	var c Course
	var more bool
	for {
		// reads in a course that is ready for insertion
		c, more = <-courseChan
		// sends course to be inserted to the database
		dbQueue <- c
		if !more {
			break
		}
	}
	wg.Wait()

	getESData(db)
}
