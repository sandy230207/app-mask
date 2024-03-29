package routes

import (
	"fmt"
	"net/http"

	controller "app-mask/controller"

	"github.com/gorilla/mux"
)

type Route struct {
	Method     string
	Pattern    string
	Handler    http.HandlerFunc
	Middleware mux.MiddlewareFunc
}

var routes []Route

func init() {
	register("POST", "/api/signIn", controller.SignIn, nil)                           // ok
	register("POST", "/api/signUp", controller.SignUp, nil)                           // ok
	register("GET", "/api/queryStockByDate/{date}", controller.QueryStockByDate, nil) // ok
	register("GET", "/api/queryStore", controller.QueryStore, nil)                    // ok
	register("GET", "/api/queryStockByStore/{id}", controller.QueryStockByStore, nil) // ok
	register("POST", "/api/queryHistoryOrder", controller.QueryHistoryOrder, nil)     // ok
	register("POST", "/api/book", controller.Book, nil)                               // ok
	register("POST", "/api/cancelOrder", controller.CancelOrder, nil)                 // ok

	register("GET", "/api/queryUser", controller.QueryUser, nil)           // ok
	register("GET", "/api/queryOrder", controller.QueryOrder, nil)         // ok
	register("GET", "/api/queryInventory", controller.QueryInventory, nil) // ok

	register("POST", "/api/insertInventory", controller.InsertInventory, nil) // ok
	register("POST", "/api/insertStore", controller.InsertStore, nil)         // ok
	register("POST", "/api/pickUp", controller.PickUp, nil)                   // ok

	register("GET", "/api/healthCheck", controller.HealthCheck, nil) // ok

}

func NewRouter() *mux.Router {
	r := mux.NewRouter()
	for _, route := range routes {
		r.Methods(route.Method).
			Path(route.Pattern).
			Handler(route.Handler)
		if route.Middleware != nil {
			r.Use(route.Middleware)
		}
	}
	return r
}

// http to https
func RedirectToHTTPSRouter(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		proto := req.Header.Get("x-forwarded-proto")
		if proto == "http" || proto == "HTTP" {
			http.Redirect(res, req, fmt.Sprintf("https://%s%s", req.Host, req.URL), http.StatusPermanentRedirect)
			return
		}

		next.ServeHTTP(res, req)

	})
}

func register(method, pattern string, handler http.HandlerFunc, middleware mux.MiddlewareFunc) {
	routes = append(routes, Route{method, pattern, handler, middleware})
}
