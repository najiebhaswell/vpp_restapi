package api

import (
    api "go.fd.io/govpp/api"
    "go.fd.io/govpp/adapter/socketclient"
    "go.fd.io/govpp/core"
)

type VPPClient struct {
    conn *core.Connection
}

func NewVPPClient(socketPath string) (*VPPClient, error) {
    adapter := socketclient.NewVppClient(socketPath)
    conn, err := core.Connect(adapter)
    if err != nil {
        return nil, err
    }
    return &VPPClient{conn: conn}, nil
}

func (c *VPPClient) Close() {
    if c.conn != nil {
        c.conn.Disconnect()
    }
}

func (c *VPPClient) NewAPIChannel() (api.Channel, error) {
    return c.conn.NewAPIChannel()
}
