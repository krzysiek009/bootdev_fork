package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/spf13/viper"
)

type LoginRequest struct {
	Otp string `json:"otp"`
}

type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func FetchAccessToken() (*LoginResponse, error) {
	api_url := viper.GetString("api_url")
	client := &http.Client{}
	r, err := http.NewRequest("POST", api_url+"/v1/auth/refresh", bytes.NewBuffer([]byte{}))
	r.Header.Add("X-Refresh-Token", viper.GetString("refresh_token"))
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(r)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, errors.New("Invalid refresh token")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var creds LoginResponse
	err = json.Unmarshal(body, &creds)
	return &creds, nil
}

func LoginWithCode(code string) (*LoginResponse, error) {
	api_url := viper.GetString("api_url")
	req, err := json.Marshal(LoginRequest{Otp: code})
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(api_url+"/v1/auth/otp/login", "application/json", bytes.NewReader(req))
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, errors.New("Invalid login code")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var creds LoginResponse
	err = json.Unmarshal(body, &creds)
	if err != nil {
		return nil, err
	}

	return &creds, nil
}

func fetchWithAuth(method string, url string) ([]byte, error) {
	body, code, err := fetchWithAuthAndPayload(method, url, []byte{})
	if err != nil {
		return nil, err
	}
	if code != 200 {
		return nil, errors.New(fmt.Sprintf("Failed to %s to %s\nResponse: %d %s\n", method, url, code, string(body)))
	}
	return body, err
}

func fetchWithAuthAndPayload(method string, url string, payload []byte) ([]byte, int, error) {
	api_url := viper.GetString("api_url")
	client := &http.Client{}
	r, err := http.NewRequest(method, api_url+url, bytes.NewBuffer(payload))
	r.Header.Add("X-Access-Token", viper.GetString("access_token"))
	if err != nil {
		return nil, 0, err
	}
	resp, err := client.Do(r)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, err
	}

	return body, resp.StatusCode, nil
}
