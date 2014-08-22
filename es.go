package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/natebrennand/pg_array"
)

var (
	esURL     string
	esIndex   string
	esType    = "courses"
	batchSize = 200
)

func init() {
	esIndex = getEnvVar("ES_INDEX")
	esURL = fmt.Sprintf("http://%s:%s/",
		getEnvVar("ES_HOST"),
		getEnvVar("ES_PORT"),
	)
}

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

type esMetadata struct {
	Index string `json:"es_index"`
	Type  string `json:"_type"`
	ID    string `json:"_id"` // will be the 'Course' attribute
}

type bulkItem struct {
	MetaData esMetadata
	Data     esData
}

type bulkInsert []bulkItem

func (b bulkInsert) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer
	var encoder = json.NewEncoder(&buf)

	for _, item := range b {
		if err := encoder.Encode(item.MetaData); err != nil {
			return []byte{}, fmt.Errorf("Failed to encode MetaData => %s", err.Error())
		}
		if err := encoder.Encode(item.Data); err != nil {
			return []byte{}, fmt.Errorf("Failed to encode item data => %s", err.Error())
		}
	}
	return buf.Bytes(), nil
}

func (d esData) NewBulkItem() bulkItem {
	return bulkItem{
		MetaData: esMetadata{
			Index: esIndex,
			Type:  esType,
			ID:    d.Course,
		},
		Data: d,
	}
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

	var batchBuffer = make([]bulkItem, batchSize)
	var bufferIndex = 0

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

		batchBuffer[bufferIndex] = data.NewBulkItem()
		bufferIndex++
		if bufferIndex == batchSize {
			insertEsData(bulkInsert(batchBuffer))
			bufferIndex = 0
		}
	}
	// insert remainder of buffer
	insertEsData(bulkInsert(batchBuffer[0:bufferIndex]))

	return esDataList
}

func deleteIndex() error {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/%s", esURL, esIndex), nil)

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("Problem deleting ES index => %s", err.Error())
	} else if resp.StatusCode/100 != 2 {
		return fmt.Errorf("Problem deleting ES index => status code = %d", resp.StatusCode)
	}
	return nil
}

func insertEsData(data bulkInsert) error {
	if len(data) == 0 { // don't insert if no data
		return nil
	}

	var buf bytes.Buffer
	var encoder = json.NewEncoder(&buf)
	encoder.Encode(data)

	client := http.Client{}
	resp, err := client.Post(fmt.Sprintf("%s_bulk", esURL), "application/json", &buf)
	if err != nil {
		return fmt.Errorf("Failure stuffing data into ES => %s", err.Error())
	} else if resp.StatusCode/100 != 2 {
		return fmt.Errorf("Problem stuffing data into  ES => status code = %d", resp.StatusCode)
	}

	return nil
}
