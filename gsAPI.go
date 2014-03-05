package gsAPI

import(
	"log"
	"net/http"
	"bytes"
	"crypto/hmac"
	"crypto/md5"
	"encoding/json"
	"io/ioutil"
	"errors"
	"strconv"
	"encoding/hex"
)

type ErrorResponse struct {
	Code int `json:"code"`
	Message string `json:"Message"`
}

type HeaderResponse struct {
	Hostname string `json:"hostname"`
}

type ResultResponse struct {
	Success	bool `json:"success"`
	SessionId string `json:"sessionID"`
}

type JsonResponse struct {
	Error []ErrorResponse `json:"errors"`
	Header HeaderResponse `json:"header"`
	Result ResultResponse `json:"result"`
}

var apiScheme = "http://"
var apiHost = "api.grooveshark.com"
var apiEndpoint = "/ws3.php"

var WsKey ,WsSecret string

func New(key string, secret string) {
	WsKey = key
	WsSecret = secret
}

func StartSession() (string, error) {
	result, err := makeCall("startsession","", "sessionID", true, "")
	if err != nil {
		return "",err
	}
	return result,nil
}

func createMessageSig(params string, secret string) string {
	return ComputeHmacMD5(params, secret)
}

func ComputeHmacMD5(message string, secret string) string {
	key := []byte(secret)
	h := hmac.New(md5.New, key)
	h.Write([]byte(message))
    return hex.EncodeToString(h.Sum(nil))

}

func makeCall(method string, args interface{}, resultkey string, https bool, sessionID string) (string, error) {

	if https {
		apiScheme = "https://"
	}
	postData := `{"method":"`+method+`","header":{"wsKey":"`+WsKey+`"},"parameters":[]}`
	content := bytes.NewBuffer([]byte(postData))

	sig := createMessageSig(postData, WsSecret)
	log.Print(sig)
	requestUrl := apiScheme+apiHost+apiEndpoint+"?sig="+sig
	log.Print(requestUrl)
	resp, err := http.Post(requestUrl,"text/json charset=UTF-8",content)
	if err != nil {
		log.Fatal("Err POST: "+ err.Error())
	}
	body, readerr := ioutil.ReadAll(resp.Body)
	if readerr != nil {
		log.Fatal("Err READALL: "+ err.Error())
	}
	log.Print(string(body))
	var jsonresponse JsonResponse
	err = json.Unmarshal(body,&jsonresponse)
	log.Print(jsonresponse)
	if err != nil {
		log.Fatal("Err : "+ err.Error())
	}
	if len(jsonresponse.Error) != 0 {
		return "", errors.New(strconv.Itoa(jsonresponse.Error[0].Code)+" - "+jsonresponse.Error[0].Message)
	}

	return "yes",errors.New(strconv.Itoa(1))
}

func handleError(v interface{}) {
	log.Fatal(v)

}
