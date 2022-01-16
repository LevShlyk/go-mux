package main_test

import (
	"bytes"
	"encoding/json"
	"log"
	"lshlyk/case"
	"lshlyk/case/internal/shorter"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
)

var a main.App

func TestMain(m *testing.M) {
	a.Initialize(
		os.Getenv("DB"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_DBNAME"),
	)

	ensureTableExists()
	code := m.Run()
	clearTable()
	os.Exit(code)
}

func ensureTableExists() {
	if _, err := a.DatabaseSQL.Exec(tableCreationQuery); err != nil {
		log.Fatal(err)
	}
}

func clearTable() {
	a.DatabaseSQL.Exec("DELETE FROM shorts")
	a.DatabaseSQL.Exec("ALTER SEQUENCE shorts_id_seq RESTART WITH 1")
}

const tableCreationQuery = `CREATE TABLE IF NOT EXISTS shorts
(
    id BIGSERIAL,
    source TEXT NOT NULL,
    short VARCHAR(10) NOT NULL,
    CONSTRAINT shorts_pkey PRIMARY KEY (id)
)`

func TestEmptyTable(t *testing.T) {
	clearTable()

	req, _ := http.NewRequest("GET", "/shorts", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	if body := response.Body.String(); body != "[]" {
		t.Errorf("Expected an empty array. Got %s", body)
	}
}

func executeRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	a.Router.ServeHTTP(rr, req)

	return rr
}

func checkResponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected response code %d. Got %d\n", expected, actual)
	}
}

func TestGetNonExistentShort(t *testing.T) {
	clearTable()

	req, _ := http.NewRequest("GET", "/shorts/11", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusNotFound, response.Code)

	var m map[string]string
	json.Unmarshal(response.Body.Bytes(), &m)
	if m["error"] != "Short not found" {
		t.Errorf("Expected the 'error' key of the response to be set to 'Short not found'. Got '%s'", m["error"])
	}
}

func TestCreateShort(t *testing.T) {

	clearTable()

	var jsonStr = []byte(`{"source":"http://google.com"}`)
	req, _ := http.NewRequest("POST", "/shorts", bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	response := executeRequest(req)
	checkResponseCode(t, http.StatusCreated, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	if m["source"] != "http://google.com" {
		t.Errorf("Expected short source to be 'http://google.com'. Got '%s'", m["source"])
	}

	if m["short"] != "aaaaaaaaaa" {
		t.Errorf("Expected short short to be 'aaaaaaaaaa'. Got '%s'", m["short"])
	}

	// the id is compared to 1.0 because JSON unmarshaling converts numbers to
	// floats, when the target is a map[string]interface{}
	if m["id"] != 1.0 {
		t.Errorf("Expected short ID to be '1'. Got '%v'", m["id"])
	}
}

func TestGetShort(t *testing.T) {
	clearTable()
	addShorts(1)

	req, _ := http.NewRequest("GET", "/shorts/1", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)
}

func TestGetShortByShort(t *testing.T) {
	clearTable()
	addShorts(1)

	req, _ := http.NewRequest("GET", "/shorts/aaaaaaaaaa", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)
}

func addShorts(count int) {
	if count < 1 {
		count = 1
	}

	for i := 1; i <= count; i++ {
		s := shorter.BuildShorter()
		a.DatabaseSQL.Exec("INSERT INTO shorts(source, short) VALUES($1, $2)", "http://googlepage"+strconv.Itoa(i)+".com", s.GetShortByID(uint64(i)))
	}
}

func TestUpdateShort(t *testing.T) {

	clearTable()
	addShorts(1)

	req, _ := http.NewRequest("GET", "/shorts/1", nil)
	response := executeRequest(req)
	var originalShort map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &originalShort)

	var jsonStr = []byte(`{"source":"http://ozon.ru"}`)
	req, _ = http.NewRequest("PUT", "/shorts/1", bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	response = executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	if m["id"] != originalShort["id"] {
		t.Errorf("Expected the id to remain the same (%v). Got %v", originalShort["id"], m["id"])
	}

	if m["source"] == originalShort["source"] {
		t.Errorf("Expected the source to change from '%v' to '%v'. Got '%v'", originalShort["source"], m["source"], m["source"])
	}
}

func TestDeleteShort(t *testing.T) {
	clearTable()
	addShorts(1)

	req, _ := http.NewRequest("GET", "/shorts/1", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	req, _ = http.NewRequest("DELETE", "/shorts/1", nil)
	response = executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	req, _ = http.NewRequest("GET", "/shorts/1", nil)
	response = executeRequest(req)
	checkResponseCode(t, http.StatusNotFound, response.Code)
}