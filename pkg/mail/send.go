package mail

import (
	"bytes"
	"fmt"
	"net/smtp"
	"os"
	"strings"
	"text/template"
)

type Config struct {
	SMTPAddr   string // "localhost:25"
	SMTPUser   string // optional
	SMTPPass   string // optional
	FromHeader string // "My App <noreply@example.com>"
}

type EmailTemplate struct {
	TemplatePath string
	Variables    map[string]string
	To           string
}

func Send(cfg Config, et EmailTemplate) error {
	// Read template file
	tplContent, err := os.ReadFile(et.TemplatePath)
	if err != nil {
		return fmt.Errorf("failed to read template: %w", err)
	}

	// Create template with custom delimiters
	tpl := template.Must(template.New("email").
		Delims("{%", "%}").
		Parse(string(tplContent)))

	// Add required From header to variables
	et.Variables["FROM_HEADER"] = cfg.FromHeader

	// Execute template with variables
	var buf bytes.Buffer
	if err := tpl.Execute(&buf, et.Variables); err != nil {
		return fmt.Errorf("template execution failed: %w", err)
	}
	// Connect to SMTP server
	auth := smtp.PlainAuth("", cfg.SMTPUser, cfg.SMTPPass, strings.Split(cfg.SMTPAddr, ":")[0])
	msg := buf.Bytes()
	from := "noreply@example.com"
	return smtp.SendMail(
		cfg.SMTPAddr,
		auth,
		from,
		[]string{et.To},
		msg,
	)
}
