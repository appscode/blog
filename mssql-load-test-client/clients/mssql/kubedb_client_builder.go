/*
Copyright AppsCode Inc. and Contributors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package mssql

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"

	dbapi "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"

	core "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const DefaultMSSQLDatabase = "master"

// KubeDBClientBuilder builds a SQL Server client from a KubeDB MSSQLServer CR.
// It reads the SA credentials from the referenced Kubernetes Secret and constructs
// the connection string, including TLS settings if configured in the CR.
type KubeDBClientBuilder struct {
	kc      client.Client
	db      *dbapi.MSSQLServer
	url     string
	podName string
	dbName  string
	ctx     context.Context
}

// NewKubeDBClientBuilder creates a new KubeDBClientBuilder for the given MSSQLServer CR.
func NewKubeDBClientBuilder(kc client.Client, db *dbapi.MSSQLServer) *KubeDBClientBuilder {
	return &KubeDBClientBuilder{
		kc: kc,
		db: db,
	}
}

// WithPod sets a specific pod to connect to (using the governing service DNS name).
func (o *KubeDBClientBuilder) WithPod(podName string) *KubeDBClientBuilder {
	o.podName = podName
	return o
}

// WithURL overrides the host URL to connect to.
func (o *KubeDBClientBuilder) WithURL(url string) *KubeDBClientBuilder {
	o.url = url
	return o
}

// WithDatabase sets the target database name (default: "master").
func (o *KubeDBClientBuilder) WithDatabase(dbName string) *KubeDBClientBuilder {
	o.dbName = dbName
	return o
}

// WithContext sets the context used for connection and credential fetching.
func (o *KubeDBClientBuilder) WithContext(ctx context.Context) *KubeDBClientBuilder {
	o.ctx = ctx
	return o
}

// GetMSSQLClient builds the SQL Server *Client (wraps *sql.DB) from the KubeDB CR.
func (o *KubeDBClientBuilder) GetMSSQLClient() (*Client, error) {
	if o.ctx == nil {
		o.ctx = context.Background()
	}

	connStr, err := o.getConnectionString()
	if err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlserver", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open SQL Server connection: %w", err)
	}

	if err := db.PingContext(o.ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to ping SQL Server: %w", err)
	}

	return &Client{db}, nil
}

// getURL constructs the pod DNS name for direct pod connections.
func (o *KubeDBClientBuilder) getURL() string {
	return fmt.Sprintf("%s.%s.%s.svc", o.podName, o.db.GoverningServiceName(), o.db.Namespace)
}

// getMSSQLAuthCredentials reads username and password from the MSSQLServer auth Secret.
func (o *KubeDBClientBuilder) getMSSQLAuthCredentials() (string, string, error) {
	db := o.db
	if db.Spec.AuthSecret == nil {
		return "", "", fmt.Errorf("no authSecret configured on MSSQLServer %s/%s", db.Namespace, db.Name)
	}

	secretName := db.GetAuthSecretName()

	var secret core.Secret
	if err := o.kc.Get(o.ctx, client.ObjectKey{Namespace: db.Namespace, Name: secretName}, &secret); err != nil {
		return "", "", fmt.Errorf("failed to get auth secret %s/%s: %w", db.Namespace, secretName, err)
	}

	user, ok := secret.Data[core.BasicAuthUsernameKey]
	if !ok {
		return "", "", fmt.Errorf("username key %q not found in secret %s/%s",
			core.BasicAuthUsernameKey, db.Namespace, secretName)
	}
	pass, ok := secret.Data[core.BasicAuthPasswordKey]
	if !ok {
		return "", "", fmt.Errorf("password key %q not found in secret %s/%s",
			core.BasicAuthPasswordKey, db.Namespace, secretName)
	}

	return string(user), string(pass), nil
}

// getConnectionString builds the go-mssqldb URL DSN for the SQL Server.
//
// TLS behaviour (mirrors how KubeDB deploys SQL Server):
//   - If Spec.TLS is nil               → encrypt=disable  (plain text, typical for dev/test)
//   - If Spec.TLS != nil               → encrypt=true      (server always uses TLS)
//   - If Spec.TLS.ClientTLS == true    → TrustServerCertificate=true  (cert auth enabled)
//   - If Spec.TLS.ClientTLS == false   → TrustServerCertificate=true  (server TLS but no client cert)
func (o *KubeDBClientBuilder) getConnectionString() (string, error) {
	user, pass, err := o.getMSSQLAuthCredentials()
	if err != nil {
		return "", err
	}

	if o.podName != "" {
		o.url = o.getURL()
	}

	dbName := o.dbName
	if dbName == "" {
		dbName = DefaultMSSQLDatabase
	}

	klog.Infof("Connecting to SQL Server %s/%s at %s (database: %s)", o.db.Namespace, o.db.Name, o.url, dbName)

	q := url.Values{}
	q.Set("database", dbName)
	q.Set("connection timeout", "30")
	q.Set("dial timeout", "30")

	if o.db.Spec.TLS != nil {
		// Server uses TLS; always encrypt, trust the server cert
		// (for mutual/client TLS the SQL Server operator issues certs automatically)
		q.Set("encrypt", "true")
		q.Set("TrustServerCertificate", "true")
	} else {
		q.Set("encrypt", "disable")
	}

	u := &url.URL{
		Scheme:   "sqlserver",
		User:     url.UserPassword(user, pass),
		Host:     o.url,
		RawQuery: q.Encode(),
	}

	return u.String(), nil
}

