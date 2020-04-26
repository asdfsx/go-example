package gormexample_test

import (
	"fmt"
	"strconv"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	gorm_example "../gorm"
)

var _ = Describe("gorm test", func() {
	Context("module user test", func() {
		It("create 1st record", func() {
			u := gorm_example.User{Name: "0"}
			Expect(db.NewRecord(u)).To(BeTrue())
			err = db.Create(&u).Error
			Expect(err.Error()).To(ContainSubstring("Column 'member_number' cannot be null"))

			u.MemberNumber = &u.Name
			u.Email = u.Name
			err = db.Create(&u).Error
			Expect(err).NotTo(HaveOccurred())
		})
		It("create record, from 1 to 9", func() {
			for i := 1; i < 10; i++ {
				name := strconv.Itoa(i)
				err = db.Create(&gorm_example.User{Name: name, MemberNumber: &name, Email: name}).Error
				fmt.Println(err, name)
				Expect(err).NotTo(HaveOccurred())
			}
		})
		It("query", func() {
			u := gorm_example.User{}
			err = db.First(&u).Error
			Expect(err).NotTo(HaveOccurred())
			Expect(u.Name).To(Equal("0"))

			u = gorm_example.User{}
			db.Last(&u)
			Expect(u.Name).To(Equal("9"))

			var userSlice []gorm_example.User
			err = db.Find(&userSlice).Error
			Expect(err).NotTo(HaveOccurred())
			Expect(len(userSlice)).To(Equal(10))

			err = db.Where("name>?", "4").Find(&userSlice).Error
			Expect(err).NotTo(HaveOccurred())
			Expect(len(userSlice)).To(Equal(5))

			err = db.Where(&gorm_example.User{Name: "7"}).Find(&userSlice).Error
			Expect(err).NotTo(HaveOccurred())
			Expect(userSlice[0].Name).To(Equal("7"))
		})
		It("update", func() {
			var user gorm_example.User
			err = db.Where("name=?", 8).Find(&user).Error
			Expect(err).NotTo(HaveOccurred())
			Expect(user.Name).To(Equal("8"))
			Expect(user.Address).To(Equal(""))

			user.Address = "test"
			err = db.Save(&user).Error
			Expect(err).NotTo(HaveOccurred())
			err = db.Where("name=?", "8").Find(&user).Error
			Expect(err).NotTo(HaveOccurred())
			Expect(user.Name).To(Equal("8"))
			Expect(user.Address).To(Equal("test"))
			Expect(user.Role).To(Equal(""))

			err = db.Model(&user).Where("name = ?", "8").Update("role", "hello").Error
			Expect(err).NotTo(HaveOccurred())
			err = db.Where("name=?", "8").Find(&user).Error
			Expect(user.Name).To(Equal("8"))
			Expect(user.Address).To(Equal("test"))
			Expect(user.Role).To(Equal("hello"))
		})
		It("delete", func() {
			var user gorm_example.User
			var userSlice []gorm_example.User

			err = db.Where("name=?", "9").Delete(&user).Error
			Expect(err).NotTo(HaveOccurred())
			err = db.Unscoped().Where("name=?", "9").Find(&user).Error
			Expect(err).NotTo(HaveOccurred())
			Expect(user.Name).To(Equal("9"))
			Expect(user.DeletedAt).NotTo(BeZero())

			err = db.Unscoped().Where("name=?", "9").Delete(&user).Error
			Expect(err).NotTo(HaveOccurred())
			err = db.Unscoped().Where("name=?", "9").Find(&user).Error
			Expect(err.Error()).To(Equal("record not found"))

			err = db.Unscoped().Delete(&gorm_example.User{}).Error
			Expect(err).NotTo(HaveOccurred())
			err = db.Find(&userSlice).Error
			Expect(err).NotTo(HaveOccurred())
			Expect(len(userSlice)).To(Equal(0))
		})
	})
})
