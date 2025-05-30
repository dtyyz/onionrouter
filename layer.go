package onionrouter

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

type Map map[string]interface{} // helper for inline json objects
type List []interface{}         // helper for inline json arrays

const (
	LOG_ALL  int = iota // default
	LOG_NONE            // hide errors
)

// LOG_ALL or LOG_NONE
var LogLevel = LOG_ALL

// logger used for internal errors
var Logger = log.New(os.Stderr, "onion: ", log.LstdFlags|log.Lmsgprefix)

// simplify callback args
type Data struct {
	Writer  http.ResponseWriter
	Request *http.Request
}

// callback func wrapper
type Callback func(*Data)

// write status code
func (d *Data) Status(code int) {
	d.Writer.WriteHeader(code)
}

// write string response
func (d *Data) Write(str string) {
	io.WriteString(d.Writer, str)
}

// send redirect
func (d *Data) Redirect(url string, code int) {
	http.Redirect(d.Writer, d.Request, url, code)
}

// helper for params
func (d *Data) Param(id string) string {
	return d.Request.PathValue(id)
}

// sends and logs a generic '500 internal server error'
func (d *Data) Error(err error) {
	Logger.Println("internal server error:", err)
	d.Status(http.StatusInternalServerError)
}

// get json request data
func (d *Data) Json(v interface{}) error {
	obj := json.NewDecoder(d.Request.Body)
	obj.DisallowUnknownFields()

	err := obj.Decode(v)
	if err != nil {
		d.Status(http.StatusBadRequest)
		if LogLevel != LOG_NONE {
			Logger.Printf("invalid request %s", err)
		}
		return err
	}

	if obj.More() {
		d.Status(http.StatusBadRequest)
		if LogLevel != LOG_NONE {
			Logger.Println("extra data in request")
		}
		return err
	}

	return nil
}

// write json response
func (d *Data) WriteJson(v interface{}) error {
	b, err := json.Marshal(v)
	if err != nil {
		d.Error(fmt.Errorf("invalid json object %+v", v))
		return err
	}
	d.Write(string(b))
	return nil
}

// wrapper for handler func
func Route(pattern string, cb Callback) {
	http.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		data := &Data{w, r}
		cb(data)
	})
}
