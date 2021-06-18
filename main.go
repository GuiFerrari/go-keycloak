package main

import (
	"context"
	oidc "github.com/coreos/go-oidc"
	oauth2 "golang.org/x/oauth2"
	"log"
	"net/http"
	"encoding/json"
)

var (
	clientID = 			"myclient"
	clientSecret = 	"acfdee62-9c7e-4435-be89-8484f76a0e9b"
)

func main() {
	ctx := context.Background()

	provider, err := oidc.NewProvider(ctx, "http://localhost:8080/auth/realms/myrealm")

	if err != nil {
		log.Fatal(err)
	}

	config := oauth2.Config{
		ClientID: 			clientID,
		ClientSecret: 	clientSecret,
		Endpoint: 			provider.Endpoint(),
		RedirectURL: 		"http://localhost:8081/auth/callback",
		Scopes: 				[]string{oidc.ScopeOpenID, "profile", "email", "roles"},
	}

	state := "123"

	http.HandleFunc("/", func (writer http.ResponseWriter, request *http.Request) {
		http.Redirect(writer, request, config.AuthCodeURL(state), http.StatusFound)
	})

	http.HandleFunc("/auth/callback", func (writer http.ResponseWriter, request *http.Request) {
		if request.URL.Query().Get("state") != state {
			http.Error(writer, "Invalid state!", http.StatusBadRequest)
			return
		}

		token, err := config.Exchange(ctx, request.URL.Query().Get("code"))

		if err != nil {
			http.Error(writer, "Change token error!", http.StatusInternalServerError)
			return
		}

		idToken, ok := token.Extra("id_token").(string)

		if !ok {
			http.Error(writer, "Error to generate iDToken", http.StatusInternalServerError)
			return
		}

		userInfo, err := provider.UserInfo(ctx, oauth2.StaticTokenSource(token))

		if err != nil {
			http.Error(writer, "Error to get UserInfo!", http.StatusInternalServerError)
			return
		}

		resp := struct {
			AccessToken *oauth2.Token
			IDToken string
			UserInfo *oidc.UserInfo
		}{
			token,
			idToken,
			userInfo,
		}

		data, err := json.Marshal(resp);

		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}

		writer.Write(data)

	})

	log.Fatal(http.ListenAndServe(":8081", nil))
}