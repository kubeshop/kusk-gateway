/*
MIT License

# Copyright (c) 2022 Kubeshop

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/
package config

import (
	"fmt"

	accesslog "github.com/envoyproxy/go-control-plane/envoy/config/accesslog/v3"
	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	accesslogstream "github.com/envoyproxy/go-control-plane/envoy/extensions/access_loggers/stream/v3"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/structpb"
)

const (
	AccessLogFormatJson string = "json"
	AccessLogFormatText string = "text"
)

type AccessLogBuilder struct {
	al *accesslog.AccessLog
}

func (a *AccessLogBuilder) ValidateAll() error {
	return a.al.ValidateAll()
}

func (a *AccessLogBuilder) GetAccessLog() *accesslog.AccessLog {
	return a.al
}

func NewJSONAccessLog(template map[string]string) (*AccessLogBuilder, error) {
	var (
		// See https://istio.io/latest/docs/tasks/observability/logs/access-log/#default-access-log-format.
		defaultJsonLogTemplate = &structpb.Struct{
			Fields: map[string]*structpb.Value{
				"start_time":                        {Kind: &structpb.Value_StringValue{StringValue: "%START_TIME%"}},
				"method":                            {Kind: &structpb.Value_StringValue{StringValue: "%REQ(:METHOD)%"}},
				"path":                              {Kind: &structpb.Value_StringValue{StringValue: "%REQ(X-ENVOY-ORIGINAL-PATH?:PATH)%"}},
				"protocol":                          {Kind: &structpb.Value_StringValue{StringValue: "%PROTOCOL%"}},
				"response_code":                     {Kind: &structpb.Value_StringValue{StringValue: "%RESPONSE_CODE%"}},
				"response_flags":                    {Kind: &structpb.Value_StringValue{StringValue: "%RESPONSE_FLAGS%"}},
				"response_code_details":             {Kind: &structpb.Value_StringValue{StringValue: "%RESPONSE_CODE_DETAILS%"}},
				"connection_termination_details":    {Kind: &structpb.Value_StringValue{StringValue: "%CONNECTION_TERMINATION_DETAILS%"}},
				"upstream_transport_failure_reason": {Kind: &structpb.Value_StringValue{StringValue: "%UPSTREAM_TRANSPORT_FAILURE_REASON%"}},
				"bytes_received":                    {Kind: &structpb.Value_StringValue{StringValue: "%BYTES_RECEIVED%"}},
				"bytes_sent":                        {Kind: &structpb.Value_StringValue{StringValue: "%BYTES_SENT%"}},
				"duration":                          {Kind: &structpb.Value_StringValue{StringValue: "%DURATION%"}},
				"upstream_service_time":             {Kind: &structpb.Value_StringValue{StringValue: "%RESP(X-ENVOY-UPSTREAM-SERVICE-TIME)%"}},
				"x_forwarded_for":                   {Kind: &structpb.Value_StringValue{StringValue: "%REQ(X-FORWARDED-FOR)%"}},
				"user_agent":                        {Kind: &structpb.Value_StringValue{StringValue: "%REQ(USER-AGENT)%"}},
				"request_id":                        {Kind: &structpb.Value_StringValue{StringValue: "%REQ(X-REQUEST-ID)%"}},
				"authority":                         {Kind: &structpb.Value_StringValue{StringValue: "%REQ(:AUTHORITY)%"}},
				"upstream_host":                     {Kind: &structpb.Value_StringValue{StringValue: "%UPSTREAM_HOST%"}},
				"upstream_cluster":                  {Kind: &structpb.Value_StringValue{StringValue: "%UPSTREAM_CLUSTER%"}},
				"upstream_local_address":            {Kind: &structpb.Value_StringValue{StringValue: "%UPSTREAM_LOCAL_ADDRESS%"}},
				"downstream_local_address":          {Kind: &structpb.Value_StringValue{StringValue: "%DOWNSTREAM_LOCAL_ADDRESS%"}},
				"downstream_remote_address":         {Kind: &structpb.Value_StringValue{StringValue: "%DOWNSTREAM_REMOTE_ADDRESS%"}},
				"requested_server_name":             {Kind: &structpb.Value_StringValue{StringValue: "%REQUESTED_SERVER_NAME%"}},
				"route_name":                        {Kind: &structpb.Value_StringValue{StringValue: "%ROUTE_NAME%"}},
			},
		}
	)

	// Default template on the start
	var formatTemplate *structpb.Struct = defaultJsonLogTemplate

	if len(template) != 0 {
		// convert map[string]string to map[string]interface{} for the structpb.Struct conversion
		interfaceMap := make(map[string]interface{}, len(template))
		for k, v := range template {
			interfaceMap[k] = v
		}

		var err error
		if formatTemplate, err = structpb.NewStruct(interfaceMap); err != nil {
			return nil, fmt.Errorf("cannot convert log format template to struct")
		}
	}

	accessLogFormatString := &core.SubstitutionFormatString{
		Format: &core.SubstitutionFormatString_JsonFormat{
			JsonFormat: formatTemplate,
		},
	}

	return accessLogFinalize(accessLogFormatString)
}

func NewTextAccessLog(template string) (*AccessLogBuilder, error) {
	const (
		// See https://istio.io/latest/docs/tasks/observability/logs/access-log/#default-access-log-format
		defaultTextLogTemplate = `[%START_TIME%] "%REQ(:METHOD)% %REQ(X-ENVOY-ORIGINAL-PATH?:PATH)% %PROTOCOL%" %RESPONSE_CODE% %RESPONSE_FLAGS% %RESPONSE_CODE_DETAILS% %CONNECTION_TERMINATION_DETAILS%
"%UPSTREAM_TRANSPORT_FAILURE_REASON%" %BYTES_RECEIVED% %BYTES_SENT% %DURATION% %RESP(X-ENVOY-UPSTREAM-SERVICE-TIME)% "%REQ(X-FORWARDED-FOR)%" "%REQ(USER-AGENT)%" "%REQ(X-REQUEST-ID)%"
"%REQ(:AUTHORITY)%" "%UPSTREAM_HOST%" %UPSTREAM_CLUSTER% %UPSTREAM_LOCAL_ADDRESS% %DOWNSTREAM_LOCAL_ADDRESS% %DOWNSTREAM_REMOTE_ADDRESS% %REQUESTED_SERVER_NAME% %ROUTE_NAME%\n`
	)

	var formatTemplate string = defaultTextLogTemplate
	if template != "" {
		formatTemplate = template

	}
	accessLogFormatString := &core.SubstitutionFormatString{
		Format: &core.SubstitutionFormatString_TextFormatSource{
			TextFormatSource: &core.DataSource{
				Specifier: &core.DataSource_InlineString{
					InlineString: formatTemplate,
				},
			},
		},
	}

	return accessLogFinalize(accessLogFormatString)
}

// This block is shared between JSON and Text types of AccessLog creation
func accessLogFinalize(accessLogFormatString *core.SubstitutionFormatString) (*AccessLogBuilder, error) {
	accessLogConfig := &accesslogstream.StdoutAccessLog{
		AccessLogFormat: &accesslogstream.StdoutAccessLog_LogFormat{
			LogFormat: accessLogFormatString,
		},
	}

	anyAccessLog, err := anypb.New(accessLogConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to convert access log config to Any message type: %w", err)
	}

	accessLog := &accesslog.AccessLog{
		Name: "envoy.access_loggers.stdout",
		ConfigType: &accesslog.AccessLog_TypedConfig{
			TypedConfig: anyAccessLog,
		},
	}
	if err := accessLog.ValidateAll(); err != nil {
		return nil, fmt.Errorf("failed validation of the new access log: %w", err)
	}

	return &AccessLogBuilder{al: accessLog}, nil
}
