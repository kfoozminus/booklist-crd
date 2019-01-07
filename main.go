package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	corejennyv1 "github.com/kfoozminus/booklist-crd/pkg/apis/corejenny/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
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
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	podjenniesClient := clientset.CorejennyV1().Podjennies(corejennyv1.NamespaceDefault)

	podjenny := &corejennyv1.Podjenny{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Podjenny",
			APIVersion: corejennyv1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "podjenny-crd",
			Namespace: "default",
			Labels: map[string]string{
				"app": "booklistkube-client",
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32ptr(3),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "booklistkube-client",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name: "booklistkube-client",
					Labels: map[string]string{
						"app": "booklistkube-client",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:            "booklistkube-client",
							Image:           "kfoozminus/booklist:alpine",
							ImagePullPolicy: corev1.PullIfNotPresent,
							//Command:         []string{"/bin/sh", "-c", "echo hello; sleep 36000"},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "task-pv-storage-client",
									MountPath: "/etc/pvc",
								},
							},
							Ports: []corev1.ContainerPort{
								{
									Name:          "exposedc",
									ContainerPort: 4321,
									Protocol:      corev1.ProtocolTCP,
								},
							},
						},
					},
					RestartPolicy: corev1.RestartPolicyAlways,
					Volumes: []corev1.Volume{
						{
							Name: "task-pv-storage-client",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: "task-pv-claim-client",
								},
							},
						},
					},
				},
			},
		},
	}
	waitForEnter()
	fmt.Println("Creating Deployment...")
	resultDeployment, err := deploymentsClient.Create(deployment)
	if err != nil {
		panic(fmt.Errorf("Error while creating Deployment - %v\n", err))
	}
	fmt.Printf("Created Deployment - Name: %q, UID: %q\n", resultDeployment.GetObjectMeta().GetName(), resultDeployment.GetObjectMeta().GetUID())

	//create or patch deployment via appscode/kutil
	waitForEnter()
	fmt.Println("Patching Deployment...")
	deploymentPatch, kutilVerb, kutilErr := appsv1Kutil.CreateOrPatchDeployment(clientset, deployment.ObjectMeta, func(deploymentTransformed *appsv1.Deployment) *appsv1.Deployment {
		deploymentTransformed.Spec.Replicas = int32ptr(4)
		return deploymentTransformed
	})
	if kutilErr != nil {
		panic(fmt.Errorf("Error while patching Deployment - %v\n", kutilErr))
	}
	fmt.Printf("%v - Name: %q, UID: %q\n", kutilVerb, deploymentPatch.GetObjectMeta().GetName(), deploymentPatch.GetObjectMeta().GetUID())

	//update the deployment via Update method
	waitForEnter()
	fmt.Println("Updating Deployment...")
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		resultDeployment, getErr := deploymentsClient.Get("booklistkube-client", metav1.GetOptions{})
		if getErr != nil {
			panic(fmt.Errorf("Error while getting the Deployment object - %v\n", getErr))
		}

		resultDeployment.Spec.Replicas = int32ptr(5)
		resultDeployment.Spec.Template.Spec.Containers[0].Image = "kfoozminus/booklist:ubuntu"

		_, updateErr := deploymentsClient.Update(resultDeployment)
		return updateErr
	})
	if retryErr != nil {
		panic(fmt.Errorf("Error in updating the Deployment object - %v\n", retryErr))
	}
	fmt.Printf("Updated Deployment - Name: %q, UID: %q\n", resultDeployment.GetObjectMeta().GetName(), resultDeployment.GetObjectMeta().GetUID())

	//list pv via List Method
	waitForEnter()
	fmt.Println("Listing PVs...")
	resultPvList, listErr := pvsClient.List(metav1.ListOptions{})
	if listErr != nil {
		panic(fmt.Errorf("Error while listing the pvs - %v\n", listErr))
	}
	for _, pv := range resultPvList.Items {
		name := "None"
		if pv.Spec.ClaimRef != nil {
			name = pv.Spec.ClaimRef.Name
		}
		fmt.Printf("Persistent Volume - Name %v - Capacity %v - AccessModes %v - ReclaimPolicy %v - Status %v - Claim %v - StorageClass %v - Reason %v\n", pv.Name, pv.Spec.Capacity[corev1.ResourceStorage], pv.Spec.AccessModes, pv.Spec.PersistentVolumeReclaimPolicy, pv.Status.Phase, name, pv.Spec.StorageClassName, pv.Status.Reason)
	}

	//delete objects via Delete Method
	fmt.Println("Deleting All the objects...")
	waitForEnter()
	deletePolicy := metav1.DeletePropagationForeground
	if deleteErr := deploymentsClient.Delete("booklistkube-client", &metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}); deleteErr != nil {
		panic(fmt.Errorf("Error while deleting deployment - %v\n", deleteErr))
	}
	fmt.Println("Deleted Deployment")
}

func waitForEnter() {
	fmt.Println("..... Press Enter to Continue .....")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	if err := scanner.Err(); err != nil {
		panic(err)
	}
}
