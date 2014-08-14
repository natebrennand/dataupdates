package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/coopernurse/gorp"
	_ "github.com/lib/pq"
)

const (
	courses_table   = "courses_t"
	courses_table_2 = "courses_v2_t"
)

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
		getEnvVar("PG_DB"),
		getEnvVar("PG_PASSWORD"),
		getEnvVar("PG_HOST"),
		getEnvVar("PG_PORT"),
	))
	if err != nil {
		log.Fatalf("failed to connect to Postgres => %s", err.Error())
	}

	gorpdb := &gorp.DbMap{Db: db, Dialect: gorp.PostgresDialect{}}

	gorpdb.AddTableWithName(Course{}, "courses_t").SetKeys(false, "Course", "Term")
	gorpdb.AddTableWithName(Course2{}, "courses_v2_t").SetKeys(false, "CourseFull")
	gorpdb.AddTableWithName(Section{}, "sections_v2_t").SetKeys(false, "SectionFull", "Term")

	if gorpdb.CreateTablesIfNotExists() != nil {
		log.Fatalf("Error creating databases => %s", err.Error())
	}
	return gorpdb
}

func dbWorker(db *gorp.DbMap, readyCourse chan Course, wg *sync.WaitGroup, descCache map[string]string) {
	var (
		c         Course
		more      bool
		v2Course  Course2
		v2Section Section
	)

	for {
		c, more = <-readyCourse
		courseFull, err := c.getCourseFull()
		if err != nil {
			log.Printf("failed to insert course,  %s", c.Course)
			continue
		}

		cachedDesc, exists := descCache[courseFull]
		if exists {
			c.Description = cachedDesc
		} else {
			// now we must get the description
			if err := c.getDescription(); err != nil {
				log.Printf("Could not get description for %s, %s", c.Course, err.Error())
				continue
			}
			descCache[courseFull] = c.Description
			v2Course, _ = c.split()
			if err := db.Insert(&v2Course); err != nil {
				log.Printf("Failed to insert course_v2, %s, err => %s", v2Course.CourseFull, err.Error())
			}
		}

		_, v2Section = c.split()
		if err := db.Insert(&v2Section); err != nil {
			log.Printf("Failed to insert section, %s, err => %s", v2Section.SectionFull, err.Error())
		}

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
	if len(os.Args) != 2 {
		panic("must pass path to JSON file as an argument")
	}
	jsonFile := os.Args[1]

	// open database connection
	var wg sync.WaitGroup
	db := connectPG()
	defer db.Db.Close()

	// parse the json file of courses
	courseChan := make(chan Course)
	go parseCourses(jsonFile, courseChan, &wg)

	// readyCourse receives courses that are ready for database insertion
	readyCourse := make(chan Course, 50)

	// db worker reads from readyCourse and inserts to the database
	wg.Add(1)
	descCache := make(map[string]string) //  CourseFull --> description
	for _ = range make([]interface{}, 1) {
		go dbWorker(db, readyCourse, &wg, descCache)
	}

	// process courses as they come from the parser
	var c Course
	var more bool
	for {
		c, more = <-courseChan
		// uncomment to get description and insert to DB
		// TODO: consider ways to send courses in sequential groups
		readyCourse <- c
		if !more {
			break
		}
		log.Printf("prepped, %s", c.Course)
	}
	wg.Wait()
}
