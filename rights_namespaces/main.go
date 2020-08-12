package main

import (
	"flag"
	"fmt"
	RbacV1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"os"
	"path/filepath"
	"reflect"
)

type RoleInstans struct {
	RoleBindings map[string]interface{}
}

const userName = "test_wtp"


func main() {
	var rBind RoleInstans
	var userBind = make(map[string]interface{})
	var kubeconfig *string

	rBind.RoleBindings = userBind

	ns := []string{"wgie-wtp-dataplatform"}

	rBind.RoleBindings["ed-ks1"] = ns

	if home := homeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()


	// create the RoleBindings from
	for cl, ns := range rBind.RoleBindings {

		log.Printf("Use the %s context in kubeconfig", cl)
		config, err :=  buildConfigFromFlags(cl, *kubeconfig)
		if err != nil {
			panic(err.Error())
		}

		// create the clientset
		clientset, err := kubernetes.NewForConfig(config)
		if err != nil {
			panic(err.Error())
		}

		log.Println("we create these namespaces:")

		s := reflect.ValueOf(ns)

		for i := 0; i < s.Len(); i++ {

			roleBindingName := fmt.Sprintf("%v-%v", userName, s.Index(i))

			nsName := fmt.Sprintf("%v", s.Index(i))

			_, err := clientset.RbacV1().RoleBindings(nsName).Get(roleBindingName, metav1.GetOptions{})

			if err == nil {
				log.Printf("Existing RoleBinding for %v in namespace %v \n", userName, nsName)
				continue
			}

			log.Printf("Creating RoleBinding for %v in namespace %v \n in cluster %v", userName, nsName, cl)

			roleBinding := RbacV1.RoleBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name: roleBindingName,
				},
				Subjects: []RbacV1.Subject{
					{
						Kind:      "User",
						Namespace: nsName,
						Name:      userName,
					},
				},
				RoleRef: RbacV1.RoleRef{
					APIGroup: "rbac.authorization.k8s.io",
					Kind:     "ClusterRole",
					Name:     "admin",
				},
			}

			_, err = clientset.RbacV1().RoleBindings(nsName).Create(&roleBinding)

			if err != nil {
				panic(err.Error())
			}

		}

	}

}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}

func buildConfigFromFlags(context, kubeconfigPath string) (*rest.Config, error) {
	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfigPath},
		&clientcmd.ConfigOverrides{
			CurrentContext: context,
		}).ClientConfig()
}