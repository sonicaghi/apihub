package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"

	. "github.com/albertoleal/backstage/account"
	"github.com/albertoleal/backstage/api/context"
	"github.com/albertoleal/backstage/api/helpers"
	"github.com/albertoleal/backstage/errors"
	"github.com/zenazn/goji/web"
)

type UsersController struct {
	ApiController
}

func (controller *UsersController) CreateUser(c *web.C, w http.ResponseWriter, r *http.Request) (*HTTPResponse, error) {
	body, err := controller.getPayload(c, r)
	if err != nil {
		return nil, err
	}
	var erro *errors.HTTPError
	user := &User{}
	if err := json.Unmarshal(body, user); err != nil {
		fmt.Print("It was not possible to create a new user.")
		return nil, err
	}

	err = user.Save()
	if err != nil {
		e := err.(*errors.ValidationError)
		erro = &errors.HTTPError{StatusCode: http.StatusBadRequest, Message: e.Message}
		context.AddRequestError(c, erro)
		return nil, erro
	}
	user.Password = ""
	payload, _ := json.Marshal(user)
	response := &HTTPResponse{StatusCode: http.StatusCreated, Payload: string(payload)}
	return response, nil
}

func (controller *UsersController) DeleteUser(c *web.C, w http.ResponseWriter, r *http.Request) (*HTTPResponse, error) {
	user, err := context.GetCurrentUser(c)
	if err != nil {
		erro := &errors.HTTPError{StatusCode: http.StatusBadRequest, Message: err.Error()}
		context.AddRequestError(c, erro)
		return nil, erro
	}

	user.Delete()
	user.Password = ""
	payload, _ := json.Marshal(user)
	response := &HTTPResponse{StatusCode: http.StatusOK, Payload: string(payload)}
	return response, nil
}

func (controller *UsersController) SignIn(c *web.C, w http.ResponseWriter, r *http.Request) (*HTTPResponse, error) {
	username, password := r.FormValue("username"), r.FormValue("password")

	token, err := helpers.SignIn(username, password)
	if err != nil {
		var erro *errors.HTTPError
		erro = &errors.HTTPError{StatusCode: http.StatusBadRequest, Message: "Invalid Username or Password."}
		context.AddRequestError(c, erro)
		return nil, erro
	}

	payload, _ := json.Marshal(token)
	response := &HTTPResponse{StatusCode: http.StatusOK, Payload: string(payload)}
	return response, nil
}
