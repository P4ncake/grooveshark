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
	return result.(string), nil
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

func LogOut(SessionID string) error {
	args := map[string]interface{}{}
	result, err := makeCall("logout",args,"success",false,SessionID)

	if err != nil {
		return err
	}
	if result.(map[string]interface{})["success"].(bool) {
		return nil
	}
	return errors.New("Logout Failed!")
}

func GetUserInfo() (interface{}, error) {
	return makeCall("getuserinfo", map[string]interface{}{},"",false,sessionID)
}

func Authenticate(username string, password string) (interface{}, error) {
	if username == "" || password == "" {
		return nil, errors.New("Empty username or password")
	}

	h := md5.New()
	h.Write([]byte(password))
	hpass := hex.EncodeToString(h.Sum(nil))
	log.Print(hpass)
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

/*
 *	Get the playlists owned by the given UserId
 *	Requires a valid sessionID and authenticated user
 *
 *	Set limit to 0 for no limit
 *	Set UserId to 0 to get the logged-in user's playlists
 */
func GetUserPlaylists(limit int,UserId int) (interface{}, error) {
	var method string
	args := map[string]interface{}{
		"limit": limit,
	}

	if UserId != 0 {
		method = "getUserPlaylistsByUserID"
		args["userID"] = UserId
	} else {
		method = "getUserPlaylists"

	}

	return makeCall(method,args,"playlists",false,sessionID)
}

func GetUserLibrary(limit int) (interface{}, error) {

	args := map[string]interface{}{
		"limit": limit,
	}

	return makeCall("getUserLibrarySongs",args,"songs",false,sessionID)

}

func GetUserFavorites(limit int) (interface{}, error) {

	args := map[string]interface{}{
		"limit": limit,
	}

	return makeCall("getUserFavoriteSongs",args,"songs",false,sessionID)

}

func AddUserFavoriteSong(songId int) (interface{}, error) {
	args := map[string]interface{}{
		"songID": songId,
	}
	return makeCall("addUserFavoriteSong",args,"success",false,sessionID)
}

func AddUserLibrarySongs(songId int) (interface{}, error) {
	args := map[string]interface{}{
		"songID": songId,
	}
	return makeCall("addUserLibrarySongsEx",args,"success",false,sessionID)
}

func CreatePlaylist(name string,songIds []int) (interface{}, error) {
	if name == "" {
		return nil, errors.New("Playlist's Name empty")
	}
	args := map[string]interface{}{
		"name": name,
		"songIDs":songIds,
	}

	return makeCall("createPlaylist",args,"",false,sessionID)
}


func AddSongToPlaylist(playlistId int, songId int) (interface{}, error) {
	songs, err := GetPlaylistSongs(playlistId,0)
	if err != nil {
		return nil, err
	}
	log.Print(songs)
	return nil,nil
}


func GetPlaylistSongs(playlistId int, limit int) (interface{}, error) {
	args := map[string]interface{}{
		"playlistID": playlistId,
		"limit": limit,
	}
	songs, err :=  makeCall("getPlaylistSongs",args,"songs",false,sessionID)
	if err != nil {
		return nil, err
	}
	var SongIds []float64
	for _,v := range songs.(map[string]interface{})["songs"].([]interface{}) {
		songId := v.(map[string]interface{})["SongID"].(float64)
		SongIds = append(SongIds,songId)
	}
	return SongIds ,nil
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
	//log.Print(string(body))
	var jsonresponse map[string]interface{}
	err = json.Unmarshal(body,&jsonresponse)
	log.Print(jsonresponse)
	if err != nil {
		log.Fatal("Err : "+ err.Error())
	}

	jsonErr := jsonresponse["errors"]
	if jsonErr != nil {
		Err := jsonresponse["errors"].([]interface{})[0].(map[string]interface{})
		return nil, errors.New(strconv.FormatFloat(Err["code"].(float64),'g',3,64)+" - " + Err["message"].(string))
		return nil,errors.New("TODO")
	}

	return jsonresponse["result"], nil
}

