package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
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
	Course         string // EX: ZULU336
	CourseFull     string // EX: ZULUW336
	CourseSubtitle string
	CourseTitle    string
	Description    string
	Term           pg_array.SqlIntArray
	CallNumber     pg_array.SqlIntArray
	Instructor     pg_array.SqlStringArray
}

type esMetadata struct {
	Index string `json:"_index"`
	Type  string `json:"_type"`
	ID    string `json:"_id"` // will be the 'Course' attribute
}

type esAction struct {
	Index esMetadata `json:"index"`
}

type bulkItem struct {
	Index esAction
	Data  esData
}

type bulkInsert []bulkItem

func (b bulkInsert) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer
	var encoder = json.NewEncoder(&buf)

	for _, item := range b {
		if err := encoder.Encode(item.Index); err != nil {
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
		Index: esAction{
			Index: esMetadata{
				Index: esIndex,
				Type:  esType,
				ID:    d.Course,
			},
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

func updateES(db *sql.DB) []esData {
	// remove the existing ES index
	if err := deleteIndex(); err != nil {
		log.Printf("WARNING: %s", err.Error())
	}
	if err := createIndex(); err != nil {
		log.Fatalf("WARNING: %s", err.Error())
	}

	// query for the new data used in the index
	var esDataList []esData
	rows, err := db.Query(esQuery)
	if err != nil {
		log.Fatalf("Error while querying Postgres for ES data => %s", err.Error())
	}
	defer rows.Close()

	// process each record to be inserted to ES
	var batchBuffer = make([]bulkItem, batchSize)
	var bufferIndex = 0
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

		// add to buffer
		batchBuffer[bufferIndex] = data.NewBulkItem()
		bufferIndex++
		if bufferIndex == batchSize {
			log.Printf("Inserting batch of %d\n", batchSize)
			err := insertEsData(bulkInsert(batchBuffer))
			bufferIndex = 0
			if err != nil {
				log.Printf("WARNING: failed to run batch insert => %s\n", err.Error())
			}
		}
	}
	// insert remainder of buffer
	insertEsData(bulkInsert(batchBuffer[0:bufferIndex]))

	return esDataList
}

func deleteIndex() error {
	log.Println("Attempting to delete ES index")
	req, err := http.NewRequest("DELETE", esURL+esIndex, nil)

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("Problem deleting ES index => %s", err.Error())
	} else if resp.StatusCode/100 != 2 {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Printf("Failed to read in response body => %s\n", err.Error())
		}
		log.Println(string(bodyBytes))
		return fmt.Errorf("Problem deleting ES index => status code = %d", resp.StatusCode)
	}
	log.Println("ES index deleted")
	return nil
}

func createIndex() error {
	log.Println("Attempting to create new ES Index")

	req, err := http.NewRequest("PUT", esURL+esIndex, nil)
	if err != nil {
		return fmt.Errorf("Failed to create PUT request => %s", err.Error())
	}

	client := http.Client{}
	if resp, err := client.Do(req); err != nil {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Printf("Failed to read in response body => %s\n", err.Error())
		}
		log.Println(string(bodyBytes))
		return fmt.Errorf("Failed to create new ES Index => %s", err.Error())
	}

	log.Printf("ES Index, %s, created", esIndex)
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
	req, err := http.NewRequest("POST", esURL+"_bulk", &buf)
	if err != nil {
		return fmt.Errorf("Failed to form HTTP request.")
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("Failure stuffing data into ES => %s", err.Error())
	} else if resp.StatusCode/100 != 2 {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Printf("Failed to read in response body => %s\n", err.Error())
		}
		log.Println(string(bodyBytes))
		return fmt.Errorf("Problem stuffing data into  ES => status code = %d", resp.StatusCode)
	}

	return nil
}
