package main

import (
	"log"

	"github.com/coopernurse/gorp"
)

type esData struct {
	Course         string   `db:"course"`     // EX: ZULU336
	CourseFull     string   `db:"coursefull"` // EX: ZULUW336
	CourseSubtitle string   `db:"coursesubtitle"`
	CourseTitle    string   `db:"coursetitle"`
	Description    string   `db:"description"`
	Term           string   `db:"term"`
	CallNumber     []int    `db:"callnumber"`
	Instructor     []string `db:"instructor"`
}

var query = `
SELECT
	C.course,
	C.coursefull,
	C.coursesubtitle,
	C.coursetitle,
	C.description,
	array_agg(DISTINCT S.term) as "term",
	array_agg(DISTINCT S.callnumber) as "callnumber",
	array_agg(DISTINCT S.instructor1name) as "instructor"
 FROM courses_v2_t C JOIN sections_v2_t S
 ON C.course = S.course
 GROUP BY
	C.course,
	C.coursefull,
	C.coursesubtitle,
	C.coursetitle,
	C.description;
`

func getESData(db *gorp.DbMap) []esData {
	var esDataList []esData
	_, err := db.Select(&esDataList, query)
	if err != nil {
		log.Printf("Error while querying for ES data => %s", err.Error())
	}
	return esDataList
}
