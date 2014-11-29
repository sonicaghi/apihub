package system

import (
	"io"
	"net/http"
	"reflect"

	"github.com/albertoleal/backstage/api/controllers"
	"github.com/zenazn/goji"
	"github.com/zenazn/goji/web"
	"github.com/zenazn/goji/web/middleware"
)

type Application struct {
}

func (app *Application) DrawRoutes() {
	goji.NotFound(NotFoundHandler)

	// Controllers
	serviceController := &controllers.ServicesController{}
	debugController := &controllers.DebugController{}

	// Public Routes
	goji.Get("/", app.Route(serviceController, "Index"))

	// Private Routes
	api := web.New()
	goji.Handle("/api/*", api)
	api.Use(middleware.SubRouter)
	api.NotFound(NotFoundHandler)
	api.Use(AuthorizationMiddleware)
	api.Use(ErrorHandlerMiddleware)
	api.Get("/helloworld", app.Route(debugController, "HelloWorld"))
}

func (app *Application) Route(controller interface{}, route string) interface{} {
	fn := func(c web.C, w http.ResponseWriter, r *http.Request) {
		c.Env["Content-Type"] = "application/json"

		methodValue := reflect.ValueOf(controller).MethodByName(route)
		addHeaders := reflect.ValueOf(controller).MethodByName("AddHeaders")
		methodInterface := methodValue.Interface()
		addHeadersInterface := addHeaders.Interface()

		method := methodInterface.(func(c web.C, w http.ResponseWriter, r *http.Request) (string, int))

		addmethod := addHeadersInterface.(func(w http.ResponseWriter))
		addmethod(w)
		body, code := method(c, w, r)

		switch code {
		case http.StatusOK:
			if _, exists := c.Env["Content-Type"]; exists {
				w.Header().Set("Content-Type", c.Env["Content-Type"].(string))
			}
			io.WriteString(w, body)
		case http.StatusSeeOther, http.StatusFound:
			http.Redirect(w, r, body, code)
		}
	}
	return fn
}