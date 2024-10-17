variable "schema_src" {
  type = string
  default = "file://internal/mysql/schema.sql"
}

variable "port" {
  type string
  default = getenv("DB_PORT")
}

variable "user" {
  type = string
  default = getenv("DB_USER")
}

variable "pass" {
  type = string
  default = getenv("DB_PASSWORD")
}

env "local" {
  src = var.schema_src
  url = "mysql://go-api-starter-user:go-api-starter-password@localhost:${port}"
  dev = "docker://mysql/8"
}
