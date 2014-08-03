package main

import (
	"fmt"
	"github.com/kennygrant/sanitize"
	"io/ioutil"
	"log"
	"net/http"
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

func bulletinWorker(prepChan, readyChan chan Course) {
	var url string
	var course Course
	for {
		course = <-prepChan
		url = course.getDescriptionURL()

		resp, err := http.Get(url)
		if err != nil {
			log.Printf("Error getting bulletin page, %s => %s", url, err.Error())
		}

		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Printf("Error reading in page (%s) body => %s", url, err.Error())
		}

		course.Description = getDesc(string(bodyBytes))
		if course.Description != "" {
			log.Printf("Found description for %s\n", course.Course)
		} else {
			log.Printf("No description for %s\n", course.Course)
		}
		readyChan <- course
	}
}

func getDesc(page string) string {
	res := desc.FindStringSubmatch(page)

	if len(res) == 1 {
		s := tags.ReplaceAllString(res[0], "")
		s = strings.TrimSpace(strings.Replace(strings.Replace(s, web, "", 1), courseDesc, "", 1))
		return sanitize.Accents(s)
	}
	return ""
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
