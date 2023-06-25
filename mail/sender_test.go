package mail

import (
	"github.com/patchbrain/simple-bank/util"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSendEmail(t *testing.T) {
	cfg, err := util.LoadConfig("..")
	require.NoError(t, err)
	sender := NewWangYiEmailSender(cfg.FromEmailAddress, cfg.FromEmailPassword, "simple-bank")
	err = sender.SendEmail("test email",
		[]string{cfg.FromEmailAddress},
		`<h1></h1>`,
		nil,
		nil,
		[]string{"../main.go"})
	require.NoError(t, err)
}
