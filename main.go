package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"

	_ "github.com/denisenkom/go-mssqldb"
	"github.com/go-playground/validator"
	"github.com/golang/gddo/httputil/header"
)

type ClientRequest struct {
	Request_id  int    `json:"request_id"`
	Url_package []int  `json:"url_package" validate:"required"`
	Ip          string `json:"ip" validate:"ipv6,required"`
}
type GetResponse struct {
	Price float64 `json:"price" validate:"required"`
}

// Getting URL from localDB (ms sql)
func URLFromDB(id int) string {

	db, err := sql.Open("sqlserver", "vlad:123123v@HOME-PC:1433?database=test_1")
	if err != nil {
		log.Fatal("Error creating connection pool:" + err.Error())
	}
	defer db.Close()
	var queryString = fmt.Sprintf("select UrlString from test_1.dbo.UrlPackage where ID = %+v", id)
	row := db.QueryRow(queryString)

	var res string
	err = row.Scan(&res)
	if err != nil {
		return ""
	}
	return res
}

// Parsing POST body to ClientRequest struct
func ParsePOST(writer http.ResponseWriter, request *http.Request, obj interface{}) {
	if request.Header.Get("Content-type") != "" {
		value, _ := header.ParseValueAndParams(request.Header, "Content-type")
		if value != "application/json" {
			http.Error(writer, " ", http.StatusNoContent)
		}
	}
	decoder := json.NewDecoder(request.Body)
	err := decoder.Decode(&obj)
	if err != nil {
		http.Error(writer, " ", http.StatusNoContent)
	}
	validate := validator.New()
	err = validate.Struct(obj)
	if err != nil {
		http.Error(writer, " ", http.StatusNoContent)
	}
}

// Get responce on http.GET and parsing to FinalResponse
func Get(url string, j *GetResponse) {
	resp, err := http.Get(url)
	if err != nil {
		j.Price = 0
	}
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&j)
	if err != nil {
		j.Price = 0
	}
	validate := validator.New()
	err = validate.Struct(j)
	if err != nil {
		j.Price = 0
	}
}

// Handler that serve server
func createResponce(writer http.ResponseWriter, request *http.Request) {
	var resp ClientRequest
	ParsePOST(writer, request, &resp)
	//Array of Get responces from url_package
	var result []GetResponse
	//Counting how many urls from url_package is in database
	IDinDB := 0
	for _, item := range resp.Url_package {
		url := URLFromDB(item)
		if len(url) > 0 {
			getRes := GetResponse{}
			Get(url, &getRes)
			result = append(result, getRes)
			IDinDB += 1
		}
	}
	//Checking that any of url_package ID is in DB
	if IDinDB == 0 {
		http.Error(writer, "", http.StatusNoContent)
	} else {
		//Sort array by descending -> result[0] is max in array
		sort.SliceStable(result, func(i, j int) bool { return result[i].Price > result[j].Price })
		//Seting up responce
		writer.Header().Set("Content-type", "application/json")
		json.NewEncoder(writer).Encode(result[0])
	}
}
func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/test", createResponce)
	log.Print("Starting server on :4000...")
	err := http.ListenAndServe(":4000", mux)
	log.Fatal(err)
}
