// +build csi

package e2e

import (
	goctx "context"
	"strings"
	"testing"

	framework "github.com/operator-framework/operator-sdk/pkg/test"
	storageos "github.com/storageos/cluster-operator/pkg/apis/storageos/v1alpha1"
	deploy "github.com/storageos/cluster-operator/pkg/storageos"
	testutil "github.com/storageos/cluster-operator/test/e2e/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func TestClusterCSI(t *testing.T) {
	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup()

	namespace, err := ctx.GetNamespace()
	if err != nil {
		t.Fatalf("could not get namespace: %v", err)
	}

	clusterSpec := storageos.StorageOSClusterSpec{
		SecretRefName:      "storageos-api",
		SecretRefNamespace: "default",
		ResourceNS:         "storageos",
		CSI: storageos.StorageOSClusterCSI{
			Enable: true,
		},
	}

	testStorageOS := testutil.NewStorageOSCluster(namespace, clusterSpec)

	testutil.SetupOperator(t, ctx)
	err = testutil.DeployCluster(t, ctx, testStorageOS)
	if err != nil {
		t.Fatal(err)
	}

	f := framework.Global

	err = f.Client.Get(goctx.TODO(), types.NamespacedName{Name: "example-storageos", Namespace: namespace}, testStorageOS)
	if err != nil {
		t.Fatal(err)
	}

	testutil.ClusterStatusCheck(t, testStorageOS.Status, 1)

	daemonset, err := f.KubeClient.AppsV1().DaemonSets("storageos").Get("storageos-daemonset", metav1.GetOptions{IncludeUninitialized: true})
	if err != nil {
		t.Fatalf("failed to get storageos-daemonset: %v", err)
	}

	info, err := f.KubeClient.Discovery().ServerVersion()
	if err != nil {
		t.Fatalf("failed to get version info: %v", err)
	}

	version := strings.TrimLeft(info.String(), "v")

	//Check the number of containers in daemonset pod spec.
	if deploy.CSIV1Supported(version) {
		if len(daemonset.Spec.Template.Spec.Containers) != 3 {
			t.Errorf("unexpected number of daemonset pod containers:\n\t(GOT) %d\n\t(WNT) %d", len(daemonset.Spec.Template.Spec.Containers), 2)
		}
	} else {
		if len(daemonset.Spec.Template.Spec.Containers) != 2 {
			t.Errorf("unexpected number of daemonset pod containers:\n\t(GOT) %d\n\t(WNT) %d", len(daemonset.Spec.Template.Spec.Containers), 2)
		}
	}
}
