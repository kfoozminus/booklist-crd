package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	corejennyv1 "github.com/kfoozminus/booklist-crd/pkg/apis/corejenny/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
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
			Name:      "booklistkube-client",
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

	service := &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: corev1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "booklistkube-client",
			Namespace: "default",
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app": "booklistkube-client",
			},
			Ports: []corev1.ServicePort{
				{
					Name:       "exposeds",
					Port:       1234,
					TargetPort: intstr.IntOrString{StrVal: "exposedc", Type: 1},
					//TargetPort: intstr.IntOrString{IntVal: 4321},
				},
			},
			Type: corev1.ServiceTypeNodePort,
		},
	}
	waitForEnter()
	fmt.Println("Creating Service...")
	//time.Sleep(60 * time.Second)
	resultService, err := servicesClient.Create(service)
	if err != nil {
		panic(fmt.Errorf("Error while creating Service - %v\n", err))
	}
	fmt.Printf("Created Service - Name: %q, UID: %q\n", resultService.GetObjectMeta().GetName(), resultService.GetObjectMeta().GetUID())

	//create or patch service via appscode/kutil
	waitForEnter()
	fmt.Println("Patching Service...")
	servicePatch, kutilVerb, kutilErr := corev1Kutil.CreateOrPatchService(clientset, service.ObjectMeta, func(serviceTransformed *corev1.Service) *corev1.Service {
		//serviceTransformed = resultService
		serviceTransformed.Spec.Ports[0].Port = 2345
		return serviceTransformed
	})
	if kutilErr != nil {
		panic(fmt.Errorf("Error while patching Service - %v\n", kutilErr))
	}
	fmt.Printf("%v - Name: %q, UID: %q\n", kutilVerb, servicePatch.GetObjectMeta().GetName(), servicePatch.GetObjectMeta().GetUID())

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
	waitForEnter()
	fmt.Println("Deleting All the objects...")
	deletePolicy := metav1.DeletePropagationForeground
	if deleteErr := deploymentsClient.Delete("booklistkube-client", &metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}); deleteErr != nil {
		panic(fmt.Errorf("Error while deleting deployment - %v\n", deleteErr))
	}
	fmt.Println("Deleted Deployment")

	if deleteErr := pvcsClient.Delete("task-pv-claim-client", &metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}); deleteErr != nil {
		panic(fmt.Errorf("Error while deleting pvc - %v\n", deleteErr))
	}
	fmt.Println("Deleted PVC")

	if deleteErr := pvsClient.Delete("task-pv-volume-client", &metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}); deleteErr != nil {
		panic(fmt.Errorf("Error while deleting pv - %v\n", deleteErr))
	}
	fmt.Println("Deleted PV")

	if deleteErr := servicesClient.Delete("booklistkube-client", &metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}); deleteErr != nil {
		panic(fmt.Errorf("Error while deleting service - %v\n", deleteErr))
	}
	fmt.Println("Deleted Service")
}

func waitForEnter() {
	fmt.Println("..... Press Enter to Continue .....")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	if err := scanner.Err(); err != nil {
		panic(err)
	}
}

func int32ptr(i int32) *int32                                               { return &i }
func hostpathtypeptr(hostpathtype corev1.HostPathType) *corev1.HostPathType { return &hostpathtype }
