///////////////////////////////////////////////////////////////////////////////////////
//                     gesedels · a key-value API in Go · v0.0.0                     //
///////////////////////////////////////////////////////////////////////////////////////

package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"

	"go.etcd.io/bbolt"
)

///////////////////////////////////////////////////////////////////////////////////////
//                          part one · constants and globals                         //
///////////////////////////////////////////////////////////////////////////////////////

// DB is the global database connection object.
var DB *bbolt.DB

///////////////////////////////////////////////////////////////////////////////////////
//                      part two · string sanitisation functions                     //
///////////////////////////////////////////////////////////////////////////////////////

// IsPrivate returns true if a name string is surrounded with two leading underscores.
func IsPrivate(name string) bool {
	return strings.HasPrefix(name, "__") && strings.HasSuffix(name, "__")
}

// PairKey returns a lowercase pair key string from user and name strings.
func PairKey(user, name string) []byte {
	user = strings.ToLower(user)
	name = strings.ToLower(name)
	return []byte(user + ":" + name)
}

// PairValue returns a whitespace-trimmed pair value string.
func PairValue(text string) []byte {
	return []byte(strings.TrimSpace(text) + "\n")
}

///////////////////////////////////////////////////////////////////////////////////////
//                      part three · database handling functions                     //
///////////////////////////////////////////////////////////////////////////////////////

// DeletePair deletes an existing pair from a database.
func DeletePair(db *bbolt.DB, user, name string) error {
	return db.Update(func(tx *bbolt.Tx) error {
		if buck := tx.Bucket([]byte("main")); buck != nil {
			return buck.Delete(PairKey(user, name))
		}

		return nil
	})
}

// GetPair returns the value of an existing pair from a database and a boolean
// indicating if the pair exists.
func GetPair(db *bbolt.DB, user, name string) (string, bool, error) {
	var pval string
	var okay = false

	return pval, okay, db.View(func(tx *bbolt.Tx) error {
		if buck := tx.Bucket([]byte("main")); buck != nil {
			bytes := buck.Get(PairKey(user, name))
			pval = string(bytes)
			okay = bytes != nil
		}

		return nil
	})
}

// SetPair sets the value of a new or existing pair in a database.
func SetPair(db *bbolt.DB, user, name, pval string) error {
	return db.Update(func(tx *bbolt.Tx) error {
		buck, err := tx.CreateBucketIfNotExists([]byte("main"))
		if err != nil {
			return err
		}

		return buck.Put(PairKey(user, name), PairValue(pval))
	})
}

///////////////////////////////////////////////////////////////////////////////////////
//                        part four · http response functions                        //
///////////////////////////////////////////////////////////////////////////////////////

// WriteHTTP writes a plaintext response to a ResponseWriter.
func WriteHTTP(w http.ResponseWriter, code int, form string, elems ...any) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(code)
	fmt.Fprintf(w, form+"\n", elems...)
}

// WriteError writes a plaintext error response to a ResponseWriter.
func WriteError(w http.ResponseWriter, code int, form string, elems ...any) {
	form = fmt.Sprintf("server error %d: %s", code, form)
	WriteHTTP(w, code, form, elems...)
}

// WriteFailure writes a plaintext failure response to a ResponseWriter.
func WriteFailure(w http.ResponseWriter, code int, form string, elems ...any) {
	form = fmt.Sprintf("client error %d: %s", code, form)
	WriteHTTP(w, code, form, elems...)
}

///////////////////////////////////////////////////////////////////////////////////////
//                       part five · server endpoint functions                       //
///////////////////////////////////////////////////////////////////////////////////////

// GetIndex returns the index page.
func GetIndex(w http.ResponseWriter, r *http.Request) {
	WriteHTTP(w, http.StatusOK, "Hello.")
}

///////////////////////////////////////////////////////////////////////////////////////
//                         part six · main runtime functions                         //
///////////////////////////////////////////////////////////////////////////////////////

// try panics on a non-nil error.
func try(err error) {
	if err != nil {
		panic(err)
	}
}

// main runs the main Gesedels program.
func main() {
	// Define and parse command-line functions.
	fset := flag.NewFlagSet("gesedels", flag.ExitOnError)
	addr := fset.String("addr", "127.0.0.1:8080", "set server address")
	path := fset.String("path", "./gesedels.db", "set database path")
	fset.Parse(os.Args[1:])

	// Connect to and set database.
	db, err := bbolt.Open(*path, 0666, nil)
	try(err)
	DB = db

	// Initialise mux and register endpoints.
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", GetIndex)

	// Initialise and run server.
	srv := &http.Server{Addr: *addr, Handler: mux}
	try(srv.ListenAndServe())
}
