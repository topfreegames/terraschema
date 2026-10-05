variable "tier" {
  type        = string
  default     = "small"
  description = "A preset whose valid values are the keys of local.tiers"
  validation {
    condition     = contains(keys(local.tiers), var.tier)
    error_message = "Invalid value for tier"
  }
}

variable "region" {
  type        = string
  description = "A region from local.regions"
  validation {
    condition     = contains(local.regions, var.region)
    error_message = "Invalid value for region"
  }
}

variable "prefixed_region" {
  type        = string
  default     = "aws-us-east-1"
  description = "A value from a local built on another local"
  validation {
    condition     = contains(local.prefixed_regions, var.prefixed_region)
    error_message = "Invalid value for prefixed_region"
  }
}

variable "default_value" {
  type        = string
  default     = "a"
  description = "A plain string"
}

variable "not_static" {
  type        = string
  default     = "a"
  description = "Validated against a local that depends on a variable, so it gets no enum"
  validation {
    condition     = contains(local.from_variable, var.not_static)
    error_message = "Invalid value for not_static"
  }
}
