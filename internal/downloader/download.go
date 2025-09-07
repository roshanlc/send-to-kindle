package downloader

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/roshanlc/send-to-kindle/internal/helper"
	"resty.dev/v3"
)

const adsTerm = "ads.php?"

var downloadDir = "./downloads" // should be configurable, take from .env file // Also add a cleanup logic cron

var (
	NilRestyClientErr     = errors.New("no resty clienty provided, expected a resty client reference")
	Non200StatusErr       = errors.New("recived a non 200 status code")
	NoLinkFoundAdsPageErr = errors.New("could not extract download link from ads page")
)

// SetDownloadDir sets the directory for saving the downloaded file
func SetDownloadDirectory(dir string) {
	downloadDir = dir
}

// Process takes the url and attempts to download the file
func Process(ctx context.Context, client *resty.Client, url string) (string, context.Context, error) {
	// 1. Check if the url is libgen ads site from which download link is to be extracted or it is download url itself
	// 2. if libgen ads site, extract the download link
	// 3. Proceed to download from the url
	// 4. Save the file to a directory
	if client == nil {
		return "", ctx, NilRestyClientErr
	}

	taskID := helper.GetIDFromContext(ctx).String()
	downloadLink := url
	if isAdsPage(url) {
		slog.Info("Extracting download link from ads page:", slog.String("url", url), slog.String("taskID", taskID))
		resp, err := client.R().Get(url)
		if err != nil {
			return "", ctx, err
		}

		if resp.StatusCode() != http.StatusOK {
			return "", ctx, fmt.Errorf("%w, got: %d", Non200StatusErr, resp.StatusCode())
		}

		downloadLink, err = extractDownloadLink(resp)
		if err != nil {
			return "", ctx, err
		}
	}

	return downloadAndSave(ctx, client, downloadLink)
}

// downloadAndSave downloads a file form the provided link and saves it under
// the downloads(read from env var during start) directory
func downloadAndSave(ctx context.Context, client *resty.Client, url string) (string, context.Context, error) {
	resp, err := client.R().Get(url)
	if err != nil {
		return "", ctx, err
	}
	taskID := helper.GetIDFromContext(ctx).String()
	slog.Info("Downloaded file from url", slog.String("url", url), slog.String("taskID", taskID))

	contentDisposition := resp.Header().Get("Content-Disposition")
	var filename string

	if contentDisposition != "" {
		_, params, err := mime.ParseMediaType(contentDisposition)
		if err == nil && params["filename"] != "" {
			filename = params["filename"]
		}
	}

	// Fallback to use task ID
	if filename == "" {
		filename = taskID // taskID as fallback name
		t, _, _ := mime.ParseMediaType(resp.Header().Get("Content-Type"))
		if t != "" {
			exts, _ := mime.ExtensionsByType(t)
			if len(exts) > 0 {
				filename += exts[0]
			}
		}
	}

	filename = strings.TrimSpace(filename) // trim the spaces
	filePath := downloadDir + "/" + filename
	out, err := os.Create(filePath)

	if err != nil {
		return "", ctx, fmt.Errorf("error while creating file, %w", err)
	}

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return "", ctx, fmt.Errorf("error while saving response, %w", err)
	}

	slog.Info("Saved file", slog.String("filepath", filePath), slog.String("taskID", taskID))
	newCtx := helper.NewContextWithFilePath(ctx, filePath)
	return filename, newCtx, nil
}

// deleteDownloadedFile deletes the download file
func DeleteDownloadedFile(path string) error {
	err := os.Remove(path)
	if err != nil {
		return err
	}
	slog.Info("Deleted file", slog.String("filepath", path))
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
