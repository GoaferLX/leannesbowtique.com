package controllers

import (
	"net/http"
	"net/url"

	"github.com/gorilla/schema"
)

// Helper functions for form handling

// Parse forms from /POST requests
func parsePostForm(r *http.Request, dst interface{}) error {
	if err := r.ParseForm(); err != nil {
		return err
	}
	return parseFormValues(r.PostForm, dst)
}

// Parse forms from /GET requests
// Used for retreiving URL params
func parseGetForm(r *http.Request, dst interface{}) error {
	if err := r.ParseForm(); err != nil {
		return err
	}
	return parseFormValues(r.Form, dst)
}

// Fill destination struct with parsed values
func parseFormValues(values url.Values, dst interface{}) error {
	var dec *schema.Decoder = schema.NewDecoder()
	dec.IgnoreUnknownKeys(true)
	if err := dec.Decode(dst, values); err != nil {
		return err
	}
	return nil
}
