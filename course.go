package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/kennygrant/sanitize"
)

const (
	MAX_REQUESTS = 10
)

var (
	httpSemaphore = make(chan int, MAX_REQUESTS)                    // used to limit to 1 HTTP request at a time
	re            = regexp.MustCompile(`(\w{4})(\w{4})(\w)(\w{3})`) // EXAMPLE:  COMS4995W001 => [COMS, 4995, W, 001]
	tags          = regexp.MustCompile(`(?s:<.+?>)`)                // meant to match all HTML tags
	// TODO: repent for this hidiousness
	desc       = regexp.MustCompile(`[.\n]*Course Description</td>\n <td bgcolor=#DADADA>(?s:.*)<tr valign=top><td bgcolor=#99CCFF>Web Site</td>[.\n]*`)
	web        = "Web Site"
	courseDesc = "Course Description"

	meetsOn   = window{0, 7}
	startTime = window{7, 13}
	endTime   = window{14, 20}
	building  = window{24, 35}
	room      = window{35, -1}
)

func (c Course) getDescriptionURL() string {
	s := strings.Replace(c.Course, " ", "_", 6)
	res := re.FindStringSubmatch(s)
	dept, deptNum, symbol, section := res[1], res[2], res[3], res[4]

	// GOAL: http://www.columbia.edu/cu/bulletin/uwb/subj/COMS/W4995-20143-001/
	return fmt.Sprintf("http://www.columbia.edu/cu/bulletin/uwb/subj/%s/%s-%s-%s/",
		dept,
		symbol+deptNum,
		c.Term,
		section,
	)
}

func zeroInt(s string) string {
	n, _ := strconv.Atoi(s)
	return strconv.FormatInt(int64(n), 10)
}
func parseDate(t string) string {
	if tm, err := time.Parse("15:04P", t); err == nil {
		return tm.Format("15:04:05")
	}
	return "00:00:00"
}

func (c *Course) fill() {
	if c.Meets1 == "" {
		c.StartTime1 = "00:00:00"
		c.EndTime1 = "00:00:00"
	} else {

		s := c.Meets1
		c.MeetsOn1 = meetsOn.parse(s)
		c.StartTime1 = parseDate(startTime.parse(s))
		c.EndTime1 = parseDate(endTime.parse(s))
		c.Building1 = building.parse(s)
		c.Room1 = room.parse(s)
	}

	c.NumFixedUnits = zeroInt(c.NumFixedUnits)
	c.MinUnits = zeroInt(c.MinUnits)
	c.MaxUnits = zeroInt(c.MaxUnits)
	c.CallNumber = zeroInt(c.CallNumber)
	c.NumEnrolled = zeroInt(c.NumEnrolled)
	c.MaxSize = zeroInt(c.MaxSize)

	c.setCourseFull()
}

func (c *Course) setCourseFull() (string, error) {
	res := re.FindStringSubmatch(strings.Replace(c.Course, " ", "_", 6))
	if len(res) != 5 {
		return "", fmt.Errorf("Failed to parse given 'Course', %s. found %#v", c.Course, res)
	}

	// set up the "Course Full"
	dept, deptNum, symbol := res[1], res[2], res[3]
	courseFull := dept + symbol + deptNum
	c.CourseFull = courseFull
	return courseFull, nil
}

func (c *Course) getDescription() error {
	url := c.getDescriptionURL()

	httpSemaphore <- 1
	resp, err := http.Get(url)
	<-httpSemaphore

	if err != nil {
		log.Printf("Error getting bulletin page, %s => %s", url, err.Error())
		return fmt.Errorf("Error querying bulletin for course, %s", c.Course)
	} else if resp.StatusCode/100 != 2 {
		log.Printf("Error getting bulletin page, %s, status code => %d", url, resp.StatusCode)
		return fmt.Errorf("Error querying bulletin for course, %s", c.Course)
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading in page (%s) body => %s", url, err.Error())
	}

	res := desc.FindStringSubmatch(string(bodyBytes))

	if len(res) != 1 {
		log.Printf("no description for course, %s", c.Course)
		c.Description = "no description"
		return nil
	}
	s := tags.ReplaceAllString(res[0], "")
	s = strings.TrimSpace(strings.Replace(strings.Replace(s, web, "", 1), courseDesc, "", 1))
	c.Description = sanitize.Accents(s)
	return nil
}

type window struct {
	lower, upper int
}

func (w window) parse(s string) string {
	if w.lower > len(s)-1 {
		return ""
	} else if w.upper > len(s)-1 {
		return strings.Replace(s[w.lower:], " ", "", -1)
	}
	if w.lower < 0 {
		return strings.Replace(s[:w.upper], " ", "", -1)
	} else if w.upper < 0 {
		return strings.Replace(s[w.lower:], " ", "", -1)
	}
	return strings.Replace(s[w.lower:w.upper], " ", "", -1)
}

type Course struct {
	Course2Contents
	SectionContents
	Course     string `json:",omitempty",db:"course"`
	ChargeMsg1 string `json:",omitempty",db:"chargemsg1"`
	ChargeAmt1 string `json:",omitempty",db:"chargeamt1"`
	ChargeMsg2 string `json:",omitempty",db:"chargemsg2"`
	ChargeAmt2 string `json:",omitempty",db:"chargeamt2"`
}

type Course2 struct {
	Course string `json:",omitempty",db:"course"`
	Course2Contents
}

type Course2Contents struct {
	CourseFull       string `json:",omitempty"`
	PrefixName       string `json:",omitempty"`
	DivisionCode     string `json:",omitempty"`
	DivisionName     string `json:",omitempty"`
	SchoolCode       string `json:",omitempty"`
	SchoolName       string `json:",omitempty"`
	DepartmentCode   string `json:",omitempty"`
	DepartmentName   string `json:",omitempty"`
	SubtermCode      string `json:",omitempty"`
	SubtermName      string `json:",omitempty"`
	EnrollmentStatus string `json:",omitempty"`
	NumFixedUnits    string `json:",omitempty,`
	MinUnits         string `json:",omitempty,`
	MaxUnits         string `json:",omitempty,`
	CourseTitle      string `json:",omitempty"`
	CourseSubtitle   string `json:",omitempty"`
	Approval         string `json:",omitempty"`
	BulletinFlags    string `json:",omitempty"`
	ClassNotes       string `json:",omitempty"`
	PrefixLongname   string `json:",omitempty"`
	Description      string `json:",omitempty"`
}

type Section struct {
	Course string `json:",omitempty",db:"course"`
	SectionContents
}

type SectionContents struct {
	SectionFull     string `json:",omitempty"`
	Term            string `json:",omitempty"`
	MeetsOn1        string `json:",omitempty"`
	StartTime1      string `json:",omitempty"`
	EndTime1        string `json:",omitempty"`
	Building1       string `json:",omitempty"`
	Room1           string `json:",omitempty"`
	CallNumber      string `json:",omitempty,int"`
	CampusCode      string `json:",omitempty"`
	CampusName      string `json:",omitempty"`
	NumEnrolled     string `json:",omitempty,int"`
	MaxSize         string `json:",omitempty,int"`
	TypeCode        string `json:",omitempty"`
	TypeName        string `json:",omitempty"`
	Meets1          string `json:",omitempty"`
	Meets2          string `json:",omitempty"`
	Meets3          string `json:",omitempty"`
	Meets4          string `json:",omitempty"`
	Meets5          string `json:",omitempty"`
	Meets6          string `json:",omitempty"`
	Instructor1Name string `json:",omitempty"`
	Instructor2Name string `json:",omitempty"`
	Instructor3Name string `json:",omitempty"`
	Instructor4Name string `json:",omitempty"`
	ExamMeet        string `json:",omitempty"`
	ExamDate        string `json:",omitempty"`
}

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
		c.PrefixName,
		c.CourseFull,
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
		return fmt.Errorf("Failed to insert courses_v2_t, %#v, => %s", c.Course2Contents, err.Error())
	}

	return nil
}

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
		return fmt.Errorf("Failed to insert sections_v2_t, %#v, => %s", c.SectionContents, err.Error())
	}
	return nil
}
