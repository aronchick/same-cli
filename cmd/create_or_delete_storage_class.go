package cmd

import (
	"context"
	"fmt"

	apiv1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// StorageClassSetupInterface : Interface for creating and deleting Storage Classes in Kubernetes
type StorageClassSetupInterface interface {
	Create()
	Delete()

	configureStorageClass(config map[string]string)
}

type azureBlobStorageClassSetup struct {
	storageClass *storagev1.StorageClass
	clientset    *kubernetes.Clientset
}

func (setup *azureBlobStorageClassSetup) configureStorageClass(config map[string]string) {

	setup.storageClass = new(storagev1.StorageClass)
	setup.storageClass.ObjectMeta.Name = "blob"
	setup.storageClass.Provisioner = "blob.csi.azure.com"
	setup.storageClass.Parameters = map[string]string{
		"resourceGroup":  config["blobStorageResourceGroup"],
		"storageAccount": config["blobStorageAccountName"],
		"containerName":  config["blobStorageContainerName"]}
	// "server":         blobStorageAccountName + ".blob.core.windows.net"} // optional, provide a new address to replace default "accountname.blob.core.windows.net"
	setup.storageClass.ReclaimPolicy = new(apiv1.PersistentVolumeReclaimPolicy)
	*setup.storageClass.ReclaimPolicy = apiv1.PersistentVolumeReclaimRetain
	setup.storageClass.VolumeBindingMode = new(storagev1.VolumeBindingMode)
	*setup.storageClass.VolumeBindingMode = storagev1.VolumeBindingImmediate
	setup.storageClass.MountOptions = []string{
		"-o allow_other",
		"--file-cache-timeout-in-seconds=120",
		"--use-attr-cache=true",
		"-o attr_timeout=120",
		"-o entry_timeout=120",
		"-o negative_timeout=120"}

}

// Create : Create Storage Class in Kubernetes Cluster
func (setup azureBlobStorageClassSetup) Create() {
	fmt.Println("Creating storage class...")
	storageClassClient := setup.clientset.StorageV1().StorageClasses()
	result, err := storageClassClient.Create(context.TODO(), setup.storageClass, metav1.CreateOptions{})
	if err != nil {
		panic(err)
	}
	fmt.Printf("Created storage class %q.\n", result.GetObjectMeta().GetName())
}

// Delete : Delete Storage Class from Kubernetes Cluster
func (setup azureBlobStorageClassSetup) Delete() {
	fmt.Println("Deleting storage class...")
	deletePolicy := metav1.DeletePropagationForeground
	storageClassClient := setup.clientset.StorageV1().StorageClasses()
	if err := storageClassClient.Delete(context.TODO(), setup.storageClass.ObjectMeta.Name, metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}); err != nil {
		panic(err)
	}
	fmt.Println("Deleted Storage Class.")
}

// NewAzureBlobStorageSetup : Create new Setup client to enable Blob Storage support in AKS
func NewAzureBlobStorageSetup(clientset *kubernetes.Clientset, blobStorageResourceGroup string, blobStorageAccountName string, blobStorageContainerName string) StorageClassSetupInterface {
	setup := new(azureBlobStorageClassSetup)
	config := map[string]string{
		"blobStorageResourceGroup": blobStorageResourceGroup,
		"blobStorageAccountName":   blobStorageAccountName,
		"blobStorageContainerName": blobStorageContainerName}
	setup.configureStorageClass(config)
	setup.clientset = clientset
	return setup

}

// Below is not yet implemented - blocking linting
// func createStorage() {

// 	// Load kubeconfig
// 	var kubeconfig string
// 	if home := homedir.HomeDir(); home != "" {
// 		kubeconfig = filepath.Join(home, ".kube", "config")
// 	} else {
// 		panic("Could not find kube config!")
// 	}
// 	flag.Parse()

// 	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
// 	if err != nil {
// 		panic(err)
// 	}
// 	clientset, err := kubernetes.NewForConfig(config)
// 	if err != nil {
// 		panic(err)
// 	}

// 	// Create the Azure Blob Storage Setup here
// 	storageSetup := NewAzureBlobStorageSetup(clientset, "BlobStorageResourceGroup", "BlobStorageAccountName", "BlobStorageContainerName")
// 	storageSetup.Create()

// }
