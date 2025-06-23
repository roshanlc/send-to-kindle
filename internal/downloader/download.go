package downloader

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"resty.dev/v3"
)

const adsTerm = "ads.php?"

var (
	NilRestyClientErr     = errors.New("no resty clienty provided, expected a resty client reference")
	Non200StatusErr       = errors.New("recived a non 200 status code")
	NoLinkFoundAdsPageErr = errors.New("could not extract download link from ads page")
)

func Process(client *resty.Client, url string) error {
	if client == nil {
		return NilRestyClientErr
	}
	// 1. Check if the url is libgen ads site from which download link is to be extracted or it is download url itself
	// 2. if libgen ads site, extract the download link
	// 3. Proceed to download from the url
	// 4. Save the file to a directory

	downloadLink := url
	if isAdsPage(url) {
		resp, err := client.R().Get(url)
		if err != nil {
			return err
		}

		if resp.StatusCode() != http.StatusOK {
			return fmt.Errorf("%w, got: %d", Non200StatusErr, resp.StatusCode())
		}

		downloadLink, err = extractDownloadLink(resp)
		if err != nil {
			return err
		}
	}
	_ = downloadLink
	return nil

}

// isAdsPage checks if the given urls if an ads page (not advertisement, more like file description)
// which contains download link to file.
// Example: https://libgen.li/ads.php?md5=7e5412b8ece1fe49f7bfbc6e5ab77809
func isAdsPage(url string) bool {
	if strings.Contains(url, adsTerm) {
		return true
	}
	return false
}

// extractDownloadLink extracts download link from ads page of libgen by parsing the html doc.
// Example: https://libgen.li/get.php?md5=7e5412b8ece1fe49f7bfbc6e5ab77809&key=HIUDRYJW8JKVNH5A
func extractDownloadLink(resp *resty.Response) (string, error) {
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", err
	}
	val, exists := doc.Find("a[href] > h2").Parent().Attr("href")
	if !exists {
		return "", NoLinkFoundAdsPageErr
	}

	parsedURL, err := url.Parse(resp.Request.URL)
	if err != nil {
		return "", fmt.Errorf("could not parse url: %w", err)
	}

	link := fmt.Sprintf("%s://%s/%s", parsedURL.Scheme, parsedURL.Host, val)
	return link, nil
}

// SaveToDirectory saves the file to the given path
func SaveToDirectory(path string, content bytes.Buffer) error {

	return nil
}
