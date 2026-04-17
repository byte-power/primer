package primer

import "net/url"

// ===== Client Session API =====

// CreateClientSession 创建客户端会话
// POST /client-session
func (c *Client) CreateClientSession(req *CreateClientSessionRequest) (*ClientSessionWithTokenResponse, *Error) {
	var resp ClientSessionWithTokenResponse
	if err := c.doPost("/client-session", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// UpdateClientSession 更新客户端会话
// PATCH /client-session
func (c *Client) UpdateClientSession(req *UpdateClientSessionRequest) (*ClientSessionWithTokenResponse, *Error) {
	var resp ClientSessionWithTokenResponse
	if err := c.doPatch("/client-session", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetClientSession 检索客户端会话
// GET /client-session?clientToken=xxx
func (c *Client) GetClientSession(clientToken string) (*ClientSessionResponse, *Error) {
	params := url.Values{}
	params.Set("clientToken", clientToken)

	var resp ClientSessionResponse
	if err := c.doGet("/client-session", params, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
