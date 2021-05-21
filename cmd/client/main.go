package main

import (
	"bytes"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
)

func main() {
	config, err := getConfigAndRequiredValues()
	if err != nil {
		log.Fatalf("Unable to load config: %s", err.Error())
	}
	if len(os.Args) == 1 {
		log.Fatalf("Filename is required")
	}
	fileName := os.Args[1]
	_, err = os.Stat(fileName)
	if os.IsNotExist(err) {
		log.Fatalf("File not found: %s", err.Error())
		return
	}
	file, err := os.Open(fileName)
	if os.IsNotExist(err) {
		log.Fatalf("Unable to open file: %s", err.Error())
		return
	}
	defer func() {
		_ = file.Close()
	}()
	body, err := doRequest(config, file)
	if err != nil {
		log.Fatalf("Unable to read body: %s", err.Error())
	}
	log.Printf("Response: %s", string(body))
}

func doRequest(config *Config, payload io.Reader) ([]byte, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "upload")
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(part, payload)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = writer.Close()
	}()
	_ = writer.WriteField("expiry", "0")
	_ = writer.WriteField("randomise", "true")
	req, err := http.NewRequest(http.MethodPost, config.URL+"/upload/file", body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-API-KEY", config.APIKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	log.Printf("Response: %d", resp.StatusCode)
	defer func() {
		_ = resp.Body.Close()
	}()
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return responseBody, nil
}

func getConfigAndRequiredValues() (*Config, error) {
	config, err := getConfig()
	if err != nil {
		return nil, err
	}
	err = getOrPromptForValue(&config.APIKey, "API Key")
	if err != nil {
		return nil, err
	}
	err = getOrPromptForValue(&config.URL, "Base URL")
	if err != nil {
		return nil, err
	}
	err = putConfig(config)
	if err != nil {
		return nil, err
	}
	return config, nil
}
