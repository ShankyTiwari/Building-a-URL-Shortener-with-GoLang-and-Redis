package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/gorilla/mux"
	"github.com/ventu-io/go-shortid"
)

// ErrorResponse is interface for sending error message with code.
type ErrorResponse struct {
	Code    int
	Message string
}

// RenderHome Rendering the Home Page
func RenderHome(response http.ResponseWriter, request *http.Request) {
	http.ServeFile(response, request, "views/index.html")
}

// RedirectIfURLFound This function will redirect to actual website
func RedirectIfURLFound(response http.ResponseWriter, request *http.Request) {
	var httpError = ErrorResponse{
		Code: http.StatusInternalServerError, Message: "Something went wrong at our end",
	}
	urlShortCode := mux.Vars(request)["urlShortCode"]
	if urlShortCode == "" {
		httpError.Code = http.StatusBadRequest
		httpError.Message = "URL Code can't be empty"
		returnErrorResponse(response, request, httpError)
	} else {
		ActualURL, err := Client.Get(urlShortCode).Result()
		if ActualURL == "" || err != nil {
			httpError.Code = http.StatusNotFound
			httpError.Message = "An invalid/expired URL Code found"
			returnErrorResponse(response, request, httpError)
		} else {
			http.Redirect(response, request, ActualURL, http.StatusSeeOther)
		}
	}
}

// GetShortURLHandler This function will return the response based ono user found in Database
func GetShortURLHandler(response http.ResponseWriter, request *http.Request) {
	type URLRequestObject struct {
		URL string `json:"url"`
	}
	type URLCollection struct {
		ActualURL string
		ShortURL  string
	}
	type SuccessResponse struct {
		Code     int
		Message  string
		Response URLCollection
	}
	var urlRequest URLRequestObject
	var httpError = ErrorResponse{
		Code: http.StatusInternalServerError, Message: "Something went wrong at our end",
	}

	decoder := json.NewDecoder(request.Body)
	err := decoder.Decode(&urlRequest)

	if err != nil {
		httpError.Message = "URL can't be empty"
		returnErrorResponse(response, request, httpError)
	} else if !isURL(urlRequest.URL) {
		httpError.Message = "An invalid URL found, provide a valid URL"
		returnErrorResponse(response, request, httpError)
	} else {
		uniqueID, idError := shortid.Generate()
		if idError != nil {
			returnErrorResponse(response, request, httpError)
		} else {
			err := Client.Set(uniqueID, urlRequest.URL, 0).Err()
			if err != nil {
				fmt.Println(err)
				returnErrorResponse(response, request, httpError)
			}
			var successResponse = SuccessResponse{
				Code:    http.StatusOK,
				Message: "Short URL generated",
				Response: URLCollection{
					ActualURL: urlRequest.URL,
					ShortURL:  request.Host + "/" + uniqueID,
				},
			}
			jsonResponse, err := json.Marshal(successResponse)
			if err != nil {
				panic(err)
			}
			response.Header().Set("Content-Type", "application/json")
			response.WriteHeader(successResponse.Code)
			response.Write(jsonResponse)
		}
	}
}

// Helper function to handle the HTTP response

func isURL(str string) bool {
	u, err := url.Parse(str)
	return err == nil && u.Scheme != "" && u.Host != ""
}

func returnErrorResponse(response http.ResponseWriter, request *http.Request, errorMesage ErrorResponse) {
	httpResponse := &ErrorResponse{Code: errorMesage.Code, Message: errorMesage.Message}
	jsonResponse, err := json.Marshal(httpResponse)
	if err != nil {
		panic(err)
	}
	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(errorMesage.Code)
	response.Write(jsonResponse)
}
