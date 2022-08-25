package rpc

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func wrapperHandlerFunc(f func(http.ResponseWriter, *http.Request)) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		f(w, r)
	}
}

func wrapperHandler(h http.Handler) httprouter.Handle {
	f := func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r)
	}
	return wrapperHandlerFunc(f)
}

func cors(w *http.ResponseWriter, r *http.Request) (isOptions bool) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "POST")
	(*w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
	return ((*r).Method == "OPTIONS")
}

func WriteResponse(w http.ResponseWriter, jsn, path, ip string) {
	b, err := json.Marshal(jsn)
	if err != nil {
		WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
		fmt.Println(err.Error())
	} else {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		_, err := w.Write(b)
		if err != nil {
			fmt.Println(fmt.Errorf("error in RPC Handler WriteResponse: %v", err))
		}
	}
}

func WriteRaw(w http.ResponseWriter, jsn, path, ip string) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte(jsn))
	if err != nil {
		fmt.Println(fmt.Errorf("error in RPC Handler WriteRaw: %v", err))
	}
}

func WriteJSONResponse(w http.ResponseWriter, jsn, path, ip string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	var raw map[string]interface{}
	if err := json.Unmarshal([]byte(jsn), &raw); err != nil {
		WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
		fmt.Println(fmt.Errorf("error in RPC Handler WriteJSONResponse: %v", err))
		return
	}
	w.WriteHeader(http.StatusOK)
	err := json.NewEncoder(w).Encode(raw)
	if err != nil {
		fmt.Println(fmt.Errorf("error in RPC Handler WriteJSONResponse: %v", err))
		return
	}
}

func WriteJSONResponseWithCode(w http.ResponseWriter, jsn, path, ip string, code int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	var raw map[string]interface{}
	if err := json.Unmarshal([]byte(jsn), &raw); err != nil {
		WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
		fmt.Println(fmt.Errorf("error in RPC Handler WriteJSONResponse: %v", err))
		return
	}
	w.WriteHeader(code)
	err := json.NewEncoder(w).Encode(raw)
	if err != nil {
		fmt.Println(fmt.Errorf("error in RPC Handler WriteJSONResponse: %v", err))
		return
	}
}

func WriteErrorResponse(w http.ResponseWriter, errorCode int, errorMsg string) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(errorCode)
	err := json.NewEncoder(w).Encode(&rpcError{
		Code:    errorCode,
		Message: errorMsg,
	})
	if err != nil {
		fmt.Println(fmt.Errorf("error in RPC Handler WriteErrorResponse: %v", err))
	}
}

func WriteOKResponse(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
}
