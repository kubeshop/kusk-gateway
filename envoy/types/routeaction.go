package types

//func generateRoute(
//	clusterName string,
//	corsPolicy *route.CorsPolicy,
//	rewritePathRegex *envoytypematcher.RegexMatchAndSubstitute,
//	timeout int64,
//	idleTimeout int64,
//	retries uint32) (*route.Route_Route, error) {
//
//	routeRoute := &route.Route_Route{
//		Route: &route.RouteAction{
//			ClusterSpecifier: &route.RouteAction_Cluster{
//				Cluster: clusterName,
//			},
//		},
//	}
//
//	if corsPolicy != nil {
//		routeRoute.Route.Cors = corsPolicy
//	}
//	if rewritePathRegex != nil {
//		routeRoute.Route.RegexRewrite = rewritePathRegex
//	}
//
//	if timeout != 0 {
//		routeRoute.Route.Timeout = &durationpb.Duration{Seconds: timeout}
//	}
//	if idleTimeout != 0 {
//		routeRoute.Route.IdleTimeout = &durationpb.Duration{Seconds: idleTimeout}
//	}
//
//	if retries != 0 {
//		routeRoute.Route.RetryPolicy = &route.RetryPolicy{
//			RetryOn:    "5xx",
//			NumRetries: &wrappers.UInt32Value{Value: retries},
//		}
//	}
//	if err := routeRoute.Route.Validate(); err != nil {
//		return nil, fmt.Errorf("incorrect Route Action: %w", err)
//	}
//	return routeRoute, nil
//}
