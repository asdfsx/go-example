package minio_test

import (
	"log"
	"testing"

	"github.com/minio/minio-go"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestMinio(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Minio Suite")
}

var (
	endpoint        = "127.0.0.1:9000"
	accessKeyID     = "minioadmin"
	secretAccessKey = "minioadmin"
	minioClient     *minio.Client
	useSSL          = false
	err             error
	bucketName      = "testminio"
	regionName      = "testminio"
	uploadfile      = "minio_test.go"
)

var _ = BeforeSuite(func() {
	endpoint := "play.min.io"
	accessKeyID := "Q3AM3UQ867SPQQA43P2F"
	secretAccessKey := "zuf+tfteSlswRu7BJ86wekitnifILbZam1KYY3TG"

	// Initialize minio client object.
	minioClient, err = minio.New(endpoint, accessKeyID, secretAccessKey, useSSL)
	if err != nil {
		log.Fatalln(err)
	}
})

var _ = AfterSuite(func() {
	isExists := false
	isExists, err = minioClient.BucketExists(bucketName)
	Expect(err).NotTo(HaveOccurred())

	if isExists {
		err = minioClient.RemoveObject(bucketName, uploadfile)
		Expect(err).NotTo(HaveOccurred())
		err = minioClient.RemoveBucket(bucketName)
		Expect(err).NotTo(HaveOccurred())
	}
})
