package j8a

import (
	"github.com/rs/zerolog/log"
	"net/http"
	"regexp"
	"time"
)

//Aboutj8a special Resource alias for internal endpoint
const about string = "about"

type Routes []Route

func (s Routes) Len() int {
	return len(s)
}
func (s Routes) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s Routes) Less(i, j int) bool {
	return len(s[i].Path) > len(s[j].Path)
}

//Route maps a Path to an upstream resource
type Route struct {
	Path      string
	PathRegex *regexp.Regexp
	Transform string
	Resource  string
	Policy    string
	Jwt       string
}

func (route Route) matchURI(request *http.Request) bool {
	matched := false
	if route.PathRegex != nil {
		matched = route.PathRegex.MatchString(request.RequestURI)
	} else {
		matched, _ = regexp.MatchString("^"+route.Path, request.RequestURI)
	}

	return matched
}

// maps a route to a URL. Returns the URL, the name of the mapped policy and whether mapping was successful
func (route Route) mapURL(proxy *Proxy) (*URL, string, bool) {
	var policy Policy
	var policyLabel string
	if len(route.Policy) > 0 {
		policy = Runner.Policies[route.Policy]
		policyLabel = policy.resolveLabel()
	}

	resource := Runner.Resources[route.Resource]
	if resource == nil {
		return nil, "", false
	}
	//if a policy exists, we match resources with a label. TODO: this should be an interface
	if len(route.Policy) > 0 {
		for _, resourceMapping := range resource {
			for _, resourceLabel := range resourceMapping.Labels {
				if policyLabel == resourceLabel {
					log.Trace().
						Str("route", route.Path).
						Str("upRes", resourceMapping.URL.String()).
						Str("label", resourceLabel).
						Str("policy", route.Policy).
						Str(XRequestID, proxy.XRequestID).
						Int64("dwnElapsedMicros", time.Since(proxy.Dwn.startDate).Microseconds()).
						Msg("upstream resource mapped")
					return &resourceMapping.URL, policyLabel, true
				}
			}
		}
	} else {
		log.Trace().
			Str("routePath", route.Path).
			Str("policy", "default").
			Str(XRequestID, proxy.XRequestID).
			Str("upstream", resource[0].URL.String()).
			Msg("route mapped")
		return &resource[0].URL, "default", true
	}

	log.Trace().
		Str("routePath", route.Path).
		Str(XRequestID, proxy.XRequestID).
		Msg("route not mapped")

	return nil, "", false
}

func (route Route) hasJwt() bool {
	return len(route.Jwt) > 0
}
