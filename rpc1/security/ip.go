// 拷贝自 grpc@v1.50.1/credentials/local/local.go

// 提示：该软件包是实验性的，可能在以后的版本中更改或删除。
// Notice: This package is EXPERIMENTAL and may be changed or removed in a later release.

package security

import (
	"context"
	"fmt"
	"net"
	"strings"

	"google.golang.org/grpc/credentials"
)

// info contains the auth information for a local connection.
// It implements the AuthInfo interface.
type info struct {
	credentials.CommonAuthInfo
}

// AuthType returns the type of info as a string.
func (info) AuthType() string {
	return "local"
}

// 实现功能：gRpc ip 白名单

// 在建立连接时候检查
// 客户端认证失败错误提示：
// Unavailable:connection closed before server preface received

type ipWhiteListTC struct {
	info credentials.ProtocolInfo
}

func (c *ipWhiteListTC) Info() credentials.ProtocolInfo {
	return c.info
}

// getSecurityLevel returns the security level for a local connection.
// It returns an error if a connection is not local.
func getSecurityLevel(network, addr string) (credentials.SecurityLevel, error) {
	switch {
	// Local TCP connection
	case strings.HasPrefix(addr, "127."), strings.HasPrefix(addr, "[::1]:"):
		return credentials.NoSecurity, nil

	// TODO: impl

	// // UDS connection
	// case network == "unix":
	// 	return credentials.PrivacyAndIntegrity, nil
	// Not a local connection and should fail
	default:
		return credentials.InvalidSecurityLevel, fmt.Errorf("local credentials rejected connection to non-local address %q", addr)
	}
}

func (*ipWhiteListTC) ClientHandshake(ctx context.Context, authority string, conn net.Conn) (net.Conn, credentials.AuthInfo, error) {
	secLevel, err := getSecurityLevel(conn.RemoteAddr().Network(), conn.RemoteAddr().String())
	if err != nil {
		return nil, nil, err
	}
	return conn, info{credentials.CommonAuthInfo{SecurityLevel: secLevel}}, nil
}

func (*ipWhiteListTC) ServerHandshake(conn net.Conn) (net.Conn, credentials.AuthInfo, error) {
	secLevel, err := getSecurityLevel(conn.RemoteAddr().Network(), conn.RemoteAddr().String())
	if err != nil {
		return nil, nil, err
	}
	return conn, info{credentials.CommonAuthInfo{SecurityLevel: secLevel}}, nil
}

// NewCredentials returns a local credential implementing credentials.TransportCredentials.
func NewCredentials() credentials.TransportCredentials {
	return &ipWhiteListTC{
		info: credentials.ProtocolInfo{
			SecurityProtocol: "local",
		},
	}
}

// Clone makes a copy of Local credentials.
func (c *ipWhiteListTC) Clone() credentials.TransportCredentials {
	return &ipWhiteListTC{info: c.info}
}

// OverrideServerName overrides the server name used to verify the hostname on the returned certificates from the server.
// Since this feature is specific to TLS (SNI + hostname verification check), it does not take any effet for local credentials.
func (c *ipWhiteListTC) OverrideServerName(serverNameOverride string) error {
	c.info.ServerName = serverNameOverride
	return nil
}
