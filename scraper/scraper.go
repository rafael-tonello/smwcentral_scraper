package scraper

import (
	"io"
	"net/http"
	"smwex/misc"
	"smwex/pkv"
	"strconv"
	"strings"
	"sync"
)

type LinkFoundResult uint8

const (
	LF_DO_NOT_PROCESS LinkFoundResult = iota
	LF_PROCEED        LinkFoundResult = iota + 1
)

type DownloadedPage struct {
	Url     string
	Content string
}

type Scraper struct {
	Domains        []string
	OnPageDownload misc.Stream[DownloadedPage]
	Storage        *pkv.PrefixTreeKeyValue
	LinkFound      func(link string) LinkFoundResult
	ValidLinks     int
	VisitedLinks   int
	locker         sync.Mutex
}

func NewScraper(databaseFile string) *Scraper {
	ret := Scraper{
		OnPageDownload: *misc.NewStream[DownloadedPage](),
		LinkFound:      func(link string) LinkFoundResult { return LF_PROCEED },
		Storage:        pkv.New(databaseFile, 64),
	}

	count, _ := strconv.ParseInt(ret.Storage.Get("count", "0"), 10, 64)
	ret.ValidLinks = int(count)

	visitedLinks, _ := strconv.ParseInt(ret.Storage.Get("nextToConsume", "0"), 10, 64)
	ret.VisitedLinks = int(visitedLinks)

	return &ret
}

func (s *Scraper) AddLimitingDomain(domain string) {
	s.Domains = append(s.Domains, domain)
}

func (s *Scraper) Visit(initialLink string) {
	s.saveLink(initialLink)
	s.visitLinks()
}

func (s *Scraper) visitLinks() {

	for {
		if !s.visitNextLink() {
			break
		}
	}
}

func (s *Scraper) saveLink(link string) {
	s.locker.Lock()
	if s.linkIsAlreadySaved(link) {
		s.locker.Unlock()
		return
	}
	//check if the link is in the limiting domains
	if len(s.Domains) > 0 {
		found := false
		linkDomain := GetDomain(link)
		for _, domain := range s.Domains {
			if strings.Contains(linkDomain, domain) {
				found = true
			}
		}

		if !found {
			s.locker.Unlock()
			return
		}
	}

	if s.LinkFound(link) == LF_DO_NOT_PROCESS {
		s.locker.Unlock()
		return
	}

	count, _ := strconv.ParseInt(s.Storage.Get("count", "0"), 10, 64)
	s.Storage.Set("link"+strconv.FormatInt(count, 10), link)
	s.Storage.Set("savedLinks."+link, "1")
	count++
	s.Storage.Set("count", strconv.FormatInt(count, 10))
	s.ValidLinks = int(count)
	s.locker.Unlock()
}

func (s *Scraper) getNextLink() string {
	s.locker.Lock()
	index, _ := strconv.ParseInt(s.Storage.Get("nextToConsume", "0"), 10, 64)
	ret := s.Storage.Get("link"+strconv.FormatInt(index, 10), "")
	index++
	s.Storage.Set("nextToConsume", strconv.FormatInt(index, 10))
	s.locker.Unlock()
	return ret
}

func (s *Scraper) visitNextLink() bool {

	nextLink := s.getNextLink()
	if nextLink == "" {
		return false
	}
	s.visitAndProcessLink(nextLink)
	return true
}

func (s *Scraper) linkIsAlreadySaved(link string) bool {
	return s.Storage.Get("savedLinks."+link, "") != ""
}

func (s *Scraper) visitAndProcessLink(link string) {
	content := DownloadPage(link)
	s.VisitedLinks++
	s.OnPageDownload.Stream(DownloadedPage{Url: link, Content: content})
	if content != "" {
		links := extractLinks(content, link)
		//count := 0
		for _, link := range links {
			s.saveLink(link)
		}
	}
}

func extractLinks(content string, currentUrl string) []string {
	var result []string

	for {
		//find next 'href="' in the string
		hrefIndex := strings.Index(content, "href=\"")
		if hrefIndex == -1 {
			break
		}

		content = content[hrefIndex+6:]
		//find next '"' in the string
		endHrefIndex := strings.Index(content, "\"")
		if endHrefIndex == -1 {
			break
		}

		//extract the link
		link := content[:endHrefIndex]
		link = FixLink(link, currentUrl)

		result = append(result, link)
	}

	return result
}

func DownloadPage(link string) string {

	resp, error := http.Get(link)
	if error != nil {
		return ""
	}

	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return ""
		}

		return string(bodyBytes)
	}
	return ""

}

func FixLink(link string, currentUrl string) string {
	if strings.HasPrefix(link, "http") {
		return link
	}

	if !strings.HasPrefix(link, "/") {
		link = "/" + link
	}

	result := GetParentPage(currentUrl) + link

	return result
}

func GetParentPage(url string) string {
	//check if url ends with '/'
	if strings.HasSuffix(url, "/") {
		url = url[:len(url)-1]
	}

	//find last '/' in the url
	lastSlashIndex := strings.LastIndex(url, "/")
	if lastSlashIndex == -1 {
		return url
	}

	//cut string until the lastSlashIndex
	url = url[:lastSlashIndex]

	return url
}

func GetDomain(url string) string {
	//check contains '://' in the url
	if strings.Contains(url, "://") {
		//find the "://"
		firstSlashIndex := strings.Index(url, "://")
		if firstSlashIndex == -1 {
			return url
			//TODO: ERROR (STRING CONTAISN :// BUT CANT CUT?)
		}

		//remove the first part of string (untion the end of ://)
		url = url[firstSlashIndex+3:]
	}

	//find the first '/' in the url
	firstSlashIndex := strings.Index(url, "/")
	if firstSlashIndex == -1 {
		return url
	}

	url = url[:firstSlashIndex]

	return url
}
