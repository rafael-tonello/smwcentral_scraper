package misc

import (
	"errors"
	"io"
	"net/http"
	"os"
	"strings"
)

func DerivateError(e error, msg string) error {
	prevErrors := e.Error()
	prevErrors = strings.ReplaceAll(prevErrors, "  >", "    >")
	err := errors.New(msg + "\n  > " + prevErrors)
	return err
}

func DerivateError2(msg string, e error) error {
	return DerivateError(e, msg)
}

func SeparateKeyAndValue(keyValuePair string, possibleCharSeps string) (string, string) {

	for _, char := range possibleCharSeps {
		if strings.Contains(keyValuePair, string(char)) {
			after, before, _ := strings.Cut(keyValuePair, string(char))
			return after, before
		}
	}
	return keyValuePair, ""
}

func DownloadFile(filepath string, url string) (err error) {

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		return errors.New("bad status: " + resp.Status)
	}

	// Writer the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}
