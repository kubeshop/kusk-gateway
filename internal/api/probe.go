package api

import "net/http"

type ProbeController struct{}

func NewProbeController() Router {
	return &ProbeController{}
}

func (c *ProbeController) Routes() Routes {
	return Routes{
		{
			"Health",
			http.MethodGet,
			"/",
			c.Healthz,
		},
	}
}

func (c *ProbeController) Healthz(w http.ResponseWriter, r *http.Request) {
	_ = EncodeJSONResponse("200 OK", nil, w)
}
