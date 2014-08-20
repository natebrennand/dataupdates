package main

import (
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

var (
	httpSemaphore = make(chan int, MAX_HTTP_REQUESTS)               // used to limit the # of HTTP requests at a time
	re            = regexp.MustCompile(`(\w{4})(\w{4})(\w)(\w{3})`) // EXAMPLE:  COMS4995W001 => [COMS, 4995, W, 001]
	tags          = regexp.MustCompile(`(?s:<.+?>)`)                // meant to match all HTML tags
	// TODO: repent for this hidiousness
	desc = regexp.MustCompile(`[.\n]*Course Description</td>\n <td bgcolor=#DADADA>(?s:.*)<tr valign=top><td bgcolor=#99CCFF>Web Site</td>[.\n]*`)

	// used to parse the Meets1 parameter into useful pieces
	meetsOn   = window{0, 7}
	startTime = window{7, 13}
	endTime   = window{14, 20}
	building  = window{24, 35}
	room      = window{35, -1}
)

// helper struct for fill()
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

// helper method for fill()
func zeroInt(s string) string {
	n, _ := strconv.Atoi(s)
	return strconv.FormatInt(int64(n), 10)
}

// helper method for fill()
func parseDate(t string) string {
	if tm, err := time.Parse("15:04P", t); err == nil {
		return tm.Format("15:04:05")
	}
	return "00:00:00"
}

// standardizes information in a Course
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

// parses the 'CourseFull' attribute
func (c *Course) setCourseFull() {
	res := re.FindStringSubmatch(strings.Replace(c.Course, " ", "_", 6))
	if len(res) != 5 {
		log.Printf("Failed to parse given 'Course', %s. found %#v", c.Course, res)
	}

	// set up the "Course Full"
	dept, deptNum, symbol := res[1], res[2], res[3]
	courseFull := dept + symbol + deptNum
	c.CourseFull = courseFull
}

func (c *Course) getBulletinURL() string {
	courseRegex := re.FindStringSubmatch(strings.Replace(c.Course, " ", "_", 6))
	dept, deptNum, symbol, section := courseRegex[1], courseRegex[2], courseRegex[3], courseRegex[4]

	// GOAL: http://www.columbia.edu/cu/bulletin/uwb/subj/COMS/W4995-20143-001/
	return fmt.Sprintf("http://www.columbia.edu/cu/bulletin/uwb/subj/%s/%s-%s-%s/",
		dept,
		symbol+deptNum,
		c.Term,
		section,
	)
}

func parsePage(page []byte) string {
	res := desc.FindStringSubmatch(string(page))
	if len(res) != 1 {
		return ""
	}

	// remove all tags
	s := tags.ReplaceAllString(res[0], "")
	// remove static words
	s = strings.TrimSpace(strings.Replace(strings.Replace(s, "Web Site", "", 1), "Course Description", "", 1))
	// remove special characters and return
	return sanitize.Accents(s)
}

// scrapes the bulletin to get the description of a course
func (c *Course) getDescription() error {
	// first find the URL of the course's bulletin page
	url := c.getBulletinURL()

	// locks while requesting
	httpSemaphore <- 1
	resp, err := http.Get(url)
	<-httpSemaphore

	// check for errors
	if err != nil {
		log.Printf("Error getting bulletin page, %s => %s", url, err.Error())
		return fmt.Errorf("Error querying bulletin for course, %s", c.Course)
	} else if resp.StatusCode/100 != 2 {
		log.Printf("Error getting bulletin page, %s, status code => %d", url, resp.StatusCode)
		return fmt.Errorf("Error querying bulletin for course, %s", c.Course)
	}

	// read in then sanitize description
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading in page (%s) body => %s", url, err.Error())
	}

	// parse the page for the description
	courseDesc := parsePage(bodyBytes)
	if courseDesc == "" { // set to 'no description' if there is not one
		log.Printf("no description for course, %s", c.Course)
		log.Printf(url)
		log.Print(string(bodyBytes))
		c.Description = "no description"
		return nil
	}
	c.Description = courseDesc
	return nil
}

// Course holds all information about an instance of a course
type Course struct {
	Course2
	Section
	Course     string `json:",omitempty"`
	ChargeMsg1 string `json:",omitempty"`
	ChargeAmt1 string `json:",omitempty"`
	ChargeMsg2 string `json:",omitempty"`
	ChargeAmt2 string `json:",omitempty"`
}

// Course2 holds all information a Course offered (ignoring section details)
type Course2 struct {
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

// Section holds all information about a course's individual section
type Section struct {
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
