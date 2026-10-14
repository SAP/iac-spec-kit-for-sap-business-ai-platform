package scaffold

// backendTemplates maps a backend type identifier to a placeholder backend.tf.
// Values are intentionally incomplete — users must fill in project-specific
// details.  No secrets are stored here; credentials must come from environment
// variables or a separate secrets manager.
// See: https://developer.hashicorp.com/terraform/language/backend
var backendTemplates = map[string]string{
	"local": `# Explicit local backend — state is stored in terraform.tfstate next to your
# configuration. Suitable for single-developer use; not recommended for teams.
# See: https://developer.hashicorp.com/terraform/language/backend/local
terraform {
  backend "local" {
    path = "terraform.tfstate"
  }
}
`,
	"remote": `# TODO: fill in your HCP Terraform / Terraform Enterprise details.
# Authentication uses a token stored in ~/.terraform.d/credentials.tfrc.json
# or TF_TOKEN_<hostname> env var — never hardcode it here.
# See: https://developer.hashicorp.com/terraform/language/backend/remote
terraform {
  backend "remote" {
    hostname     = "app.terraform.io"
    organization = "<your-organization>"

    workspaces {
      name = "<your-workspace>"
    }
  }
}
`,
	"azurerm": `# TODO: fill in your Azure Blob Storage backend details.
# This template uses the recommended OpenID Connect (OIDC) authentication.
# For other auth methods (Managed Identity, Access Key, SAS Token, Client Secret)
# see: https://developer.hashicorp.com/terraform/language/backend/azurerm
# Credentials must never be hardcoded here — use environment variables or a
# secrets manager.
terraform {
  backend "azurerm" {
    storage_account_name = "<your-storage-account>"
    container_name       = "<your-container>"
    key                  = "terraform.tfstate"
    use_azuread_auth     = true
    use_oidc             = true
    tenant_id            = "<your-tenant-id>"
    client_id            = "<your-client-id>"
  }
}
`,
	"consul": `# TODO: fill in your Consul backend details.
# Only path is required. address and scheme are optional.
# Token must be supplied via CONSUL_HTTP_TOKEN env var — never hardcode it here.
# For all options see: https://developer.hashicorp.com/terraform/language/backend/consul
terraform {
  backend "consul" {
    path = "<your-app>/terraform.tfstate"
  }
}
`,
	"cos": `# TODO: fill in your Tencent Cloud COS backend details.
# Only bucket is required. region and prefix are optional but recommended.
# Credentials must be supplied via environment variables (TENCENTCLOUD_SECRET_ID,
# TENCENTCLOUD_SECRET_KEY) — never hardcode them here.
# See: https://developer.hashicorp.com/terraform/language/backend/cos
terraform {
  backend "cos" {
    bucket = "<your-cos-bucket>"
  }
}
`,
	"gcs": `# TODO: fill in your Google Cloud Storage backend details.
# Credentials must be supplied via GOOGLE_CREDENTIALS or Application Default
# Credentials — never hardcode them here.
# See: https://developer.hashicorp.com/terraform/language/backend/gcs
terraform {
  backend "gcs" {
    bucket = "<your-gcs-bucket>"
    prefix = "terraform/state"
  }
}
`,
	"http": `# TODO: fill in your HTTP backend details.
# Only address is required. lock_address and unlock_address are optional.
# Credentials must be supplied via environment variables (TF_HTTP_USERNAME,
# TF_HTTP_PASSWORD) — never hardcode them here.
# See: https://developer.hashicorp.com/terraform/language/backend/http
terraform {
  backend "http" {
    address = "https://your-backend-url/terraform.tfstate"
  }
}
`,
	"kubernetes": `# TODO: fill in your Kubernetes backend details.
# Credentials are taken from the active kubeconfig or in-cluster config.
# For other auth options (in-cluster, exec) see:
# https://developer.hashicorp.com/terraform/language/backend/kubernetes
terraform {
  backend "kubernetes" {
    secret_suffix = "<your-app>"
    config_path   = "~/.kube/config"
  }
}
`,
	"oci": `# TODO: fill in your Oracle Cloud Infrastructure backend details.
# bucket and namespace are required. region and key are optional.
# Credentials must be supplied via environment variables or instance principal
# — never hardcode them here.
# See: https://developer.hashicorp.com/terraform/language/backend/oci
terraform {
  backend "oci" {
    bucket    = "<your-oci-bucket>"
    namespace = "<your-object-storage-namespace>"
  }
}
`,
	"oss": `# TODO: fill in your Alibaba Cloud OSS backend details.
# Credentials must be supplied via environment variables
# — never hardcode them here.
# See: https://developer.hashicorp.com/terraform/language/backend/oss
terraform {
  backend "oss" {
    bucket = "<your-oss-bucket>"
    prefix = "terraform/state"
    region = "<alibaba-region>"
  }
}
`,
	"pg": `# TODO: fill in your PostgreSQL backend details.
# Connection string must be supplied via PG_CONN_STR env var
# — never hardcode credentials here.
# See: https://developer.hashicorp.com/terraform/language/backend/pg
terraform {
  backend "pg" {
    conn_str    = "postgres://<user>:<password>@<host>/<dbname>"
    schema_name = "terraform_remote_state"
  }
}
`,
	"s3": `# TODO: fill in your AWS S3 backend details.
# Credentials must be supplied via environment variables (AWS_ACCESS_KEY_ID,
# AWS_SECRET_ACCESS_KEY) or an IAM role — never hardcode them here.
# See: https://developer.hashicorp.com/terraform/language/backend/s3
terraform {
  backend "s3" {
    bucket = "<your-state-bucket>"
    key    = "<path/to/terraform.tfstate>"
    region = "<aws-region>"
  }
}
`,
}

// KnownBackends lists the backend type identifiers accepted by --backend and
// the interactive prompt.
var KnownBackends = []string{"local", "remote", "azurerm", "consul", "cos", "gcs", "http", "kubernetes", "oci", "oss", "pg", "s3"}
