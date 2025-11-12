package transformer

import (
	"encoding/json"
	"log"
	"strconv"
	"strings"

	trafficpb "client-message-transformer/protobuf/traffic_payload"
)

// TransformToProto converts the transformed message to protobuf format
func TransformToProto(data []byte, clientID string) (*trafficpb.HttpResponseParam, error) {
	log.Printf("🔄 [PROTO TRANSFORMER] Starting protobuf transformation for client: %s", clientID)

	var input map[string]interface{}
	err := json.Unmarshal(data, &input)
	if err != nil {
		log.Printf("❌ [PROTO TRANSFORMER] JSON parse error: %v", err)
		return nil, err
	}

	// Helper to safely get nested value
	getNestedString := func(parent map[string]interface{}, keys ...string) string {
		current := parent
		for i, key := range keys {
			if i == len(keys)-1 {
				if val, ok := current[key]; ok {
					if strVal, ok := val.(string); ok {
						return strVal
					}
				}
				return ""
			}
			if next, ok := current[key].(map[string]interface{}); ok {
				current = next
			} else {
				return ""
			}
		}
		return ""
	}

	getNestedFloat := func(parent map[string]interface{}, keys ...string) float64 {
		current := parent
		for i, key := range keys {
			if i == len(keys)-1 {
				if val, ok := current[key]; ok {
					if floatVal, ok := val.(float64); ok {
						return floatVal
					}
				}
				return 0
			}
			if next, ok := current[key].(map[string]interface{}); ok {
				current = next
			} else {
				return 0
			}
		}
		return 0
	}

	// Parse headers helper
	parseHeaders := func(headersStr string) map[string]*trafficpb.StringList {
		headers := make(map[string]*trafficpb.StringList)
		if headersStr == "" {
			return headers
		}

		var headersMap map[string]interface{}
		err := json.Unmarshal([]byte(headersStr), &headersMap)
		if err != nil {
			log.Printf("⚠️  [PROTO TRANSFORMER] Failed to parse headers: %v", err)
			return headers
		}

		for name, value := range headersMap {
			var values []string
			switch v := value.(type) {
			case string:
				values = []string{v}
			case []interface{}:
				for _, item := range v {
					if str, ok := item.(string); ok {
						values = append(values, str)
					}
				}
			}
			headers[strings.ToLower(name)] = &trafficpb.StringList{
				Values: values,
			}
		}
		return headers
	}

	// Extract from nested payload structure
	request, _ := input["request"].(map[string]interface{})
	fullURL := getNestedString(request, "url")
	path := extractURI(fullURL)
	method := getNestedString(request, "method")
	requestHeaders := getNestedString(request, "headers")
	requestPayload := getNestedString(request, "body")

	// Response fields
	response, _ := input["response"].(map[string]interface{})
	responseHeaders := getNestedString(response, "headers")
	responsePayload := getNestedString(response, "body")
	statusCode := int32(getNestedFloat(response, "statusCode"))

	// Info fields
	info, _ := input["info"].(map[string]interface{})
	clientIP := getNestedString(info, "ip")
	dateTime := int64(getNestedFloat(info, "dateTime"))

	// Parse headers into protobuf format
	reqHeaderMap := parseHeaders(requestHeaders)

	// Add host header
	if host := extractHostFromURL(fullURL); host != "" {
		reqHeaderMap["host"] = &trafficpb.StringList{
			Values: []string{host},
		}
	}

	respHeaderMap := parseHeaders(responseHeaders)

	// Build protobuf message
	payload := &trafficpb.HttpResponseParam{
		Method:          method,
		Path:            path,
		Type:            "HTTP/1.1",
		RequestHeaders:  reqHeaderMap,
		RequestPayload:  requestPayload,
		ResponseHeaders: respHeaderMap,
		ResponsePayload: responsePayload,
		Ip:              clientIP,
		Time:            int32(dateTime / 1000), // Convert to seconds
		StatusCode:      statusCode,
		Status:          getStatus(int(statusCode)),
		AktoAccountId:   clientID,
		AktoVxlanId:     "0",      // Default value
		IsPending:       false,     // Default value
		Source:          "MIRRORING",
		Direction:       "",        // Not available in client message
		DestIp:          "",        // Not available in client message
	}

	log.Printf("✅ [PROTO TRANSFORMER] Protobuf transformation completed - Method: %s, Path: %s, Status: %d", method, path, statusCode)

	return payload, nil
}

// TransformToProtoFromFlat converts the flat JSON format to protobuf format
func TransformToProtoFromFlat(flatData map[string]interface{}) (*trafficpb.HttpResponseParam, error) {
	// Helper to safely get string from map
	getString := func(key string) string {
		if val, ok := flatData[key]; ok {
			if strVal, ok := val.(string); ok {
				return strVal
			}
		}
		return ""
	}

	// Helper to safely get int32 from map
	getInt32 := func(key string) int32 {
		if val, ok := flatData[key]; ok {
			switch v := val.(type) {
			case int:
				return int32(v)
			case int32:
				return v
			case int64:
				return int32(v)
			case float64:
				return int32(v)
			case string:
				if i, err := strconv.ParseInt(v, 10, 32); err == nil {
					return int32(i)
				}
			}
		}
		return 0
	}

	// Parse headers helper
	parseHeaders := func(headersStr string) map[string]*trafficpb.StringList {
		headers := make(map[string]*trafficpb.StringList)
		if headersStr == "" {
			return headers
		}

		var headersMap map[string]interface{}
		err := json.Unmarshal([]byte(headersStr), &headersMap)
		if err != nil {
			log.Printf("⚠️  [PROTO TRANSFORMER] Failed to parse headers: %v", err)
			return headers
		}

		for name, value := range headersMap {
			var values []string
			switch v := value.(type) {
			case string:
				values = []string{v}
			case []interface{}:
				for _, item := range v {
					if str, ok := item.(string); ok {
						values = append(values, str)
					}
				}
			}
			headers[strings.ToLower(name)] = &trafficpb.StringList{
				Values: values,
			}
		}
		return headers
	}

	requestHeaders := getString("requestHeaders")
	responseHeaders := getString("responseHeaders")

	// Build protobuf message
	payload := &trafficpb.HttpResponseParam{
		Method:          getString("method"),
		Path:            getString("path"),
		Type:            getString("type"),
		RequestHeaders:  parseHeaders(requestHeaders),
		RequestPayload:  getString("requestPayload"),
		ResponseHeaders: parseHeaders(responseHeaders),
		ResponsePayload: getString("responsePayload"),
		Ip:              getString("ip"),
		Time:            getInt32("time"),
		StatusCode:      getInt32("statusCode"),
		Status:          getString("status"),
		AktoAccountId:   getString("akto_account_id"),
		AktoVxlanId:     "0",
		IsPending:       false,
		Source:          getString("source"),
		Direction:       "",
		DestIp:          "",
	}

	return payload, nil
}

// extractHostFromURL extracts the host from a URL
func extractHostFromURL(fullURL string) string {
	if fullURL == "" {
		return ""
	}

	// Simple host extraction
	parts := strings.Split(fullURL, "/")
	if len(parts) >= 3 {
		// URL format: scheme://host/path
		return parts[2]
	}

	return ""
}
