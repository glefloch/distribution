package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/docker/distribution"
	"github.com/docker/distribution/registry/api/errcode"
	"github.com/docker/distribution/registry/api/v2"
	"github.com/gorilla/handlers"
)

// tagsDispatcher constructs the tags handler api endpoint.
func tagsDispatcher(ctx *Context, r *http.Request) http.Handler {
	tagsHandler := &tagsHandler{
		Context: ctx,
	}

	return handlers.MethodHandler{
		"GET": http.HandlerFunc(tagsHandler.GetTags),
	}
}

// tagsHandler handles requests for lists of tags under a repository name.
type tagsHandler struct {
	*Context
}

type tagsAPIResponse struct {
	Name string   `json:"name"`
	Tags []string `json:"tags"`
}

// GetTags returns a json list of tags for a specific image name.
func (th *tagsHandler) GetTags(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	tagService := th.Repository.Tags(th)
	tags, err := tagService.All(th)
	if err != nil {
		switch err := err.(type) {
		case distribution.ErrRepositoryUnknown:
			th.Errors = append(th.Errors, v2.ErrorCodeNameUnknown.WithDetail(map[string]string{"name": th.Repository.Named().Name()}))
		case errcode.Error:
			th.Errors = append(th.Errors, err)
		default:
			th.Errors = append(th.Errors, errcode.ErrorCodeUnknown.WithDetail(err))
		}
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	// retrieve pagination parameters
	if pagingEnabled(r.URL) {
		p := pagingParameters(r.URL)

		page := make([]string, p.n)
		filled, err := pageFilter(tags, page, p.last)
		if err == io.EOF {
			p.last = page[len(page)-1]
			urlStr, err := p.createLink(r.URL.String())
			if err != nil {
				th.Errors = append(th.Errors, errcode.ErrorCodeUnknown.WithDetail(err))
				return
			}
			w.Header().Set("Link", urlStr)
		}
		tags = page[0:filled]
	}

	enc := json.NewEncoder(w)
	if err := enc.Encode(tagsAPIResponse{
		Name: th.Repository.Named().Name(),
		Tags: tags,
	}); err != nil {
		th.Errors = append(th.Errors, errcode.ErrorCodeUnknown.WithDetail(err))
		return
	}
}

// pageFilter fills 'page' with tags from the 'last' up to the size of 'page'
// and return 'n' for the number of tags which were filled.
//'last' contains an offset in the tags and 'err' will be set to io.EOF
// if there are no more tags to obtain
func pageFilter(tags, page []string, last string) (n int, err error) {
	var foundTags []string
	var lastPage bool
	for _, tag := range tags {
		if strings.Compare(tag, last) > 0 {
			foundTags = append(foundTags, tag)
			if len(foundTags) == len(page) {
				lastPage = true
				break
			}
		}
	}

	n = copy(page, foundTags)
	if !lastPage {
		return n, io.EOF
	}
	return n, nil
}
