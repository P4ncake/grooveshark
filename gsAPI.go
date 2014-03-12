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
	_ "strconv"
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
	Result map[string]interface{} `json:"result"`
}


type Headers struct {
	WsKey string `json:"wsKey"`
	SessionID string `json:"sessionID"`
}

type Payload struct {
	Method string `json:"method"`
	Parameters interface{} `json:"parameters"`
	Header Headers `json:"header"`
}

var apiScheme = "http://"
var apiHost = "api.grooveshark.com"
var apiEndpoint = "/ws3.php"

var WsKey ,WsSecret string
var sessionID string

func New(key string, secret string) {
	WsKey = key
	WsSecret = secret
}

func PingService() (string, error){
	args := map[string]interface{}{}
	result, err := makeCall("pingService",args,"",true,"")
	if err != nil {
		return "", err
	}
	log.Fatal(result)
	return "", nil
}

func StartSession() (string, error) {

	args := map[string]interface{}{}
	result, err := makeCall("startSession",args, "sessionID", true, "")
	if err != nil {
		return "",err
	}
	sessionID = result.(map[string]interface{})["sessionID"].(string)
	return sessionID, nil
}

func Authenticate(username string, password string) (interface{}, error) {
	if username == "" || password == "" {
		return nil, errors.New("Empty username or password")
	}

	h := md5.New()
	h.Write([]byte(password))
	hpass := hex.EncodeToString(h.Sum(nil))

	args := map[string]interface{}{
		"login":username,
		"password" : hpass,
	}

	result, err := makeCall("authenticate",args,"",true,sessionID)
	if err != nil {
		return nil, err
	}

	if !result.(map[string]interface{})["success"].(bool) {
		return nil, errors.New("Login failed. invalid username/password")
	}
	return result, nil
}

func createMessageSig(params string, secret string) string {
	return computeHmacMD5(params, secret)
}

func computeHmacMD5(message string, secret string) string {
	key := []byte(secret)
	h := hmac.New(md5.New, key)
	h.Write([]byte(message))
    return hex.EncodeToString(h.Sum(nil))

}

func makeCall(method string, args interface{}, resultkey string, https bool, sessionID string) (interface{}, error) {

	if https {
		apiScheme = "https://"
	}
	headers := Headers{WsKey: WsKey, SessionID: sessionID}
	post := Payload{Method: method,Parameters: args,Header: headers }
	d,e := json.Marshal(post)
	if e !=nil {
		log.Fatal(e.Error())
	}
//	log.Print(string(d))
	content := bytes.NewBuffer([]byte(d))

	sig := createMessageSig(string(d), WsSecret)
	requestUrl := apiScheme+apiHost+apiEndpoint+"?sig="+sig
	resp, err := http.Post(requestUrl,"text/json charset=UTF-8",content)
	if err != nil {
		log.Fatal("Err POST: "+ err.Error())
	}
	body, readerr := ioutil.ReadAll(resp.Body)
	if readerr != nil {
		log.Fatal("Err READALL: "+ err.Error())
	}
	log.Print(string(body))
	var jsonresponse map[string]interface{}
	err = json.Unmarshal(body,&jsonresponse)
	log.Print(jsonresponse)
	if err != nil {
		log.Fatal("Err : "+ err.Error())
	}

/*	if len(jsonresponse["Error"]) != 0 {
		return nil, errors.New(strconv.Itoa(jsonresponse.Error[0].Code)+" - "+jsonresponse.Error[0].Message)
	}
*/
	return jsonresponse["result"], nil
}

