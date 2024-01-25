package v1alpha1

// This file contains a collection of methods that can be used from go-restful to
// generate Swagger API documentation for its models. Please read this PR for more
// information on the implementation: https://github.com/emicklei/go-restful/pull/215
//
// TODOs are ignored from the parser (e.g. TODO(andronat):... || TODO:...) if and only if
// they are on one line! For multiple line or blocks that you want to ignore use ---.
// Any context after a --- is ignored.
//
// Those methods can be generated by using hack/update-swagger-docs.sh

// AUTO-GENERATED FUNCTIONS START HERE
var map_Backup = map[string]string{
	"":         "\n\nBackup provides configuration for performing backups of the openshift cluster.\n\nCompatibility level 4: No compatibility is provided, the API can change at any point for any reason. These capabilities should not be used by applications needing long term support.",
	"metadata": "metadata is the standard object's metadata. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata",
	"spec":     "spec holds user settable values for configuration",
	"status":   "status holds observed values from the cluster. They may not be overridden.",
}

func (Backup) SwaggerDoc() map[string]string {
	return map_Backup
}

var map_BackupList = map[string]string{
	"":         "BackupList is a collection of items\n\nCompatibility level 4: No compatibility is provided, the API can change at any point for any reason. These capabilities should not be used by applications needing long term support.",
	"metadata": "metadata is the standard list's metadata. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata",
}

func (BackupList) SwaggerDoc() map[string]string {
	return map_BackupList
}

var map_BackupSpec = map[string]string{
	"etcd": "etcd specifies the configuration for periodic backups of the etcd cluster",
}

func (BackupSpec) SwaggerDoc() map[string]string {
	return map_BackupSpec
}

var map_EtcdBackupSpec = map[string]string{
	"":                "EtcdBackupSpec provides configuration for automated etcd backups to the cluster-etcd-operator",
	"schedule":        "Schedule defines the recurring backup schedule in Cron format every 2 hours: 0 */2 * * * every day at 3am: 0 3 * * * Empty string means no opinion and the platform is left to choose a reasonable default which is subject to change without notice. The current default is \"no backups\", but will change in the future.",
	"timeZone":        "The time zone name for the given schedule, see https://en.wikipedia.org/wiki/List_of_tz_database_time_zones. If not specified, this will default to the time zone of the kube-controller-manager process. See https://kubernetes.io/docs/concepts/workloads/controllers/cron-jobs/#time-zones",
	"retentionPolicy": "RetentionPolicy defines the retention policy for retaining and deleting existing backups.",
	"pvcName":         "PVCName specifies the name of the PersistentVolumeClaim (PVC) which binds a PersistentVolume where the etcd backup files would be saved The PVC itself must always be created in the \"openshift-etcd\" namespace If the PVC is left unspecified \"\" then the platform will choose a reasonable default location to save the backup. In the future this would be backups saved across the control-plane master nodes.",
}

func (EtcdBackupSpec) SwaggerDoc() map[string]string {
	return map_EtcdBackupSpec
}

var map_RetentionNumberConfig = map[string]string{
	"":                   "RetentionNumberConfig specifies the configuration of the retention policy on the number of backups",
	"maxNumberOfBackups": "MaxNumberOfBackups defines the maximum number of backups to retain. If the existing number of backups saved is equal to MaxNumberOfBackups then the oldest backup will be removed before a new backup is initiated.",
}

func (RetentionNumberConfig) SwaggerDoc() map[string]string {
	return map_RetentionNumberConfig
}

var map_RetentionPolicy = map[string]string{
	"":                "RetentionPolicy defines the retention policy for retaining and deleting existing backups. This struct is a discriminated union that allows users to select the type of retention policy from the supported types.",
	"retentionType":   "RetentionType sets the type of retention policy. Currently, the only valid policies are retention by number of backups (RetentionNumber), by the size of backups (RetentionSize). More policies or types may be added in the future. Empty string means no opinion and the platform is left to choose a reasonable default which is subject to change without notice. The current default is RetentionNumber with 15 backups kept.",
	"retentionNumber": "RetentionNumber configures the retention policy based on the number of backups",
	"retentionSize":   "RetentionSize configures the retention policy based on the size of backups",
}

func (RetentionPolicy) SwaggerDoc() map[string]string {
	return map_RetentionPolicy
}

var map_RetentionSizeConfig = map[string]string{
	"":                   "RetentionSizeConfig specifies the configuration of the retention policy on the total size of backups",
	"maxSizeOfBackupsGb": "MaxSizeOfBackupsGb defines the total size in GB of backups to retain. If the current total size backups exceeds MaxSizeOfBackupsGb then the oldest backup will be removed before a new backup is initiated.",
}

func (RetentionSizeConfig) SwaggerDoc() map[string]string {
	return map_RetentionSizeConfig
}

var map_ClusterImagePolicy = map[string]string{
	"":         "ClusterImagePolicy holds cluster-wide configuration for image signature verification\n\nCompatibility level 4: No compatibility is provided, the API can change at any point for any reason. These capabilities should not be used by applications needing long term support.",
	"metadata": "metadata is the standard object's metadata. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata",
	"spec":     "spec contains the configuration for the cluster image policy.",
	"status":   "status contains the observed state of the resource.",
}

func (ClusterImagePolicy) SwaggerDoc() map[string]string {
	return map_ClusterImagePolicy
}

var map_ClusterImagePolicyList = map[string]string{
	"":         "ClusterImagePolicyList is a list of ClusterImagePolicy resources\n\nCompatibility level 4: No compatibility is provided, the API can change at any point for any reason. These capabilities should not be used by applications needing long term support.",
	"metadata": "metadata is the standard list's metadata. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata",
}

func (ClusterImagePolicyList) SwaggerDoc() map[string]string {
	return map_ClusterImagePolicyList
}

var map_ClusterImagePolicySpec = map[string]string{
	"":       "CLusterImagePolicySpec is the specification of the ClusterImagePolicy custom resource.",
	"scopes": "scopes defines the list of image identities assigned to a policy. Each item refers to a scope in a registry implementing the \"Docker Registry HTTP API V2\". Scopes matching individual images are named Docker references in the fully expanded form, either using a tag or digest. For example, docker.io/library/busybox:latest (not busybox:latest). More general scopes are prefixes of individual-image scopes, and specify a repository (by omitting the tag or digest), a repository namespace, or a registry host (by only specifying the host name and possibly a port number) or a wildcard expression starting with `*.`, for matching all subdomains (not including a port number). Wildcards are only supported for subdomain matching, and may not be used in the middle of the host, i.e.  *.example.com is a valid case, but example*.*.com is not. Please be aware that the scopes should not be nested under the repositories of OpenShift Container Platform images. If configured, the policies for OpenShift Container Platform repositories will not be in effect. For additional details about the format, please refer to the document explaining the docker transport field, which can be found at: https://github.com/containers/image/blob/main/docs/containers-policy.json.5.md#docker",
	"policy": "policy contains configuration to allow scopes to be verified, and defines how images not matching the verification policy will be treated.",
}

func (ClusterImagePolicySpec) SwaggerDoc() map[string]string {
	return map_ClusterImagePolicySpec
}

var map_ClusterImagePolicyStatus = map[string]string{
	"conditions": "conditions provide details on the status of this API Resource.",
}

func (ClusterImagePolicyStatus) SwaggerDoc() map[string]string {
	return map_ClusterImagePolicyStatus
}

var map_FulcioCAWithRekor = map[string]string{
	"":              "FulcioCAWithRekor defines the root of trust based on the Fulcio certificate and the Rekor public key.",
	"fulcioCAData":  "fulcioCAData contains inline base64-encoded data for the PEM format fulcio CA. fulcioCAData must be at most 8192 characters.",
	"rekorKeyData":  "rekorKeyData contains inline base64-encoded data for the PEM format from the Rekor public key. rekorKeyData must be at most 8192 characters.",
	"fulcioSubject": "fulcioSubject specifies OIDC issuer and the email of the Fulcio authentication configuration.",
}

func (FulcioCAWithRekor) SwaggerDoc() map[string]string {
	return map_FulcioCAWithRekor
}

var map_ImagePolicy = map[string]string{
	"":         "ImagePolicy holds namespace-wide configuration for image signature verification\n\nCompatibility level 4: No compatibility is provided, the API can change at any point for any reason. These capabilities should not be used by applications needing long term support.",
	"metadata": "metadata is the standard object's metadata. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata",
	"spec":     "spec holds user settable values for configuration",
	"status":   "status contains the observed state of the resource.",
}

func (ImagePolicy) SwaggerDoc() map[string]string {
	return map_ImagePolicy
}

var map_ImagePolicyList = map[string]string{
	"":         "ImagePolicyList is a list of ImagePolicy resources\n\nCompatibility level 4: No compatibility is provided, the API can change at any point for any reason. These capabilities should not be used by applications needing long term support.",
	"metadata": "metadata is the standard list's metadata. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata",
}

func (ImagePolicyList) SwaggerDoc() map[string]string {
	return map_ImagePolicyList
}

var map_ImagePolicySpec = map[string]string{
	"":       "ImagePolicySpec is the specification of the ImagePolicy CRD.",
	"scopes": "scopes defines the list of image identities assigned to a policy. Each item refers to a scope in a registry implementing the \"Docker Registry HTTP API V2\". Scopes matching individual images are named Docker references in the fully expanded form, either using a tag or digest. For example, docker.io/library/busybox:latest (not busybox:latest). More general scopes are prefixes of individual-image scopes, and specify a repository (by omitting the tag or digest), a repository namespace, or a registry host (by only specifying the host name and possibly a port number) or a wildcard expression starting with `*.`, for matching all subdomains (not including a port number). Wildcards are only supported for subdomain matching, and may not be used in the middle of the host, i.e.  *.example.com is a valid case, but example*.*.com is not. Please be aware that the scopes should not be nested under the repositories of OpenShift Container Platform images. If configured, the policies for OpenShift Container Platform repositories will not be in effect. For additional details about the format, please refer to the document explaining the docker transport field, which can be found at: https://github.com/containers/image/blob/main/docs/containers-policy.json.5.md#docker",
	"policy": "policy contains configuration to allow scopes to be verified, and defines how images not matching the verification policy will be treated.",
}

func (ImagePolicySpec) SwaggerDoc() map[string]string {
	return map_ImagePolicySpec
}

var map_ImagePolicyStatus = map[string]string{
	"conditions": "conditions provide details on the status of this API Resource.",
}

func (ImagePolicyStatus) SwaggerDoc() map[string]string {
	return map_ImagePolicyStatus
}

var map_Policy = map[string]string{
	"":               "Policy defines the verification policy for the items in the scopes list.",
	"rootOfTrust":    "rootOfTrust specifies the root of trust for the policy.",
	"signedIdentity": "signedIdentity specifies what image identity the signature claims about the image. The required matchPolicy field specifies the approach used in the verification process to verify the identity in the signature and the actual image identity, the default matchPolicy is \"MatchRepoDigestOrExact\".",
}

func (Policy) SwaggerDoc() map[string]string {
	return map_Policy
}

var map_PolicyFulcioSubject = map[string]string{
	"":            "PolicyFulcioSubject defines the OIDC issuer and the email of the Fulcio authentication configuration.",
	"oidcIssuer":  "oidcIssuer contains the expected OIDC issuer. It will be verified that the Fulcio-issued certificate contains a (Fulcio-defined) certificate extension pointing at this OIDC issuer URL. When Fulcio issues certificates, it includes a value based on an URL inside the client-provided ID token. Example: \"https://expected.OIDC.issuer/\"",
	"signedEmail": "signedEmail holds the email address the the Fulcio certificate is issued for. Example: \"expected-signing-user@example.com\"",
}

func (PolicyFulcioSubject) SwaggerDoc() map[string]string {
	return map_PolicyFulcioSubject
}

var map_PolicyIdentity = map[string]string{
	"":                "PolicyIdentity defines image identity the signature claims about the image. When omitted, the default matchPolicy is \"MatchRepoDigestOrExact\".",
	"matchPolicy":     "matchPolicy sets the type of matching to be used. Valid values are \"MatchRepoDigestOrExact\", \"MatchRepository\", \"ExactRepository\", \"RemapIdentity\". When omitted, the default value is \"MatchRepoDigestOrExact\". If set matchPolicy to ExactRepository, then the exactRepository must be specified. If set matchPolicy to RemapIdentity, then the remapIdentity must be specified. \"MatchRepoDigestOrExact\" means that the identity in the signature must be in the same repository as the image identity if the image identity is referenced by a digest. Otherwise, the identity in the signature must be the same as the image identity. \"MatchRepository\" means that the identity in the signature must be in the same repository as the image identity. \"ExactRepository\" means that the identity in the signature must be in the same repository as a specific identity specified by \"repository\". \"RemapIdentity\" means that the signature must be in the same as the remapped image identity. Remapped image identity is obtained by replacing the \"prefix\" with the specified “signedPrefix” if the the image identity matches the specified remapPrefix.",
	"exactRepository": "exactRepository is required if matchPolicy is set to \"ExactRepository\".",
	"remapIdentity":   "remapIdentity is required if matchPolicy is set to \"RemapIdentity\".",
}

func (PolicyIdentity) SwaggerDoc() map[string]string {
	return map_PolicyIdentity
}

var map_PolicyMatchExactRepository = map[string]string{
	"repository": "repository is the reference of the image identity to be matched. The value should be a repository name (by omitting the tag or digest) in a registry implementing the \"Docker Registry HTTP API V2\". For example, docker.io/library/busybox",
}

func (PolicyMatchExactRepository) SwaggerDoc() map[string]string {
	return map_PolicyMatchExactRepository
}

var map_PolicyMatchRemapIdentity = map[string]string{
	"prefix":       "prefix is the prefix of the image identity to be matched. If the image identity matches the specified prefix, that prefix is replaced by the specified “signedPrefix” (otherwise it is used as unchanged and no remapping takes place). This useful when verifying signatures for a mirror of some other repository namespace that preserves the vendor’s repository structure. The prefix and signedPrefix values can be either host[:port] values (matching exactly the same host[:port], string), repository namespaces, or repositories (i.e. they must not contain tags/digests), and match as prefixes of the fully expanded form. For example, docker.io/library/busybox (not busybox) to specify that single repository, or docker.io/library (not an empty string) to specify the parent namespace of docker.io/library/busybox.",
	"signedPrefix": "signedPrefix is the prefix of the image identity to be matched in the signature. The format is the same as \"prefix\". The values can be either host[:port] values (matching exactly the same host[:port], string), repository namespaces, or repositories (i.e. they must not contain tags/digests), and match as prefixes of the fully expanded form. For example, docker.io/library/busybox (not busybox) to specify that single repository, or docker.io/library (not an empty string) to specify the parent namespace of docker.io/library/busybox.",
}

func (PolicyMatchRemapIdentity) SwaggerDoc() map[string]string {
	return map_PolicyMatchRemapIdentity
}

var map_PolicyRootOfTrust = map[string]string{
	"":                  "PolicyRootOfTrust defines the root of trust based on the selected policyType.",
	"policyType":        "policyType serves as the union's discriminator. Users are required to assign a value to this field, choosing one of the policy types that define the root of trust. \"PublicKey\" indicates that the policy relies on a sigstore publicKey and may optionally use a Rekor verification. \"FulcioCAWithRekor\" indicates that the policy is based on the Fulcio certification and incorporates a Rekor verification.",
	"publicKey":         "publicKey defines the root of trust based on a sigstore public key.",
	"fulcioCAWithRekor": "fulcioCAWithRekor defines the root of trust based on the Fulcio certificate and the Rekor public key. For more information about Fulcio and Rekor, please refer to the document at: https://github.com/sigstore/fulcio and https://github.com/sigstore/rekor",
}

func (PolicyRootOfTrust) SwaggerDoc() map[string]string {
	return map_PolicyRootOfTrust
}

var map_PublicKey = map[string]string{
	"":             "PublicKey defines the root of trust based on a sigstore public key.",
	"keyData":      "keyData contains inline base64-encoded data for the PEM format public key. KeyData must be at most 8192 characters.",
	"rekorKeyData": "rekorKeyData contains inline base64-encoded data for the PEM format from the Rekor public key. rekorKeyData must be at most 8192 characters.",
}

func (PublicKey) SwaggerDoc() map[string]string {
	return map_PublicKey
}

var map_GatherConfig = map[string]string{
	"":                  "gatherConfig provides data gathering configuration options.",
	"dataPolicy":        "dataPolicy allows user to enable additional global obfuscation of the IP addresses and base domain in the Insights archive data. Valid values are \"None\" and \"ObfuscateNetworking\". When set to None the data is not obfuscated. When set to ObfuscateNetworking the IP addresses and the cluster domain name are obfuscated. When omitted, this means no opinion and the platform is left to choose a reasonable default, which is subject to change over time. The current default is None.",
	"disabledGatherers": "disabledGatherers is a list of gatherers to be excluded from the gathering. All the gatherers can be disabled by providing \"all\" value. If all the gatherers are disabled, the Insights operator does not gather any data. The particular gatherers IDs can be found at https://github.com/openshift/insights-operator/blob/master/docs/gathered-data.md. Run the following command to get the names of last active gatherers: \"oc get insightsoperators.operator.openshift.io cluster -o json | jq '.status.gatherStatus.gatherers[].name'\" An example of disabling gatherers looks like this: `disabledGatherers: [\"clusterconfig/machine_configs\", \"workloads/workload_info\"]`",
}

func (GatherConfig) SwaggerDoc() map[string]string {
	return map_GatherConfig
}

var map_InsightsDataGather = map[string]string{
	"":         "\n\nInsightsDataGather provides data gather configuration options for the the Insights Operator.\n\nCompatibility level 4: No compatibility is provided, the API can change at any point for any reason. These capabilities should not be used by applications needing long term support.",
	"metadata": "metadata is the standard object's metadata. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata",
	"spec":     "spec holds user settable values for configuration",
	"status":   "status holds observed values from the cluster. They may not be overridden.",
}

func (InsightsDataGather) SwaggerDoc() map[string]string {
	return map_InsightsDataGather
}

var map_InsightsDataGatherList = map[string]string{
	"":         "InsightsDataGatherList is a collection of items\n\nCompatibility level 4: No compatibility is provided, the API can change at any point for any reason. These capabilities should not be used by applications needing long term support.",
	"metadata": "metadata is the standard list's metadata. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata",
}

func (InsightsDataGatherList) SwaggerDoc() map[string]string {
	return map_InsightsDataGatherList
}

var map_InsightsDataGatherSpec = map[string]string{
	"gatherConfig": "gatherConfig spec attribute includes all the configuration options related to gathering of the Insights data and its uploading to the ingress.",
}

func (InsightsDataGatherSpec) SwaggerDoc() map[string]string {
	return map_InsightsDataGatherSpec
}

// AUTO-GENERATED FUNCTIONS END HERE
