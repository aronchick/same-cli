package utils_test

import (
	"io/ioutil"

	log "github.com/sirupsen/logrus"

	"github.com/azure-octo/same-cli/pkg/utils"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func badURL(thisUrl string) {
	isRemote, err := utils.IsRemoteFilePath(thisUrl)
	Expect(isRemote).Should(Equal(false))
	Expect(err).Should(Not(BeNil()))
}

var _ = Describe("IsRemoteFilePath", func() {

	BeforeSuite(func() {
		log.SetOutput(ioutil.Discard)
	})

	Context("can identify", func() {
		It("a GH url with no org or repo", func() {
			Expect(utils.IsRemoteFilePath("http://github.com")).Should(Equal(true))
		})
		It("a GH url with org but no repo", func() {
			Expect(utils.IsRemoteFilePath("http://github.com/contoso")).Should(Equal(true))
		})
		It("a GH url with org and repo", func() {
			Expect(utils.IsRemoteFilePath("http://github.com/contoso/sameple-repo")).Should(Equal(true))
		})
		It("a GH url with org and repo and file", func() {
			Expect(utils.IsRemoteFilePath("http://github.com/contoso/sameple-repo/same.yaml")).Should(Equal(true))
		})
		It("a URL should start with a scheme (e.g. http://, https:// or git://", func() {
			badURL("github.com/contoso/sameple-repo/same.yaml")
		})
		It("a badly formed URL", func() {
			badURL("github/contoso/sample-repo/same.yaml")
		})
		It("a local relative file", func() {
			badURL("../abc.txt")
		})
		It("a local absolute file", func() {
			badURL("/ab/c.txt")
		})
		It("a file with no path", func() {
			badURL("c.txt")
		})
	})

})

// var _ = Describe("SplitYAML", func() {

// 	tests := []struct {
// 		name     string
// 		yaml     []byte
// 		expected [][]byte
// 	}{
// 		{
// 			name:     "simple",
// 			yaml:     []byte("a: b\n---\nc: d"),
// 			expected: [][]byte{[]byte("a: b\n"), []byte("c: d\n")},
// 		},
// 	}

// 	for _, test := range tests {
// 		t.Run(test.name, func(t *testing.T) {
// 			resources, err := SplitYAML(test.yaml)
// 			if err != nil {
// 				t.Fatalf("Unexpected error: %v", err)
// 			}
// 			for idx := range resources {
// 				if string(resources[idx]) != string(test.expected[idx]) {
// 					t.Fatalf("Resource in place %v. Got '%s', Want '%s'.", idx, resources[idx], test.expected[idx])
// 				}
// 			}
// 		})
// 	}
// })
