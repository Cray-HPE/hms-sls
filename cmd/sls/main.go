// MIT License
//
// (C) Copyright [2019, 2021-2022] Hewlett Packard Enterprise Development LP
//
// Permission is hereby granted, free of charge, to any person obtaining a
// copy of this software and associated documentation files (the "Software"),
// to deal in the Software without restriction, including without limitation
// the rights to use, copy, modify, merge, publish, distribute, sublicense,
// and/or sell copies of the Software, and to permit persons to whom the
// Software is furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included
// in all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
// THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
// OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
// ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
// OTHER DEALINGS IN THE SOFTWARE.

package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/namsral/flag"

	"github.com/Cray-HPE/hms-sls/v2/internal/database"
	"github.com/gorilla/mux"
	cache "github.com/victorspringer/http-cache"
	"github.com/victorspringer/http-cache/adapter/memory"
)

type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

type Routes []Route

const (
	API_ROOT      = "/v1"
	API_READINESS = API_ROOT + "/readiness"
	API_LIVENESS  = API_ROOT + "/liveness"
	API_HEALTH    = API_ROOT + "/health"
	API_READY     = API_ROOT + "/ready" // DEPREICATED
	API_VERSION   = API_ROOT + "/version"
	API_HARDWARE  = API_ROOT + "/hardware"
	API_NETWORKS  = API_ROOT + "/networks"
	API_SEARCH    = API_ROOT + "/search"
	API_DUMPSTATE = API_ROOT + "/dumpstate"
	API_LOADSTATE = API_ROOT + "/loadstate"
)

var httpAddr string
var cacheLayerEnable bool
var cacheTTLSeconds int
var cacheCapacity int
var debugLevel int

var vaultKeypath string

var Running = true

// Generate the API routes
func newRouter(routes []Route) *mux.Router {
	router := mux.NewRouter().StrictSlash(true)
	for _, route := range routes {
		var handler http.Handler
		handler = route.HandlerFunc
		router.
			Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(handler)
	}
	return router
}

// Create the API route descriptors.

func generateRoutes() Routes {
	return Routes{
		Route{"doReadinessGet",
			strings.ToUpper("Get"),
			API_READINESS,
			doReadinessGet,
		},
		Route{"doLivenessGet",
			strings.ToUpper("Get"),
			API_LIVENESS,
			doLivenessGet,
		},
		Route{"doHealthGet",
			strings.ToUpper("Get"),
			API_HEALTH,
			doHealthGet,
		},
		Route{"doVersionGet",
			strings.ToUpper("Get"),
			API_VERSION,
			doVersionGet,
		},

		// Hardware
		Route{"doHardwarePost",
			strings.ToUpper("Post"),
			API_HARDWARE,
			doHardwarePost,
		},
		Route{"doHardwareGet",
			strings.ToUpper("Get"),
			API_HARDWARE,
			doHardwareGet,
		},
		Route{"doHardwareObjGet",
			strings.ToUpper("Get"),
			API_HARDWARE + "/{xname}",
			doHardwareObjGet,
		},
		Route{"doHardwareObjPut",
			strings.ToUpper("Put"),
			API_HARDWARE + "/{xname}",
			doHardwareObjPut,
		},
		Route{"doHardwareObjDelete",
			strings.ToUpper("Delete"),
			API_HARDWARE + "/{xname}",
			doHardwareObjDelete,
		},

		// Networks
		Route{"doNetworksGet",
			strings.ToUpper("Get"),
			API_NETWORKS,
			doNetworksGet,
		},
		Route{"doNetworksPost",
			strings.ToUpper("Post"),
			API_NETWORKS,
			doNetworksPost,
		},
		Route{"doNetworkObjGet",
			strings.ToUpper("Get"),
			API_NETWORKS + "/{network}",
			doNetworkObjGet,
		},
		Route{"doNetworkObjPut",
			strings.ToUpper("Put"),
			API_NETWORKS + "/{network}",
			doNetworkObjPut,
		},
		Route{"doNetworkObjPatch",
			strings.ToUpper("Patch"),
			API_NETWORKS + "/{network}",
			doNetworkObjPatch,
		},
		Route{"doNetworkObjDelete",
			strings.ToUpper("Delete"),
			API_NETWORKS + "/{network}",
			doNetworkObjDelete,
		},

		Route{"doHardwareSearch",
			strings.ToUpper("Get"),
			API_SEARCH + "/hardware",
			doHardwareSearch,
		},
		Route{"doNetworksSearch",
			strings.ToUpper("Get"),
			API_SEARCH + "/networks",
			doNetworksSearch,
		},
		Route{"doDumpState",
			strings.ToUpper("Get"),
			API_DUMPSTATE,
			doDumpState,
		},
		Route{"doDumpStateWithVaultData",
			strings.ToUpper("Post"),
			API_DUMPSTATE,
			doPostDumpState,
		},
		Route{"doLoadState",
			strings.ToUpper("Post"),
			API_LOADSTATE,
			doLoadState,
		},
	}
}

// Grab relevant environment variables.  Each one meant for sls starts
// with SLS_ and is named the same as its cmdline argument equivalent,
// all in upper case.

func envVars() {
	var envstr string

	envstr = os.Getenv("SLS_HTTP_LISTEN_ADDR")
	if envstr != "" {
		httpAddr = envstr
	}
	envstr = os.Getenv("SLS_DEBUG")
	if envstr != "" {
		var err error
		debugLevel, err = strconv.Atoi(envstr)
		if err != nil {
			log.Printf("Setting env var SLS_DEBUG bad value (%s), setting to 0.\n",
				envstr)
			debugLevel = 0
		}
	}
}

func main() {
	log.Printf("INFO: Starting SLS")

	var httpAddr string
	var datastoreBase string

	flag.StringVar(&httpAddr, "http_listen_addr", ":8376",
		"The address (in [address]:port) on which to expose SLS's HTTP interface")
	flag.BoolVar(&cacheLayerEnable, "cache_layer_enable", false,
		"Enable the HTTP caching middleware")
	flag.IntVar(&cacheTTLSeconds, "cache_ttl_seconds", 15,
		"Set the caches Time To Live (TTL) in seconds")
	flag.IntVar(&cacheCapacity, "cache_capacity", 1000, // TODO tune
		"Set the size of the cache")
	flag.IntVar(&debugLevel, "debug", 0, "Debug level")
	flag.Parse()
	envVars()

	// Hook up the API routes
	routes := generateRoutes()
	router := newRouter(routes)

	srv := &http.Server{
		Addr:    httpAddr,
		Handler: router,
	}

	if cacheLayerEnable {
		log.Printf("INFO: Enabling cache layer with capacity %d and TTL of %d seconds\n", cacheCapacity, cacheTTLSeconds)

		// Setup Caching
		memcached, err := memory.NewAdapter(
			memory.AdapterWithAlgorithm(memory.LRU),
			memory.AdapterWithCapacity(cacheCapacity),
		)
		if err != nil {
			log.Fatalf("ERROR: Failed to setup memory cache adapter. Error: %s\n", err.Error())
		}

		cacheClient, err := cache.NewClient(
			cache.ClientWithMethods([]string{http.MethodGet}),
			cache.ClientWithAdapter(memcached),
			cache.ClientWithTTL(time.Second*time.Duration(cacheTTLSeconds)),
			// cache.ClientWithRefreshKey("opn"), // TODO
		)
		if err != nil {
			log.Fatalf("ERROR: Failed to setup cache client. Error: %s\n", err.Error())
		}

		// Apply caching middleware, but exclude it from liveness readiness
		router.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == API_READINESS || r.URL.Path == API_LIVENESS {
					// For readiness and liveness bypass the caching layer, as this would obscure and delay responses that k8s
					// needs for livesness and readiness probes
					next.ServeHTTP(w, r)
				} else {
					// Continue onto the caching middle ware
					cacheClient.Middleware(next).ServeHTTP(w, r)
				}
			})
		})
	} else {
		log.Println("INFO: Caching layer is disabled")
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	idleConnsClosed := make(chan struct{})
	go func() {
		<-c
		Running = false

		// Gracefully shutdown the HTTP server.
		if err := srv.Shutdown(context.Background()); err != nil {
			// Error from closing listeners, or context timeout:
			log.Printf("HTTP server Shutdown: %v", err)
		}
		close(idleConnsClosed)
	}()

	log.Printf("DEBUG: Parsing command line options")
	log.Printf("INFO: Configuration: HTTP Listen Address: %s", httpAddr)
	log.Printf("INFO: Backing datastore: %s", datastoreBase)
	log.Printf("DEBUG: Done parsing command line options")

	log.Printf("DEBUG: Connecting to database...")
	err := database.NewDatabase()
	if err != nil {
		// The NewDatabase method will try forever to connect, if we get to this point it really is time to panic.
		panic(err)
	}

	log.Printf("INFO: Beginning to serve HTTP")
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		// Error starting or closing listener:
		log.Fatalf("HTTP server ListenAndServe: %v", err)
	}

	log.Printf("HTTP server shutdown, waiting for idle connection to close...")

	<-idleConnsClosed

	log.Printf("Done. Exiting.")

	_ = database.CloseDatabase()
}
