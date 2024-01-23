# Corefile example - Environment Variables

This example sources the values from environment variables using the convention built in to Coredns.

* `controller_url` -> from the `OMADA_URL` environment variable
* `site` -> from the `OMADA_SITE` environment variable
* `username` -> from the `OMADA_URL` environment variable
* `password` -> from the `OMADA_URL` environment variable
* `forward` server -> from the `UPSTREAM_DNS` environment variable


> # Environment Variables
> CoreDNS supports environment variables >substitution in its configuration. They can be >used anywhere in > the Corefile. The syntax is `{$ENV_VAR}`` (a more Windows-like syntax `{%ENV_VAR%}`` is also supported).  > CoreDNS substitutes the contents of the variable while parsing the Corefile.

Reference: https://coredns.io/manual/configuration/#environment-variables