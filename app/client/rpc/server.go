package rpc

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"
)

//TODO(deblasis): still unused, need to gather feedback on RPC routes below before implementing everything, for now I copied some code from v0 over as a blueprint for style
// currently perhaps only GetRoutes is important in this file

func StartRPC(port string, timeout uint64) {
	log.Printf("Starting RPC on port %s...\n", port)

	routes := GetRoutes()

	srv := &http.Server{
		ReadTimeout:       30 * time.Second,
		ReadHeaderTimeout: 20 * time.Second,
		WriteTimeout:      60 * time.Second,
		Addr:              ":" + port,
		Handler:           http.TimeoutHandler(Router(routes), time.Duration(timeout)*time.Millisecond, "Server Timeout Handling Request"),
	}
	log.Fatal(srv.ListenAndServe())
}

func Router(routes Routes) *httprouter.Router {
	router := httprouter.New()
	for _, route := range routes {
		router.Handle(route.Method, route.Path, route.HandlerFunc)
	}
	return router
}

func cors(w *http.ResponseWriter, r *http.Request) (isOptions bool) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "POST")
	(*w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
	return ((*r).Method == "OPTIONS")
}

type Route struct {
	Name        string
	Method      string
	Path        string
	HandlerFunc httprouter.Handle
}

type Routes []Route

func GetRoutes() Routes {
	routes := Routes{
		Route{Name: "Send", Method: "POST", Path: "/v1/client/send", HandlerFunc: TODOHandler},
		Route{Name: "Stake", Method: "POST", Path: "/v1/client/stake", HandlerFunc: TODOHandler},
		Route{Name: "EditStake", Method: "POST", Path: "/v1/client/editstake", HandlerFunc: TODOHandler},
		Route{Name: "UnStake", Method: "POST", Path: "/v1/client/unstake", HandlerFunc: TODOHandler},
		Route{Name: "Unpause", Method: "POST", Path: "/v1/client/unpause", HandlerFunc: TODOHandler},
		Route{Name: "SendRaw", Method: "POST", Path: "/v1/client/sendraw", HandlerFunc: TODOHandler},
		Route{Name: "ChangeParameter", Method: "POST", Path: "/v1/gov/changeparam", HandlerFunc: TODOHandler},
	}
	return routes
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func PopModel(_ http.ResponseWriter, r *http.Request, _ httprouter.Params, model interface{}) error {
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	if err != nil {
		return err
	}
	if len(body) == 0 {
		return nil
	}
	if err := r.Body.Close(); err != nil {
		return err
	}
	if err := json.Unmarshal(body, model); err != nil {
		return err
	}
	return nil
}

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

func TODOHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	WriteErrorResponse(w, 503, "not implemented")
}
