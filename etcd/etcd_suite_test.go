package etcd_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/coreos/etcd/clientv3"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestEtcd(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Etcd Suite")
}

var (
	keyprefix           = "/etcdtest"
	endpoint   []string = []string{"http://127.0.0.1:2379"}
	etcdCli    *clientv3.Client
	kvCli      clientv3.KV
	watcherCli clientv3.Watcher
	leaseCli   clientv3.Lease
	clusterCli clientv3.Cluster

	err error
)

var _ = BeforeSuite(func() {
	etcdCli, err = clientv3.New(clientv3.Config{
		Endpoints: endpoint,
	})
	Expect(err).NotTo(HaveOccurred())

	kvCli = clientv3.NewKV(etcdCli)
	watcherCli = clientv3.NewWatcher(etcdCli)
	leaseCli = clientv3.NewLease(etcdCli)
	clusterCli = clientv3.NewCluster(etcdCli)
})

var _ = AfterSuite(func() {
	ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
	defer cancel()
	response, err := etcdCli.Delete(ctx, keyprefix, clientv3.WithPrefix())
	Expect(err).NotTo(HaveOccurred())
	fmt.Printf("deleted: %d\n", response.Deleted)
	etcdCli.Close()
})
