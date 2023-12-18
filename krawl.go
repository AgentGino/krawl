package krawl

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/chromedp/chromedp"
	// Additional imports may be required for HTTP requests, concurrency handling, etc.
)

// Global variables
var (
	baseDomain   string
	visitedURLs  = make(map[string]bool)
	mux          sync.Mutex
	matchPattern *regexp.Regexp
)

// RunnerOutput represents the data structure for storing information from a web page.
type RunnerOutput struct {
	URL     string
	Content string
	Title   string
}

// RunnerInput represents the input parameters for the Runner function.
type RunnerInput struct {
	StartUrl     string
	PathPatterns []string
	Depth        int
}

type CrawlPageInput struct {
	PageURL      string
	Depth        int
	PathPatterns []string
	RunnerOutput *[]*RunnerOutput
}

// validateURL checks if the provided string is a valid URL.
func validateURL(u string) error {
	url, err := url.ParseRequestURI(u)
	fmt.Println(url)

	// if error or url is nill or invalid throw error
	if err != nil || url == nil || url.Host == "" {
		return fmt.Errorf("invalid URL")

	}
	return nil
}

func validatePathPatterns(input RunnerInput) error {
	// If the array is empty, no validation is needed
	if len(input.PathPatterns) == 0 {
		return nil
	}

	// Parse the domain from StartUrl
	startURLDomain, err := extractDomain(input.StartUrl)
	if err != nil {
		return fmt.Errorf("error extracting domain from start URL: %v", err)
	}

	// Validate each pattern
	for _, pattern := range input.PathPatterns {
		patternDomain, err := extractDomain(pattern)
		if err != nil {
			// If the pattern is not a valid URL, assume it's just a path
			continue
		}

		if patternDomain == "" {
			continue
		}

		// Check if the domain in the pattern matches the StartUrl domain
		if patternDomain != startURLDomain {
			return fmt.Errorf("domain in pattern %s does not match start URL domain", pattern)
		}
	}

	return nil
}

// extractDomain extracts the domain from a URL.
func extractDomain(urlStr string) (string, error) {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return "", err
	}
	return parsedURL.Hostname(), nil
}

// validateInputs performs validation on the RunnerInput struct.
func validateInputs(input RunnerInput) error {
	if err := validateURL(input.StartUrl); err != nil {
		return err
	}
	if err := validatePathPatterns(input); err != nil {
		return err
	}

	return nil
}

func CrawlPage(
	ctx context.Context,
	Input CrawlPageInput,
) error {
	if Input.Depth <= 0 {
		return nil
	}

	mux.Lock()
	if visitedURLs[Input.PageURL] {
		mux.Unlock()
		return nil
	}
	visitedURLs[Input.PageURL] = true
	mux.Unlock()

	pageData, err := ParsePage(ctx, Input.PageURL)
	if err != nil {
		return err
	}

	*Input.RunnerOutput = append(*Input.RunnerOutput, pageData)

	// Extract and filter links
	var links []string
	err = chromedp.Run(ctx,
		chromedp.Evaluate(`Array.from(document.querySelectorAll('a[href]')).map(a => a.href);`, &links),
	)
	if err != nil {
		return err
	}

	for _, link := range links {
		if IsInternalLink(link,
			RunnerInput{
				StartUrl:     Input.PageURL,
				PathPatterns: Input.PathPatterns,
			},
		) && !strings.Contains(link, "#") {
			err := CrawlPage(
				ctx,
				CrawlPageInput{
					PageURL:      link,
					Depth:        Input.Depth - 1,
					RunnerOutput: Input.RunnerOutput,
				},
			)
			if err != nil {
				fmt.Println("Error crawling %s: %v\n", link, err)
			}
		}
	}

	return nil
}

func ParsePage(ctx context.Context, url string) (*RunnerOutput, error) {
	var title, data string
	// println("  Parsing : ", url)
	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.WaitVisible(`body`, chromedp.ByQuery),
		chromedp.Title(&title),
		chromedp.Text(`body`, &data, chromedp.ByQuery),
	)
	if err != nil {
		return nil, err
	}

	return &RunnerOutput{
		Title:   title,
		URL:     url,
		Content: data,
	}, nil
}

// Update the isInternalLink function to also check the match pattern
func IsInternalLink(link string, Input RunnerInput) bool {
	parsedURL, err := url.Parse(link)
	if err != nil {
		return false
	}
	parsedBaseURL, err := url.Parse(Input.StartUrl)
	if err != nil {
		return false
	}

	// default pattern match is false
	patternMatched := false

	// compile regex for each pattern
	for _, pattern := range Input.PathPatterns {
		matchPattern, err = regexp.Compile(pattern)
		if err != nil {
			return false
		}
		if matchPattern.MatchString(link) {
			patternMatched = true
		}
	}

	return parsedURL.Hostname() == parsedBaseURL.Hostname() && patternMatched
}

// Runner starts a web crawling/processing task with given parameters.
// Returns a slice of RunnerOutput containing information from processed pages, or an error if something goes wrong.
func Runner(input RunnerInput) ([]*RunnerOutput, error) {

	// Set default depth to 3
	depth := input.Depth
	if depth < 1 {
		depth = 3
	}

	// Collecting data
	var allData []*RunnerOutput

	// Validate inputs
	if err := validateInputs(input); err != nil {
		return nil, err
	}

	// fmt.Println("  ✔️  Runner started with valid inputs")

	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 60*time.Minute)
	defer cancel()

	if err := CrawlPage(ctx, CrawlPageInput{
		PageURL:      input.StartUrl,
		Depth:        depth,
		PathPatterns: input.PathPatterns,
		RunnerOutput: &allData,
	},
	); err != nil {
		return nil, fmt.Errorf("Crawl Error", err)
	}

	// fmt.Println("  ✔️  Runner finished successfully")
	return allData, nil
}
