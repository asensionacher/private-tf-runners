package backends

import (
	"encoding/json"
	"errors"
	"strings"
)

type Field struct {
	Key         string `json:"key"`
	Description string `json:"description"`
	Required    bool   `json:"required"`
	Sensitive   bool   `json:"sensitive"`
	EnvVar      string `json:"env_var,omitempty"`
	Default     string `json:"default,omitempty"`
}

type AuthMethod struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Fields      []Field `json:"fields"`
}

type BackendSchema struct {
	Name           string       `json:"name"`
	Description    string       `json:"description"`
	AuthMethods    []AuthMethod `json:"auth_methods"`
	CommonFields   []Field      `json:"common_fields,omitempty"`
	RequiredFields []Field      `json:"required_fields,omitempty"`
	OptionalFields []Field      `json:"optional_fields,omitempty"`
}

var Azurerm = BackendSchema{
	Name:        "azurerm",
	Description: "Microsoft Azure Blob Storage",
	RequiredFields: []Field{
		{Key: "storage_account_name", Description: "The name of the Storage Account", Required: true},
		{Key: "container_name", Description: "The name of the Storage Container", Required: true},
		{Key: "key", Description: "The name of the Blob used to store the state", Required: true},
	},
	OptionalFields: []Field{
		{Key: "environment", Description: "The Azure Environment (public, china, usgovernment)", Required: false, EnvVar: "ARM_ENVIRONMENT", Default: "public"},
		{Key: "metadata_host", Description: "The Hostname of the Azure Metadata Service", Required: false, EnvVar: "ARM_METADATA_HOSTNAME"},
		{Key: "lookup_blob_endpoint", Description: "Set to true to lookup Storage Account Data Plane URI from Management Plane", Required: false, EnvVar: "ARM_USE_DNS_ZONE_ENDPOINT", Default: "false"},
		{Key: "snapshot", Description: "Set to true to snapshot the Blob before use", Required: false, EnvVar: "ARM_SNAPSHOT", Default: "false"},
	},
	AuthMethods: []AuthMethod{
		{
			Name:        "oidc",
			Description: "Microsoft Entra ID with OpenID Connect / Workload Identity Federation (Recommended)",
			Fields: []Field{
				{Key: "use_azuread_auth", Description: "Set to true to use Microsoft Entra ID authentication", Required: true, EnvVar: "ARM_USE_AZUREAD"},
				{Key: "use_oidc", Description: "Set to true to use OpenID Connect / Workload identity federation", Required: true, EnvVar: "ARM_USE_OIDC"},
				{Key: "tenant_id", Description: "The tenant ID of the Microsoft Entra ID principal", Required: true, EnvVar: "ARM_TENANT_ID"},
				{Key: "client_id", Description: "The client ID of the Service Principal or Managed Identity", Required: false, EnvVar: "ARM_CLIENT_ID"},
				{Key: "subscription_id", Description: "The subscription ID (required for lookup_blob_endpoint)", Required: false, EnvVar: "ARM_SUBSCRIPTION_ID"},
				{Key: "resource_group_name", Description: "The resource group name (required for lookup_blob_endpoint)", Required: false, EnvVar: "ARM_RESOURCE_GROUP_NAME"},
				{Key: "oidc_azure_service_connection_id", Description: "Azure DevOps Pipeline Service Connection ID for OIDC", Required: false, EnvVar: "ARM_OIDC_AZURE_SERVICE_CONNECTION_ID"},
				{Key: "oidc_request_url", Description: "The URL for the OpenID Connect provider", Required: false, EnvVar: "ARM_OIDC_REQUEST_URL"},
				{Key: "oidc_request_token", Description: "The bearer token for the OIDC request", Required: false, Sensitive: true, EnvVar: "ARM_OIDC_REQUEST_TOKEN"},
				{Key: "oidc_token", Description: "The ID Token for OIDC authentication", Required: false, Sensitive: true, EnvVar: "ARM_OIDC_TOKEN"},
				{Key: "oidc_token_file_path", Description: "Path to a file containing an ID token", Required: false, EnvVar: "ARM_OIDC_TOKEN_FILE_PATH"},
				{Key: "use_aks_workload_identity", Description: "Use Azure AKS Workload Identity", Required: false, EnvVar: "ARM_USE_AKS_WORKLOAD_IDENTITY"},
			},
		},
		{
			Name:        "msi",
			Description: "Microsoft Entra ID with Compute Attached Managed Identity",
			Fields: []Field{
				{Key: "use_azuread_auth", Description: "Set to true to use Microsoft Entra ID authentication", Required: true, EnvVar: "ARM_USE_AZUREAD"},
				{Key: "use_msi", Description: "Set to true to use a Compute Attached Managed Identity", Required: true, EnvVar: "ARM_USE_MSI"},
				{Key: "tenant_id", Description: "The tenant ID of the Microsoft Entra ID principal", Required: true, EnvVar: "ARM_TENANT_ID"},
				{Key: "client_id", Description: "The client ID of the User Assigned Managed Identity (not required for System Assigned)", Required: false, EnvVar: "ARM_CLIENT_ID"},
				{Key: "subscription_id", Description: "The subscription ID (required for lookup_blob_endpoint)", Required: false, EnvVar: "ARM_SUBSCRIPTION_ID"},
				{Key: "resource_group_name", Description: "The resource group name (required for lookup_blob_endpoint)", Required: false, EnvVar: "ARM_RESOURCE_GROUP_NAME"},
				{Key: "msi_endpoint", Description: "Custom Managed Service Identity endpoint", Required: false, EnvVar: "ARM_MSI_ENDPOINT"},
			},
		},
		{
			Name:        "cli",
			Description: "Microsoft Entra ID with Azure CLI",
			Fields: []Field{
				{Key: "use_azuread_auth", Description: "Set to true to use Microsoft Entra ID authentication", Required: true, EnvVar: "ARM_USE_AZUREAD"},
				{Key: "use_cli", Description: "Set to true to use the Azure CLI session", Required: true, EnvVar: "ARM_USE_CLI"},
				{Key: "tenant_id", Description: "The tenant ID of the Microsoft Entra ID principal (Azure CLI will fallback to connected tenant if not supplied)", Required: false, EnvVar: "ARM_TENANT_ID"},
				{Key: "subscription_id", Description: "The subscription ID (Azure CLI will fallback to connected subscription if not supplied)", Required: false, EnvVar: "ARM_SUBSCRIPTION_ID"},
				{Key: "resource_group_name", Description: "The resource group name", Required: false, EnvVar: "ARM_RESOURCE_GROUP_NAME"},
			},
		},
		{
			Name:        "client_secret",
			Description: "Microsoft Entra ID with Service Principal Client Secret (deprecated, use OIDC)",
			Fields: []Field{
				{Key: "use_azuread_auth", Description: "Set to true to use Microsoft Entra ID authentication", Required: true, EnvVar: "ARM_USE_AZUREAD"},
				{Key: "tenant_id", Description: "The tenant ID of the Microsoft Entra ID principal", Required: true, EnvVar: "ARM_TENANT_ID"},
				{Key: "client_id", Description: "The client ID of the Service Principal", Required: true, EnvVar: "ARM_CLIENT_ID"},
				{Key: "client_secret", Description: "The client secret of the Service Principal", Required: true, Sensitive: true, EnvVar: "ARM_CLIENT_SECRET"},
			},
		},
		{
			Name:        "client_certificate",
			Description: "Microsoft Entra ID with Service Principal Client Certificate (deprecated, use OIDC)",
			Fields: []Field{
				{Key: "use_azuread_auth", Description: "Set to true to use Microsoft Entra ID authentication", Required: true, EnvVar: "ARM_USE_AZUREAD"},
				{Key: "tenant_id", Description: "The tenant ID of the Microsoft Entra ID principal", Required: true, EnvVar: "ARM_TENANT_ID"},
				{Key: "client_id", Description: "The client ID of the Service Principal", Required: true, EnvVar: "ARM_CLIENT_ID"},
				{Key: "client_certificate_path", Description: "Path to the PFX file used as the Client Certificate", Required: true, EnvVar: "ARM_CLIENT_CERTIFICATE_PATH"},
				{Key: "client_certificate_password", Description: "Password for the Client Certificate", Required: false, Sensitive: true, EnvVar: "ARM_CLIENT_CERTIFICATE_PASSWORD"},
				{Key: "client_certificate", Description: "Base64 encoded PKCS#12 certificate bundle", Required: false, Sensitive: true, EnvVar: "ARM_CLIENT_CERTIFICATE"},
			},
		},
		{
			Name:        "access_key",
			Description: "Access Key (not recommended for new workloads)",
			Fields: []Field{
				{Key: "access_key", Description: "The Access Key of the storage account", Required: true, Sensitive: true, EnvVar: "ARM_ACCESS_KEY"},
			},
		},
		{
			Name:        "sas_token",
			Description: "SAS Token (not recommended for new workloads)",
			Fields: []Field{
				{Key: "sas_token", Description: "The SAS Token for the storage account container or blob", Required: true, Sensitive: true, EnvVar: "ARM_SAS_TOKEN"},
			},
		},
		{
			Name:        "access_key_lookup_oidc",
			Description: "Access Key Lookup with OpenID Connect / Workload Identity Federation",
			Fields: []Field{
				{Key: "use_oidc", Description: "Set to true to use OpenID Connect / Workload identity federation", Required: true, EnvVar: "ARM_USE_OIDC"},
				{Key: "tenant_id", Description: "The tenant ID of the Microsoft Entra ID principal", Required: true, EnvVar: "ARM_TENANT_ID"},
				{Key: "subscription_id", Description: "The subscription ID of the storage account", Required: true, EnvVar: "ARM_SUBSCRIPTION_ID"},
				{Key: "resource_group_name", Description: "The resource group name of the storage account", Required: true, EnvVar: "ARM_RESOURCE_GROUP_NAME"},
				{Key: "client_id", Description: "The client ID of the Service Principal or Managed Identity", Required: true, EnvVar: "ARM_CLIENT_ID"},
				{Key: "oidc_azure_service_connection_id", Description: "Azure DevOps Pipeline Service Connection ID for OIDC", Required: false, EnvVar: "ARM_OIDC_AZURE_SERVICE_CONNECTION_ID"},
			},
		},
		{
			Name:        "access_key_lookup_msi",
			Description: "Access Key Lookup with Compute Attached Managed Identity",
			Fields: []Field{
				{Key: "use_msi", Description: "Set to true to use a Compute Attached Managed Identity", Required: true, EnvVar: "ARM_USE_MSI"},
				{Key: "tenant_id", Description: "The tenant ID of the Microsoft Entra ID principal", Required: true, EnvVar: "ARM_TENANT_ID"},
				{Key: "subscription_id", Description: "The subscription ID of the storage account", Required: true, EnvVar: "ARM_SUBSCRIPTION_ID"},
				{Key: "resource_group_name", Description: "The resource group name of the storage account", Required: true, EnvVar: "ARM_RESOURCE_GROUP_NAME"},
				{Key: "client_id", Description: "The client ID of the User Assigned Managed Identity (not required for System Assigned)", Required: false, EnvVar: "ARM_CLIENT_ID"},
			},
		},
		{
			Name:        "access_key_lookup_cli",
			Description: "Access Key Lookup with Azure CLI",
			Fields: []Field{
				{Key: "use_cli", Description: "Set to true to use the Azure CLI session", Required: true, EnvVar: "ARM_USE_CLI"},
				{Key: "tenant_id", Description: "The tenant ID of the Microsoft Entra ID principal (Azure CLI will fallback if not supplied)", Required: false, EnvVar: "ARM_TENANT_ID"},
				{Key: "subscription_id", Description: "The subscription ID of the storage account (Azure CLI will fallback if not supplied)", Required: false, EnvVar: "ARM_SUBSCRIPTION_ID"},
				{Key: "resource_group_name", Description: "The resource group name of the storage account", Required: true, EnvVar: "ARM_RESOURCE_GROUP_NAME"},
			},
		},
		{
			Name:        "access_key_lookup_client_secret",
			Description: "Access Key Lookup with Service Principal Client Secret (deprecated, use OIDC)",
			Fields: []Field{
				{Key: "tenant_id", Description: "The tenant ID of the Microsoft Entra ID principal", Required: true, EnvVar: "ARM_TENANT_ID"},
				{Key: "subscription_id", Description: "The subscription ID of the storage account", Required: true, EnvVar: "ARM_SUBSCRIPTION_ID"},
				{Key: "resource_group_name", Description: "The resource group name of the storage account", Required: true, EnvVar: "ARM_RESOURCE_GROUP_NAME"},
				{Key: "client_id", Description: "The client ID of the Service Principal", Required: true, EnvVar: "ARM_CLIENT_ID"},
				{Key: "client_secret", Description: "The client secret of the Service Principal", Required: true, Sensitive: true, EnvVar: "ARM_CLIENT_SECRET"},
			},
		},
		{
			Name:        "access_key_lookup_client_certificate",
			Description: "Access Key Lookup with Service Principal Client Certificate (deprecated, use OIDC)",
			Fields: []Field{
				{Key: "tenant_id", Description: "The tenant ID of the Microsoft Entra ID principal", Required: true, EnvVar: "ARM_TENANT_ID"},
				{Key: "subscription_id", Description: "The subscription ID of the storage account", Required: true, EnvVar: "ARM_SUBSCRIPTION_ID"},
				{Key: "resource_group_name", Description: "The resource group name of the storage account", Required: true, EnvVar: "ARM_RESOURCE_GROUP_NAME"},
				{Key: "client_id", Description: "The client ID of the Service Principal", Required: true, EnvVar: "ARM_CLIENT_ID"},
				{Key: "client_certificate_path", Description: "Path to the PFX file used as the Client Certificate", Required: true, EnvVar: "ARM_CLIENT_CERTIFICATE_PATH"},
				{Key: "client_certificate_password", Description: "Password for the Client Certificate", Required: false, Sensitive: true, EnvVar: "ARM_CLIENT_CERTIFICATE_PASSWORD"},
			},
		},
	},
}

var S3 = BackendSchema{
	Name:        "s3",
	Description: "Amazon S3",
	RequiredFields: []Field{
		{Key: "bucket", Description: "Name of the S3 Bucket", Required: true, EnvVar: "AWS_BUCKET"},
		{Key: "key", Description: "Path to the state file inside the S3 Bucket", Required: true, EnvVar: "AWS_KEY"},
		{Key: "region", Description: "AWS Region of the S3 Bucket", Required: true, EnvVar: "AWS_DEFAULT_REGION"},
	},
	AuthMethods: []AuthMethod{
		{
			Name:        "default",
			Description: "Default AWS credentials (environment variables, shared config, etc.)",
			Fields: []Field{
				{Key: "access_key", Description: "AWS access key", Required: false, Sensitive: true, EnvVar: "AWS_ACCESS_KEY_ID"},
				{Key: "secret_key", Description: "AWS secret key", Required: false, Sensitive: true, EnvVar: "AWS_SECRET_ACCESS_KEY"},
				{Key: "token", Description: "MFA token", Required: false, Sensitive: true, EnvVar: "AWS_SESSION_TOKEN"},
				{Key: "profile", Description: "AWS profile name", Required: false, EnvVar: "AWS_PROFILE"},
				{Key: "shared_credentials_files", Description: "List of paths to AWS shared credentials files", Required: false, EnvVar: "AWS_SHARED_CREDENTIALS_FILES"},
				{Key: "shared_config_files", Description: "List of paths to AWS shared config files", Required: false, EnvVar: "AWS_CONFIG_FILES"},
			},
		},
		{
			Name:        "assume_role",
			Description: "Assume Role with IAM",
			Fields: []Field{
				{Key: "role_arn", Description: "ARN of the IAM Role to assume", Required: true, EnvVar: "AWS_ROLE_ARN"},
				{Key: "duration", Description: "Duration of credentials (e.g. 1h30m)", Required: false, EnvVar: "AWS_DURATION"},
				{Key: "external_id", Description: "External identifier for assume role", Required: false, EnvVar: "AWS_EXTERNAL_ID"},
				{Key: "policy", Description: "IAM Policy JSON describing further restricting permissions", Required: false},
				{Key: "policy_arns", Description: "ARNs of IAM Policies describing further restricting permissions", Required: false},
				{Key: "session_name", Description: "Session name to use when assuming the role", Required: false, EnvVar: "AWS_SESSION_NAME"},
				{Key: "source_identity", Description: "Source identity specified by the principal assuming the role", Required: false, EnvVar: "AWS_SOURCE_IDENTITY"},
				{Key: "tags", Description: "Map of assume role session tags", Required: false},
				{Key: "transitive_tag_keys", Description: "Set of assume role session tag keys to pass to subsequent sessions", Required: false},
			},
		},
		{
			Name:        "assume_role_with_web_identity",
			Description: "Assume Role with Web Identity (OIDC/OAuth)",
			Fields: []Field{
				{Key: "role_arn", Description: "ARN of the IAM Role to assume", Required: true, EnvVar: "AWS_ROLE_ARN"},
				{Key: "duration", Description: "Duration of credentials (e.g. 1h30m)", Required: false, EnvVar: "AWS_DURATION"},
				{Key: "policy", Description: "IAM Policy JSON describing further restricting permissions", Required: false},
				{Key: "policy_arns", Description: "ARNs of IAM Policies describing further restricting permissions", Required: false},
				{Key: "session_name", Description: "Session name to use when assuming the role", Required: false, EnvVar: "AWS_ROLE_SESSION_NAME"},
				{Key: "web_identity_token", Description: "Web identity token from OIDC/OAuth provider", Required: false, Sensitive: true, EnvVar: "AWS_WEB_IDENTITY_TOKEN"},
				{Key: "web_identity_token_file", Description: "File containing web identity token from OIDC/OAuth provider", Required: false, EnvVar: "AWS_WEB_IDENTITY_TOKEN_FILE"},
			},
		},
	},
	OptionalFields: []Field{
		{Key: "acl", Description: "Canned ACL to be applied to state and lock files", Required: false},
		{Key: "encrypt", Description: "Enable server side encryption of state files", Required: false, Default: "true"},
		{Key: "endpoint", Description: "Custom endpoint URL for the AWS S3 API (deprecated, use endpoints.s3)", Required: false},
		{Key: "force_path_style", Description: "Enable path-style S3 URLs", Required: false},
		{Key: "kms_key_id", Description: "ARN of a KMS Key for encrypting state files", Required: false, EnvVar: "AWS_KMS_KEY_ID"},
		{Key: "sse_customer_key", Description: "Key for SSE-C encryption (base64-encoded, 256 bits)", Required: false, Sensitive: true, EnvVar: "AWS_SSE_CUSTOMER_KEY"},
		{Key: "use_path_style", Description: "Enable path-style S3 URLs", Required: false},
		{Key: "workspace_key_prefix", Description: "Prefix for state path in non-default workspaces", Required: false, Default: "env:"},
		{Key: "custom_ca_bundle", Description: "File containing custom root and intermediate certificates", Required: false, EnvVar: "AWS_CA_BUNDLE"},
		{Key: "ec2_metadata_service_endpoint", Description: "Custom endpoint URL for EC2 Metadata Service", Required: false, EnvVar: "AWS_EC2_METADATA_SERVICE_ENDPOINT"},
		{Key: "ec2_metadata_service_endpoint_mode", Description: "Mode for EC2 Metadata Service (IPv4 or IPv6)", Required: false, EnvVar: "AWS_EC2_METADATA_SERVICE_ENDPOINT_MODE"},
		{Key: "http_proxy", Description: "HTTP proxy URL for AWS API requests", Required: false, EnvVar: "HTTP_PROXY"},
		{Key: "https_proxy", Description: "HTTPS proxy URL for AWS API requests", Required: false, EnvVar: "HTTPS_PROXY"},
		{Key: "no_proxy", Description: "Hosts that should not use proxies", Required: false, EnvVar: "NO_PROXY"},
		{Key: "max_retries", Description: "Maximum retry attempts for AWS API requests", Required: false, Default: "5"},
		{Key: "retry_mode", Description: "Retry mode (standard or adaptive)", Required: false, EnvVar: "AWS_RETRY_MODE"},
		{Key: "skip_credentials_validation", Description: "Skip credentials validation via STS API", Required: false},
		{Key: "skip_region_validation", Description: "Skip validation of provided region name", Required: false},
		{Key: "skip_requesting_account_id", Description: "Skip requesting the account ID", Required: false},
		{Key: "skip_metadata_api_check", Description: "Skip usage of EC2 Metadata API", Required: false},
		{Key: "skip_s3_checksum", Description: "Do not include checksum when uploading S3 Objects", Required: false},
		{Key: "sts_endpoint", Description: "Custom endpoint URL for STS API (deprecated)", Required: false},
		{Key: "sts_region", Description: "AWS region for STS", Required: false, EnvVar: "AWS_STS_REGION"},
		{Key: "use_dualstack_endpoint", Description: "Force DualStack endpoint resolution", Required: false, EnvVar: "AWS_USE_DUALSTACK_ENDPOINT"},
		{Key: "use_fips_endpoint", Description: "Force FIPS endpoint resolution", Required: false, EnvVar: "AWS_USE_FIPS_ENDPOINT"},
		{Key: "allowed_account_ids", Description: "List of allowed AWS account IDs", Required: false},
		{Key: "forbidden_account_ids", Description: "List of forbidden AWS account IDs", Required: false},
		{Key: "use_lockfile", Description: "Whether to use a lockfile for state file locking", Required: false, Default: "false"},
		{Key: "insecure", Description: "Whether to allow insecure SSL requests", Required: false, Default: "false"},
		{Key: "iam_endpoint", Description: "Custom endpoint URL for IAM API (deprecated, use endpoints.iam)", Required: false, EnvVar: "AWS_IAM_ENDPOINT"},
		{Key: "dynamodb_endpoint", Description: "Custom endpoint URL for DynamoDB API (deprecated)", Required: false, EnvVar: "AWS_DYNAMODB_ENDPOINT"},
		{Key: "dynamodb_table", Description: "Name of the DynamoDB table for state locking (deprecated)", Required: false, EnvVar: "AWS_DYNAMODB_TABLE"},
		{Key: "endpoints_dynamodb", Description: "Custom endpoint URL for DynamoDB API", Required: false, EnvVar: "AWS_ENDPOINT_URL_DYNAMODB"},
		{Key: "endpoints_iam", Description: "Custom endpoint URL for IAM API", Required: false, EnvVar: "AWS_ENDPOINT_URL_IAM"},
		{Key: "endpoints_s3", Description: "Custom endpoint URL for S3 API", Required: false, EnvVar: "AWS_ENDPOINT_URL_S3"},
		{Key: "endpoints_sso", Description: "Custom endpoint URL for SSO API", Required: false, EnvVar: "AWS_ENDPOINT_URL_SSO"},
		{Key: "endpoints_sts", Description: "Custom endpoint URL for STS API", Required: false, EnvVar: "AWS_ENDPOINT_URL_STS"},
	},
}

var HTTP = BackendSchema{
	Name:        "http",
	Description: "HTTP REST endpoint",
	RequiredFields: []Field{
		{Key: "address", Description: "The address of the REST endpoint", Required: true, EnvVar: "TF_HTTP_ADDRESS"},
	},
	OptionalFields: []Field{
		{Key: "update_method", Description: "HTTP method to use when updating state", Required: false, Default: "POST", EnvVar: "TF_HTTP_UPDATE_METHOD"},
		{Key: "lock_address", Description: "The address of the lock REST endpoint", Required: false, EnvVar: "TF_HTTP_LOCK_ADDRESS"},
		{Key: "lock_method", Description: "HTTP method to use when locking", Required: false, Default: "LOCK", EnvVar: "TF_HTTP_LOCK_METHOD"},
		{Key: "unlock_address", Description: "The address of the unlock REST endpoint", Required: false, EnvVar: "TF_HTTP_UNLOCK_ADDRESS"},
		{Key: "unlock_method", Description: "HTTP method to use when unlocking", Required: false, Default: "UNLOCK", EnvVar: "TF_HTTP_UNLOCK_METHOD"},
		{Key: "username", Description: "The username for HTTP basic authentication", Required: false, EnvVar: "TF_HTTP_USERNAME"},
		{Key: "password", Description: "The password for HTTP basic authentication", Required: false, Sensitive: true, EnvVar: "TF_HTTP_PASSWORD"},
		{Key: "skip_cert_verification", Description: "Whether to skip TLS verification", Required: false, Default: "false"},
		{Key: "retry_max", Description: "The number of HTTP request retries", Required: false, Default: "2", EnvVar: "TF_HTTP_RETRY_MAX"},
		{Key: "retry_wait_min", Description: "The minimum time in seconds to wait between HTTP request attempts", Required: false, Default: "1", EnvVar: "TF_HTTP_RETRY_WAIT_MIN"},
		{Key: "retry_wait_max", Description: "The maximum time in seconds to wait between HTTP request attempts", Required: false, Default: "30", EnvVar: "TF_HTTP_RETRY_WAIT_MAX"},
	},
	AuthMethods: []AuthMethod{
		{
			Name:        "none",
			Description: "No authentication",
			Fields:      []Field{},
		},
		{
			Name:        "basic",
			Description: "HTTP Basic authentication",
			Fields: []Field{
				{Key: "username", Description: "Username for basic auth", Required: true, EnvVar: "TF_HTTP_USERNAME"},
				{Key: "password", Description: "Password for basic auth", Required: true, Sensitive: true, EnvVar: "TF_HTTP_PASSWORD"},
			},
		},
		{
			Name:        "mtls",
			Description: "Mutual TLS authentication",
			Fields: []Field{
				{Key: "client_certificate_pem", Description: "PEM-encoded certificate for mTLS authentication", Required: true, EnvVar: "TF_HTTP_CLIENT_CERTIFICATE_PEM"},
				{Key: "client_private_key_pem", Description: "PEM-encoded private key for mTLS authentication", Required: true, Sensitive: true, EnvVar: "TF_HTTP_CLIENT_PRIVATE_KEY_PEM"},
				{Key: "client_ca_certificate_pem", Description: "PEM-encoded CA certificate chain for TLS verification", Required: false, EnvVar: "TF_HTTP_CLIENT_CA_CERTIFICATE_PEM"},
			},
		},
	},
}

var Pg = BackendSchema{
	Name:        "pg",
	Description: "PostgreSQL",
	RequiredFields: []Field{
		{Key: "conn_str", Description: "PostgreSQL connection string (postgres:// URL)", Required: true, Sensitive: true, EnvVar: "PG_CONN_STR"},
	},
	AuthMethods: []AuthMethod{
		{
			Name:        "default",
			Description: "Default PostgreSQL connection",
			Fields:      []Field{},
		},
	},
	OptionalFields: []Field{
		{Key: "schema_name", Description: "Name of the Postgres schema", Required: false, Default: "terraform_remote_state", EnvVar: "PG_SCHEMA_NAME"},
		{Key: "skip_schema_creation", Description: "Skip creation of the schema if it does not exist", Required: false, Default: "false", EnvVar: "PG_SKIP_SCHEMA_CREATION"},
		{Key: "skip_table_creation", Description: "Skip creation of the states table if it does not exist", Required: false, Default: "false", EnvVar: "PG_SKIP_TABLE_CREATION"},
		{Key: "skip_index_creation", Description: "Skip creation of the index if it does not exist", Required: false, Default: "false", EnvVar: "PG_SKIP_INDEX_CREATION"},
		{Key: "skip_lock_table_creation", Description: "Skip creation of the locks table if it does not exist", Required: false, Default: "false"},
		{Key: "workspace_table_name", Description: "The table name for workspaces", Required: false, Default: "terraform_workspaces"},
		{Key: "states_table_name", Description: "The table name for states", Required: false, Default: "terraform_states"},
		{Key: "locks_table_name", Description: "The table name for locks", Required: false, Default: "terraform_locks"},
		{Key: "skip_gss_principals_creation", Description: "Skip creation of GSS principals", Required: false},
	},
}

var Kubernetes = BackendSchema{
	Name:        "kubernetes",
	Description: "Kubernetes",
	RequiredFields: []Field{
		{Key: "secret_suffix", Description: "Suffix used when creating the secret", Required: true},
	},
	OptionalFields: []Field{
		{Key: "labels", Description: "Map of additional labels to be applied to the secret and lease", Required: false},
		{Key: "namespace", Description: "Namespace to store the secret and lease in", Required: false, EnvVar: "KUBE_NAMESPACE"},
		{Key: "in_cluster_config", Description: "Use in-cluster service account for authentication", Required: false, Default: "false", EnvVar: "KUBE_IN_CLUSTER_CONFIG"},
		{Key: "host", Description: "Hostname (in form of URI) of Kubernetes master", Required: false, Default: "https://localhost", EnvVar: "KUBE_HOST"},
		{Key: "username", Description: "Username for HTTP basic authentication", Required: false, EnvVar: "KUBE_USER"},
		{Key: "password", Description: "Password for HTTP basic authentication", Required: false, Sensitive: true, EnvVar: "KUBE_PASSWORD"},
		{Key: "insecure", Description: "Whether to skip TLS verification", Required: false, Default: "false", EnvVar: "KUBE_INSECURE"},
		{Key: "client_certificate", Description: "PEM-encoded client certificate for TLS authentication", Required: false, EnvVar: "KUBE_CLIENT_CERT_DATA"},
		{Key: "client_key", Description: "PEM-encoded client certificate key for TLS authentication", Required: false, Sensitive: true, EnvVar: "KUBE_CLIENT_KEY_DATA"},
		{Key: "cluster_ca_certificate", Description: "PEM-encoded root certificates bundle for TLS authentication", Required: false, EnvVar: "KUBE_CLUSTER_CA_CERT_DATA"},
		{Key: "config_path", Description: "Path to the kube config file", Required: false, EnvVar: "KUBE_CONFIG_PATH"},
		{Key: "config_paths", Description: "List of paths to kube config files", Required: false, EnvVar: "KUBE_CONFIG_PATHS"},
		{Key: "config_context", Description: "Context to choose from the config file", Required: false, EnvVar: "KUBE_CTX"},
		{Key: "config_context_auth_info", Description: "Authentication info context of the kube config", Required: false, EnvVar: "KUBE_CTX_AUTH_INFO"},
		{Key: "config_context_cluster", Description: "Cluster context of the kube config", Required: false, EnvVar: "KUBE_CTX_CLUSTER"},
		{Key: "token", Description: "Token of your service account", Required: false, Sensitive: true, EnvVar: "KUBE_TOKEN"},
		{Key: "load_config_file", Description: "Load kube config file from disk", Required: false, Default: "true"},
	},
	AuthMethods: []AuthMethod{
		{
			Name:        "exec",
			Description: "Exec-based credential plugin",
			Fields: []Field{
				{Key: "api_version", Description: "API version to use when decoding ExecCredentials", Required: true, EnvVar: "KUBE_EXEC_API_VERSION"},
				{Key: "command", Description: "Command to execute", Required: true, EnvVar: "KUBE_EXEC_COMMAND"},
				{Key: "args", Description: "JSON list of arguments to pass when executing the plugin", Required: false, EnvVar: "KUBE_EXEC_ARGS"},
				{Key: "env", Description: "JSON map of environment variables to set when executing", Required: false, EnvVar: "KUBE_EXEC_ENV"},
			},
		},
	},
}

var GCS = BackendSchema{
	Name:        "gcs",
	Description: "Google Cloud Storage",
	RequiredFields: []Field{
		{Key: "bucket", Description: "The name of the GCS bucket", Required: true},
	},
	OptionalFields: []Field{
		{Key: "prefix", Description: "GCS prefix inside the bucket for state files", Required: false},
		{Key: "credentials", Description: "Local path to GCP account credentials JSON", Required: false, Sensitive: true, EnvVar: "GOOGLE_BACKEND_CREDENTIALS"},
		{Key: "access_token", Description: "OAuth 2.0 access token for GCP", Required: false, Sensitive: true, EnvVar: "GOOGLE_ACCESS_TOKEN"},
		{Key: "impersonate_service_account", Description: "Service account to impersonate", Required: false, EnvVar: "GOOGLE_BACKEND_IMPERSONATE_SERVICE_ACCOUNT"},
		{Key: "impersonate_service_account_delegates", Description: "Delegation chain for service account impersonation", Required: false, EnvVar: "GOOGLE_IMPERSONATE_SERVICE_ACCOUNT_DELEGATES"},
		{Key: "encryption_key", Description: "32 byte base64-encoded customer-supplied encryption key", Required: false, Sensitive: true, EnvVar: "GOOGLE_ENCRYPTION_KEY"},
		{Key: "kms_encryption_key", Description: "Cloud KMS key for encryption (projects/{{project}}/locations/{{location}}/keyRings/{{keyRing}}/cryptoKeys/{{name}})", Required: false, EnvVar: "GOOGLE_KMS_ENCRYPTION_KEY"},
		{Key: "storage_custom_endpoint", Description: "Private Service Connect endpoint URL for Cloud Storage API", Required: false, EnvVar: "GOOGLE_BACKEND_STORAGE_CUSTOM_ENDPOINT"},
	},
	AuthMethods: []AuthMethod{
		{
			Name:        "default",
			Description: "Google Application Default Credentials",
			Fields:      []Field{},
		},
	},
}

func GetBackendSchema(backendType string) *BackendSchema {
	switch backendType {
	case "azurerm":
		return &Azurerm
	case "s3":
		return &S3
	case "http":
		return &HTTP
	case "pg":
		return &Pg
	case "kubernetes":
		return &Kubernetes
	case "gcs":
		return &GCS
	default:
		return nil
	}
}

func GetAllBackendSchemas() map[string]*BackendSchema {
	return map[string]*BackendSchema{
		"azurerm":    &Azurerm,
		"s3":         &S3,
		"http":       &HTTP,
		"pg":         &Pg,
		"kubernetes": &Kubernetes,
		"gcs":        &GCS,
	}
}

func ValidateConfig(backendType, configJSON string) error {
	schema := GetBackendSchema(backendType)
	if schema == nil {
		return errors.New("unknown backend type: " + backendType)
	}

	if configJSON == "" {
		return errors.New("config is required")
	}

	var config map[string]string
	if err := json.Unmarshal([]byte(configJSON), &config); err != nil {
		return errors.New("invalid config JSON")
	}

	for _, field := range schema.RequiredFields {
		if config[field.Key] == "" && field.Default == "" {
			return errors.New("field '" + field.Key + "' is required")
		}
	}

	for _, method := range schema.AuthMethods {
		methodFieldsSet := 0
		for _, field := range method.Fields {
			if config[field.Key] != "" {
				methodFieldsSet++
			}
		}
		if methodFieldsSet > 0 && methodFieldsSet < len(method.Fields) {
			missing := []string{}
			for _, field := range method.Fields {
				if field.Required && config[field.Key] == "" {
					missing = append(missing, field.Key)
				}
			}
			if len(missing) > 0 {
				return errors.New("auth method '" + method.Name + "' requires fields: " + strings.Join(missing, ", "))
			}
		}
	}

	return nil
}