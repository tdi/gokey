package keystok

import (
	"code.google.com/p/go.crypto/pbkdf2"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"strings"
	"time"
)

var Version string = "0.35"

type KeystokClient struct {
	Access_token AccessToken
	Opts         KeystokOptions
}

func GetKeystokClient(access_token string) KeystokClient {
  access_token = access_token
  if access_token == "" {
	  access_token = os.Getenv("KEYSTOK_ACCESS_TOKEN")
  }
  if access_token == "" {
    panic("No access token given")
  }
	options := KeystokOptions{"https://api.keystok.com", "https://keystok.com", "", true}
	atk := decode_access_token(access_token)
	return KeystokClient{Access_token: atk, Opts: options}
}

func (k *KeystokClient) GetKey(name string) string {
  k.setup_cache()
	return k.get_key(k.Access_token, name)
}

func (k *KeystokClient) ListKeys() map[string]string {
  k.setup_cache()
	return k.list_keys(k.Access_token)

}

type KeystokOptions struct {
	APIHost  string
	AuthHost string
	CacheDir string
	UseCache bool
}

type AccessToken struct {
	Id            int
	RefreshToken  string
	DecryptionKey string
	AccessToken   string
}

func (k *KeystokClient) get_key(atk AccessToken, key_id string) string {
	api_host := k.Opts.APIHost
	atk.AccessToken = k.refresh_access_token(atk)
	var url string = fmt.Sprintf("%s/apps/%d/deploy/%s?access_token=%s", api_host, atk.Id, key_id,
		atk.AccessToken)
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	var dat map[string]map[string]interface{}
	if err := json.Unmarshal(body, &dat); err != nil {
		panic(err)
	}
	key := k.decrypt_key(atk, string(dat[key_id]["key"].(string)))
	return key
}

func (k *KeystokClient) decrypt_key(atk AccessToken, key string) string {
	if !strings.HasPrefix(key, ":aes256:") {
		panic("Not supported")
	}
	key_data := key[8:]
	key_bytes, _ := base64.StdEncoding.DecodeString(key_data)
	var dat map[string]interface{}
	if err := json.Unmarshal(key_bytes, &dat); err != nil {
		panic(err)
	}

	salt, _ := base64.StdEncoding.DecodeString(dat["salt"].(string))

	dk := pbkdf2.Key([]byte(atk.DecryptionKey), salt, 1000, 32, sha1.New)
	iv, _ := base64.StdEncoding.DecodeString(dat["iv"].(string))
	ct, _ := base64.StdEncoding.DecodeString(dat["ct"].(string))
	block, err := aes.NewCipher(dk)
	if err != nil {
		fmt.Println(err)
	}
	aes := cipher.NewCBCDecrypter(block, iv)
	aes.CryptBlocks(ct, ct)
	padding := ct[len(ct)-1]
	ct = ct[0 : len(ct)-int(padding)]
	return string(ct)
}

func (k *KeystokClient) list_keys(atk AccessToken) map[string]string {
	api_host := k.Opts.APIHost
	atk.AccessToken = k.refresh_access_token(atk)
	var url string = fmt.Sprintf("%s/apps/%d/keys?access_token=%s", api_host, atk.Id,
		atk.AccessToken)
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	var dat []map[string]interface{}
	if err := json.Unmarshal(body, &dat); err != nil {
		panic(err)
	}
	m := make(map[string]string)
	for _, v := range dat {
		m[v["id"].(string)] = v["description"].(string)
	}
	return m
}

func (k *KeystokClient) refresh_access_token(atk AccessToken) string {
	auth_host := k.Opts.AuthHost + "/oauth/token"
	access_token := atk.AccessToken
	refresh_token := atk.RefreshToken
	if access_token != "" {
		return access_token
	}
	var dat map[string]interface{}

	cache_file, err := ioutil.ReadFile(k.Opts.CacheDir + "/access_token")
	if err == nil {
		if err := json.Unmarshal(cache_file, &dat); err != nil {
			panic(err)
		}
		if int32(dat["expires_at"].(float64)) > int32(time.Now().Unix()) {
			return dat["access_token"].(string)
		}
	}

	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Add("refresh_token", refresh_token)

	resp, err := http.PostForm(auth_host, data)
	if err != nil {
		fmt.Println("error")
		panic(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err := json.Unmarshal(body, &dat); err != nil {
		panic(err)
	}
	dat["expires_at"] = int32(dat["expires_in"].(float64)) + int32(time.Now().Unix())
	dat2, _ := json.Marshal(dat)
	err = ioutil.WriteFile(k.Opts.CacheDir+"/access_token", dat2, 0666)
	if err != nil {
		//
	}
	return dat["access_token"].(string)
}

func decode_access_token(at string) AccessToken {
	var access_token string = strings.Replace(at, "-", "+", -1)
	access_token = strings.Replace(access_token, "_", "/", -1)
	data, err := base64.StdEncoding.DecodeString(access_token)
	if err != nil {
		fmt.Println("error", err)
		panic(err)
	}
	var dat map[string]interface{}
	if err := json.Unmarshal(data, &dat); err != nil {
		panic(err)
	}
	atk := AccessToken{int(dat["id"].(float64)), string(dat["rt"].(string)),
		string(dat["dk"].(string)), string("")}
	return atk
}

func (k *KeystokClient) setup_cache() {
	if k.Opts.UseCache == false {
		return
	}
	if k.Opts.CacheDir != "" {
		return
	}
	usr, _ := user.Current()
	dir := usr.HomeDir
	k.Opts.CacheDir = fmt.Sprintf("%s/.keystok", dir)
	_, err := os.Stat(k.Opts.CacheDir)
	if err == nil {
	  // no directory, we create one
	}
	if os.IsNotExist(err) {
		err := os.Mkdir(k.Opts.CacheDir, 0777)
		if err == nil {
			return
		}
	}
}
