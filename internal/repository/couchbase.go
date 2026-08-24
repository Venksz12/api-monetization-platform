package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/couchbase/gocb/v2"
)

type Couchbase struct {
	Cluster *gocb.Cluster
	Bucket  *gocb.Bucket
	Scope   *gocb.Scope
}

func Connect(ctx context.Context, connstr, username, password, bucket, scope string) (*Couchbase, error) {
	cluster, err := gocb.Connect(connstr, gocb.ClusterOptions{
		Username: username, Password: password,
	})
	if err != nil {
		return nil, err
	}
	if err := cluster.WaitUntilReady(15*time.Second, nil); err != nil {
		cluster.Close(nil)
		return nil, err
	}
	b := cluster.Bucket(bucket)
	if err := b.WaitUntilReady(15*time.Second, nil); err != nil {
		cluster.Close(nil)
		return nil, err
	}
	return &Couchbase{Cluster: cluster, Bucket: b, Scope: b.Scope(scope)}, nil
}

func (c *Couchbase) Close() {
	c.Cluster.Close(nil)
}

func (c *Couchbase) Keyspace(collection string) *gocb.Collection {
	return c.Scope.Collection(collection)
}

func (c *Couchbase) Ping(ctx context.Context) error {
	return c.Cluster.Ping(&gocb.PingOptions{Context: ctx})
}

func IsNotFound(err error) bool {
	return err == gocb.ErrDocumentNotFound
}

func IsCasMismatch(err error) bool {
	return err == gocb.ErrCasMismatch
}

func ScopeExists(ctx context.Context, c *Couchbase, collection string) error {
	if c.Scope == nil {
		return fmt.Errorf("scope unavailable")
	}
	_, err := c.Scope.Collection(collection).Get("health-check", &gocb.GetOptions{Context: ctx})
	return err
}
