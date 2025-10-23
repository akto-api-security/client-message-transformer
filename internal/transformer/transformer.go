package transformer

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
)

// extractURI extracts only the path/URI from a full URL
func extractURI(fullURL string) string {
	if fullURL == "" {
		return ""
	}

	// Parse the URL
	parsedURL, err := url.Parse(fullURL)
	if err != nil {
		// If it fails to parse, assume it's already just a URI/path
		return fullURL
	}

	// Return path with query string if present
	if parsedURL.RawQuery != "" {
		return parsedURL.Path + "?" + parsedURL.RawQuery
	}
	return parsedURL.Path
}

// TransformMessage transforms from client nested format to standard flat format
func TransformMessage(data []byte, clientID string) (map[string]interface{}, error) {
	log.Printf("🔄 [TRANSFORMER] Starting transformation for client: %s", clientID)
	log.Printf("🔄 [TRANSFORMER] Input size: %d bytes", len(data))

	previewSize := len(data)
	if previewSize > 100 {
		previewSize = 100
	}
	log.Printf("🔄 [TRANSFORMER] Input preview: %s...", string(data[:previewSize]))

	var input map[string]interface{}
	err := json.Unmarshal(data, &input)
	if err != nil {
		log.Printf("❌ [TRANSFORMER] JSON parse error: %v", err)
		return nil, err
	}

	log.Printf("✅ [TRANSFORMER] JSON parsed successfully")

	// Extract nested payload structure
	output := make(map[string]interface{})

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

	// Extract from nested payload structure

	log.Printf("✅ [TRANSFORMER] Payload structure found")

	// Request fields
	request, _ := input["request"].(map[string]interface{})
	fullURL := getNestedString(request, "url")
	fmt.Println("[DEBUG] Full URL value:", fullURL)
	path := extractURI(fullURL)
	fmt.Println("[DEBUG] Extracted URI value:", path)
	method := getNestedString(request, "method")
	requestHeaders := request["headers"].(string)
	requestPayload := request["body"].(string)

	output["path"] = path
	output["method"] = method
	output["requestHeaders"] = requestHeaders
	output["requestPayload"] = requestPayload
	output["type"] = "HTTP/1.1"

	log.Printf("📥 [TRANSFORMER] Request extracted - Method: %s, Path: %s", method, path)

	// Response fields
	response, _ := input["response"].(map[string]interface{})
	responseHeaders := getNestedString(response, "headers")
	responsePayload := getNestedString(response, "body")
	statusCode := int(getNestedFloat(response, "statusCode"))

	output["responseHeaders"] = responseHeaders
	output["responsePayload"] = responsePayload
	output["statusCode"] = fmt.Sprintf("%d", statusCode)
	output["status"] = getStatus(statusCode)
	output["contentType"] = responseHeaders // Would need to parse from headers

	log.Printf("📤 [TRANSFORMER] Response extracted - Status: %d, Response size: %d bytes", statusCode, len(responsePayload))

	// Info fields
	info, _ := input["info"].(map[string]interface{})
	clientIP := getNestedString(info, "ip")
	dateTime := int64(getNestedFloat(info, "dateTime"))
	responseTime := int(getNestedFloat(info, "responseTime"))

	output["ip"] = clientIP
	output["time"] = fmt.Sprintf("%d", dateTime/1000) // Convert to seconds
	output["akto_account_id"] = clientID
	output["responseTime"] = responseTime
	output["source"] = "MIRRORING"

	log.Printf("ℹ️  [TRANSFORMER] Info extracted - IP: %s, Client ID: %s, Response Time: %dms", clientIP, clientID, responseTime)
	log.Printf("✅ [TRANSFORMER] Transformation completed successfully - Output has %d fields", len(output))

	return output, nil
}

// getStatus converts HTTP status code to status message
func getStatus(code int) string {
	statusMap := map[int]string{
		200: "OK",
		201: "Created",
		204: "No Content",
		400: "Bad Request",
		401: "Unauthorized",
		403: "Forbidden",
		404: "Not Found",
		500: "Internal Server Error",
		502: "Bad Gateway",
		503: "Service Unavailable",
	}
	if msg, ok := statusMap[code]; ok {
		return msg
	}
	return "Unknown"
}
