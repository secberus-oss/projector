package main_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	projector "github.com/secberus-oss/projector"
)

var _ = Describe("Projector", func() {
	Describe("Healthcheck", func() {
		Context("Server works correctly", func() {
			It("should return a 200", func() {
				prj := projector.NewPRJ()
				Expect(prj.CheckHealth()).To(Equal(200))
			})
		})
	})
})
