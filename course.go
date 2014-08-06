package main

import (
	"fmt"
	"github.com/kennygrant/sanitize"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	MAX_REQUESTS = 10
)

var (
	httpSemaphore = make(chan int, MAX_REQUESTS)
	// EXAMPLE:  COMS4995W001 => [COMS, 4995, W, 001]
	re = regexp.MustCompile(`(\w{4})(\w{4})(\w)(\w{3})`)
	// meant to match all HTML tags
	tags = regexp.MustCompile(`(?s:<.+?>)`)
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

type window struct {
	lower, upper int
}

func (w window) parse(s string) string {
	if w.lower < len(s)-1 {
		return ""
	} else if w.upper < len(s)-1 {
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
	Course     string `json:",omitempty",db:"course"`
	CourseFull string `json:",omitempty",db:"coursefull"`
	Course2Contents
}
type Section struct {
	Course      string `json:",omitempty",db:"course"`
	SectionFull string `json:",omitempty",db:"sectionfull"`
	SectionContents
}
type Course2Contents struct {
	PrefixName       string `json:",omitempty",db:"prefixname"`
	DivisionCode     string `json:",omitempty",db:"divisioncode"`
	DivisionName     string `json:",omitempty",db:"divisionname"`
	SchoolCode       string `json:",omitempty",db:"schoolcode"`
	SchoolName       string `json:",omitempty",db:"schoolname"`
	DepartmentCode   string `json:",omitempty",db:"departmentcode"`
	DepartmentName   string `json:",omitempty",db:"departmentname"`
	SubtermCode      string `json:",omitempty",db:"subtermcode"`
	SubtermName      string `json:",omitempty",db:"subtermname"`
	EnrollmentStatus string `json:",omitempty",db:"enrollmentstatus"`
	NumFixedUnits    string `json:",omitempty,int",db:"numfixedunits"`
	MinUnits         string `json:",omitempty,int",db:"minunits"`
	MaxUnits         string `json:",omitempty,int",db:"maxunits"`
	CourseTitle      string `json:",omitempty",db:"coursetitle"`
	CourseSubtitle   string `json:",omitempty",db:"coursesubtitle"`
	Approval         string `json:",omitempty",db:"approval"`
	BulletinFlags    string `json:",omitempty",db:"bulletinflags"`
	ClassNotes       string `json:",omitempty",db:"classnotes"`
	PrefixLongname   string `json:",omitempty",db:"prefixlongname"`
	Description      string `json:",omitempty",db:"description"`
}

type SectionContents struct {
	Term            string `json:",omitempty",db:"term"`
	MeetsOn1        string `json:",omitempty",db:"meetson1"`
	StartTime1      string `json:",omitempty",db:"starttime1"`
	EndTime1        string `json:",omitempty",db:"endtime1"`
	Building1       string `json:",omitempty",db:"building1"`
	Room1           string `json:",omitempty",db:"room1"`
	CallNumber      string `json:",omitempty,int",db:"callnumber"`
	CampusCode      string `json:",omitempty",db:"campuscode"`
	CampusName      string `json:",omitempty",db:"campusname"`
	NumEnrolled     string `json:",omitempty,int",db:"numenrolled"`
	MaxSize         string `json:",omitempty,int",db:"maxsize"`
	TypeCode        string `json:",omitempty",db:"typecode"`
	TypeName        string `json:",omitempty",db:"typename"`
	Meets1          string `json:",omitempty",db:"meets1"`
	Meets2          string `json:",omitempty",db:"meets2"`
	Meets3          string `json:",omitempty",db:"meets3"`
	Meets4          string `json:",omitempty",db:"meets4"`
	Meets5          string `json:",omitempty",db:"meets5"`
	Meets6          string `json:",omitempty",db:"meets6"`
	Instructor1Name string `json:",omitempty",db:"instructor1name"`
	Instructor2Name string `json:",omitempty",db:"instructor2name"`
	Instructor3Name string `json:",omitempty",db:"instructor3name"`
	Instructor4Name string `json:",omitempty",db:"instructor4name"`
	ExamMeet        string `json:",omitempty",db:"exammeet"`
	ExamDate        string `json:",omitempty",db:"examdate"`
}

func (c Course) split() (Course2, Section) {

	// finding the proper 'Course'
	res := re.FindStringSubmatch(strings.Replace(c.Course, " ", "_", 6))

	// set up the "Course Full"
	dept, deptNum, symbol := res[1], res[2], res[3]

	return Course2{
			Course:          dept + deptNum,
			CourseFull:      dept + symbol + deptNum,
			Course2Contents: c.Course2Contents,
		}, Section{
			Course:          dept + deptNum,
			SectionFull:     c.Course,
			SectionContents: c.SectionContents,
		}
}

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

func (c *Course) fill() {
	// t, err := time.Parse("15:04P", )
	if c.Meets1 == "" {
		c.StartTime1 = "00:00:00"
		c.EndTime1 = "00:00:00"
	} else {

		s := c.Meets1
		c.MeetsOn1 = meetsOn.parse(s)

		t := startTime.parse(s)
		if tm, err := time.Parse("15:04P", t); err == nil {
			c.StartTime1 = tm.Format("15:04:05")
		} else {
			c.StartTime1 = "00:00:00"
		}

		t = endTime.parse(s)
		if tm, err := time.Parse("15:04P", t); err == nil {
			c.EndTime1 = tm.Format("15:04:05")
		} else {
			c.EndTime1 = "00:00:00"
		}

		c.Building1 = building.parse(s)
		c.Room1 = room.parse(s)
	}

	c.NumFixedUnits = zeroInt(c.NumFixedUnits)
	c.MinUnits = zeroInt(c.MinUnits)
	c.MaxUnits = zeroInt(c.MaxUnits)
	c.CallNumber = zeroInt(c.CallNumber)
	c.NumEnrolled = zeroInt(c.NumEnrolled)
	c.MaxSize = zeroInt(c.MaxSize)
}

func (c Course) getCourseFull() (string, error) {
	res := re.FindStringSubmatch(strings.Replace(c.Course, " ", "_", 6))
	if len(res) != 5 {
		return "", fmt.Errorf("Failed to parse given 'Course', %s. found %#v", c.Course, res)
	}

	// set up the "Course Full"
	dept, deptNum, symbol := res[1], res[2], res[3]
	return dept + symbol + deptNum, nil
}

func (c Course) getDescription() error {
	url := c.getDescriptionURL()

	httpSemaphore <- 1
	resp, err := http.Get(url)
	<-httpSemaphore

	if err != nil {
		log.Printf("Error getting bulletin page, %s => %s", url, err.Error())
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
