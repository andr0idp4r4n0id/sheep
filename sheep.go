package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"sync"

	"github.com/labstack/gommon/random"
)

func CheckContains(url_t string) bool {
	re := regexp.MustCompile(`\?\w.+`)
	matched := re.MatchString(url_t)
	return matched
}

func SetPayloads(parameters url.Values) (map[string]string, url.Values) {
	payloads := url.Values{}
	reversed_payload := make(map[string]string)
	for name := range parameters {
		payload := random.New().String(10, "abcdefghklqpoirykmnbv")
		payloads.Set(name, payload)
		reversed_payload[payload] = name
	}
	return (reversed_payload), (payloads)
}

func EncodePayloads(payloads url.Values) string {
	return payloads.Encode()
}

func SendHttpRequestReadResponseBody(new_url string) []byte {
	resp, err := http.Get(new_url)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil
	}
	return bodyBytes
}

func CheckMatches(bodyString string, reversed_payload map[string]string) []string {
	var matches []string
	for payload := range reversed_payload {
		re, _ := regexp.Compile(payload)
		matches = append(matches, re.FindString(bodyString))
	}

	return matches
}

func FindPayloadInReversePayloads(matches []string, reversed_payload map[string]string) url.Values {
	reflected_url := url.Values{}
	for _, match := range matches {
		for payload, name := range reversed_payload {
			if match == payload {
				reflected_url.Set(name, "1")
				break
			}
		}
	}
	return reflected_url
}

func PrintReflections(reflected_url url.Values, new_url string, url_t string) {
	if len(reflected_url) > 0 {
		encoded_reflected_payloads := EncodePayloads(reflected_url)
		url := strings.Split(url_t, "?")[0]
		url = fmt.Sprintf("%s?%s", url, encoded_reflected_payloads)
		fmt.Println(url)
	} else {
		return
	}
}

func CheckReflectedParameters(url_t string, parameters url.Values, sem chan bool) {
	reversed_payload, payloads := SetPayloads(parameters)
	encoded_payloads := EncodePayloads(payloads)
	var new_url string
	if CheckContains(url_t) {
		new_url = fmt.Sprintf("%s&%s", url_t, encoded_payloads)
	} else {
		new_url = fmt.Sprintf("%s?%s", url_t, encoded_payloads)
	}
	bodyBytes := SendHttpRequestReadResponseBody(new_url)
	if bodyBytes == nil {
		return
	}
	bodyString := string(bodyBytes)
	matches := CheckMatches(bodyString, reversed_payload)
	reflected_url := FindPayloadInReversePayloads(matches, reversed_payload)
	PrintReflections(reflected_url, new_url, url_t)
}

func main() {
	reader := bufio.NewScanner(os.Stdin)
	conc := flag.Int("concurrency", 10, "concurrency level.")
	sem := make(chan bool, *conc)
	var wg sync.WaitGroup
	for reader.Scan() {
		url_t := reader.Text()
		uri, _ := url.Parse(url_t)
		sem <- true
		wg.Add(1)
		go func() {
			CheckReflectedParameters(url_t, uri.Query(), sem)
			<-sem
		}()
		wg.Done()
	}
	wg.Wait()
}
