package config

import (
	"fmt"

	accesslog "github.com/envoyproxy/go-control-plane/envoy/config/accesslog/v3"
	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	accesslogstream "github.com/envoyproxy/go-control-plane/envoy/extensions/access_loggers/stream/v3"
	"github.com/envoyproxy/go-control-plane/pkg/conversion"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/structpb"
)

const (
	AccessLogFormatJson string = "json"
	AccessLogFormatText string = "text"

	// The name of the Envoy extention that handles stdout access log
	AccessLogStdoutName string = "envoy.access_loggers.stdout"

	defaultTextLogTemplate = "[%START_TIME%] \"%REQ(:METHOD)% %REQ(X-ENVOY-ORIGINAL-PATH?:PATH)% " +
		"%PROTOCOL%\" %RESPONSE_CODE% %RESPONSE_FLAGS% " +
		"%RESPONSE_CODE_DETAILS% %CONNECTION_TERMINATION_DETAILS% " +
		"\"%UPSTREAM_TRANSPORT_FAILURE_REASON%\" %BYTES_RECEIVED% %BYTES_SENT% " +
		"%DURATION% %RESP(X-ENVOY-UPSTREAM-SERVICE-TIME)% \"%REQ(X-FORWARDED-FOR)%\" " +
		"\"%REQ(USER-AGENT)%\" \"%REQ(X-REQUEST-ID)%\" \"%REQ(:AUTHORITY)%\" \"%UPSTREAM_HOST%\" " +
		"%UPSTREAM_CLUSTER% %UPSTREAM_LOCAL_ADDRESS% %DOWNSTREAM_LOCAL_ADDRESS% " +
		"%DOWNSTREAM_REMOTE_ADDRESS% %REQUESTED_SERVER_NAME% %ROUTE_NAME%\n"
)

var (
	defaultJsonLogTemplate = &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"start_time":                        {Kind: &structpb.Value_StringValue{StringValue: "%START_TIME%"}},
			"route_name":                        {Kind: &structpb.Value_StringValue{StringValue: "%ROUTE_NAME%"}},
			"method":                            {Kind: &structpb.Value_StringValue{StringValue: "%REQ(:METHOD)%"}},
			"path":                              {Kind: &structpb.Value_StringValue{StringValue: "%REQ(X-ENVOY-ORIGINAL-PATH?:PATH)%"}},
			"protocol":                          {Kind: &structpb.Value_StringValue{StringValue: "%PROTOCOL%"}},
			"response_code":                     {Kind: &structpb.Value_StringValue{StringValue: "%RESPONSE_CODE%"}},
			"response_flags":                    {Kind: &structpb.Value_StringValue{StringValue: "%RESPONSE_FLAGS%"}},
			"response_code_details":             {Kind: &structpb.Value_StringValue{StringValue: "%RESPONSE_CODE_DETAILS%"}},
			"connection_termination_details":    {Kind: &structpb.Value_StringValue{StringValue: "%CONNECTION_TERMINATION_DETAILS%"}},
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
			"upstream_transport_failure_reason": {Kind: &structpb.Value_StringValue{StringValue: "%UPSTREAM_TRANSPORT_FAILURE_REASON%"}},
		},
	}
)

// AddLogger adds stdout logging configuration to HTTPManager
func (e *envoyConfiguration) AddLogger(logFormat string, logFormatTemplate interface{}) error {
	// Creating stdout logging configuration
	var accessLogFormatString *core.SubstitutionFormatString
	var err error
	// Depending on the logFormat (json or text), we parse logFormatTemplate for the type we expect there (string or map[string]string)
	switch logFormat {
	case AccessLogFormatJson:
		// If template is not the empty interface - check for provided template
		// otherwise use default template
		var formatTemplate *structpb.Struct = defaultJsonLogTemplate
		if logFormatTemplate != nil {
			jsonFormatMap, ok := logFormatTemplate.(map[string]string)
			if !ok {
				return fmt.Errorf("cannot assert log format template to map")
			}
			if len(jsonFormatMap) != 0 {
				// convert map[string]string to map[string]interface{} for the structpb.Struct conversion
				interfaceMap := make(map[string]interface{}, len(jsonFormatMap))
				for k, v := range jsonFormatMap {
					interfaceMap[k] = v
				}
				if formatTemplate, err = structpb.NewStruct(interfaceMap); err != nil {
					return fmt.Errorf("cannot convert log format template to struct")
				}
			}
		}
		accessLogFormatString = &core.SubstitutionFormatString{
			Format: &core.SubstitutionFormatString_JsonFormat{
				JsonFormat: formatTemplate,
			},
		}
	case AccessLogFormatText:
		// If template is not the empty interface - check for provided template
		// otherwise - use the default
		var formatTemplate string = defaultTextLogTemplate
		if logFormatTemplate != nil {
			textFormat, ok := logFormatTemplate.(string)
			if !ok {
				return fmt.Errorf("cannot assert log format template to string")
			}
			if textFormat != "" {
				formatTemplate = textFormat
			}
		}
		accessLogFormatString = &core.SubstitutionFormatString{
			Format: &core.SubstitutionFormatString_TextFormatSource{
				TextFormatSource: &core.DataSource{
					Specifier: &core.DataSource_InlineString{
						InlineString: formatTemplate,
					},
				},
			},
		}
	default:
		return fmt.Errorf("unknown log format type %s", logFormat)
	}
	accessLogConfig, err := conversion.MessageToStruct(
		&accesslogstream.StdoutAccessLog{
			AccessLogFormat: &accesslogstream.StdoutAccessLog_LogFormat{
				LogFormat: accessLogFormatString,
			},
		},
	)
	if err != nil {
		return fmt.Errorf("failed to convert access log config to struct: %w", err)
	}
	anyAccessLog, err := anypb.New(accessLogConfig)
	if err != nil {
		return fmt.Errorf("failed to convert access log config to Any message type: %w", err)
	}
	e.httpManager.AccessLog = append(e.httpManager.AccessLog,
		&accesslog.AccessLog{
			Name: AccessLogStdoutName,
			ConfigType: &accesslog.AccessLog_TypedConfig{
				TypedConfig: anyAccessLog,
			},
		},
	)
	return nil
}
