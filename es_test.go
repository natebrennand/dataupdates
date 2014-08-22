package main

import (
	"bytes"
	"testing"

	"github.com/natebrennand/pg_array"
)

var (
	testEsData = esData{
		Course:         "test",
		CourseFull:     "123",
		CourseSubtitle: "test course subtitle",
		CourseTitle:    "test",
		Description:    "a course for testing",
		Term: pgarray.SqlIntArray{
			Data: []int64{1, 2, 3},
		},
		CallNumber: pgarray.SqlIntArray{
			Data: []int64{4, 5, 6},
		},
		Instructor: pgarray.SqlStringArray{
			Data: []string{"teacher1", "teacher2"},
		},
	}
	testEsAction = esAction{
		Index: esMetadata{
			Index: "test",
			Type:  "courses",
			ID:    "123",
		},
	}
	testBulkItem = bulkItem{
		Index: testEsAction,
		Data:  testEsData,
	}
	testBulkInsert  = bulkInsert([]bulkItem{testBulkItem})
	testBulkInsert2 = bulkInsert([]bulkItem{testBulkItem, testBulkItem})
)

func TestBulkInsertMarshal(t *testing.T) {
	jsonBytes, err := testBulkInsert.MarshalJSON()
	if err != nil {
		t.Errorf("Error encoding JSON => %s", err.Error())
	}

	buf := bytes.NewBuffer(jsonBytes)
	if buf.String() != expectedJSON {
		t.Error("ES JSON not encoded as expected.")
	}
}

func TestBulkInsertMarshal2(t *testing.T) {
	jsonBytes, err := testBulkInsert2.MarshalJSON()
	if err != nil {
		t.Errorf("Error encoding JSON => %s", err.Error())
	}

	buf := bytes.NewBuffer(jsonBytes)
	if buf.String() != expectedJSON2 {
		t.Error("ES JSON not encoded as expected.")
	}
}

// Newlines are expected after each of the json segments
//
// Spec: http://www.elasticsearch.org/guide/en/elasticsearch/reference/current/docs-bulk.html
var expectedJSON = `{"index":{"_index":"test","_type":"courses","_id":"123"}}
{"Course":"test","CourseFull":"123","DespartmentCode":"","DespartmentName":"","CourseTitle":"test","CourseSubtitle":"test course subtitle","Description":"a course for testing","Term":[1,2,3],"CallNumber":[4,5,6],"Instructor":["teacher1","teacher2"]}
`

var expectedJSON2 = `{"index":{"_index":"test","_type":"courses","_id":"123"}}
{"Course":"test","CourseFull":"123","DespartmentCode":"","DespartmentName":"","CourseTitle":"test","CourseSubtitle":"test course subtitle","Description":"a course for testing","Term":[1,2,3],"CallNumber":[4,5,6],"Instructor":["teacher1","teacher2"]}
{"index":{"_index":"test","_type":"courses","_id":"123"}}
{"Course":"test","CourseFull":"123","DespartmentCode":"","DespartmentName":"","CourseTitle":"test","CourseSubtitle":"test course subtitle","Description":"a course for testing","Term":[1,2,3],"CallNumber":[4,5,6],"Instructor":["teacher1","teacher2"]}
`
