/*
k2: Reverse whois lookup
by github.com/harleo — MIT License
*/

package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/puerkitobio/goquery"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

type sliceVal []string

func (s sliceVal) String() string {
	var str string
	for _, i := range s {
		str += fmt.Sprintf("%s\n", i)
	}
	return str
}

func writeLines(lines []string, path string) error {
	file, err := os.Create(path)
	if err != nil {
		log.Fatalf("[!] Couldn't create file: %s\n", err.Error())
	}
	defer file.Close()

	w := bufio.NewWriter(file)
	for _, line := range lines {
		fmt.Fprintln(w, line)
	}
	return w.Flush()
}

func writeJson(lines []string, path string) {
	file, err := json.MarshalIndent(lines, "", "  ")

	if err != nil {
		log.Fatalf("[!] Couldn't create file: %s\n", err.Error())
	}

	err = ioutil.WriteFile(path, file, 0644)

	if err != nil {
		log.Fatal(err)
	}
}

func httpRequest(URI string) string {
	response, errGet := http.Get(URI)
	if errGet != nil {
		log.Fatalf("[!] Error sending request: %s\n", errGet.Error())
	}

	responseText, errRead := ioutil.ReadAll(response.Body)
	if errRead != nil {
		log.Fatalf("[!] Error reading response: %s\n", errRead.Error())
	}

	defer response.Body.Close()
	return string(responseText)
}

func parseTable(data string) []string {
	var domains []string

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(data))
	if err != nil {
		log.Fatalf("[!] Parsing issue: %s\n", err.Error())
	}

	doc.Find("tr").Each(func(_ int, tr *goquery.Selection) {
		tr.Find("td").Each(func(i int, row *goquery.Selection) {
			if i == 1 {
				domain := row.Text()
				if domain != "" {
					domains = append(domains, domain)
				}
			}
		})
	})

	return domains[1:]
}

func main() {
	namePtr := flag.String("n", "", "Registrant name, email or domain name of the target (Required)")
	printPtr := flag.Bool("p", false, "Print results")
	jsonF := flag.Bool("j", false, "Write to a json file")
	flag.Parse()

	if *namePtr == "" {
		flag.PrintDefaults()
		log.Fatal("Registrant name, email or domain name of the target is required")
	}

	fmt.Println("[:] Sending query...")
	request := httpRequest(fmt.Sprintf("https://www.reversewhois.io/?searchterm=%s", *namePtr))
	domains := parseTable(request)

	if len(domains) > 0 {

		if *printPtr {
			fmt.Print(sliceVal(domains))
		}

		fmt.Printf("[:] Writing %d domain(s) to file...\n", len(domains))

		if *jsonF {
			writeJson(domains, "domains.json")
		} else {
			writeLines(domains, "domains.txt")
		}
	} else {
		fmt.Println("[!] No domains found")
	}

	fmt.Println("[:] Done.")
}
