package mcpemapcore

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

type createMapRequest struct {
	Map NewMap
}

type createMapResponse struct {
	Code  int
	MapID string
	Error string
}

func HandleCreateMap(wr http.ResponseWriter, r *http.Request) {
	userid := r.Header.Get("userid")
	u, err := LoadUserInfo(userid)

	mapResponse := createMapResponse{}
	theMapRequest, err := deserializeRequest(r)
	if err != nil {
		writeError(&mapResponse, err, wr)
		return
	}
	mapService := NewCreateMapService()
	mapResponse.MapID, err = mapService.CreateMap(u, &theMapRequest.Map)

	if err != nil {
		writeError(&mapResponse, err, wr)
		return
	} else {
		writeSuccess(&mapResponse, wr)
	}
}

func serializeResponse(r createMapResponse) (string, error) {
	st, err := json.Marshal(r)
	if err != nil {
		return "", err
	}
	return string(st), nil
}

func deserializeRequest(r *http.Request) (*createMapRequest, error) {
	bytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	var theMap createMapRequest
	err = json.Unmarshal(bytes, &theMap)
	if err != nil {
		return nil, err
	}
	return &theMap, nil
}

func writeSuccess(response *createMapResponse, wr http.ResponseWriter) {
	response.Code = http.StatusOK
	sr, err := serializeResponse(*response)
	if err != nil {
		//TODO: handle this log properly by using system error logging
		log.Fatal(err.Error())
		wr.WriteHeader(http.StatusBadRequest)
	}
	wr.WriteHeader(http.StatusOK)
	fmt.Fprintf(wr, "%v", sr)
}

func writeError(response *createMapResponse, err error, wr http.ResponseWriter) {
	//TODO replace with an interface to make generic
	wr.WriteHeader(http.StatusBadRequest)
	response.Error = err.Error()
	response.Code = http.StatusBadRequest
	//TODO make this a member of response struct so it can be made generic
	sr, err := serializeResponse(*response)
	if err != nil {
		//TODO: handle this log properly by using system error logging
		log.Fatal(err.Error())
		wr.WriteHeader(http.StatusBadRequest)
	}
	fmt.Fprintf(wr, "%v", sr)
}
