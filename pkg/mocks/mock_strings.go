package mocks

var (
	INIT_TEST_K3S_STARTED_BUT_SERVICES_FAILED_PROBE  string = "k3s-started-but-services-failed"
	INIT_TEST_K3S_STARTED_BUT_SERVICES_FAILED_RESULT string = "K3S STARTED BUT SERVICES FAILED"

	CREATE_TEST_KUBECTL_MISSING_PROBE  string = "kubectl-not-on-path"
	CREATE_TEST_KUBECTL_MISSING_RESULT string = "KUBECTL NOT ON PATH"

	DEPENDENCY_CHECKER_KUBECTL_ON_PATH_PROBE  string = "dependency-checker-kubectl-not-on-path"
	DEPENDENCY_CHECKER_KUBECTL_ON_PATH_RESULT string = "DEPENDENCY CHECKER KUBECTL NOT ON PATH"

	DEPENDENCY_CHECKER_CANNOT_CONNECT_TO_K8S_PROBE  string = "dependency-checker-cannot-connect-to-kubernetes"
	DEPENDENCY_CHECKER_CANNOT_CONNECT_TO_K8S_RESULT string = "DEPENDENCY CHECKER CANNOT CONNECT TO KUBERNETES"

	DEPENDENCY_CHECKER_MISSING_KUBEFLOW_NAMESPACE_PROBE  string = "dependency-checker-missing-kubeflow-namespace"
	DEPENDENCY_CHECKER_MISSING_KUBEFLOW_NAMESPACE_RESULT string = "DEPENDENCY CHECKER MISSING KUBEFLOW NAMESPACE"

	UTILS_TEST_K3S_RUNNING_FAILED_PROBE  string = "k3s-running-test-failed"
	UTILS_TEST_K3S_RUNNING_FAILED_RESULT string = "K3S RUNNING TEST FAILED"
)
