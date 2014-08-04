package main

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	// EXAMPLE:  COMS4995W001 => [COMS, 4995, W, 001]
	re   = regexp.MustCompile(`(\w{4})(\w{4})(\w)(\w{3})`)
	tags = regexp.MustCompile(`(?s:<.+?>)`)
	// TODO: repent for this hidiousness
	desc       = regexp.MustCompile(`[.\n]*Course Description</td>\n <td bgcolor=#DADADA>(?s:.*)<tr valign=top><td bgcolor=#99CCFF>Web Site</td>[.\n]*`)
	web        = "Web Site"
	courseDesc = "Course Description"
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

func (c Course) getDescriptionURL() string {
	s := strings.Replace(c.Course, " ", "_", 6)
	res := re.FindStringSubmatch(s)

	// http://www.columbia.edu/cu/bulletin/uwb/subj/COMS/W4995-20143-001/
	return fmt.Sprintf("http://www.columbia.edu/cu/bulletin/uwb/subj/%s/%s-%s-%s/",
		res[1],
		res[3]+res[2],
		c.Term,
		res[4],
	)
}
