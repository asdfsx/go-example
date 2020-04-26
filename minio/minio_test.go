package minio_test

import (
	"bytes"
	"fmt"
	"io"

	"github.com/minio/minio-go"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Minio", func() {
	Context("test minio", func() {
		It("bucket operation", func() {
			err := minioClient.MakeBucket(bucketName, regionName)
			Expect(err).NotTo(HaveOccurred())

			var result []minio.BucketInfo
			resultNames := []string{}
			result, err = minioClient.ListBuckets()
			Expect(err).NotTo(HaveOccurred())
			for i := range result {
				resultNames = append(resultNames, result[i].Name)
			}
			Expect(resultNames).To(ContainElement(bucketName))

			isExists := false
			isExists, err = minioClient.BucketExists(bucketName)
			Expect(err).NotTo(HaveOccurred())
			Expect(isExists).To(Equal(true))

			err = minioClient.RemoveBucket(bucketName)
			Expect(err).NotTo(HaveOccurred())

			resultNames = []string{}
			result, err = minioClient.ListBuckets()
			Expect(err).NotTo(HaveOccurred())
			for i := range result {
				resultNames = append(resultNames, result[i].Name)
			}
			Expect(resultNames).NotTo(ContainElement(bucketName))
		})
		It("test Object operation", func() {
			err := minioClient.MakeBucket(bucketName, regionName)
			Expect(err).NotTo(HaveOccurred())

			_, err = minioClient.FPutObject(bucketName, uploadfile, uploadfile, minio.PutObjectOptions{
				ContentType: "application/text",
			})
			Expect(err).NotTo(HaveOccurred())

			done := make(chan struct{})
			objectNames := []string{}
			objectCh := minioClient.ListObjectsV2(bucketName, "", false, done)
			for object := range objectCh {
				if object.Err != nil {
					fmt.Println(object.Err)
					return
				}
				objectNames = append(objectNames, object.Key)
			}
			Expect(objectNames).To(ContainElement(uploadfile))

			var info minio.ObjectInfo
			info, err = minioClient.StatObject(bucketName, uploadfile, minio.StatObjectOptions{})
			Expect(err).NotTo(HaveOccurred())
			Expect(info.Key).To(Equal(uploadfile))
			Expect(info.ContentType).To(Equal("application/text"))

			var object *minio.Object
			var buff bytes.Buffer
			object, err = minioClient.GetObject(bucketName, uploadfile, minio.GetObjectOptions{})
			Expect(err).NotTo(HaveOccurred())
			_, err = io.Copy(&buff, object)
			Expect(err).NotTo(HaveOccurred())
			fmt.Println(buff.String())
			Expect(buff.String()).To(ContainSubstring("test Object operation"))

			err = minioClient.RemoveObject(bucketName, uploadfile)
			Expect(err).NotTo(HaveOccurred())

			_, err = minioClient.StatObject(bucketName, uploadfile, minio.StatObjectOptions{})
			Expect(err).To(HaveOccurred())
			fmt.Println(err)
		})
	})
})
