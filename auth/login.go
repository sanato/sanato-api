package auth

import (
	"encoding/json"
	"github.com/Sirupsen/logrus"
	"github.com/dgrijalva/jwt-go"
	"github.com/julienschmidt/httprouter"
	"io/ioutil"
	"net/http"
	"time"
)

func (api *API) login(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	config, err := api.configProvider.Parse()
	if err != nil {
		logrus.Error(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		logrus.Error(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	type logrusinParams struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	params := logrusinParams{}
	err = json.Unmarshal(body, &params)
	if err != nil {
		logrus.Warn(err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	authRes, err := api.authProvider.Authenticate(params.Username, params.Password)
	if err != nil {
		logrus.Error(err)
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}
	token := jwt.New(jwt.GetSigningMethod(config.TokenCipherSuite))
	token.Claims["iss"] = "blacksync"
	token.Claims["exp"] = time.Now().Add(time.Minute * 480).Unix()
	token.Claims["username"] = authRes.Username
	token.Claims["displayName"] = authRes.DisplayName
	token.Claims["email"] = authRes.Email
	tokenString, err := token.SignedString([]byte(config.TokenSecret))
	if err != nil {
		logrus.Error(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	data := make(map[string]string)
	data["token"] = tokenString
	tokenJSON, err := json.Marshal(data)
	if err != nil {
		logrus.Error(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	w.Write(tokenJSON)
	return
}
