package httptoreq

import (
	"bufio"
	"bytes"
	"fmt"
	"net/http"
	"strings"
)

func GetOne(data []byte) (*http.Request, error) {
	file := bytes.NewReader(data)
	scanner := bufio.NewScanner(file)
	if !scanner.Scan() {
		return nil, fmt.Errorf("No request")
	}
	firstLine := scanner.Text()
	if !isReq(firstLine) {
		return nil, fmt.Errorf("invalid request")
	}
	method, url, err := lineToReq(firstLine)
	if err != nil {
		return nil, err
	}
	headers, err := readHeaders(scanner)
	if err != nil {
		return nil, err
	}
	var req *http.Request

	if method == http.MethodPost {
		var body bytes.Buffer
		for scanner.Scan() {
			line := scanner.Text()
			if strings.TrimSpace(line) == "" {
				continue
			}
			if strings.TrimSpace(line) == "###" {
				break
			}
			body.WriteString(strings.TrimSpace(line))
		}
		req, err = http.NewRequest(method, url, &body)
		if err != nil {
			return nil, err
		}
	} else {
		req, err = http.NewRequest(method, url, nil)
		if err != nil {
			return nil, err
		}
	}
	req.Header = headers
	return req, nil
}

func ParseFile(data []byte) ([]*http.Request, error) {
	file := bytes.NewReader(data)
	scanner := bufio.NewScanner(file)
	reqsts := []*http.Request{}
	for {
		req, hasN, err := ParseReq(scanner)
		if err != nil {
			return nil, err
		}
		reqsts = append(reqsts, req)
		if !hasN {
			break
		}
	}
	return reqsts, nil
}

func ParseReq(scanner *bufio.Scanner) (*http.Request, bool, error) {
	var req bytes.Buffer
	var hasNewReq bool
	var startLine bool
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !startLine {
			startLine = isReq(line)
		}
		if line == "###" {
			fmt.Println("new request")
			hasNewReq = true
			break
		}
		if startLine {
			req.WriteString(line + "\n")
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, false, fmt.Errorf("Scanner err: %w", err)
	}
	request, err := http.ReadRequest(bufio.NewReader(&req))
	if err != nil {
		return nil, false, err
	}
	return request, hasNewReq, nil
}

func checkMethod(method string) bool {
	validMethods := []string{
		http.MethodGet,
		http.MethodPost,
		http.MethodPut,
		http.MethodDelete,
		http.MethodPatch,
		http.MethodHead,
		http.MethodOptions,
	}
	for _, validMethod := range validMethods {
		if method == validMethod {
			return true
		}
	}
	return false
}
func isReq(line string) bool {
	return checkMethod(strings.Split(line, " ")[0])
}
func lineToReq(line string) (string, string, error) {
	v := strings.Split(line, " ")
	if len(v) < 2 {
		return "", "", fmt.Errorf("Invalid request line must contain method url")
	}
	return v[0], v[1], nil
}

func readHeaders(scanner *bufio.Scanner) (http.Header, error) {
	headers := make(http.Header)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			break
		}
		headerParts := strings.SplitN(line, ": ", 2)
		if len(headerParts) < 2 {
			return nil, fmt.Errorf("Invalid header: %s", line)
		}
		headers.Add(headerParts[0], headerParts[1])
	}
	return headers, nil
}
