package curlconverter

import (
	"strconv"
	"strings"

	"github.com/mdtosif/go-curlconverter/pkg/request"
)

type support struct {
	multipleURLs   bool
	dataReadsFile  bool
	queryReadsFile bool
	cookieFiles    bool
}

func getFirstRequest(reqs []*request.Request, warnings Warnings, support support) (*request.Request, Warnings, error) {
	if len(reqs) == 0 {
		return nil, warnings, ErrNoRequests
	}
	if len(reqs) > 1 {
		warnings = append(warnings, Warning{
			"next",
			"got " + strconv.Itoa(len(reqs)) + " curl requests, only converting the first one",
		})
	}
	req := reqs[0]
	warnings = warnIfPartsIgnored(req, warnings, support)
	return req, warnings, nil
}

func warnIfPartsIgnored(req *request.Request, warnings Warnings, support support) Warnings {
	if req == nil {
		return warnings
	}
	if len(req.URLs) > 1 && !support.multipleURLs {
		urls := make([]string, 0, len(req.URLs))
		for _, u := range req.URLs {
			urls = append(urls, strconv.Quote(u.OriginalURL))
		}
		warnings = append(warnings, Warning{
			"multiple-urls",
			"found " + strconv.Itoa(len(req.URLs)) + " URLs, only the first one will be used: " + strings.Join(urls, ", "),
		})
	}
	if req.BodyFile != "" && !support.dataReadsFile {
		warnings = append(warnings, Warning{
			"unsafe-data",
			"the generated data content is wrong, " + strconv.Quote("@"+req.BodyFile) +
				" means read the file " + strconv.Quote(req.BodyFile),
		})
	}
	if len(req.CookieFiles) > 0 && !support.cookieFiles {
		files := make([]string, 0, len(req.CookieFiles))
		for _, file := range req.CookieFiles {
			files = append(files, strconv.Quote(file))
		}
		warnings = append(warnings, Warning{
			"cookie-files",
			"passing a file for --cookie/-b is not supported: " + strings.Join(files, ", "),
		})
	}
	return warnings
}
