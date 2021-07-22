package collectors_test

import (
	"testing"

	"github.com/cri-o/cri-o/server/metrics/collectors"
	. "github.com/cri-o/cri-o/test/framework"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

// TestCollectors runs the created specs
func TestCollectors(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Collectors")
}

// nolint: gochecknoglobals
var t *TestFramework

var _ = BeforeSuite(func() {
	t = NewTestFramework(NilFunc, NilFunc)
	t.Setup()
})

var _ = AfterSuite(func() {
	t.Teardown()
})

// The actual test suite
var _ = t.Describe("Collectors", func() {
	t.Describe("All", func() {
		It("should contain all available metrics", func() {
			// Given
			all := collectors.All()

			// When / Then
			for _, collector := range []collectors.Collector{
				collectors.Operations,
				collectors.OperationsLatencyTotal,
				collectors.OperationsLatency,
				collectors.OperationsErrors,
				collectors.ImagePullsByDigest,
				collectors.ImagePullsByName,
				collectors.ImagePullsByNameSkipped,
				collectors.ImagePullsFailures,
				collectors.ImagePullsSuccesses,
				collectors.ImagePullsLayerSize,
				collectors.ImageLayerReuse,
				collectors.ContainersOOMTotal,
				collectors.ContainersOOM,
				collectors.ProcessesDefunct,
			} {
				Expect(all.Contains(collector)).To(BeTrue())
			}

			Expect(all).To(HaveLen(14))
		})
	})

	t.Describe("Stripped", func() {
		It("should remove all prefixes", func() {
			// Given
			const s = "image_pulls_by_digest"

			// When / Then
			Expect(collectors.Collector("container_runtime_crio_" + s).
				Stripped().String()).To(Equal(s))
			Expect(collectors.Collector("crio_" + s).
				Stripped().String()).To(Equal(s))
			Expect(collectors.Collector(s).
				Stripped().String()).To(Equal(s))
		})
	})

	t.Describe("FromSlice", func() {
		It("should convert from slice", func() {
			// Given
			sut := []string{
				"test",
				"crio_sample",
				"container_runtime_crio_example",
			}

			// When
			res := collectors.FromSlice(sut)

			// Then
			Expect(res).To(HaveLen(3))
			Expect(res.Contains("test")).To(BeTrue())
			Expect(res.Contains("sample")).To(BeTrue())
			Expect(res.Contains("crio_sample")).To(BeTrue())
			Expect(res.Contains("crio_example")).To(BeTrue())
			Expect(res.Contains("container_runtime_crio_example")).To(BeTrue())
		})
	})

	t.Describe("ToSlice", func() {
		It("should convert to slice", func() {
			// Given
			sut := collectors.Collectors{
				"test",
				"crio_sample",
				"container_runtime_crio_example",
			}

			// When
			res := sut.ToSlice()

			// Then
			Expect(res).To(HaveLen(3))
			Expect(res[0]).To(Equal("test"))
			Expect(res[1]).To(Equal("sample"))
			Expect(res[2]).To(Equal("example"))
		})
	})
})
