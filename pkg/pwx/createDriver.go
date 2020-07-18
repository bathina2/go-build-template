package pwx

import (
	"fmt"

	"github.com/libopenstorage/openstorage/api"
	volumeclient "github.com/libopenstorage/openstorage/api/client/volume"
	"github.com/libopenstorage/openstorage/volume"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const (
	// serviceName is the name of the portworx service
	serviceName = "portworx-service"

	// namespace is the kubernetes namespace in which portworx
	// daemon set
	// runs
	namespace = "kube-system"

	// Config parameters
	configType = "type"

	typeLocal = "local"
	typeCloud = "cloud"
)

// GetClient get k8s client
func GetClient() (*kubernetes.Clientset, error) {
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	// creates the clientset
	return kubernetes.NewForConfig(config)
}

// GetVolumeDriver get the driver
func GetVolumeDriver() (volume.VolumeDriver, error) {
	k8sClient, err := GetClient()
	if err != nil {
		return nil, err
	}
	var endpoint string
	svc, err := k8sClient.CoreV1().Services(namespace).Get(serviceName, v1.GetOptions{})
	if err == nil {
		endpoint = svc.Spec.ClusterIP
	} else {
		return nil, fmt.Errorf("Failed to get k8s service spec: %v", err)
	}

	if len(endpoint) == 0 {
		return nil, fmt.Errorf("Failed to get endpoint for portworx volume driver")
	}

	clnt, err := volumeclient.NewDriverClient("http://"+endpoint+":9001", "pxd", "", "stork")
	if err != nil {
		return nil, err
	}
	return volumeclient.VolumeDriver(clnt), nil
}

// CreateSnapshot takes a volumeID as input and takes a snapshot
func CreateSnapshot(volumeID string) (string, error) {
	volDriver, err := GetVolumeDriver()
	if err != nil {
		return "", err
	}

	vols, err := volDriver.Inspect([]string{volumeID})
	if err != nil {
		return "", err
	}
	if len(vols) == 0 {
		return "", fmt.Errorf("Volume %v not found", volumeID)
	}

	// tags["pvName"] = vols[0].Locator.Name
	// l.log.Infof("Tags: %v", tags)
	locator := &api.VolumeLocator{
		Name: "pwxdemo_" + vols[0].Locator.Name,
		//VolumeLabels: tags,
	}
	snapshotID, err := volDriver.Snapshot(volumeID, true, locator, true)
	if err != nil {
		return "", err
	}

	return snapshotID, err
}

// CreateVolumeFromSnapshot creates volume from snapshot
func CreateVolumeFromSnapshot(snapshotID string) (string, error) {
	volDriver, err := GetVolumeDriver()
	if err != nil {
		return "", err
	}
	vols, err := volDriver.Inspect([]string{snapshotID})
	if err != nil {
		return "", nil
	}
	if len(vols) == 0 {
		return "", fmt.Errorf("Snapshot %v not found", snapshotID)
	}

	locator := &api.VolumeLocator{
		Name: vols[0].Locator.VolumeLabels["pvName"],
	}
	volumeID, err := volDriver.Snapshot(snapshotID, false, locator, true)
	if err != nil {
		return "", err
	}
	return volumeID, err
}

// CreatePVCFromSnapshot sadfasdf
func CreatePVCFromSnapshot(name string, storageclass string, snapshotID string, namespace string) error {
	volID, err := CreateVolumeFromSnapshot(snapshotID)
	if err != nil {
		return err
	}
	k8sClient, err := GetClient()
	if err != nil {
		return err
	}

	if len(name) == 0 || len(storageclass) == 0 {
		return fmt.Errorf("pvc/storageclass needs name")
	}
	pvName := name + "-pv"
	pvcName := name + "-pvc"

	pvc := corev1.PersistentVolumeClaim{
		ObjectMeta: v1.ObjectMeta{
			Name: pvcName,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			VolumeName:       pvName,
			StorageClassName: &storageclass,
		},
	}
	pvcRet, err := k8sClient.CoreV1().PersistentVolumeClaims(namespace).Create(&pvc)
	if err != nil {
		return err
	}
	fmt.Println(pvcRet)

	pv := corev1.PersistentVolume{
		ObjectMeta: v1.ObjectMeta{
			Name: pvName,
		},
		Spec: corev1.PersistentVolumeSpec{
			StorageClassName: storageclass,
		},
	}
	pv.Spec.PersistentVolumeSource.PortworxVolume = &corev1.PortworxVolumeSource{
		VolumeID: volID,
	}
	pvRet, err := k8sClient.CoreV1().PersistentVolumes().Create(&pv)
	if err != nil {
		return err
	}
	fmt.Println(pvRet)

	return nil
}
