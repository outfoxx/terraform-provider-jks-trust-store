terraform {
  required_providers {
    tls = {
      source  = "hashicorp/tls"
      version = "~> 3.1.0"
    }
    jks = {
      # Locally compiled provider
      #source  = "hashicorp.com/outfox/jks-trust-store"

      # Terraform forge hosted provider
      source  = "outfoxx/jks-trust-store"
      version = "~> 0.1"
    }
  }
}

provider "tls" {
}

provider "jks" {
}

resource "tls_private_key" "ca" {
  algorithm = "ECDSA"
}

resource "tls_self_signed_cert" "ca" {
  key_algorithm   = tls_private_key.ca.algorithm
  private_key_pem = tls_private_key.ca.private_key_pem

  validity_period_hours = 12
  early_renewal_hours   = 3

  is_ca_certificate = true
  allowed_uses = [
    "cert_signing",
    "crl_signing",
  ]

  subject {
    common_name  = "Cluster TLS Root CA"
    organization = "Outfox, Inc."
  }

  set_subject_key_id = true
}

resource "jks_trust_store" "ca" {
  certificates = [
    tls_self_signed_cert.ca.cert_pem
  ]
  password = "none"
}

output "ca_jks" {
  value = jks_trust_store.ca.jks
}
