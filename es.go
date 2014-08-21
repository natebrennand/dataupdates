package main

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/natebrennand/pg_array"
)

type esData struct {
	Course         string                  `db:"course"`     // EX: ZULU336
	CourseFull     string                  `db:"coursefull"` // EX: ZULUW336
	CourseSubtitle string                  `db:"coursesubtitle"`
	CourseTitle    string                  `db:"coursetitle"`
	Description    string                  `db:"description"`
	Term           pg_array.SqlIntArray    `db:"term"`
	CallNumber     pg_array.SqlIntArray    `db:"callnumber"`
	Instructor     pg_array.SqlStringArray `db:"instructor"`
}

var esQuery = `
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

func getESData(db *sql.DB) []esData {
	var esDataList []esData
	rows, err := db.Query(esQuery)
	if err != nil {
		log.Fatalf("Error while querying Postgres for ES data => %s", err.Error())
	}
	defer rows.Close()

	// process records
	var data esData
	for rows.Next() {
		err := rows.Scan(
			&data.Course,
			&data.CourseFull,
			&data.CourseSubtitle,
			&data.Description,
			&data.CourseFull,
			&data.Term,
			&data.CallNumber,
			&data.Instructor,
		)
		if err != nil {
			log.Fatalf("Error while processing PG data => %s", err.Error())
		}
		fmt.Println(data)
	}

	return esDataList
}
