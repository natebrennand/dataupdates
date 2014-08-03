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
)

const (
	jsonFile        = "./doc2.json"
	courses_table   = "courses_t"
	courses_table_2 = "courses_v2_t"
)

type Course struct {
	Term             string `json:",omitempty",db:"term"`
	Course           string `json:",omitempty",db:"course"`
	PrefixName       string `json:",omitempty",db:"prefixname"`
	DivisionCode     string `json:",omitempty",db:"divisioncode"`
	DivisionName     string `json:",omitempty",db:"divisionname"`
	CampusCode       string `json:",omitempty",db:"campuscode"`
	CampusName       string `json:",omitempty",db:"campusname"`
	SchoolCode       string `json:",omitempty",db:"schoolcode"`
	SchoolName       string `json:",omitempty",db:"schoolname"`
	DepartmentCode   string `json:",omitempty",db:"departmentcode"`
	DepartmentName   string `json:",omitempty",db:"departmentname"`
	SubtermCode      string `json:",omitempty",db:"subtermcode"`
	SubtermName      string `json:",omitempty",db:"subtermname"`
	CallNumber       string `json:",omitempty,int",db:"callnumber"`
	NumEnrolled      string `json:",omitempty,int",db:"numenrolled"`
	MaxSize          string `json:",omitempty,int",db:"maxsize"`
	EnrollmentStatus string `json:",omitempty",db:"enrollmentstatus"`
	NumFixedUnits    string `json:",omitempty,int",db:"numfixedunits"`
	MinUnits         string `json:",omitempty,int",db:"minunits"`
	MaxUnits         string `json:",omitempty,int",db:"maxunits"`
	CourseTitle      string `json:",omitempty",db:"coursetitle"`
	CourseSubtitle   string `json:",omitempty",db:"coursesubtitle"`
	TypeCode         string `json:",omitempty",db:"typecode"`
	TypeName         string `json:",omitempty",db:"typename"`
	Approval         string `json:",omitempty",db:"approval"`
	BulletinFlags    string `json:",omitempty",db:"bulletinflags"`
	ClassNotes       string `json:",omitempty",db:"classnotes"`
	Meets1           string `json:",omitempty",db:"meets1"`
	Meets2           string `json:",omitempty",db:"meets2"`
	Meets3           string `json:",omitempty",db:"meets3"`
	Meets4           string `json:",omitempty",db:"meets4"`
	Meets5           string `json:",omitempty",db:"meets5"`
	Meets6           string `json:",omitempty",db:"meets6"`
	Instructor1Name  string `json:",omitempty",db:"instructor1name"`
	Instructor2Name  string `json:",omitempty",db:"instructor2name"`
	Instructor3Name  string `json:",omitempty",db:"instructor3name"`
	Instructor4Name  string `json:",omitempty",db:"instructor4name"`
	PrefixLongname   string `json:",omitempty",db:"prefixlongname"`
	ExamMeet         string `json:",omitempty",db:"exammeet"`
	ExamDate         string `json:",omitempty",db:"examdate"`
	ChargeMsg1       string `json:",omitempty",db:"chargemsg1"`
	ChargeAmt1       string `json:",omitempty",db:"chargeamt1"`
	ChargeMsg2       string `json:",omitempty",db:"chargemsg2"`
	ChargeAmt2       string `json:",omitempty",db:"chargeamt2"`
	Description      string `json:",omitempty",db:"description"`
}

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

	err = gorpdb.CreateTablesIfNotExists()
	if err != nil {
		log.Fatalf("Error creating databases => %s", err.Error())
	}
	return gorpdb
}

func parseCourseList(filename string) chan Course {
	courseChan := make(chan Course)

	go func() {
		file, err := os.OpenFile(filename, os.O_RDONLY, 0644)
		if err != nil {
			log.Fatalf("Failed to open file, %s, with error: %s", filename, err.Error())
		}

		dec := json.NewDecoder(file)
		for {
			var c Course
			if err := dec.Decode(&c); err == io.EOF {
				log.Printf("Finished reading json file, %s\n", filename)
				close(courseChan)
				break
			} else if err != nil {
				log.Fatalf("Error decoding => %s", err.Error())
			}
			courseChan <- c
		}
	}()
	return courseChan
}

func dbWorker(db *gorp.DbMap, readyCourse chan Course, done chan bool) {
	var c Course
	var more bool
	for {
		c, more = <-readyCourse
		err := db.Insert(&c)
		log.Printf("Saving %s to the db\n", c.Course)
		if err != nil {
			log.Printf("Inserting course => %#v\n", c)
			log.Fatalf("database error during insertion => %s", err.Error())
		}
		if !more {
			done <- true
		}
	}
}

func main() {
	db := connectPG()
	defer db.Db.Close()

	courseChan := parseCourseList(jsonFile)
	var c Course
	var more bool

	newCourse := make(chan Course, 100)
	readyCourse := make(chan Course)
	done := make(chan bool)

	for i := 0; i < 10; i++ {
		go bulletinWorker(newCourse, readyCourse)
	}
	go dbWorker(db, readyCourse, done)

	for {
		c, more = <-courseChan
		if !more {
			break
		}
		newCourse <- c
	}
	<-done
}
