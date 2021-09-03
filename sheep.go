package main

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
)

func OrganizeInputTags(url string, wg *sync.WaitGroup) {
	defer wg.Done()
	resp, err := http.Get(url)
	if err != nil {
		return
	} else {
		defer resp.Body.Close()
		doc, err := goquery.NewDocumentFromReader(resp.Body)
		if err != nil {
			return
		} else {
			forms := doc.Find("form")
			if forms.Size() > 0 {
				var complete_input_tags_name string
				forms.Each(func(_ int, selection *goquery.Selection) {
					method, _ := selection.Attr("method")
					if strings.ToLower(method) == "get" {
						input_tags := selection.Find("input")
						if input_tags.Size() > 0 {
							input_tags.Each(func(_ int, s *goquery.Selection) {
								input_tags_name, _ := s.Attr("name")
								if input_tags_name != "" {
									complete_input_tags_name += "," + input_tags_name + "=1"
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
			} else {
				return
			}
		}
	}
}
func main() {
	var url string
	var wg sync.WaitGroup
	reader := bufio.NewScanner(os.Stdin)
	for reader.Scan() {
		url = reader.Text()
		wg.Add(1)
		go OrganizeInputTags(url, &wg)
	}
	wg.Wait()
}
