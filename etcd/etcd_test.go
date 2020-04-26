package etcd_test

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/coreos/etcd/clientv3"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var revMap map[int64]string = make(map[int64]string)

var _ = Describe("Etcd", func() {
	Context("cluster api", func() {
		It("list", func() {
			ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
			defer cancel()
			response, err := clusterCli.MemberList(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(response.Members)).To(Equal(1))
		})
	})
	Context("key-value api", func() {
		It("put", func() {
			ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
			defer cancel()
			path := filepath.Join(keyprefix, "kvtest")
			response, err := kvCli.Put(ctx, path, "1")
			Expect(err).NotTo(HaveOccurred())
			fmt.Printf("put revision-----%d\n", response.Header.GetRevision())
		})
		It("get", func() {
			ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
			defer cancel()
			path := filepath.Join(keyprefix, "kvtest")
			response, err := kvCli.Get(ctx, path)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(response.Kvs[0].Value)).To(Equal("1"))
			fmt.Printf("get revision-----%d\n", response.Header.GetRevision())
		})
		It("put with option", func() {
			ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
			defer cancel()
			path := filepath.Join(keyprefix, "kvtest")
			response, err := kvCli.Put(ctx, path, "2", clientv3.WithPrevKV())
			Expect(err).NotTo(HaveOccurred())
			Expect(string(response.PrevKv.Value)).To(Equal("1"))
			fmt.Printf("puth with option revision-----%d\n", response.Header.GetRevision())
			revMap[response.Header.GetRevision()] = "2"

			response, err = kvCli.Put(ctx, path, "3", clientv3.WithPrevKV())
			Expect(err).NotTo(HaveOccurred())
			Expect(string(response.PrevKv.Value)).To(Equal("2"))
			fmt.Printf("puth with option revision-----%d\n", response.Header.GetRevision())
			revMap[response.Header.GetRevision()] = "3"
		})
		It("get with options", func() {
			ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
			defer cancel()
			path := filepath.Join(keyprefix, "kvtest")

			for rev, value := range revMap {
				response, err := kvCli.Get(ctx, path, clientv3.WithRev(rev))
				fmt.Println(response)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(response.Kvs[0].Value)).To(Equal(value))
			}
		})
	})
	Context("watch api", func() {
		It("watch prefix", func() {
			var result = []*clientv3.Event{}
			path := filepath.Join(keyprefix, "watchtest")
			rch := watcherCli.Watch(context.Background(), path, clientv3.WithPrefix())
			go func() {
				for wresp := range rch {
					for _, ev := range wresp.Events {
						result = append(result, ev)
						fmt.Printf("%s %q : %q\n", ev.Type, ev.Kv.Key, ev.Kv.Value)
					}
				}
			}()

			_, _ = kvCli.Put(context.TODO(), path, "1")
			_, _ = kvCli.Put(context.TODO(), path, "2")
			_, _ = kvCli.Put(context.TODO(), path, "3")

			_, _ = kvCli.Put(context.TODO(), path+"/1", "1")
			_, _ = kvCli.Put(context.TODO(), path+"/2", "2")
			_, _ = kvCli.Put(context.TODO(), path+"/3", "3")

			_, _ = kvCli.Delete(context.TODO(), path+"/1")
			_, _ = kvCli.Delete(context.TODO(), path+"/1")
			_, _ = kvCli.Delete(context.TODO(), path+"/1")

			Expect(len(result)).To(Equal(7))

			_, _ = kvCli.Delete(context.TODO(), path+"/2")
			_, _ = kvCli.Delete(context.TODO(), path+"/3")

			time.Sleep(1 * time.Second)

			Expect(len(result)).To(Equal(9))

			for i, v := range result {
				if i > 6 {
					Expect(v.Type.String()).To(Equal("DELETE"))
				}
			}
		})
	})
	Context("lease api", func() {
		It("grant", func() {
			var result = []*clientv3.Event{}
			path := filepath.Join(keyprefix, "leasetest")
			rch := watcherCli.Watch(context.Background(), path, clientv3.WithPrefix())
			go func() {
				for wresp := range rch {
					for _, ev := range wresp.Events {
						result = append(result, ev)
						fmt.Printf("%s %q : %q\n", ev.Type, ev.Kv.Key, ev.Kv.Value)
					}
				}
			}()

			response, err := leaseCli.Grant(context.TODO(), 5)
			Expect(err).NotTo(HaveOccurred())
			_, _ = kvCli.Put(context.TODO(), path, "1", clientv3.WithLease(response.ID))

			time.Sleep(6 * time.Second)

			Expect(len(result)).To(Equal(2))

			for _, ev := range result {
				fmt.Printf("%s %q : %q\n", ev.Type, ev.Kv.Key, ev.Kv.Value)
			}
		})
	})
})
