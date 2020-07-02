package utils_test

import (
	"github.com/google/go-github/github"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/secberus-oss/projector/utils"
	"github.com/spf13/viper"
)

var _ = Describe("Utils", func() {
	Describe("All Org Hooks", func() {
		var (
			hook    *github.Hook
			hooks   []*github.Hook
			hookURL string = "http://www.test.com"
			hookID  int64  = 1234567
		)

		BeforeEach(func() {
			hookConfig := map[string]interface{}{
				"url":          hookURL,
				"content_type": "json",
			}
			hook = &github.Hook{
				ID:     &hookID,
				Config: hookConfig,
				Events: []string{"pull_request"},
			}
			hooks = append(hooks, hook)
		})
		Context("A hook that contains PRJ_HOOK_URL in its URL", func() {
			It("should not be recreated", func() {
				viper.Set("hook_url", "http://www.test.com")
				Expect(utils.NewGH().HookExists(hooks)).To(Equal(true))
			})
		})
	})
})
