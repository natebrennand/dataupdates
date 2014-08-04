package main

import (
	"github.com/kennygrant/sanitize"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
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
