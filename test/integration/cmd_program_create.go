package integration_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("same program", func() {

	var rootCmd = cmd.RootCmd
	BeforeSuite(func() {
		log.SetOutput(ioutil.Discard)
		ts, err := cmdtest.Read("testdata/integration")
	})

	BeforeEach(func() {
	})

	Context("create", func() {

		It("Should run without arguments", func() {
			execute_and_read(*rootCmd, []string{}, "same [command]")
		})

})
