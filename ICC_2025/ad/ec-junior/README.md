# EC Junior

A simple e-commerce platform for browsing products, managing shopping carts, and placing orders.

## Service Information

- **Exposed Port**: `8080/tcp`
- **Flag Stores**: 2

## Patching

Only the file `patchable/waf_rule.conf` can be modified. All other files, including the service's JavaScript code and configuration files, cannot be patched.

The `patchable/waf_rule.conf` file is loaded by the `waf` container (see `waf/` directory and `Dockerfile-waf` for implementation details) and interpreted as a ModSecurity configuration file. You may add custom ModSecurity rules to this file to defend your service against attacks.

## Important Notes

**Admin Bot Usage**: Excessive or abusive requests to the admin bot may be considered a Denial of Service (DoS) attack. Please minimize bot requests and only send them when necessary for your attempts.
