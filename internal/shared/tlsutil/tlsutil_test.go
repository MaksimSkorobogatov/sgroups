package tlsutil

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"net"
	"net/http"
	"testing"

	"github.com/H-BF/corlib/pkg/option"
	"github.com/stretchr/testify/require"
)

const (
	folder           = "testdata/"
	caCrt            = folder + "ca.crt"
	serverCrt        = folder + "server.crt"
	serverKey        = folder + "server.key"
	serverCrtExpired = folder + "server_expired.crt"
	serverKeyExpired = folder + "server_expired.key"
	clientCrt        = folder + "client.crt"
	clientKey        = folder + "client.key"
	clientCrtExpired = folder + "client_expired.crt"
	clientKeyExpired = folder + "client_expired.key"

	serverName = "localhost"
)

func Test_TlsConfigBuilders(t *testing.T) {
	testcases := [...]struct {
		name         string
		newClientTLS func(ctx context.Context) (*tls.Config, error)
		newServerTLS func(ctx context.Context) (*tls.Config, error)
		wantErr      bool
	}{
		{
			name: "skip server verify",
			newClientTLS: func(ctx context.Context) (*tls.Config, error) {
				return ClientTLS().BuildConfig(ctx)
			},
			newServerTLS: func(ctx context.Context) (*tls.Config, error) {
				return ServerTLS(ServerCertsFromFiles(serverCrtExpired, serverKeyExpired)).BuildConfig(ctx)
			},
		},
		{
			name: "server verify",
			newClientTLS: func(ctx context.Context) (*tls.Config, error) {
				return ClientTLS().VerifyServer(ServerAuthorityFromFiles{
					CAPaths: []string{caCrt},
					SAN:     option.MustNewOption(serverName),
				}).BuildConfig(ctx)
			},
			newServerTLS: func(ctx context.Context) (*tls.Config, error) {
				return ServerTLS(ServerCertsFromFiles(serverCrt, serverKey)).BuildConfig(ctx)
			},
		},
		{
			name: "server expired",
			newClientTLS: func(ctx context.Context) (*tls.Config, error) {
				return ClientTLS().VerifyServer(ServerAuthorityFromFiles{
					CAPaths: []string{caCrt},
					SAN:     option.MustNewOption(serverName),
				}).BuildConfig(ctx)
			},
			newServerTLS: func(ctx context.Context) (*tls.Config, error) {
				return ServerTLS(ServerCertsFromFiles(serverCrtExpired, serverKeyExpired)).BuildConfig(ctx)
			},
			wantErr: true,
		},
		{
			name: "mTLS",
			newClientTLS: func(ctx context.Context) (*tls.Config, error) {
				return ClientTLS().VerifyServer(ServerAuthorityFromFiles{
					CAPaths: []string{caCrt},
					SAN:     option.MustNewOption(serverName),
				}).VerifyClient(ClientCertsFromFiles{CertPath: clientCrt, KeyPath: clientKey}).
					BuildConfig(ctx)
			},
			newServerTLS: func(ctx context.Context) (*tls.Config, error) {
				return ServerTLS(ServerCertsFromFiles(serverCrt, serverKey)).
					VerifyClient(RequireAndVerifyCerts(ClientCAsFromFiles(caCrt))).
					BuildConfig(ctx)
			},
		},
		{
			name: "mTLS, but client has no certs",
			newClientTLS: func(ctx context.Context) (*tls.Config, error) {
				return ClientTLS().VerifyServer(ServerAuthorityFromFiles{
					CAPaths: []string{caCrt},
					SAN:     option.MustNewOption(serverName),
				}).BuildConfig(ctx)
			},
			newServerTLS: func(ctx context.Context) (*tls.Config, error) {
				return ServerTLS(ServerCertsFromFiles(serverCrt, serverKey)).
					VerifyClient(RequireAndVerifyCerts(ClientCAsFromFiles(caCrt))).
					BuildConfig(ctx)
			},
			wantErr: true,
		},
		{
			name: "mTLS, but client has expired certs",
			newClientTLS: func(ctx context.Context) (*tls.Config, error) {
				return ClientTLS().VerifyServer(ServerAuthorityFromFiles{
					CAPaths: []string{caCrt},
					SAN:     option.MustNewOption(serverName),
				}).VerifyClient(ClientCertsFromFiles{CertPath: clientCrtExpired, KeyPath: clientKeyExpired}).
					BuildConfig(ctx)
			},
			newServerTLS: func(ctx context.Context) (*tls.Config, error) {
				return ServerTLS(ServerCertsFromFiles(serverCrt, serverKey)).
					VerifyClient(RequireAndVerifyCerts(ClientCAsFromFiles(caCrt))).
					BuildConfig(ctx)
			},
			wantErr: true,
		},
		{
			name: "mTLS, server requires any cert, but client doesn't have certs",
			newClientTLS: func(ctx context.Context) (*tls.Config, error) {
				return ClientTLS().VerifyServer(ServerAuthorityFromFiles{
					CAPaths: []string{caCrt},
					SAN:     option.MustNewOption(serverName),
				}).BuildConfig(ctx)
			},
			newServerTLS: func(ctx context.Context) (*tls.Config, error) {
				return ServerTLS(ServerCertsFromFiles(serverCrt, serverKey)).
					VerifyClient(RequireAnyCert()).
					BuildConfig(ctx)
			},
			wantErr: true,
		},
		{
			name: "mTLS, but client has expired certs but server ignores it",
			newClientTLS: func(ctx context.Context) (*tls.Config, error) {
				return ClientTLS().VerifyServer(ServerAuthorityFromFiles{
					CAPaths: []string{caCrt},
					SAN:     option.MustNewOption(serverName),
				}).
					VerifyClient(ClientCertsFromFiles{CertPath: clientCrtExpired, KeyPath: clientKeyExpired}).
					BuildConfig(ctx)
			},
			newServerTLS: func(ctx context.Context) (*tls.Config, error) {
				return ServerTLS(ServerCertsFromFiles(serverCrt, serverKey)).
					VerifyClient(RequireAnyCert()).
					BuildConfig(ctx)
			},
			wantErr: false,
		},
	}

	ctx := context.TODO()
	if d, ok := t.Deadline(); ok {
		var cancel func()
		ctx, cancel = context.WithDeadline(ctx, d)
		defer cancel()
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			tlsClient, err := tt.newClientTLS(ctx)
			require.NoError(t, err)
			tlsServer, err := tt.newServerTLS(ctx)
			require.NoError(t, err)
			err = test(t, ctx, tlsClient, tlsServer, "localhost:0")
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func test(t *testing.T, ctx context.Context, tlsClient, tlsServer *tls.Config, addr string) error {
	type AddReq struct {
		A int `json:"a"`
		B int `json:"b"`
	}
	type AddResp struct {
		Result int `json:"result"`
	}

	mux := http.NewServeMux()
	mux.Handle("POST /add", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Log("server got req")
		var req AddReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		if err := json.NewEncoder(w).Encode(AddResp{Result: req.A + req.B}); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}
	}))

	server := &http.Server{
		Handler: mux,
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsClient,
		},
	}

	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	go func() {
		t.Log("start server")
		_ = server.Serve(tls.NewListener(l, tlsServer))
	}()
	defer server.Close()

	var reqBytes []byte
	if reqBytes, err = json.Marshal(AddReq{A: 10, B: 5}); err != nil {
		return err
	}

	var req *http.Request
	if req, err = http.NewRequestWithContext(ctx, http.MethodPost, "https://"+l.Addr().String()+"/add", bytes.NewBuffer(reqBytes)); err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var addResp AddResp
	if err := json.NewDecoder(resp.Body).Decode(&addResp); err != nil {
		return err
	}
	t.Log("client decoded resp")
	require.Equal(t, 15, addResp.Result)

	return nil
}
