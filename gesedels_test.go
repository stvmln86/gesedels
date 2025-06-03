///////////////////////////////////////////////////////////////////////////////////////
//                     gesedels · unit tests and helper functions                    //
///////////////////////////////////////////////////////////////////////////////////////

package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.etcd.io/bbolt"
)

///////////////////////////////////////////////////////////////////////////////////////
//                        part zero · testing helper functions                       //
///////////////////////////////////////////////////////////////////////////////////////

// mockPairs is a map of mock database pairs for unit testing.
var mockPairs = map[string]string{
	"0000:alpha": "Alpha.\n",
	"0000:bravo": "Bravo.\n",
}

// getResponse returns the status code and getResponse body of a ResponseRecorder.
func getResponse(w *httptest.ResponseRecorder) (int, string) {
	rslt := w.Result()
	body, _ := io.ReadAll(rslt.Body)
	return rslt.StatusCode, string(body)
}

// mockDB returns a temporary mock databae populated with mockPairs.
func mockDB(t *testing.T) *bbolt.DB {
	dest := filepath.Join(t.TempDir(), "test.db")
	db, _ := bbolt.Open(dest, 0666, nil)

	db.Update(func(tx *bbolt.Tx) error {
		buck, _ := tx.CreateBucket([]byte("main"))
		for pkey, pval := range mockPairs {
			buck.Put([]byte(pkey), []byte(pval))
		}

		return nil
	})

	return db
}

///////////////////////////////////////////////////////////////////////////////////////
//                          part one · constants and globals                         //
///////////////////////////////////////////////////////////////////////////////////////

///////////////////////////////////////////////////////////////////////////////////////
//                      part two · string sanitisation functions                     //
///////////////////////////////////////////////////////////////////////////////////////

func TestIsPrivate(t *testing.T) {
	// success - true
	ok := IsPrivate("__test__")
	assert.True(t, ok)

	// success - false
	for _, name := range []string{"__test", "test__", "test"} {
		ok := IsPrivate(name)
		assert.False(t, ok)
	}
}

func TestPairKey(t *testing.T) {
	// success
	pkey := PairKey("USER", "NAME")
	assert.Equal(t, []byte("user:name"), pkey)
}

func TestPairValue(t *testing.T) {
	// success
	pval := PairValue("\tValue.\n")
	assert.Equal(t, []byte("Value.\n"), pval)
}

///////////////////////////////////////////////////////////////////////////////////////
//                      part three · database handling functions                     //
///////////////////////////////////////////////////////////////////////////////////////

func TestDeletePair(t *testing.T) {
	// setup
	db := mockDB(t)

	// success
	err := DeletePair(db, "0000", "alpha")
	assert.NoError(t, err)

	// success - check database
	db.View(func(tx *bbolt.Tx) error {
		buck := tx.Bucket([]byte("main"))
		bytes := buck.Get([]byte("0000:alpha"))
		assert.Nil(t, bytes)
		return nil
	})
}

func TestGetPair(t *testing.T) {
	// setup
	db := mockDB(t)

	// success - pair exists
	pval, ok, err := GetPair(db, "0000", "alpha")
	assert.Equal(t, "Alpha.\n", pval)
	assert.True(t, ok)
	assert.NoError(t, err)

	// success - pair does not exist
	pval, ok, err = GetPair(db, "0000", "nope")
	assert.Empty(t, pval)
	assert.False(t, ok)
	assert.NoError(t, err)
}

func TestSetPair(t *testing.T) {
	// setup
	db := mockDB(t)

	// success
	err := SetPair(db, "0000", "test", "Test.\n")
	assert.NoError(t, err)

	// success - check database
	db.View(func(tx *bbolt.Tx) error {
		buck := tx.Bucket([]byte("main"))
		bytes := buck.Get([]byte("0000:test"))
		assert.Equal(t, []byte("Test.\n"), bytes)
		return nil
	})
}

///////////////////////////////////////////////////////////////////////////////////////
//                        part four · http response functions                        //
///////////////////////////////////////////////////////////////////////////////////////

func TestWriteHTTP(t *testing.T) {
	// setup
	w := httptest.NewRecorder()

	// success
	WriteHTTP(w, http.StatusOK, "%s", "test")
	code, body := getResponse(w)
	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, "test\n", body)
}

func TestWriteError(t *testing.T) {
	// setup
	w := httptest.NewRecorder()

	// success
	WriteError(w, http.StatusInternalServerError, "%s", "test")
	code, body := getResponse(w)
	assert.Equal(t, http.StatusInternalServerError, code)
	assert.Equal(t, "server error 500: test\n", body)
}

func TestWriteFailure(t *testing.T) {
	// setup
	w := httptest.NewRecorder()

	// success
	WriteFailure(w, http.StatusBadRequest, "%s", "test")
	code, body := getResponse(w)
	assert.Equal(t, http.StatusBadRequest, code)
	assert.Equal(t, "client error 400: test\n", body)
}

///////////////////////////////////////////////////////////////////////////////////////
//                        part five · server type and methods                        //
///////////////////////////////////////////////////////////////////////////////////////

///////////////////////////////////////////////////////////////////////////////////////
//                         part six · server endpoint methods                        //
///////////////////////////////////////////////////////////////////////////////////////

///////////////////////////////////////////////////////////////////////////////////////
//                        part seven · main runtime functions                        //
///////////////////////////////////////////////////////////////////////////////////////
