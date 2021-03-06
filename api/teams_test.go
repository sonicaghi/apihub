package api_test

import (
	"fmt"
	"net/http"

	"github.com/apihub/apihub/account"
	"github.com/apihub/apihub/requests"
	. "gopkg.in/check.v1"
)

func (s *S) TestCreateTeam(c *C) {
	alias := "apihub-team"

	defer func() {
		s.store.DeleteTeamByAlias(alias)
	}()

	headers, code, body, _ := httpClient.MakeRequest(requests.Args{
		AcceptableCode: http.StatusCreated,
		Method:         "POST",
		Path:           "/api/teams",
		Body:           `{"name": "ApiHub Team"}`,
		Headers:        http.Header{"Authorization": {s.authHeader}},
	})

	c.Assert(code, Equals, http.StatusCreated)
	c.Assert(headers.Get("Content-Type"), Equals, "application/json")
	c.Assert(string(body), Equals, fmt.Sprintf(`{"name":"ApiHub Team","alias":"%s","users":["%s"],"owner":"%s"}`, alias, user.Email, user.Email))
}

func (s *S) TestCreateTeamWithCustomAlias(c *C) {
	alias := "apihub"

	defer func() {
		s.store.DeleteTeamByAlias(alias)
	}()

	headers, code, body, _ := httpClient.MakeRequest(requests.Args{
		AcceptableCode: http.StatusCreated,
		Method:         "POST",
		Path:           "/api/teams",
		Body:           fmt.Sprintf(`{"name": "ApiHub Team", "alias": "%s"}`, alias),
		Headers:        http.Header{"Authorization": {s.authHeader}},
	})

	c.Assert(code, Equals, http.StatusCreated)
	c.Assert(headers.Get("Content-Type"), Equals, "application/json")
	c.Assert(string(body), Equals, fmt.Sprintf(`{"name":"ApiHub Team","alias":"%s","users":["%s"],"owner":"%s"}`, alias, user.Email, user.Email))
}

func (s *S) TestCreateTeamWhenAlreadyExists(c *C) {
	team := account.Team{Name: "ApiHub Team", Alias: "apihub"}
	team.Create(user)

	defer func() {
		s.store.DeleteTeamByAlias(team.Alias)
	}()

	headers, code, body, _ := httpClient.MakeRequest(requests.Args{
		AcceptableCode: http.StatusBadRequest,
		Method:         "POST",
		Path:           "/api/teams",
		Body:           fmt.Sprintf(`{"name": "ApiHub Team", "alias": "%s"}`, team.Alias),
		Headers:        http.Header{"Authorization": {s.authHeader}},
	})

	c.Assert(code, Equals, http.StatusBadRequest)
	c.Assert(headers.Get("Content-Type"), Equals, "application/json")
	c.Assert(string(body), Equals, `{"error":"bad_request","error_description":"Someone already has that team alias. Could you try another?"}`)

}

func (s *S) TestCreateTeamWithoutSignIn(c *C) {
	testWithoutSignIn(requests.Args{AcceptableCode: http.StatusUnauthorized, Method: "POST", Path: "/api/teams", Body: `{"name": "ApiHub Team"}`}, c)
}

func (s *S) TestCreateTeamWithInvalidRequest(c *C) {
	headers, code, body, _ := httpClient.MakeRequest(requests.Args{
		AcceptableCode: http.StatusBadRequest,
		Method:         "POST",
		Path:           "/api/teams",
		Body:           `"name": "ApiHub Team"`,
		Headers:        http.Header{"Authorization": {s.authHeader}},
	})

	c.Assert(code, Equals, http.StatusBadRequest)
	c.Assert(headers.Get("Content-Type"), Equals, "application/json")
	c.Assert(string(body), Equals, `{"error":"bad_request","error_description":"The request was invalid or cannot be served."}`)
}

func (s *S) TestTeamList(c *C) {
	headers, code, body, _ := httpClient.MakeRequest(requests.Args{
		AcceptableCode: http.StatusOK,
		Method:         "GET",
		Path:           "/api/teams",
		Headers:        http.Header{"Authorization": {s.authHeader}},
	})

	c.Assert(code, Equals, http.StatusOK)
	c.Assert(headers.Get("Content-Type"), Equals, "application/json")
	c.Assert(string(body), Equals, `{"items":[],"item_count":0}`)
}

func (s *S) TestTeamListWithoutSignIn(c *C) {
	testWithoutSignIn(requests.Args{AcceptableCode: http.StatusUnauthorized, Method: "GET", Path: "/api/teams"}, c)
}

func (s *S) TestDeleteTeam(c *C) {
	team := account.Team{Name: "ApiHub Team", Alias: "apihub"}
	team.Create(user)

	headers, code, body, _ := httpClient.MakeRequest(requests.Args{
		AcceptableCode: http.StatusOK,
		Method:         "DELETE",
		Path:           fmt.Sprintf("/api/teams/%s", team.Alias),
		Headers:        http.Header{"Authorization": {s.authHeader}},
	})

	c.Assert(code, Equals, http.StatusOK)
	c.Assert(headers.Get("Content-Type"), Equals, "application/json")
	c.Assert(string(body), Equals, fmt.Sprintf(`{"name":"%s","alias":"%s","users":["%s"],"owner":"%s"}`, team.Name, team.Alias, user.Email, user.Email))
}

func (s *S) TestDeleteTeamWithoutSignIn(c *C) {
	testWithoutSignIn(requests.Args{AcceptableCode: http.StatusUnauthorized, Method: "DELETE", Path: "/api/teams/apihub"}, c)
}

func (s *S) TestDeleteTeamWithoutPermission(c *C) {
	alice := account.User{Name: "alice", Email: "alice@bar.example.org", Password: "secret"}
	alice.Create()
	defer alice.Delete()

	team := account.Team{Name: "ApiHub Team", Alias: "apihub"}
	team.Create(alice)
	defer func() {
		s.store.DeleteTeamByAlias(team.Alias)
	}()

	headers, code, body, _ := httpClient.MakeRequest(requests.Args{
		AcceptableCode: http.StatusOK,
		Method:         "DELETE",
		Path:           fmt.Sprintf("/api/teams/%s", team.Alias),
		Headers:        http.Header{"Authorization": {s.authHeader}},
	})

	c.Assert(code, Equals, http.StatusForbidden)
	c.Assert(headers.Get("Content-Type"), Equals, "application/json")
	c.Assert(string(body), Equals, `{"error":"access_denied","error_description":"Only the owner has permission to perform this operation."}`)
}

func (s *S) TestDeleteTeamNotFound(c *C) {
	headers, code, body, _ := httpClient.MakeRequest(requests.Args{
		AcceptableCode: http.StatusNotFound,
		Method:         "DELETE",
		Path:           "/api/teams/not-found",
		Headers:        http.Header{"Authorization": {s.authHeader}},
	})

	c.Assert(code, Equals, http.StatusNotFound)
	c.Assert(headers.Get("Content-Type"), Equals, "application/json")
	c.Assert(string(body), Equals, `{"error":"not_found","error_description":"Team not found."}`)
}

func (s *S) TestTeamInfo(c *C) {
	team := account.Team{Name: "ApiHub Team", Alias: "apihub"}
	team.Create(user)
	defer team.Delete(user)

	headers, code, body, _ := httpClient.MakeRequest(requests.Args{
		AcceptableCode: http.StatusOK,
		Method:         "GET",
		Path:           fmt.Sprintf("/api/teams/%s", team.Alias),
		Headers:        http.Header{"Authorization": {s.authHeader}},
	})

	c.Assert(string(body), Equals, fmt.Sprintf(`{"name":"%s","alias":"%s","users":["%s"],"owner":"%s"}`, team.Name, team.Alias, user.Email, user.Email))
	c.Assert(code, Equals, http.StatusOK)
	c.Assert(headers.Get("Content-Type"), Equals, "application/json")
}

func (s *S) TestTeamInfoNotFound(c *C) {
	headers, code, body, _ := httpClient.MakeRequest(requests.Args{
		AcceptableCode: http.StatusNotFound,
		Method:         "GET",
		Path:           "/api/teams/not-found",
		Headers:        http.Header{"Authorization": {s.authHeader}},
	})

	c.Assert(code, Equals, http.StatusNotFound)
	c.Assert(headers.Get("Content-Type"), Equals, "application/json")
	c.Assert(string(body), Equals, `{"error":"not_found","error_description":"Team not found."}`)
}
func (s *S) TestTeamInfoWithoutPermission(c *C) {
	alice := account.User{Name: "alice", Email: "alice@bar.example.org", Password: "secret"}
	alice.Create()
	defer alice.Delete()

	team := account.Team{Name: "ApiHub Team", Alias: "apihub"}
	team.Create(alice)
	defer func() {
		s.store.DeleteTeamByAlias(team.Alias)
	}()

	headers, code, body, _ := httpClient.MakeRequest(requests.Args{
		AcceptableCode: http.StatusForbidden,
		Method:         "GET",
		Path:           fmt.Sprintf("/api/teams/%s", team.Alias),
		Headers:        http.Header{"Authorization": {s.authHeader}},
	})

	c.Assert(code, Equals, http.StatusForbidden)
	c.Assert(headers.Get("Content-Type"), Equals, "application/json")
	c.Assert(string(body), Equals, `{"error":"access_denied","error_description":"You do not belong to this team!"}`)
}

func (s *S) TestAddUser(c *C) {
	team := account.Team{Name: "ApiHub Team", Alias: "apihub"}
	team.Create(user)
	defer func() {
		s.store.DeleteTeamByAlias(team.Alias)
	}()

	alice := account.User{Name: "alice", Email: "alice@bar.example.org", Password: "secret"}
	alice.Create()
	defer alice.Delete()

	headers, code, body, _ := httpClient.MakeRequest(requests.Args{
		AcceptableCode: http.StatusOK,
		Method:         "PUT",
		Path:           fmt.Sprintf("/api/teams/%s/users", team.Alias),
		Headers:        http.Header{"Authorization": {s.authHeader}},
		Body:           fmt.Sprintf(`{"users": ["%s"]}`, alice.Email),
	})

	c.Assert(string(body), Equals, `{"name":"ApiHub Team","alias":"apihub","users":["bob@bar.example.org","alice@bar.example.org"],"owner":"bob@bar.example.org"}`)
	c.Assert(code, Equals, http.StatusOK)
	c.Assert(headers.Get("Content-Type"), Equals, "application/json")
}

func (s *S) TestAddUserNotMember(c *C) {
	alice := account.User{Name: "alice", Email: "alice@bar.example.org", Password: "secret"}
	alice.Create()
	defer alice.Delete()

	team := account.Team{Name: "ApiHub Team", Alias: "apihub"}
	team.Create(alice)
	defer func() {
		s.store.DeleteTeamByAlias(team.Alias)
	}()

	headers, code, body, _ := httpClient.MakeRequest(requests.Args{
		AcceptableCode: http.StatusOK,
		Method:         "PUT",
		Path:           fmt.Sprintf("/api/teams/%s/users", team.Alias),
		Headers:        http.Header{"Authorization": {s.authHeader}},
		Body:           fmt.Sprintf(`{"users": ["%s"]}`, alice.Email),
	})

	c.Assert(code, Equals, http.StatusForbidden)
	c.Assert(headers.Get("Content-Type"), Equals, "application/json")
	c.Assert(string(body), Equals, `{"error":"access_denied","error_description":"You do not belong to this team!"}`)
}

func (s *S) TestAddUserWithoutSignIn(c *C) {
	team := account.Team{Name: "ApiHub Team", Alias: "apihub"}
	team.Create(user)
	defer func() {
		s.store.DeleteTeamByAlias(team.Alias)
	}()

	testWithoutSignIn(requests.Args{
		AcceptableCode: http.StatusUnauthorized,
		Method:         "PUT",
		Path:           fmt.Sprintf("/api/teams/%s/users", team.Alias),
		Body:           `{"users": ["bob@example.org"]}`},
		c)
}

func (s *S) TestAddUserNotFound(c *C) {
	headers, code, body, _ := httpClient.MakeRequest(requests.Args{
		AcceptableCode: http.StatusOK,
		Method:         "PUT",
		Path:           "/api/teams/not-found/users",
		Body:           `{"name": "New name"}`,
		Headers:        http.Header{"Authorization": {s.authHeader}},
	})

	c.Assert(code, Equals, http.StatusNotFound)
	c.Assert(headers.Get("Content-Type"), Equals, "application/json")
	c.Assert(string(body), Equals, `{"error":"not_found","error_description":"Team not found."}`)
}

func (s *S) TestRemoveUser(c *C) {
	alice := account.User{Name: "alice", Email: "alice@bar.example.org", Password: "secret"}
	alice.Create()
	defer alice.Delete()

	team := account.Team{Name: "ApiHub Team", Alias: "apihub", Users: []string{alice.Email}}
	team.Create(user)
	defer func() {
		s.store.DeleteTeamByAlias(team.Alias)
	}()

	headers, code, body, _ := httpClient.MakeRequest(requests.Args{
		AcceptableCode: http.StatusOK,
		Method:         "DELETE",
		Path:           fmt.Sprintf("/api/teams/%s/users", team.Alias),
		Headers:        http.Header{"Authorization": {s.authHeader}},
		Body:           fmt.Sprintf(`{"users": ["%s"]}`, alice.Email),
	})

	c.Assert(code, Equals, http.StatusOK)
	c.Assert(headers.Get("Content-Type"), Equals, "application/json")
	c.Assert(string(body), Equals, `{"name":"ApiHub Team","alias":"apihub","users":["bob@bar.example.org"],"owner":"bob@bar.example.org"}`)
}

func (s *S) TestRemoveUserWithoutSignIn(c *C) {
	team := account.Team{Name: "ApiHub Team", Alias: "apihub"}
	team.Create(user)
	defer func() {
		s.store.DeleteTeamByAlias(team.Alias)
	}()

	testWithoutSignIn(requests.Args{
		AcceptableCode: http.StatusUnauthorized,
		Method:         "DELETE",
		Path:           fmt.Sprintf("/api/teams/%s/users", team.Alias),
		Body:           `{"users": ["bob@example.org"]}`},
		c)
}

func (s *S) TestRemoveUserNotMember(c *C) {
	alice := account.User{Name: "alice", Email: "alice@bar.example.org", Password: "secret"}
	alice.Create()
	defer alice.Delete()

	team := account.Team{Name: "ApiHub Team", Alias: "apihub"}
	team.Create(alice)
	defer func() {
		s.store.DeleteTeamByAlias(team.Alias)
	}()

	headers, code, body, _ := httpClient.MakeRequest(requests.Args{
		AcceptableCode: http.StatusOK,
		Method:         "DELETE",
		Path:           fmt.Sprintf("/api/teams/%s/users", team.Alias),
		Headers:        http.Header{"Authorization": {s.authHeader}},
		Body:           fmt.Sprintf(`{"users": ["%s"]}`, alice.Email),
	})

	c.Assert(code, Equals, http.StatusForbidden)
	c.Assert(headers.Get("Content-Type"), Equals, "application/json")
	c.Assert(string(body), Equals, `{"error":"access_denied","error_description":"You do not belong to this team!"}`)
}

func (s *S) TestRemoveUserNotFound(c *C) {
	headers, code, body, _ := httpClient.MakeRequest(requests.Args{
		AcceptableCode: http.StatusOK,
		Method:         "DELETE",
		Path:           "/api/teams/not-found/users",
		Body:           `{"name": "New name"}`,
		Headers:        http.Header{"Authorization": {s.authHeader}},
	})

	c.Assert(code, Equals, http.StatusNotFound)
	c.Assert(headers.Get("Content-Type"), Equals, "application/json")
	c.Assert(string(body), Equals, `{"error":"not_found","error_description":"Team not found."}`)
}

func (s *S) TestUpdateTeam(c *C) {
	team := account.Team{Name: "ApiHub Team", Alias: "apihub"}
	team.Create(user)

	defer func() {
		s.store.DeleteTeamByAlias(team.Alias)
	}()

	headers, code, body, _ := httpClient.MakeRequest(requests.Args{
		AcceptableCode: http.StatusOK,
		Method:         "PUT",
		Path:           fmt.Sprintf("/api/teams/%s", team.Alias),
		Body:           `{"name": "New name"}`,
		Headers:        http.Header{"Authorization": {s.authHeader}},
	})

	c.Assert(code, Equals, http.StatusOK)
	c.Assert(headers.Get("Content-Type"), Equals, "application/json")
	c.Assert(string(body), Equals, fmt.Sprintf(`{"name":"New name","alias":"%s","users":["%s"],"owner":"%s"}`, team.Alias, user.Email, user.Email))
}

func (s *S) TestUpdateTeamNotMember(c *C) {
	alice := account.User{Name: "alice", Email: "alice@bar.example.org", Password: "secret"}
	alice.Create()
	defer alice.Delete()

	team := account.Team{Name: "ApiHub Team", Alias: "apihub"}
	team.Create(alice)

	defer func() {
		s.store.DeleteTeamByAlias(team.Alias)
	}()

	headers, code, body, _ := httpClient.MakeRequest(requests.Args{
		AcceptableCode: http.StatusOK,
		Method:         "PUT",
		Path:           fmt.Sprintf("/api/teams/%s", team.Alias),
		Body:           `{"name": "New name"}`,
		Headers:        http.Header{"Authorization": {s.authHeader}},
	})

	c.Assert(string(body), Equals, `{"error":"access_denied","error_description":"You do not belong to this team!"}`)
	c.Assert(code, Equals, http.StatusForbidden)
	c.Assert(headers.Get("Content-Type"), Equals, "application/json")
}

func (s *S) TestUpdateTeamNotFound(c *C) {
	headers, code, body, _ := httpClient.MakeRequest(requests.Args{
		AcceptableCode: http.StatusOK,
		Method:         "PUT",
		Path:           "/api/teams/not-found",
		Body:           `{"name": "New name"}`,
		Headers:        http.Header{"Authorization": {s.authHeader}},
	})

	c.Assert(code, Equals, http.StatusNotFound)
	c.Assert(headers.Get("Content-Type"), Equals, "application/json")
	c.Assert(string(body), Equals, `{"error":"not_found","error_description":"Team not found."}`)
}
