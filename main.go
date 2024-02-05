package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"gopkg.in/yaml.v2"
)

// Yaml
type Endpoint struct {
	Name    string            `yaml:"name"`
	URL     string            `yaml:"url"`
	Method  string            `yaml:"method"`
	Headers map[string]string `yaml:"headers"`
	Body    string            `yaml:"body"`
}

type DomainAvailability struct {
	Up   int
	Down int
}

var (
	endpoints          []Endpoint
	domainAvailability = make(map[string]*DomainAvailability)
)

func checkEndpointHealth(endpoint Endpoint) {
	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest(endpoint.Method, endpoint.URL, nil)
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		return
	}

	for key, value := range endpoint.Headers {
		req.Header.Add(key, value)
	}

	startTime := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error checking endpoint: %s - %v\n", endpoint.Name, err)
		domain := getDomainName(endpoint.URL)
		domainAvailability[domain].Down++
		return
	}
	latency := time.Since(startTime)

	if resp.StatusCode >= 200 && resp.StatusCode < 300 && latency.Milliseconds() < 500 {
		domain := getDomainName(endpoint.URL)
		domainAvailability[domain].Up++
	} else {
		domain := getDomainName(endpoint.URL)
		domainAvailability[domain].Down++
	}
}

func displayAvailability() {
	for domain, counts := range domainAvailability {
		totalChecks := counts.Up + counts.Down
		if totalChecks > 0 {
			availabilityPercent := float64(counts.Up) / float64(totalChecks) * 100
			fmt.Printf("%s has %.0f%% availability\n", domain, availabilityPercent)
		}
	}
}

func getDomainName(endpointURL string) string {
	parsedURL, err := url.Parse(endpointURL)
	if err != nil {
		fmt.Printf("Error parsing URL: %v\n", err)
		return ""
	}
	return parsedURL.Hostname()
}

func runHealthChecks() {
	for {
		for _, endpoint := range endpoints {
			checkEndpointHealth(endpoint)
		}
		displayAvailability()
		time.Sleep(15 * time.Second)
	}
}

func main() {
	// Assuming 'endpoints.yaml' is your YAML file
	yamlFile, err := ioutil.ReadFile("endpoints.yaml")
	if err != nil {
		fmt.Printf("Error reading YAML file: %s\n", err)
		return
	}

	err = yaml.Unmarshal(yamlFile, &endpoints)
	if err != nil {
		fmt.Printf("Error parsing YAML file: %s\n", err)
		return
	}

	// Initialize domainAvailability map
	for _, endpoint := range endpoints {
		domain := getDomainName(endpoint.URL)
		if domainAvailability[domain] == nil {
			domainAvailability[domain] = &DomainAvailability{}
		}
	}

	runHealthChecks()
}
