package mail

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEmailSending(t *testing.T) {
	// // Setup local SMTP server
	// smtpServer := smtp.NewMockServer(
	// 	smtp.WithHostname("localhost"),
	// 	smtp.WithPort(1025), // Use non-privileged port
	// )

	// go func() {
	// 	if err := smtpServer.Start(); err != nil {
	// 		t.Fatalf("Failed to start SMTP server: %v", err)
	// 	}
	// }()
	// defer smtpServer.Stop()

	// Wait for server to start
	// time.Sleep(100 * time.Millisecond)

	// Create test template
	// tmpFile, err := os.CreateTemp("", "test-*.eml")
	// require.NoError(t, err)
	// defer os.Remove(tmpFile.Name())

	// 	_, err = tmpFile.WriteString(`From: {%FROMHEADER%}
	// To: {%TO%}
	// Subject: Password Reset
	// Content-Type: text/plain

	// Hello {%USERNAME%},
	// Your OTP is {%OTP%}`)
	// 	require.NoError(t, err)
	// 	tmpFile.Close()

	// Configure mailer
	cfg := Config{
		SMTPAddr:   "localhost:1025", // use mail-hog to test locally
		FromHeader: "Test Sender <test@example.com>",
	}

	// Test cases
	t.Run("successful send", func(t *testing.T) {
		et := EmailTemplate{
			TemplatePath: "../../reset.eml", //tmpFile.Name()
			To:           "user@example.com",
			Variables: map[string]string{
				"FROMHEADER": cfg.FromHeader,
				"USERNAME":   "John Doe",
				"OTP":        "123456",
			},
		}

		err := Send(cfg, et)
		require.NoError(t, err)

		// // Verify email in SMTP server
		// messages := smtpServer.Messages()
		// require.Len(t, messages, 1)

		// msg := messages[0].MsgRequest()
		// require.Contains(t, msg, "Hello John Doe")
		// require.Contains(t, msg, "123456")
		// require.Contains(t, msg, "To: user@example.com")
	})
}
