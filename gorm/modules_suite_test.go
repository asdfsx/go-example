package gormexample_test

import (
	"fmt"
	"testing"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	gorm_example "../gorm"
)

func TestModules(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Modules Suite")
}

var (
	db       *gorm.DB
	err      error
	dbaddr   = "127.0.0.1:3306"
	dbname   = "steel"
	username = "steel"
	password = "steel"
)

var _ = BeforeSuite(func() {
	mysqladdr := fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=True", username, password, dbaddr, dbname)
	db, err = gorm.Open("mysql", mysqladdr)
	Expect(err).NotTo(HaveOccurred())
	db.LogMode(true)
	db.AutoMigrate(&gorm_example.User{})
})

var _ = AfterSuite(func() {
	db.DropTableIfExists(&gorm_example.User{})
	if db != nil {
		err = db.Close()
		Expect(err).NotTo(HaveOccurred())
	}
})
