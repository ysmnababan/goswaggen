package testutil

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"learn-go/config"
	"log"
	"net/http"
	"net/url"

	"time"

	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type MailHogResponse struct {
	Total int `json:"total"`
	Count int `json:"count"`
	Start int `json:"start"`
	Items []struct {
		ID   string `json:"ID"`
		From struct {
			Mailbox string `json:"Mailbox"`
			Domain  string `json:"Domain"`
			Params  string `json:"Params"`
		} `json:"From"`
		To []struct { // ✅ Outer To is an array of objects
			Mailbox string `json:"Mailbox"`
			Domain  string `json:"Domain"`
			Params  string `json:"Params"`
		} `json:"To"`
		Content struct {
			Headers map[string][]string `json:"Headers"`
			Body    string              `json:"Body"`
			Size    int                 `json:"Size"`
			MIME    any                 `json:"MIME"` // You can change `any` to a more precise type if needed
		} `json:"Content"`
		Created string `json:"Created"`
		MIME    any    `json:"MIME"`
		Raw     struct {
			From string   `json:"From"`
			To   []string `json:"To"` // ✅ Raw.To is an array of strings
			Data string   `json:"Data"`
		} `json:"Raw"`
	} `json:"items"`
}

type SMTPTestContainer struct {
	ctr     *testcontainers.DockerContainer
	cfg     *config.SMTPConfig
	apiPort nat.Port
}

func StartSMTPContainer() (*SMTPTestContainer, error) {
	ctx := context.Background()
	ctr, err := testcontainers.Run(
		ctx,
		"mailhog/mailhog:v1.0.1",
		testcontainers.WithExposedPorts("1025/tcp", "8025/tcp"),
		testcontainers.WithWaitStrategy(
			wait.ForHTTP("/").WithPort("8025/tcp").
				WithStartupTimeout(10*time.Second),
		),
	)

	if err != nil {
		log.Printf("failed to start container: %s", err)
		return nil, err
	}

	smtpPort, _ := ctr.MappedPort(ctx, "1025")
	apiPort, _ := ctr.MappedPort(ctx, "8025")
	smtpHost, _ := ctr.Host(ctx)

	smtpCfg := new(config.SMTPConfig)
	smtpCfg.Host = smtpHost
	smtpCfg.Port = smtpPort.Int()
	smtpCfg.Username = ""
	smtpCfg.Password = ""
	smtpCfg.Sender = SMTP_SENDER

	return &SMTPTestContainer{
		ctr:     ctr,
		apiPort: apiPort,
		cfg:     smtpCfg,
	}, nil
}

func (p *SMTPTestContainer) GetSMTPConfig() *config.SMTPConfig {
	return p.cfg
}

func (p *SMTPTestContainer) Terminate() error {
	return testcontainers.TerminateContainer(p.ctr)
}

func (p *SMTPTestContainer) ClearMailhog() error {
	_, err := http.DefaultClient.Do(&http.Request{
		Method: "DELETE",
		URL:    mustParseURL(fmt.Sprintf("http://%s:%d/api/v1/messages", p.cfg.Host, p.apiPort.Int())),
	})
	return err
}

func mustParseURL(s string) *url.URL {
	u, err := url.Parse(s)
	if err != nil {
		panic(err)
	}
	return u
}

func (p *SMTPTestContainer) GetEmailsFromMailHog() (*MailHogResponse, error) {
	var resp *http.Response
	var err error

	// Wait for MailHog to process the email (in case of async delay)
	for range 5 {
		resp, err = http.Get(fmt.Sprintf("http://%s:%d/api/v2/messages", p.cfg.Host, p.apiPort.Int()))
		if err == nil && resp.StatusCode == http.StatusOK {
			break
		}
		time.Sleep(1 * time.Second)
	}

	if err != nil {
		return nil, fmt.Errorf("Failed to fetch messages from MailHog: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Failed to read MailHog response: %v", err)
	}
	// log.Printf("BODY: %s _________", body)
	var mailResp MailHogResponse
	err = json.Unmarshal(body, &mailResp)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse MailHog response: %v", err)
	}

	return &mailResp, nil
}
