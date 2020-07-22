/*
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package utils

import "strconv"

type HttpCodeStatus string

const (
	Http1xx HttpCodeStatus = "1xx"
	Http2xx HttpCodeStatus = "2xx"
	Http3xx HttpCodeStatus = "3xx"
	Http4xx HttpCodeStatus = "4xx"
	Http5xx HttpCodeStatus = "5xx"
)

func is5xxStatusCode(statusCode int) bool {
	return isXxxStatusCode(statusCode, Http5xx)
}

func is4xxStatusCode(statusCode int) bool {
	return isXxxStatusCode(statusCode, Http4xx)
}

func is2xxStatusCode(statusCode int) bool {
	return isXxxStatusCode(statusCode, Http2xx)
}

func isXxxStatusCode(statusCode int, xxx HttpCodeStatus) bool {
	switch convertStatusCode(statusCode) {
	case xxx:
		return true
	default:
		return false
	}
}

func convertStatusCode(statusCodeInt int) HttpCodeStatus {
	firstDigit := strconv.Itoa(statusCodeInt)[:1]
	var statusCode HttpCodeStatus

	switch firstDigit {
	case "1":
		statusCode = Http1xx
	case "2":
		statusCode = Http2xx
	case "3":
		statusCode = Http3xx
	case "4":
		statusCode = Http4xx
	default:
		statusCode = Http5xx
	}

	return statusCode
}
