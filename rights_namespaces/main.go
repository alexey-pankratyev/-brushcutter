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
	//"reflect"
	"gopkg.in/yaml.v2"
)

var (
	configPath string
    userBind = make(map[string]interface{})
    kubeconfig *string
    userName string
)


type Bindings struct {
	Rolebindings map[string][]string
}


func main() {
	flag.StringVar(&configPath, "config", "./config.yml", "path to config file")
	flag.StringVar(&userName,"username", "", "This parameter is required. The user name for which we'll create rolebindings")

	if home := homeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	if userName == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	cfgPath, err := ParseFlags()
	if err != nil {
		log.Fatal(err)
	}

	cfg, err := NewConfig(cfgPath)
	if err != nil {
		log.Fatal(err)
	}

	// create the RoleBindings from
	for cl, ns := range cfg.Rolebindings {

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

		for _, v := range ns {

			roleBindingName := fmt.Sprintf("%v-%v", userName, v)

			nsName := fmt.Sprintf("%v", v)

			_, err := clientset.RbacV1().RoleBindings(nsName).Get(roleBindingName, metav1.GetOptions{})

			if err == nil {
				log.Printf("Existing RoleBinding for %v in namespace %v \n", userName, nsName)
				continue
			}

			log.Printf("Creating RoleBinding for %v in namespace %v \n in cluster %v", userName, nsName, cl)
			// to setup manifest
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

// ValidateConfigPath just makes sure, that the path provided is a file,
// that can be read
func ValidateConfigPath(path string) error {
	s, err := os.Stat(path)
	if err != nil {
		return err
	}
	if s.IsDir() {
		return fmt.Errorf("'%s' is a directory, not a normal file", path)
	}
	return nil
}


// NewConfig returns a new decoded Config struct
func NewConfig(configPath string) (*Bindings, error) {
	// Create config structure
	config := &Bindings{}

	// Open config file
	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Init new YAML decode
	d := yaml.NewDecoder(file)

	// Start YAML decoding from file
	if err := d.Decode(&config); err != nil {
		return nil, err
	}
	return config, nil
}

// ParseFlags will create and parse the CLI flags
// and return the path to be used elsewhere
func ParseFlags() (string, error) {

	// Validate the path first
	if err := ValidateConfigPath(configPath); err != nil {
		return "", err
	}

	// Return the configuration path
	return configPath, nil
}