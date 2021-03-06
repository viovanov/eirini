package k8s_test

import (
	"code.cloudfoundry.org/eirini"
	"code.cloudfoundry.org/eirini/models/cf"
	"code.cloudfoundry.org/eirini/opi"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"

	. "code.cloudfoundry.org/eirini/k8s"
)

var _ = Describe("Service", func() {

	var (
		fakeClient     kubernetes.Interface
		serviceManager ServiceManager
	)

	const (
		namespace = "midgard"
	)

	BeforeEach(func() {
		fakeClient = fake.NewSimpleClientset()
	})

	JustBeforeEach(func() {
		serviceManager = NewServiceManager(fakeClient, namespace)
	})

	Context("When exposing an existing LRP", func() {

		var (
			lrp *opi.LRP
			err error
		)

		BeforeEach(func() {
			lrp = createLRP("baldur", "54321.0")
		})

		Context("When creating a usual service", func() {

			JustBeforeEach(func() {
				err = serviceManager.Create(lrp)
			})

			It("should not fail", func() {
				Expect(err).ToNot(HaveOccurred())
			})

			It("should create a service", func() {
				serviceName := eirini.GetInternalServiceName("baldur")
				service, err := fakeClient.CoreV1().Services(namespace).Get(serviceName, meta.GetOptions{})
				Expect(err).ToNot(HaveOccurred())
				Expect(service).To(Equal(toService(lrp, namespace)))
			})

			Context("When recreating a existing service", func() {

				BeforeEach(func() {
					lrp = createLRP("baldur", "54321.0")
				})

				JustBeforeEach(func() {
					err = serviceManager.Create(lrp)
				})

				It("should return an error", func() {
					Expect(err).To(HaveOccurred())
				})
			})
		})

		Context("When creating a headless service", func() {

			JustBeforeEach(func() {
				err = serviceManager.CreateHeadless(lrp)
			})

			It("should not fail", func() {
				Expect(err).ToNot(HaveOccurred())
			})

			It("should create a service", func() {
				serviceName := eirini.GetInternalHeadlessServiceName("baldur")
				service, err := fakeClient.CoreV1().Services(namespace).Get(serviceName, meta.GetOptions{})
				Expect(err).ToNot(HaveOccurred())
				Expect(service).To(Equal(toHeadlessService(lrp, namespace)))
			})

			Context("When recreating a existing service", func() {

				BeforeEach(func() {
					lrp = createLRP("baldur", "54321.0")
				})

				JustBeforeEach(func() {
					err = serviceManager.CreateHeadless(lrp)
				})

				It("should return an error", func() {
					Expect(err).To(HaveOccurred())
				})
			})
		})
	})

	Context("When deleting", func() {

		var service *v1.Service

		assertServiceIsDeleted := func(err error) {
			Expect(err).ToNot(HaveOccurred())

			services, err := fakeClient.CoreV1().Services(namespace).List(meta.ListOptions{})
			Expect(err).ToNot(HaveOccurred())
			Expect(services.Items).To(BeEmpty())
		}

		JustBeforeEach(func() {
			_, err := fakeClient.CoreV1().Services(namespace).Create(service)
			Expect(err).ToNot(HaveOccurred())
		})

		Context("a regular service", func() {

			var err error

			JustBeforeEach(func() {
				err = serviceManager.Delete("odin")
			})

			BeforeEach(func() {
				lrp := createLRP("odin", "1234.5")
				service = toService(lrp, namespace)
			})

			It("deletes the service", func() {
				assertServiceIsDeleted(err)
			})

			Context("when the service does not exist", func() {

				JustBeforeEach(func() {
					err = serviceManager.Delete("tyr")
				})

				It("returns an error", func() {
					Expect(err).To(HaveOccurred())
				})

			})
		})

		Context("a headless service", func() {

			var err error

			BeforeEach(func() {
				lrp := createLRP("odin", "1234.5")
				service = toHeadlessService(lrp, namespace)
			})

			JustBeforeEach(func() {
				err = serviceManager.DeleteHeadless("odin")
			})

			It("deletes the service", func() {
				assertServiceIsDeleted(err)
			})

			Context("when the service does not exist", func() {

				JustBeforeEach(func() {
					err = serviceManager.DeleteHeadless("tyr")
				})

				It("returns an error", func() {
					Expect(err).To(HaveOccurred())
				})

			})
		})
	})

})

func getServicesNames(services *v1.ServiceList) []string {
	serviceNames := []string{}
	for _, s := range services.Items {
		serviceNames = append(serviceNames, s.Name)
	}
	return serviceNames
}

func toService(lrp *opi.LRP, namespace string) *v1.Service {
	service := &v1.Service{
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{
					Name: "service",
					Port: 8080,
				},
			},
			Selector: map[string]string{
				"name": lrp.Name,
			},
		},
	}

	service.Name = eirini.GetInternalServiceName(lrp.Name)
	service.Namespace = namespace
	service.Labels = map[string]string{
		"name": lrp.Name,
	}

	service.Annotations = map[string]string{
		"routes": lrp.Metadata[cf.VcapAppUris],
	}

	return service
}

func toHeadlessService(lrp *opi.LRP, namespace string) *v1.Service {
	service := &v1.Service{
		Spec: v1.ServiceSpec{
			ClusterIP: "None",
			Ports: []v1.ServicePort{
				{
					Name: "service",
					Port: 8080,
				},
			},
			Selector: map[string]string{
				"name": lrp.Name,
			},
		},
	}

	service.Name = eirini.GetInternalHeadlessServiceName(lrp.Name)
	service.Namespace = namespace
	service.Labels = map[string]string{
		"name": lrp.Name,
	}

	return service
}
