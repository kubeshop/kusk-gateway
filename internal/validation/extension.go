package validation

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"

	v31 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	pb "github.com/envoyproxy/go-control-plane/envoy/service/ext_proc/v3"
	v32 "github.com/envoyproxy/go-control-plane/envoy/type/v3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/go-logr/logr"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	contentType     = "Content-Type"
	applicationJSON = "application/json"
)

// Server provides OpenAPI Validation and implements ext_proc GRPC service.
type Server struct {
	services map[string]*Service
	log      logr.Logger
	m        sync.RWMutex
}

// NewServer() creates new validation Server.
func NewServer(log logr.Logger) *Server {
	return &Server{
		services: map[string]*Service{},
		log:      log,
	}

}

// Start starts the GRPC Server
func (s *Server) Start(port string) error {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		s.log.Error(err, "validation server failed to start at port", port)
		return err
	}
	s.log.Info("validation server listening at", port)
	sopts := []grpc.ServerOption{grpc.MaxConcurrentStreams(1000)}
	srv := grpc.NewServer(sopts...)

	pb.RegisterExternalProcessorServer(srv, s)
	err = srv.Serve(lis)
	if err != nil {
		s.log.Error(err, "validation server failed to start")
		return err
	}
	return nil
}

func (s *Server) Process(srv pb.ExternalProcessor_ProcessServer) error {
	s.log.WithName("Request validation:")
	header := make(http.Header)
	ctx := srv.Context()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		req, err := srv.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return status.Errorf(codes.Unknown, "cannot receive stream request: %v", err)
		}

		m, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return status.Error(codes.Unknown, "cannot parse metadata")
		}

		serviceID := m.Get(HeaderServiceID)
		if len(serviceID) != 1 {
			return status.Errorf(codes.Unknown, "cannot parse X-Kusk-Service-ID metadata: %v", serviceID)
		}
		s.m.RLock()
		service, ok := s.services[serviceID[0]]
		s.m.RUnlock()
		if !ok {
			return status.Errorf(codes.Unknown, "no such service in validation proxy: %s", serviceID[0])
		}

		operationID := m.Get(HeaderOperationID)
		if len(serviceID) != 1 {
			return status.Errorf(codes.Unknown, "cannot parse X-Kusk-Operation-ID metadata: %v", serviceID)
		}

		operation, ok := service.Operations[operationID[0]]
		if !ok {
			return status.Errorf(codes.Unknown, "no such operation in validation proxy: %s", operationID[0])
		}

		resp := &pb.ProcessingResponse{}
		switch v := req.Request.(type) {
		case *pb.ProcessingRequest_RequestHeaders:
			r := req.Request
			h := r.(*pb.ProcessingRequest_RequestHeaders)
			s.log.Info("Got RequestHeaders.Headers", h.RequestHeaders.Headers)
			for _, envoyHeader := range h.RequestHeaders.GetHeaders().GetHeaders() {
				header.Add(envoyHeader.Key, envoyHeader.Value)
			}

			if h.RequestHeaders.EndOfStream {
				u := &url.URL{
					Scheme: string(header.Get(":scheme")),
					Path:   string(header.Get(":path")),
					Host:   string(header.Get(":authority")),
				}
				req := &http.Request{
					Host:   "localhost",
					URL:    u,
					Method: string(header.Get(":method")),
					Header: header,
				}
				err = s.validate(req, service, operation)
				if err != nil {
					errorMsg := NewErrorBody()
					errorMsg.SetErrorBody(err)
					s.log.Error(fmt.Errorf(errorMsg.Error), "validation failed")

					resp = &pb.ProcessingResponse{
						Response: &pb.ProcessingResponse_ImmediateResponse{
							ImmediateResponse: &pb.ImmediateResponse{
								Status: &v32.HttpStatus{Code: v32.StatusCode_BadRequest},
								Body:   errorMsg.Error,
								Headers: &pb.HeaderMutation{
									SetHeaders: []*v31.HeaderValueOption{
										{
											Header: &v31.HeaderValue{
												Key:   contentType,
												Value: applicationJSON,
											},
										},
									},
								},
							},
						},
					}

					break
				}
				resp = &pb.ProcessingResponse{
					Response: &pb.ProcessingResponse_RequestHeaders{
						RequestHeaders: &pb.HeadersResponse{
							Response: &pb.CommonResponse{
								Status: pb.CommonResponse_CONTINUE,
							},
						},
					},
				}

				break
			}
			resp = &pb.ProcessingResponse{
				Response: &pb.ProcessingResponse_RequestHeaders{
					RequestHeaders: &pb.HeadersResponse{
						Response: &pb.CommonResponse{
							Status: pb.CommonResponse_CONTINUE,
						},
					},
				},
			}

		case *pb.ProcessingRequest_RequestBody:

			r := req.Request
			b := r.(*pb.ProcessingRequest_RequestBody)

			if b.RequestBody.EndOfStream {
				u := &url.URL{
					Scheme: string(header.Get(":scheme")),
					Path:   string(header.Get(":path")),
					Host:   string(header.Get(":authority")),
				}
				req := &http.Request{
					Host:   "localhost",
					URL:    u,
					Method: string(header.Get(":method")),
					Header: header,
					Body:   ioutil.NopCloser(bytes.NewBuffer(b.RequestBody.Body)),
				}

				err = s.validate(req, service, operation)
				if err != nil {
					errorMsg := NewErrorBody()
					errorMsg.SetErrorBody(err)
					s.log.Error(fmt.Errorf(errorMsg.Error), "validation failed")
					resp = &pb.ProcessingResponse{
						Response: &pb.ProcessingResponse_ImmediateResponse{
							ImmediateResponse: &pb.ImmediateResponse{
								Status: &v32.HttpStatus{Code: v32.StatusCode_BadRequest},
								Body:   errorMsg.Error,
								Headers: &pb.HeaderMutation{
									SetHeaders: []*v31.HeaderValueOption{
										{
											Header: &v31.HeaderValue{
												Key:   contentType,
												Value: applicationJSON,
											},
										},
									},
								},
							},
						},
					}
					break
				}
				resp = &pb.ProcessingResponse{
					Response: &pb.ProcessingResponse_RequestBody{
						RequestBody: &pb.BodyResponse{
							Response: &pb.CommonResponse{
								Status: pb.CommonResponse_CONTINUE,
							},
						},
					},
				}

			}
		default:
			s.log.Info("Unknown Request type ", v)
		}

		if err := srv.Send(resp); err != nil {
			s.log.Error(err, "send error")
		}
	}
}

// UpdateServices adds or updates Services to the validation service
func (s *Server) UpdateServices(services []*Service) {
	s.m.Lock()
	defer s.m.Unlock()

	// rebuild the services map
	s.services = make(map[string]*Service, len(services))

	for _, service := range services {
		s.services[service.ID] = service
	}
}

func (s *Server) validate(r *http.Request, service *Service, operation *operation) error {
	s.m.RLock()
	defer s.m.RUnlock()

	route, pathParams, err := service.Router.FindRoute(r)
	if err != nil {
		return err
	}

	return openapi3filter.ValidateRequest(context.Background(), &openapi3filter.RequestValidationInput{
		Request:     r,
		PathParams:  pathParams,
		QueryParams: nil,
		Route:       route,
		Options: &openapi3filter.Options{
			MultiError: true,
		},
	})
}

type ErrorBody struct {
	Error string `json:"error,omitempty"`
}

func NewErrorBody() *ErrorBody {
	return &ErrorBody{}
}

func (e *ErrorBody) SetErrorBody(err error) {
	errorMsg := ErrorBody{Error: err.Error()}
	jsn, _ := json.Marshal(errorMsg)
	msg := string(jsn)
	// removing '|' as openapi3filter.Multierror when printing adds a pipe at the end of the message https://github.com/getkin/kin-openapi/blob/master/openapi3/errors.go#L16
	e.Error = strings.ReplaceAll(msg, "|", "")
}