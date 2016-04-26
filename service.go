package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/golang/glog"

	"golang.org/x/tools/imports"
)

// Service is a helper service
type Service struct {
}

// NewService creates a new service
func NewService() *Service {
	return &Service{}
}

// A FormatRequest is a request to format
type FormatRequest struct {
	File string            `json:"file_path"`
	Text string            `json:"text"`
	Env  map[string]string `json:"env"`
}

// A FormatResult is the result of formatting
type FormatResult struct {
	Text string `json:"text"`
}

// Format formats a go source file
func (s *Service) Format(req *FormatRequest, res *FormatResult) error {
	formatted, err := imports.Process(req.File, []byte(req.Text), nil)
	if err != nil {
		return err
	}
	*res = FormatResult{
		Text: string(formatted),
	}
	return nil
}

// A GoToDefinitionRequest is a go-to-definition request
type GoToDefinitionRequest struct {
	File   string            `json:"file"`
	Text   string            `json:"text"`
	Offset int               `json:"offset"`
	Env    map[string]string `json:"env"`
}

// A GoToDefinitionResult is a go-to-definition result
type GoToDefinitionResult struct {
	File   string `json:"file"`
	Line   int    `json:"line"`
	Column int    `json:"column"`
}

// GoToDefinition uses godef to find the definition for the given cursor
func (s *Service) GoToDefinition(req *GoToDefinitionRequest, res *GoToDefinitionResult) error {
	cmd := exec.Command("godef", "-f", req.File, "-i", "-o", fmt.Sprint(req.Offset))
	cmd.Stdin = strings.NewReader(req.Text)
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Run()

	parts := strings.Split(strings.TrimSpace(buf.String()), ":")
	glog.Infof("REQUEST: %v, PARTS: %v", req, parts)
	if len(parts) == 3 {
		ln, _ := strconv.Atoi(parts[1])
		col, _ := strconv.Atoi(parts[2])

		*res = GoToDefinitionResult{
			File:   parts[0],
			Line:   ln,
			Column: col,
		}
		return nil
	} else if len(parts) == 2 {
		ln, lnerr := strconv.Atoi(parts[0])
		col, colerr := strconv.Atoi(parts[1])

		if lnerr == nil && colerr == nil {
			*res = GoToDefinitionResult{
				File:   req.File,
				Line:   ln,
				Column: col,
			}
			return nil
		}

	}

	return fmt.Errorf("not implemented")
}

type (
	// InstallRequest is an install request
	InstallRequest struct {
		Directory string `json:"directory"`
	}
	// InstallResult is an install result
	InstallResult struct {
	}
)

// Install basically does the equivalent of running `go install` on a project
func (s *Service) Install(req *InstallRequest, res *InstallResult) error {
	return fmt.Errorf("install not implemented")
}

// A LintRequest is the request for lint
type LintRequest struct {
	FilePath string            `json:"file_path"`
	Env      map[string]string `json:"env"`
}

// A LintResult is the result of running Lint
type LintResult struct {
	Messages []LintMessage `json:"messages"`
}

// A LintMessage is a lint line message
type LintMessage struct {
	Type     string `json:"type"`
	Line     int    `json:"line"`
	Column   int    `json:"column"`
	FilePath string `json:"file_path"`
	Message  string `json:"message"`
}

// Lint lints the given file
func (s *Service) Lint(req *LintRequest, res *LintResult) error {
	cmd := exec.Command("gometalinter", "--fast", filepath.Dir(req.FilePath))
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Run()

	re := regexp.MustCompile("^((?:(?::\\\\)|[^:])+):([0-9]+):([0-9]+)?:[ ]?(.*)$")

	*res = LintResult{}

	scanner := bufio.NewScanner(&buf)
	for scanner.Scan() {
		glog.Infof("text: %v", scanner.Text())
		matches := re.FindStringSubmatch(scanner.Text())
		if len(matches) >= 5 {
			filePath := matches[1]
			line, _ := strconv.Atoi(matches[2])
			column, _ := strconv.Atoi(matches[3])
			msg := matches[4]
			res.Messages = append(res.Messages, LintMessage{
				Type:     "warning",
				Line:     line,
				Column:   column,
				FilePath: filePath,
				Message:  msg,
			})
		}
	}

	return nil
}

// func (s *Service) Lint(req *LintRequest, res *LintResult) error {
// 	glog.Infof("RPC::Lint: %v", req.FilePath)
//
// 	cmd := exec.Command("gofmt", "-e")
//
// 	var wg sync.WaitGroup
// 	lines := make(chan string)
//
// 	stderr, err := cmd.StderrPipe()
// 	if err != nil {
// 		return err
// 	}
// 	defer stderr.Close()
//
// 	wg.Add(1)
// 	go func() {
// 		defer wg.Done()
// 		s := bufio.NewScanner(stderr)
// 		for s.Scan() {
// 			lines <- s.Text()
// 		}
// 	}()
//
// 	stdout, err := cmd.StdoutPipe()
// 	if err != nil {
// 		return err
// 	}
// 	defer stdout.Close()
//
// 	wg.Add(1)
// 	go func() {
// 		defer wg.Done()
// 		s := bufio.NewScanner(stderr)
// 		for s.Scan() {
// 			lines <- s.Text()
// 		}
// 	}()
//
// 	stdin, err := cmd.StdinPipe()
// 	if err != nil {
// 		return err
// 	}
//
// 	err = cmd.Start()
// 	if err != nil {
// 		stdin.Close()
// 		return err
// 	}
//
// 	wg.Add(1)
// 	go func() {
// 		defer wg.Done()
// 		io.WriteString(stdin, req.Text)
// 		stdin.Close()
// 		cmd.Wait()
// 	}()
//
// 	go func() {
// 		wg.Wait()
// 		close(lines)
// 	}()
//
// 	names := map[string]int{}
// 	for i, name := range pattern.SubexpNames() {
// 		names[name] = i
// 	}
//
// 	*res = LintResult{
// 		Messages: make([]LintMessage, 0),
// 	}
//
// 	for ln := range lines {
// 		for _, m := range pattern.FindAllStringSubmatch(ln, -1) {
// 			line, _ := strconv.Atoi(m[names["line"]])
// 			column, _ := strconv.Atoi(m[names["column"]])
// 			res.Messages = append(res.Messages, LintMessage{
// 				Type:    "error",
// 				Line:    line,
// 				Column:  column,
// 				Path:    req.FilePath,
// 				Message: m[names["message"]],
// 			})
// 			glog.Infof("PATH:%s LINE:%s COLUMN:%s MESSAGE:%s",
// 				m[names["path"]],
// 				m[names["line"]],
// 				m[names["column"]],
// 				m[names["message"]])
// 		}
// 		glog.Info(ln)
// 	}
//
// 	glog.Infof("RPC::Lint: %v", res)
//
// 	return nil
// }
