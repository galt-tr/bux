package bux

import (
	"github.com/tonicpow/go-paymail"
	"github.com/tonicpow/go-paymail/server"
)

// ModifyPaymailConfig will set the new values if found
func (c *Client) ModifyPaymailConfig(config *server.Configuration,
	defaultFromPaymail, defaultNote string) {
	if c.options.paymail != nil {
		if config != nil {
			c.options.paymail.serverConfig.Configuration = config
		}
		if len(defaultFromPaymail) > 0 {
			c.options.paymail.serverConfig.DefaultFromPaymail = defaultFromPaymail
		}
		if len(defaultNote) > 0 {
			c.options.paymail.serverConfig.DefaultNote = defaultNote
		}
	}
}

// PaymailClient will return the Paymail if it exists
func (c *Client) PaymailClient() paymail.ClientInterface {
	if c.options.paymail != nil && c.options.paymail.client != nil {
		return c.options.paymail.Client()
	}
	return nil
}

// PaymailServerConfig will return the Paymail server config if it exists
func (c *Client) PaymailServerConfig() *paymailServerOptions {
	if c.options.paymail != nil && c.options.paymail.serverConfig != nil {
		return c.options.paymail.serverConfig
	}
	return nil
}

// Client will return the paymail client from the options struct
func (p *paymailOptions) Client() paymail.ClientInterface {
	return p.client
}

// FromSender will return either the configuration value or the application default
func (p *paymailOptions) FromSender() string {
	if len(p.serverConfig.DefaultFromPaymail) > 0 {
		return p.serverConfig.DefaultFromPaymail
	}
	return defaultSenderPaymail
}

// Note will return either the configuration value or the application default
func (p *paymailOptions) Note() string {
	if len(p.serverConfig.DefaultNote) > 0 {
		return p.serverConfig.DefaultNote
	}
	return defaultAddressResolutionPurpose
}

// ServerConfig will return the Paymail Server configuration from the options struct
func (p *paymailOptions) ServerConfig() *paymailServerOptions {
	return p.serverConfig
}
