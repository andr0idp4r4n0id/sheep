package main

import (
	"bufio"
	"crypto/tls"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
)

func CheckContains(url_t string) bool {
	re := regexp.MustCompile(`\?\w+=.+`)
	matched := re.MatchString(url_t)
	if matched {
		return true
	} else {
		return false
	}
}

func OrganizeInputTags(url_t string, wg *sync.WaitGroup, sem chan bool) {
	defer wg.Done()
	<-sem
	resp, err := http.Get(url_t)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return
	}
	if resp.StatusCode >= 400 {
		return
	}
	if CheckContains(url_t) {
		fmt.Println(url_t)
	}
	complete_input_tags_name := url.Values{}
	var new_url string
	var list_of_input_tags_names []string
	doc.Find("form").Each(func(_ int, selection *goquery.Selection) {
		method, _ := selection.Attr("method")
		if strings.ToLower(method) == "get" || method == "" {
			selection.Find("input").Each(func(_ int, s *goquery.Selection) {
				input_tags_name, _ := s.Attr("name")
				if input_tags_name != "" {
					list_of_input_tags_names = append(list_of_input_tags_names, input_tags_name)
					complete_input_tags_name.Set(input_tags_name, "1")
				}
			})
			if len(complete_input_tags_name) > 0 {
				name_tags_encoded := complete_input_tags_name.Encode()
				if CheckContains(url_t) {
					new_url = fmt.Sprintf("%s&%s", url_t, name_tags_encoded)
				} else {
					new_url = fmt.Sprintf("%s?%s", url_t, name_tags_encoded)
				}
				fmt.Println(new_url)
			}
		}
	})
	complete_input_tags_name = url.Values{}
	doc.Find("input").Each(func(i int, s *goquery.Selection) {
		input_tags_name, _ := s.Attr("name")
		if input_tags_name != "" {
			check := false
			for _, name := range list_of_input_tags_names {
				if input_tags_name == name {
					check = true
					break
				}
			}
			if !check {
				complete_input_tags_name.Set(input_tags_name, "1")
			}
		}
	})
	if len(complete_input_tags_name) > 0 {
		name_tags_encoded := complete_input_tags_name.Encode()
		if !CheckContains(url_t) {
			new_url = fmt.Sprintf("%s&%s", url_t, name_tags_encoded)
		} else {
			new_url = fmt.Sprintf("%s?%s", url_t, name_tags_encoded)
		}
		fmt.Println(new_url)
	}
}

func main() {
	var wg sync.WaitGroup
	conc := flag.Int("concurrency", 10, "concurrency level")
	sem := make(chan bool, *conc)
	reader := bufio.NewScanner(os.Stdin)
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	for reader.Scan() {
		url := reader.Text()
		wg.Add(1)
		sem <- true
		go OrganizeInputTags(url, &wg, sem)
	}
	for i := 0; i < cap(sem); i++ {
		sem <- true
	}
	wg.Wait()
}
