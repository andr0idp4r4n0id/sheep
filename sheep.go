package main

import (
	"bufio"
	"crypto/tls"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
)

func OrganizeInputTags(url string, wg *sync.WaitGroup, sem chan bool) {
	defer wg.Done()
	<-sem
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	forms := doc.Find("form")
	var complete_input_tags_name string
	forms.Each(func(_ int, selection *goquery.Selection) {
		method, _ := selection.Attr("method")
		if method != "" {
			method = strings.ToLower(method)
			if method == "get" {
				input_tags := selection.Find("input")
				input_tags.Each(func(_ int, s *goquery.Selection) {
					input_tags_name, _ := s.Attr("name")
					if input_tags_name != "" {
						complete_input_tags_name += "," + input_tags_name + "=1"
					} else {
						return
					}
				})
				if complete_input_tags_name != "" {
					if strings.Contains(url, "?") {
						complete_input_tags_name = strings.Replace(complete_input_tags_name, ",", "&", -1)
					} else {
						complete_input_tags_name = strings.Replace(complete_input_tags_name, ",", "?", 1)
						complete_input_tags_name = strings.Replace(complete_input_tags_name, ",", "&", -1)
					}
					new_url := fmt.Sprintf("%s%s", url, complete_input_tags_name)
					fmt.Println(new_url)
				}
			}
		}
	})
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
