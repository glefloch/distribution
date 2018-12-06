package handlers

import (
	"fmt"
	"net/url"
	"strconv"
)

const (
	// default page size
	defaultPageSize = 100
	// query page size parameter name
	sizeParameterName = "n"
	// query page offset  parameter name
	offsetParameterName = "last"
)

// page represent pagining parameters extracted from an URL
type page struct {
	last string
	n    int
}

// pageParameters extract paging parameters from queryURL
func pagingParameters(queryURL *url.URL) page {
	q := queryURL.Query()
	lastEntry := q.Get(offsetParameterName)
	pageSize, err := strconv.Atoi(q.Get(sizeParameterName))
	if err != nil || pageSize < 0 {
		pageSize = defaultPageSize
	}
	return page{last: lastEntry, n: pageSize}
}

func pagingEnabled(queryURL *url.URL) bool {
	v := queryURL.Query()
	_, nParam := v[sizeParameterName]
	_, lastParam := v[offsetParameterName]
	return nParam || lastParam
}

// createLinkTag uses the originalURL from the request to create a new URL
// for the link header.
func (p *page) createLink(origURL string) (string, error) {
	calledURL, err := url.Parse(origURL)
	if err != nil {
		return "", err
	}

	v := url.Values{}
	v.Add(sizeParameterName, strconv.Itoa(p.n))
	v.Add(offsetParameterName, p.last)

	calledURL.RawQuery = v.Encode()

	calledURL.Fragment = ""
	urlStr := fmt.Sprintf("<%s>; rel=\"next\"", calledURL.String())

	return urlStr, nil
}
