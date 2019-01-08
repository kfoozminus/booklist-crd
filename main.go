package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	kfoozminusv1 "github.com/kfoozminus/booklist-crd/pkg/apis/kfoozminus.com/v1"
	kfoozminusclientset "github.com/kfoozminus/booklist-crd/pkg/client/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"k8s.io/client-go/util/retry"
)

func main() {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err)
	}
	clientset, err := kfoozminusclientset.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	podjenniesClient := clientset.KfoozminusV1().Podjennies(kfoozminusv1.NamespaceDefault)

	podjenny := &kfoozminusv1.Podjenny{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Podjenny",
			APIVersion: kfoozminusv1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "podjenny-client",
			Namespace: kfoozminusv1.NamespaceDefault,
			Labels: map[string]string{
				"app": "podjenny-client",
			},
		},
		Spec: kfoozminusv1.PodjennySpec{
			Image: "kfoozminus/booklistgo:alpine",
		},
	}
	waitForEnter()
	fmt.Println("Creating Podjenny...")
	resultPodjenny, err := podjenniesClient.Create(podjenny)
	if err != nil {
		panic(fmt.Errorf("Error while creating Podjenny - %v\n", err))
	}
	fmt.Printf("Created Podjenny - Name: %q, UID: %q\n", resultPodjenny.GetObjectMeta().GetName(), resultPodjenny.GetObjectMeta().GetUID())

	//update the Podjenny via Update method
	waitForEnter()
	fmt.Println("Updating Podjenny...")
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		resultPodjenny, getErr := podjenniesClient.Get("podjenny-client", metav1.GetOptions{})
		if getErr != nil {
			panic(fmt.Errorf("Error while getting the Podjenny object - %v\n", getErr))
		}

		resultPodjenny.Spec.Image = "kfoozminus/booklist:ubuntu"

		_, updateErr := podjenniesClient.Update(resultPodjenny)
		return updateErr
	})
	if retryErr != nil {
		panic(fmt.Errorf("Error in updating the Podjenny object - %v\n", retryErr))
	}
	fmt.Printf("Updated Podjenny - Name: %q, UID: %q\n", resultPodjenny.GetObjectMeta().GetName(), resultPodjenny.GetObjectMeta().GetUID())

	//list pv via List Method
	waitForEnter()
	fmt.Println("Listing Podjenny...")
	resultPjList, listErr := podjenniesClient.List(metav1.ListOptions{})
	if listErr != nil {
		panic(fmt.Errorf("Error while listing the podjenny - %v\n", listErr))
	}
	for _, pj := range resultPjList.Items {
		fmt.Printf("Podjenny - Name %v - Spec.Image %v\n", pj.Name, pj.Spec.Image)
	}

	//delete objects via Delete Method
	waitForEnter()
	fmt.Println("Deleting Podjenny the objects...")
	deletePolicy := metav1.DeletePropagationForeground
	if deleteErr := podjenniesClient.Delete("podjenny-client", &metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}); deleteErr != nil {
		panic(fmt.Errorf("Error while deleting Podjenny - %v\n", deleteErr))
	}
	fmt.Println("Deleted Podjenny")
}

func waitForEnter() {
	fmt.Println("..... Press Enter to Continue .....")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	if err := scanner.Err(); err != nil {
		panic(err)
	}
}
