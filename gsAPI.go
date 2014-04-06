package gsAPI

import (
	"bytes"
	"crypto/hmac"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strconv"
)

type Headers struct {
	WsKey     string `json:"wsKey"`
	SessionID string `json:"sessionID"`
}

type Payload struct {
	Method     string      `json:"method"`
	Parameters interface{} `json:"parameters"`
	Header     Headers     `json:"header"`
}

var apiScheme = "http://"
var apiHost = "api.grooveshark.com"
var apiEndpoint = "/ws3.php"

var WsKey, WsSecret string
var sessionID string

var Logs = false

var Country string

func New(key string, secret string) {
	WsKey = key
	WsSecret = secret
}

func PingService() (string, error) {
	args := map[string]interface{}{}
	result, err := makeCall("pingService", args, "", true, "")
	if err != nil {
		return "", err
	}
	return result.(string), nil
}

func StartSession() (string, error) {

	args := map[string]interface{}{}
	result, err := makeCall("startSession", args, "sessionID", true, "")
	if err != nil {
		return "", err
	}

	sessionID = result.(map[string]interface{})["sessionID"].(string)
	return sessionID, nil
}

func LogOut(SessionID string) error {
	args := map[string]interface{}{}
	result, err := makeCall("logout", args, "success", false, SessionID)

	if err != nil {
		return err
	}
	if result.(map[string]interface{})["success"].(bool) {
		return nil
	}
	return errors.New("Logout Failed!")
}

func GetUserInfo() (interface{}, error) {
	return makeCall("getuserinfo", map[string]interface{}{}, "", false, sessionID)
}

func GetCountry(ip string) (interface{}, error) {
	log.Print(ip)
	matched, err := regexp.MatchString(`([0-9]{1,3}\.){3}[0-9]{1,3}`,ip)
	if err != nil {
		return nil, err
	}
	if !matched {
		return nil, errors.New("Invalid IP Address")
	}
	args := map[string]interface{}{
		"ip": ip,
	}

	return makeCall("getCountry",args,"",false,"")
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
		"login":    username,
		"password": hpass,
	}

	result, err := makeCall("authenticate", args, "", true, sessionID)
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
func GetUserPlaylists(limit int, UserId int) (interface{}, error) {
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

	return makeCall(method, args, "playlists", false, sessionID)
}

func GetUserLibrary(limit int) (interface{}, error) {

	args := map[string]interface{}{
		"limit": limit,
	}

	return makeCall("getUserLibrarySongs", args, "songs", false, sessionID)

}

func GetUserFavorites(limit int) (interface{}, error) {

	args := map[string]interface{}{
		"limit": limit,
	}

	return makeCall("getUserFavoriteSongs", args, "songs", false, sessionID)

}

func AddUserFavoriteSong(songId int) (interface{}, error) {
	args := map[string]interface{}{
		"songID": songId,
	}
	return makeCall("addUserFavoriteSong", args, "success", false, sessionID)
}

func AddUserLibrarySongs(songId int) (interface{}, error) {
	args := map[string]interface{}{
		"songID": songId,
	}
	return makeCall("addUserLibrarySongsEx", args, "success", false, sessionID)
}

func CreatePlaylist(name string, songIds []int) (interface{}, error) {
	if name == "" {
		return nil, errors.New("Playlist's Name empty")
	}
	args := map[string]interface{}{
		"name":    name,
		"songIDs": songIds,
	}

	return makeCall("createPlaylist", args, "", false, sessionID)
}

func AddSongToPlaylist(playlistId int, songId int) (interface{}, error) {
	songs, err := GetPlaylistSongs(playlistId, 0)
	if err != nil {
		return nil, err
	}
	songs = append(songs, songId)
	return SetPlaylistSongs(playlistId, songs)
}

func SetPlaylistSongs(playlistId int, songIds []int) (interface{}, error) {
	args := map[string]interface{}{
		"playlistID": playlistId,
		"songIDs":    songIds,
	}

	return makeCall("setPlaylistSongs", args, "", false, sessionID)
}

/**
 * Artists/Albums/Songs related methods
 */

func GetArtistInfo(artistId []int) (interface{}, error) {
	args := map[string]interface{}{
		"artistIDs": artistId,
	}
	return makeCall("getArtistsInfo", args, "artists", false, sessionID)
}

func GetSongIDFromTinysongBase62(base string) (interface{}, error) {
	args := map[string]interface{}{
		"base62": base,
	}

	matched, err := regexp.MatchString("/^[A-Za-z0-9]+$/", base)
	if err != nil {
		return nil, err
	}
	if !matched {
		return nil, errors.New("Base don't match regexp")
	}

	return makeCall("getSongIDFromTinysongBase62", args, "songID", false, sessionID)
}

func GetSongURLFromTinysongBase62(base string) (interface{}, error) {
	args := map[string]interface{}{
		"base62": base,
	}

	matched, err := regexp.MatchString("/^[A-Za-z0-9]+$/", base)
	if err != nil {
		return nil, err
	}
	if !matched {
		return nil, errors.New("Base don't match regexp")
	}

	return makeCall("getSongURLFromTinysongBase62", args, "songID", false, sessionID)
}

func GetSongURLFromSongID(songId int) (interface{}, error) {
	args := map[string]interface{}{
		"songID": songId,
	}

	return makeCall("getSongURLFromSongID", args, "url", false, sessionID)
}

func GetSongsInfo(songIds []int) (interface{}, error) {
	args := map[string]interface{}{
		"songdIDs": songIds,
	}
	return makeCall("getSongsInfo", args, "songs", false, sessionID)
}

func GetAlbumsInfo(albumIds []int) (interface{}, error) {
	args := map[string]interface{}{
		"albumIDs": albumIds,
	}
	return makeCall("getAlbumsInfo", args, "albums", false, sessionID)
}

func GetAlbumSongs(albumId int, limit int) (interface{}, error) {
	args := map[string]interface{}{
		"albumID": albumId,
		"limit":   limit,
	}

	return makeCall("getAlbumSongs", args, "songs", false, sessionID)

}

func GetPlaylistSongs(playlistId int, limit int) ([]int, error) {
	args := map[string]interface{}{
		"playlistID": playlistId,
		"limit":      limit,
	}
	songs, err := makeCall("getPlaylistSongs", args, "songs", false, sessionID)
	if err != nil {
		return nil, err
	}
	var SongIds []int
	for _, v := range songs.(map[string]interface{})["songs"].([]interface{}) {
		songId := v.(map[string]interface{})["SongID"].(float64)
		SongIds = append(SongIds, int(songId))
	}
	return SongIds, nil
}

func GetDoesExist(id int, Type string) (interface{}, error) {
	args := map[string]interface{}{}
	var call string
	switch Type {
	case "song":
		args["songID"] = id
		call = "getDoesSongExist"
	case "artist":
		args["artistID"] = id
		call = "getDoesArtistExist"
	case "Album":
		args["albumID"] = id
		call = "getDoesAlbumExist"
	}
	return makeCall(call, args, "", false, sessionID)
}

func GetArtistAlbums(artistId int, verified bool) (interface{}, error) {
	args := map[string]interface{}{
		"artistID": artistId,
	}
	if verified {
		return makeCall("getArtistVerifiedAlbums", args, "albums", false, sessionID)
	}
	return makeCall("getArtistAlbums", args, "albums", false, sessionID)
}

func GetArtistPopularSongs(artistId int) (interface{}, error) {
	args := map[string]interface{}{
		"artistID": artistId,
	}
	return makeCall("getArtistPopularSongs", args, "songs", false, sessionID)

}

func GetPopularSongsToday(limit int) (interface{}, error) {
	args := map[string]interface{}{
		"limit": limit,
	}
	return makeCall("getPopularSongsToday", args, "songs", false, sessionID)
}

func GetPopularSongsMonth(limit int) (interface{}, error) {
	args := map[string]interface{}{
		"limit": limit,
	}
	return makeCall("getPopularSongsMonth", args, "songs", false, sessionID)
}

func GetSongSearchResults(query string, country interface{},limit int, page int) (interface{}, error) {
	var offset int
	if limit != 0 {
		offset = (page - 1) * limit
	} else {
		offset = (page - 1) * 100
	}

	args := map[string]interface{}{
		"query": query,
		"country": country,
		"limit": limit,
		"offset": offset,
	}

	return makeCall("getSongSearchResults", args, "songs", false, sessionID)
}

func GetArtistSearchResults(query string, limit int) (interface{}, error) {
	args := map[string]interface{}{
		"query": query,
		"limit": limit,
	}
	return makeCall("getArtistSearchResults", args, "artists", false, sessionID)
}

func GetAlbumSearchResults(query string, limit int) (interface{}, error) {
	args := map[string]interface{}{
		"query": query,
		"limit": limit,
	}
	return makeCall("getAlbumSearchResults", args, "albums", false, sessionID)
}

func GetStreamKeyStreamServer(songId int, country interface{}, lowBitrate bool) (interface{}, error) {
	args := map[string]interface{}{
		"songID": songId,
		"country": country,
	}
	if lowBitrate {
		args["lowBitrate"] = true
	}

	return makeCall("getStreamKeyStreamServer", args, "", false, sessionID)
}
func GetSubscriberStreamKey(songId int, country interface{}, lowBitrate bool, trialUniqueID int) (interface{}, error) {
	args := map[string]interface{}{
		"songID": songId,
		"country": country,
	}
	if lowBitrate {
		args["lowBitrate"] = true
	}
	if trialUniqueID != 0 {
		args["uniqueID"] = trialUniqueID
	}

	return makeCall("getSubscriberStreamKey", args, "", false, sessionID)

}

func MarkStreamKeyOver30Secs(streamKey string, streamServerId int) (interface{}, error) {
	args := map[string]interface{}{
		"streamKey":      streamKey,
		"streamServerID": streamServerId,
	}
	return makeCall("markStreamKeyOver30Secs", args, "", false, sessionID)
}

func MarkSongComplete(songId int, streamKey string, streamServerId int) (interface{}, error) {
	args := map[string]interface{}{
		"songID":         songId,
		"streamKey":      streamKey,
		"streamServerID": streamServerId,
	}
	return makeCall("markSongComplete", args, "", false, sessionID)
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
	post := Payload{Method: method, Parameters: args, Header: headers}
	d, e := json.Marshal(post)
	if e != nil {
		log.Fatal(e.Error())
	}
	if Logs {
		log.Print(string(d))
	}
	content := bytes.NewBuffer([]byte(d))

	sig := createMessageSig(string(d), WsSecret)
	requestUrl := apiScheme + apiHost + apiEndpoint + "?sig=" + sig
	resp, err := http.Post(requestUrl, "text/json charset=UTF-8", content)
	if err != nil {
		log.Fatal("Err POST: " + err.Error())
	}
	body, readerr := ioutil.ReadAll(resp.Body)
	if readerr != nil {
		log.Fatal("Err READALL: " + err.Error())
	}

	if Logs {
		log.Print(string(body))
	}
	var jsonresponse map[string]interface{}
	err = json.Unmarshal(body, &jsonresponse)
	if Logs {
		log.Print(jsonresponse)
	}

	if err != nil {
		log.Fatal("Err : " + err.Error())
	}

	jsonErr := jsonresponse["errors"]
	if jsonErr != nil {
		Err := jsonresponse["errors"].([]interface{})[0].(map[string]interface{})
		return nil, errors.New(strconv.FormatFloat(Err["code"].(float64), 'g', 3, 64) + " - " + Err["message"].(string))
		return nil, errors.New("TODO")
	}

	return jsonresponse["result"], nil
}
