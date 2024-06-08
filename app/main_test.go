package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var testPort = 4222
var testDir = os.TempDir()

func startTestServer() {
	server := NewApp(testDir)

	if err := server.Listen(fmt.Sprintf(":%d", testPort)); err != nil {
		fmt.Printf("Error starting server: %s", err)
	}
}

func TestMain(m *testing.M) {
	go startTestServer()

	if waitForPortOpen("localhost", testPort, 100*time.Millisecond, 5*time.Second) {
		m.Run()
	} else {
		panic("Server did not start")
	}
}

func TestHandleEcho(t *testing.T) {
	resp, err := http.Get(fmt.Sprintf("http://localhost:%d/echo/hello", testPort))
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Equal(t, "hello", string(body))
}

func TestHandleUserAgent(t *testing.T) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", fmt.Sprintf("http://localhost:%d/user-agent", testPort), nil)
	assert.NoError(t, err)
	req.Header.Set("User-Agent", "TestAgent")

	resp, err := client.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Equal(t, "TestAgent", string(body))
}

func TestHandleNotFound(t *testing.T) {
	resp, err := http.Get(fmt.Sprintf("http://localhost:%d/unknown", testPort))
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestHandleFound(t *testing.T) {
	resp, err := http.Get(fmt.Sprintf("http://localhost:%d/", testPort))
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	_, err = io.ReadAll(resp.Body)
	assert.NoError(t, err)
}

func TestHandlePostFile(t *testing.T) {
	content := "test file content"
	body := strings.NewReader(content)
	resp, err := http.Post(fmt.Sprintf("http://localhost:%d/files/test.txt", testPort), "text/plain", body)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	filepath := fmt.Sprintf("%s/test.txt", testDir)
	fileContent, err := os.ReadFile(filepath)
	assert.NoError(t, err)
	assert.Equal(t, string(fileContent), content)
}

func TestHandleGetFile(t *testing.T) {
	file, _ := os.Create(fmt.Sprintf("%s/test.txt", testDir))
	file.WriteString("test file content")
	defer file.Close()

	resp, err := http.Get(fmt.Sprintf("http://localhost:%d/files/test.txt", testPort))
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "application/octet-stream", resp.Header.Get("Content-Type"))

	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Equal(t, "test file content", string(body))
}
