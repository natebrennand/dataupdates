package main

import (
	"database/sql"
	"fmt"
	"log"
	"sync"
)

func dbWorker(db *sql.DB, readyCourse chan Course, wg *sync.WaitGroup, descCache map[string]string) {
	var (
		c              Course
		more           bool
		courseInserted = make(map[string]interface{})
	)

	for {
		c, more = <-readyCourse
		if c.CourseFull == "" {
			log.Printf("failed to insert course, %s", c.Course)
			continue
		}

		// now we must get the description
		if err := c.getDescription(); err != nil {
			log.Printf("Could not get description for %s, %s", c.Course, err.Error())
		}
		fmt.Print(".")

		if err := c.Insert(db); err != nil {
			log.Printf("While inserting course => %#v\n, database error => %s", c, err.Error())
		}

		if err := c.InsertSection(db); err != nil {
			log.Printf("Failed to insert section, %s, err => %s", c.SectionFull, err.Error())
		}

		if _, exists := courseInserted[c.CourseFull]; !exists {
			if err := c.InsertCourse2(db); err != nil {
				log.Printf("Failed to insert course_v2, %s, err => %s", c.CourseFull, err.Error())
			}
			courseInserted[c.CourseFull] = 0
		}

		if !more {
			wg.Done()
		}
	}
}

// Insert inserts the Course to the 'courses_t' database
func (c Course) Insert(db *sql.DB) error {
	query := `INSERT INTO courses_t (
	course,
	ChargeMsg1,
	ChargeAmt1,
	ChargeMsg2,
	ChargeAmt2,
	prefixname,
	divisioncode,
	divisionname,
	schoolcode,
	schoolname,
	departmentcode,
	departmentname,
	subtermcode,
	subtermname,
	enrollmentstatus,
	numfixedunits,
	minunits,
	maxunits,
	coursetitle,
	coursesubtitle,
	approval,
	bulletinflags,
	classnotes,
	prefixlongname,
	description,
	term,
	meetson1,
	starttime1,
	endtime1,
	building1,
	room1,
	callnumber,
	campuscode,
	campusname,
	numenrolled,
	maxsize,
	typecode,
	typename,
	meets1,
	meets2,
	meets3,
	meets4,
	meets5,
	meets6,
	instructor1name,
	instructor2name,
	instructor3name,
	instructor4name,
	exammeet,
	examdate
	) VALUES (
	$1,
	$2,
	$3,
	$4,
	$5,
	$6,
	$7,
	$8,
	$9,
	$10,
	$11,
	$12,
	$13,
	$14,
	$15,
	$16,
	$17,
	$18,
	$19,
	$20,
	$21,
	$22,
	$23,
	$24,
	$25,
	$26,
	$27,
	$28,
	$29,
	$30,
	$31,
	$32,
	$33,
	$34,
	$35,
	$36,
	$37,
	$38,
	$39,
	$40,
	$41,
	$42,
	$43,
	$44,
	$45,
	$46,
	$47,
	$48,
	$49,
	$50
	)`
	_, err := db.Exec(
		query,
		c.Course,
		c.ChargeMsg1,
		c.ChargeAmt1,
		c.ChargeMsg2,
		c.ChargeAmt2,
		c.PrefixName,
		c.DivisionCode,
		c.DivisionName,
		c.SchoolCode,
		c.SchoolName,
		c.DepartmentCode,
		c.DepartmentName,
		c.SubtermCode,
		c.SubtermName,
		c.EnrollmentStatus,
		c.NumFixedUnits,
		c.MinUnits,
		c.MaxUnits,
		c.CourseTitle,
		c.CourseSubtitle,
		c.Approval,
		c.BulletinFlags,
		c.ClassNotes,
		c.PrefixLongname,
		c.Description,
		c.Term,
		c.MeetsOn1,
		c.StartTime1,
		c.EndTime1,
		c.Building1,
		c.Room1,
		c.CallNumber,
		c.CampusCode,
		c.CampusName,
		c.NumEnrolled,
		c.MaxSize,
		c.TypeCode,
		c.TypeName,
		c.Meets1,
		c.Meets2,
		c.Meets3,
		c.Meets4,
		c.Meets5,
		c.Meets6,
		c.Instructor1Name,
		c.Instructor2Name,
		c.Instructor3Name,
		c.Instructor4Name,
		c.ExamMeet,
		c.ExamDate,
	)
	if err != nil {
		return fmt.Errorf("Failed to insert courses_t, %#v, => %s", c, err.Error())
	}
	return nil
}

// InsertCourse2 inserts information from the course to the 'courses_v2_t' database
func (c Course) InsertCourse2(db *sql.DB) error {
	query := `INSERT INTO courses_v2_t (
	course,
	coursefull,
	prefixname,
	divisioncode,
	divisionname,
	schoolcode,
	schoolname,
	departmentcode,
	departmentname,
	subtermcode,
	subtermname,
	enrollmentstatus,
	numfixedunits,
	minunits,
	maxunits,
	coursetitle,
	coursesubtitle,
	approval,
	bulletinflags,
	classnotes,
	prefixlongname,
	description
	) VALUES (
	$1,
	$2,
	$3,
	$4,
	$5,
	$6,
	$7,
	$8,
	$9,
	$10,
	$11,
	$12,
	$13,
	$14,
	$15,
	$16,
	$17,
	$18,
	$19,
	$20,
	$21,
	$22
	)`
	_, err := db.Exec(
		query,
		c.Course,
		c.CourseFull,
		c.PrefixName,
		c.DivisionCode,
		c.DivisionName,
		c.SchoolCode,
		c.SchoolName,
		c.DepartmentCode,
		c.DepartmentName,
		c.SubtermCode,
		c.SubtermName,
		c.EnrollmentStatus,
		c.NumFixedUnits,
		c.MinUnits,
		c.MaxUnits,
		c.CourseTitle,
		c.CourseSubtitle,
		c.Approval,
		c.BulletinFlags,
		c.ClassNotes,
		c.PrefixLongname,
		c.Description,
	)
	if err != nil {
		return fmt.Errorf("Failed to insert courses_v2_t, %#v, => %s", c.Course2, err.Error())
	}

	return nil
}

// InsertSection inserts information from the course to the 'sections_v2_t' database
func (c Course) InsertSection(db *sql.DB) error {
	query := `INSERT INTO sections_v2_t (
	course,
	term,
	meetson1,
	starttime1,
	endtime1,
	building1,
	room1,
	callnumber,
	campuscode,
	campusname,
	numenrolled,
	maxsize,
	typecode,
	typename,
	meets1,
	meets2,
	meets3,
	meets4,
	meets5,
	meets6,
	instructor1name,
	instructor2name,
	instructor3name,
	instructor4name,
	exammeet,
	examdate
	) VALUES (
		$1,
		$2,
		$3,
		$4,
		$5,
		$6,
		$7,
		$8,
		$9,
		$10,
		$11,
		$12,
		$13,
		$14,
		$15,
		$16,
		$17,
		$18,
		$19,
		$20,
		$21,
		$22,
		$23,
		$24,
		$25,
		$26
	)`
	_, err := db.Exec(
		query,
		c.Course,
		c.Term,
		c.MeetsOn1,
		c.StartTime1,
		c.EndTime1,
		c.Building1,
		c.Room1,
		c.CallNumber,
		c.CampusCode,
		c.CampusName,
		c.NumEnrolled,
		c.MaxSize,
		c.TypeCode,
		c.TypeName,
		c.Meets1,
		c.Meets2,
		c.Meets3,
		c.Meets4,
		c.Meets5,
		c.Meets6,
		c.Instructor1Name,
		c.Instructor2Name,
		c.Instructor3Name,
		c.Instructor4Name,
		c.ExamMeet,
		c.ExamDate,
	)
	if err != nil {
		return fmt.Errorf("Failed to insert sections_v2_t, %#v, => %s", c.Section, err.Error())
	}
	return nil
}
