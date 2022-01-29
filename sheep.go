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
	"github.com/corpix/uarand"
)

func CheckContains(url_t string) bool {
	re := regexp.MustCompile(`\?.+=.+`)
	return re.MatchString(url_t)
}

func SendHttpRequestReadResponseBody(url_t string) *goquery.Document {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url_t, nil)
	if err != nil {
		return nil
	}
	req.Header.Set("User-Agent", uarand.GetRandom())
	resp, err := client.Do(req)
	if err != nil {
		return nil
	}
	if resp.StatusCode >= 400 {
		return nil
	}
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil
	}
	return doc
}

func FindGetInputInForms(doc *goquery.Document) (url.Values, []string) {
	var list_of_input_tags_names []string
	complete_input_tags_name := url.Values{}
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
		}
	})
	return (complete_input_tags_name), (list_of_input_tags_names)
}

func EncodeInputTagsName(complete_input_tags_name url.Values) string {
	return complete_input_tags_name.Encode()
}

func GetInputTagsWithoutForm(doc *goquery.Document, list_of_input_tags_names []string) url.Values {
	complete_input_tags_name_s := url.Values{}
	doc.Find("input").Each(func(i int, s *goquery.Selection) {
		if s.Before("form") != nil {
			return
		}
		input_tags_name, _ := s.Attr("name")
		if input_tags_name != "" {
			complete_input_tags_name_s.Set(input_tags_name, "1")
		}
	})
	return complete_input_tags_name_s
}

func OrganizeInputTags(url_t string) {
	if CheckContains(url_t) {
		fmt.Println(url_t)
	}
	doc := SendHttpRequestReadResponseBody(url_t)
	if doc == nil {
		return
	}
	var name_tags_encoded string
	var new_url string
	complete_input_tags_name, list_of_input_tags_names := FindGetInputInForms(doc)
	if len(complete_input_tags_name) == 0 {
		return
	}
	name_tags_encoded = EncodeInputTagsName(complete_input_tags_name)
	if CheckContains(url_t) {
		new_url = fmt.Sprintf("%s&%s", url_t, name_tags_encoded)
	} else {
		new_url = fmt.Sprintf("%s?%s", url_t, name_tags_encoded)
	}
	fmt.Println(new_url)
	complete_input_tags_name_s := GetInputTagsWithoutForm(doc, list_of_input_tags_names)
	if len(complete_input_tags_name_s) == 0 {
		return
	}
	name_tags_encoded = EncodeInputTagsName(complete_input_tags_name_s)
	if CheckContains(url_t) {
		new_url = fmt.Sprintf("%s&%s", url_t, name_tags_encoded)
	} else {
		new_url = fmt.Sprintf("%s?%s", url_t, name_tags_encoded)
	}
	fmt.Println(new_url)

}

func main() {
	var wg sync.WaitGroup
	conc := flag.Int("concurrency", 10, "concurrency level")
	flag.Parse()
	reader := bufio.NewScanner(os.Stdin)
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	for i := 0; i < *conc; i++ {
		wg.Add(1)
		go func() {
			for reader.Scan() {
				url := reader.Text()
				OrganizeInputTags(url)
			}
			wg.Done()
		}()
		wg.Wait()
	}

}
